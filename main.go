package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/integrii/flaggy"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Twilio TwilioConfiguration `yaml:"twilio"`
	URLs   []URLConfiguration  `yaml:"urls"`
}

type TwilioConfiguration struct {
	FromPhoneNumber string `yaml:"fromPhoneNumber"`
	ToPhoneNumber   string `yaml:"toPhoneNumber"`
	AccountSID      string `yaml:"accountSid"`
	AuthToken       string `yaml:"authToken"`
}

type URLConfiguration struct {
	Description string         `yaml:"description"`
	URL         string         `yaml:"url"`
	Pattern     string         `yaml:"pattern"`
	AlertIf     AlertCondition `yaml:"alertIf"`
}

type AlertCondition string

const (
	Match   AlertCondition = "Match"
	NoMatch                = "NoMatch"
)

func main() {
	flaggy.SetName("url-monitor")
	flaggy.SetDescription("Monitors URLs for patterns, and alerts by SMS.")
	flaggy.SetVersion("0.2.0-dev")

	var configurationFile = "configuration.yaml"
	flaggy.String(&configurationFile, "c", "configuration-file", "YAML configuration file defining URLs to monitor")
	flaggy.Parse()

	configuration, err := readConfiguration(configurationFile)
	if err != nil {
		panic(err)
	}

	monitorURLs(configuration)
}

// readConfiguration unmarshals the configuration object from its YAML representation in a text file
func readConfiguration(configurationFile string) (configuration *Configuration, err error) {
	data, err := ioutil.ReadFile(configurationFile)
	if err != nil {
		return nil, err
	}
	var conf Configuration
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	for _, urlConfiguration := range conf.URLs {
		_, err := url.Parse(urlConfiguration.URL)
		if err != nil {
			return nil, err
		}
		_, err = regexp.Compile(urlConfiguration.Pattern)
		if err != nil {
			return nil, err
		}
		switch urlConfiguration.AlertIf {
		case Match, NoMatch:
			continue
		}
		return nil, errors.New("Invalid alert condition: '" + string(urlConfiguration.AlertIf) + "'")
	}
	return &conf, nil
}

// monitorURLs concurrently calls monitorURL for each configured URL.
func monitorURLs(configuration *Configuration) {
	var wg sync.WaitGroup
	for _, urlConfiguration := range configuration.URLs {
		wg.Add(1)
		go monitorURL(&urlConfiguration, &configuration.Twilio, &wg)
	}
	wg.Wait()
}

// monitorURL retrives the content for a URL, checks if the configured pattern is a match, and sends an SMS alert if necessary.
func monitorURL(urlConfiguration *URLConfiguration, twilioConfiguration *TwilioConfiguration, wg *sync.WaitGroup) {
	pattern, err := regexp.Compile(urlConfiguration.Pattern)
	if err != nil {
		println(err)
		wg.Done()
		return
	}
	res, err := http.Get(urlConfiguration.URL)
	if err != nil {
		println(err)
		wg.Done()
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		print(err)
		wg.Done()
		return
	}
	bodyMatches := pattern.Match(body)
	if (bodyMatches && urlConfiguration.AlertIf == Match) ||
		(!bodyMatches && urlConfiguration.AlertIf == NoMatch) {
		sendSMS(urlConfiguration, twilioConfiguration)
	}

	wg.Done()
}

// sendSMS uses the Twilio HTTP API to send an SMS alert for one URL
func sendSMS(urlConfiguration *URLConfiguration, twilioConfiguration *TwilioConfiguration) {
	smsBody := smsBody(urlConfiguration)
	msgData := url.Values{}
	msgData.Set("To", twilioConfiguration.ToPhoneNumber)
	msgData.Set("From", twilioConfiguration.FromPhoneNumber)
	msgData.Set("Body", smsBody)
	msgDataReader := *strings.NewReader(msgData.Encode())
	client := &http.Client{}
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + twilioConfiguration.AccountSID + "/Messages.json"
	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(twilioConfiguration.AccountSID, twilioConfiguration.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(smsBody)
		} else {
			println(err)
		}
	} else {
		fmt.Println(resp.Status)
	}
}

// smsBody formats the text body of an SMS alert for a URL
func smsBody(urlConfiguration *URLConfiguration) string {
	return "url-monitor: alert triggered for " + urlConfiguration.Description + ". URL: " + urlConfiguration.URL
}
