
FROM alpine
ADD url-monitor /url-monitor
ADD configuration.yaml /configuration.yaml
ENTRYPOINT ["/url-monitor"]