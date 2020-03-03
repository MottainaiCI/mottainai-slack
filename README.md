# Mottainai-slack

It is a small tool that uses [mottainai-bridge](https://github.com/MottainaiCI/mottainai-bridge) and [slack-go](https://github.com/slack-go/slack)

## Deploy on Kubernetes

Example kube.yaml:

```yaml
---
apiVersion: v1
kind: Pod
metadata:
  name: slackbot
  namespace: slackbot
spec:
  containers:
    - name: slackbot
      image: quay.io/sabayon/mottainai-slack
      imagePullPolicy: Always
      args: ["event", "run"]
      env:
        - name: MOTT_MASTER
          value: "https://mottainai_instance"
        - name: MOTT_APIKEY
          value: "mottainai_api_key"
        - name: MOTT_SLACKCHANNEL
          value: "slack_channel_id"
        - name: MOTT_SLACKAPI
          value: "slack_app_api_key"
  restartPolicy: Always
```