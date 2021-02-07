# URL Monitor

Monitors URLs for a pattern, alerting by SMS.

## Build
```sh
go build
```

## Configure
Modify `configuration.yaml` as necessary.
1. **Twilio**. Add your Twilio account SID, auth token, and the *from* and *to* phone numbers for SMS alerts.
2. **URLs**. Add the URLs to monitor, with a helpful description which will be included in the SMS alert. Supported alert conditions (`alertIf`) are `Match`, `NoMatch`, and `Never`. Patterns are regular expressions supported by golang's `regexp` package.

## Run
By default, `configuration.yaml` will be used.
```sh
./url-monitor
```

## Usage
```
url-monitor - Monitors URLs for patterns, and alerts by SMS.

  Flags: 
       --version              Displays the program version string.
    -h --help                 Displays help with available flag, subcommand, and positional value parameters.
    -c --configuration-file   YAML configuration file defining URLs to monitor (default: configuration.yaml)
```

## Example alert
For the URL configuration:
```yaml
- description: Google
  url: https://google.com
  pattern: google
  alertIf: Match
```
You'll receive an SMS message that looks like:
```
url-monitor: alert triggered for Google. URL: https://google.com
```
