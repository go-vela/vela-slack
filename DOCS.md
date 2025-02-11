## Description

This plugin enables you to send data to a [Slack](https://slack.com/) channel.

Source Code: https://github.com/go-vela/vela-slack

Registry: https://hub.docker.com/r/target/vela-slack

## Usage

> **NOTE:**
>
> Users should refrain from using latest as the tag for the Docker image.
>
> It is recommended to use a semantically versioned tag instead.

Sample of sending a message:

```yaml
steps:
  - name: message
    image: target/vela-slack:latest
    pull: always
    parameters:
      text: "Hello World!"
```

Sample of sending a message with formatting:

```diff
steps:
  - name: message-with-formatting
    image: target/vela-slack:latest
    pull: not_present
    parameters:
-     text: "Hello World!"
+     text: "Hello Repo {{ .RepositoryName }}!"
```

Sample of sending a message with local attachment file:

```diff
steps:
  - name: message-with-attachment
    image: target/vela-slack:latest
    secrets: [ slack_webhook ]
    parameters:
-     text: "Hello World!"
      filepath: slack_attachment.json
      remote: false
```

Sample of sending a message with remote attachment file:

```diff
steps:
  - name: message-with-remote-attachment
    image: target/vela-slack:latest
    secrets: [ slack_webhook ]
    parameters:
-     text: "Hello World!"
      filepath: slack_attachment.json
      remote: true
      registry: https://github.com
```

content of `slack_attachment.json`:

```json
{
    "attachments": [
        {
            "fallback": "Required plain-text summary of the attachment.",
            "color": "#36a64f",
            "pretext": "Optional text that appears above the attachment block",
            "author_name": "Bobby Tables",
            "author_link": "http://flickr.com/bobby/",
            "author_icon": "http://flickr.com/icons/bobby.jpg",
            "title": "Slack API Documentation",
            "title_link": "https://api.slack.com/",
            "text": "Build: {{ .BuildNumber }}",
            "fields": [
                {
                    "title": "Priority",
                    "value": "High",
                    "short": false
                }
            ],
            "image_url": "http://my-website.com/path/to/image.jpg",
            "thumb_url": "http://example.com/path/to/thumb.png",
            "footer": "Slack API",
            "footer_icon": "https://platform.slack-edge.com/img/default_application_icon.png",
            "ts": 123456789
        }
    ]
}
```

> **NOTE:**
>
> To use any variables within `attachments` the file must be saved in the attachment format and be a JSON file.
>
> The configuration below is pulled almost directly from the Slack [Message Builder](https://api.slack.com/docs/messages/builder) attachments example.

## Secrets

> **NOTE:** Users should refrain from configuring sensitive information in your pipeline in plain text.

### Internal

The plugin accepts the following `parameters` for authentication:

| Parameter | Environment Variable Configuration   |
| --------- | ------------------------------------ |
| `webhook` | `PARAMETER_WEBHOOK`, `SLACK_WEBHOOK` |

Users can use [Vela internal secrets](https://go-vela.github.io/docs/tour/secrets/) to substitute these sensitive values at runtime:

```diff
steps:
  - name: message
    image: target/vela-slack:latest
    pull: always
+   secrets: [ slack_webhook ]
    parameters:
      text: "Hello World!"
-     webhook: http://slack.example.com
```

> This example will add the secret to the `message` step as environment variables:
>
> * `SLACK_WEBHOOK=<value>`

### External

The plugin accepts the following files for authentication:

| Parameter | Volume Configuration                                            |
| --------- | --------------------------------------------------------------- |
| `webhook` | `/vela/parameters/slack/webhook`, `/vela/secrets/slack/webhook` |

Users can use [Vela external secrets](https://go-vela.github.io/docs/concepts/pipeline/secrets/origin/) to substitute these sensitive values at runtime:

```diff
steps:
  - name: message
    image: target/vela-slack:latest
    pull: always
    parameters:
      text: "Hello World!"
-     webhook: http://slack.example.com
```

> This example will read the secret value in the volume stored at `/vela/secrets/`

## Parameters

> **NOTE:**
>
> The plugin supports reading all parameters via environment variables or files.
>
> Any values set from a file take precedence over values set from the environment.

The following parameters are used to configure the image:

| Name         | Description                          | Required | Default | Environment Variables                        |
| ------------ | ------------------------------------ | -------- | ------- | -------------------------------------------- |
| `channel`    | Slack channel to send data to        | `false`  | `N/A`   | `PARAMETER_CHANNEL`<br>`SLACK_CHANNEL`       |
| `filepath`   | file path to attachment JSON file    | `false`  | `N/A`   | `PARAMETER_FILEPATH`<br>`SLACK_FILEPATH`     |
| `icon_emoji` | Slack emoji to use for the icon      | `false`  | `N/A`   | `PARAMETER_ICON_EMOJI`<br>`SLACK_ICON_EMOJI` |
| `icon_url`   | Slack emoji URL to use for the icon  | `false`  | `N/A`   | `PARAMETER_ICON_URL`<br>`SLACK_ICON_URL`     |
| `log_level`  | set the log level for the plugin     | `true`   | `info`  | `PARAMETER_LOG_LEVEL`<br>`SLACK_LOG_LEVEL`   |
| `text`       | top level text to display in message | `false`  | `N/A`   | `PARAMETER_TEXT`<br>`SLACK_TEXT`             |
| `thread_ts`  | timestamp of the thread post         | `false`  | `N/A`   | `PARAMETER_THREAD_TS`<br>`SLACK_THREAD_TS`   |
| `webhook`    | Slack webhook url to send data to    | `true`   | `N/A`   | `PARAMETER_WEBHOOK`<br>`SLACK_WEBHOOK`       |

## Template

COMING SOON!

## Troubleshooting

You can start troubleshooting this plugin by tuning the level of logs being displayed:

```diff
steps:
  - name: message
    image: target/vela-slack:latest
    pull: always
    parameters:
+     log_level: trace
      text: "Hello World!"
```

Below are a list of common problems and how to solve them:
