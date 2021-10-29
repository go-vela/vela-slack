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

## Secrets

> **NOTE:** Users should refrain from configuring sensitive information in your pipeline in plain text.

### Internal

The plugin accepts the following `parameters` for authentication:

| Parameter | Environment Variable Configuration                                    |
| --------- | --------------------------------------------------------------------- |

Users can use [Vela internal secrets](https://go-vela.github.io/docs/tour/secrets/) to substitute these sensitive values at runtime:

### External

The plugin accepts the following files for authentication:

| Parameter | Volume Configuration                                                  |
| --------- | --------------------------------------------------------------------- |

Users can use [Vela external secrets](https://go-vela.github.io/docs/concepts/pipeline/secrets/origin/) to substitute these sensitive values at runtime:

## Parameters

> **NOTE:**
>
> The plugin supports reading all parameters via environment variables or files.
>
> Any values set from a file take precedence over values set from the environment.

The following parameters are used to configure the image:

| Name        | Description                                      | Required | Default       | Environment Variables                           |
| ----------- | ------------------------------------------------ | -------- | ------------- | ----------------------------------------------- |

## Template

COMING SOON!

## Troubleshooting

You can start troubleshooting this plugin by tuning the level of logs being displayed:

Below are a list of common problems and how to solve them: