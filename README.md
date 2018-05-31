# Appsflyer Proxy

A simple proxy server for proxying s2s in app events to appsflyer

## Getting Started

Build and run executable.

To just run locally:

```
APPODEAL_AUTH_KEY_NAME=xxx AF_DEV_KEY=yyy AF_PROXY_PORT=4001 go run main.go
```

Server expects environment variables defined:

#####APPODEAL_AUTH_KEY

Each request should send **authentication** header with APPODEAL_AUTH_KEY value.

#####AF_DEV_KEY

Authentication key required for your app in Appsflyer.

#####AF_PROXY_PORT

Proxy will run on this port.

##### Curl request example

```
curl -H "Content-Type: application/json" -H "authentication: <YOUR APPODEAL_AUTH_KEY_NAME>" \ -d '{"appsflyer_id":"<appsflyer_id>","idfa":"<ifa>","eventName":"af_test_revenue","eventCurrency":"USD","ip":"<ip>","eventTime":"2018-05-30 08:35:44.000","af_events_api":"true","eventValue":"{\"af_revenue\":0.01,\"af_currency\":\"USD\"}"}' -X POST "http://<server>:<port>/appsflyer_proxy/<app bundle>"
```


## Deployment

Watch Dockerfile.
