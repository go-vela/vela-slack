// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	registry "github.com/go-vela/server/compiler/registry/github"

	"github.com/Masterminds/sprig/v3"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type (
	// Plugin struct represents fields user can present to plugin.
	Plugin struct {
		// webhook to use
		Webhook    string
		Env        *Env
		Path       string
		WebhookMsg *slack.WebhookMessage
		Remote     bool
	}

	// Env struct represents the environment variables the Vela injects
	// https://go-vela.github.io/docs/concepts/pipeline/steps/environment/
	Env struct {
		BuildAuthor               string
		BuildAuthorEmail          string
		BuildAuthorSAMAccountName string
		BuildBranch               string
		BuildChannel              string
		BuildCommit               string
		BuildCreated              int
		BuildEnqueued             int
		BuildEvent                string
		BuildFinished             int
		BuildHost                 string
		BuildLink                 string
		BuildMessage              string
		BuildNumber               int
		BuildParent               int
		BuildRef                  string
		BuildStarted              int
		BuildSource               string
		BuildTag                  string
		BuildTitle                string
		BuildWorkspace            string
		RegistryURL               string
		RepositoryBranch          string
		RepoBranch                string
		RepositoryClone           string
		RepoClone                 string
		RepositoryFullName        string
		RepoFullName              string
		RepositoryLink            string
		RepoLink                  string
		RepositoryName            string
		RepoName                  string
		RepositoryOrg             string
		RepoOrg                   string
		RepositoryPrivate         string
		RepoPrivate               string
		RepositoryTimeout         int
		RepoTimeout               int
		RepositoryTrusted         string
		RepoTrusted               string
		Token                     string
	}
)

// Exec formats and runs the commands for sending a message via Slack.
func (p *Plugin) Exec() error {
	var (
		attachments []slack.Attachment
		err         error
	)
	logrus.Debug("running plugin with provided configuration")

	// clean up newlines that could invalidate JSON
	// BuildMessage is the only field that can have newlines;
	// typically when the commit contains a title and body message
	p.Env.BuildMessage = strings.Replace(p.Env.BuildMessage, "\n", "\\n", -1)

	// create message struct file Slack
	msg := slack.WebhookMessage{
		Username:        p.WebhookMsg.Username,
		IconEmoji:       p.WebhookMsg.IconEmoji,
		IconURL:         p.WebhookMsg.IconURL,
		Channel:         p.WebhookMsg.Channel,
		ThreadTimestamp: p.WebhookMsg.ThreadTimestamp,
		Text:            p.WebhookMsg.Text,
		Parse:           p.WebhookMsg.Parse,
	}

	// parse the slack message file
	if len(p.Path) != 0 {
		logrus.Infof("Parsing provided template file, %s", p.Path)
		if p.Remote {
			attachments, err = getRemoteAttachment(p)

		} else {
			attachments, err = getAttachmentFromFile(p)
		}

		if err != nil {
			return fmt.Errorf("unable to parse attachment file: %w", err)
		}

		msg.Attachments = append(msg.Attachments, attachments...)
	}

	logrus.Info("Marshal webhook message to bytes...")

	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal webhook message: %w", err)
	}

	// for sprig, regex to remove backslashes added when buffer compiles escaped quotes `\"` as `\\\"`
	// and the additional `\\` that are necessary to escape those added backslashes properly
	// a set of `\"` may have been used for valid json between a set of double curly braces `{{ }}`
	// sprig example with `\"`: `<@{{trimAll \"@company.com\" .BuildAuthorEmail | lower }}>`
	bStr := string(b)

	r1, err := regexp.Compile("{{.*?(\\\\\").*?(\\\\\").*?}}")
	if err != nil {
		return fmt.Errorf("unable to execute primary regex: %w", err)
	}

	r2, err := regexp.Compile("(\\\\\")")
	if err != nil {
		return fmt.Errorf("unable to execute secondary regex: %w", err)
	}

	bStr = r1.ReplaceAllStringFunc(bStr, func(m string) string {
		return r2.ReplaceAllString(m, "\"")
	})

	b = []byte(bStr)

	logrus.Info("Parse webhook message payload...")

	tmpl := template.New("slackmessage").Funcs(sprig.TxtFuncMap())

	tmpl, err = tmpl.Parse(string(b))
	if err != nil {
		return fmt.Errorf("unable to parse from webhook message: %w", err)
	}

	logrus.Info("Execute template conversion on webhook message...")

	buffer := new(bytes.Buffer)

	err = tmpl.Execute(buffer, p.Env)
	if err != nil {
		return fmt.Errorf("unable to execute template on webhook message: %w", err)
	}

	logrus.Info("Unmarshal bytes to webhook message...")

	err = json.Unmarshal(buffer.Bytes(), &msg)
	if err != nil {
		return fmt.Errorf("unable to unmarshal webhook message: %w", err)
	}

	logrus.Info("Posting webhook message...")

	err = slack.PostWebhook(p.Webhook, &msg)
	if err != nil {
		return fmt.Errorf("unable to post webhook message: %w", err)
	}

	logrus.Info("Plugin finished...")

	return nil
}

// Validate function to validate plugin configuration.
func (p *Plugin) Validate() error {
	logrus.Debug("validating plugin configuration")

	// validate that a webhook was supplied
	if len(p.Webhook) == 0 {
		return fmt.Errorf("no webhook provided")
	}

	// validate that a message was defined or
	// a path to an attachment template
	if len(p.WebhookMsg.Text) == 0 && len(p.Path) == 0 {
		return fmt.Errorf("must provide text, filepath, or both")
	}

	return nil
}

// getAttachmentFromFile function to open and parse json file into
// slack webhook message payload.
func getAttachmentFromFile(p *Plugin) ([]slack.Attachment, error) {
	// open the provided json template
	jsonFile, err := os.Open(p.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to open json file: %w", err)
	}

	defer jsonFile.Close()

	// read the contents of the json template
	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read json file: %w", err)
	}

	// Converts bytes into string and replaces {{ .BuildCreated }}
	// with a timestamp before returning it back into bytes again.
	bStr := string(bytes)
	// x := strconv.Itoa(p.Env.BuildCreated)
	bStr = strings.Replace(bStr, "{{ .BuildCreated }}", strconv.Itoa(p.Env.BuildCreated), -1)
	bStr = strings.Replace(bStr, "{{ .BuildEnqueued }}", strconv.Itoa(p.Env.BuildEnqueued), -1)
	bStr = strings.Replace(bStr, "{{ .BuildFinished }}", strconv.Itoa(p.Env.BuildFinished), -1)
	bStr = strings.Replace(bStr, "{{ .BuildNumber }}", strconv.Itoa(p.Env.BuildNumber), -1)
	bStr = strings.Replace(bStr, "{{ .BuildParent }}", strconv.Itoa(p.Env.BuildParent), -1)
	bStr = strings.Replace(bStr, "{{ .BuildStarted }}", strconv.Itoa(p.Env.BuildStarted), -1)
	bStr = strings.Replace(bStr, "{{ .RepositoryTimeout }}", strconv.Itoa(p.Env.RepositoryTimeout), -1)
	bytes = []byte(bStr)

	logrus.Info("pip")
	logrus.Info(bStr)
	logrus.Info("pip")
	// create a variable to hold our message
	var msg slack.WebhookMessage

	// cast bytes to go struct
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal json file: %w", err)
	}

	return msg.Attachments, err
}

// getRemoteAttachment function to open and parse json file into
// slack webhook message payload.
func getRemoteAttachment(p *Plugin) ([]slack.Attachment, error) {
	var (
		bytes []byte
		err   error
	)

	reg, err := registry.New(p.Env.RegistryURL, p.Env.Token)
	if err != nil {
		return nil, err
	}

	// parse source from slack attachment
	src, err := reg.Parse(p.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid slack attachment source provided: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"org":  src.Org,
		"repo": src.Repo,
		"path": src.Name,
		"host": src.Host,
	}).Tracef("Using authenticated GitHub client to pull template")

	// use private (authenticated) github instance to pull from
	bytes, err = reg.Template(nil, src)
	if err != nil {
		return nil, err
	}

	// Converts bytes into string and replaces {{ .BuildCreated }}
	// with a timestamp before returning it back into bytes again.
	bStr := string(bytes)
	// x := strconv.Itoa(p.Env.BuildCreated)
	bStr = strings.Replace(bStr, "{{ .BuildCreated }}", strconv.Itoa(p.Env.BuildCreated), -1)
	bStr = strings.Replace(bStr, "{{ .BuildEnqueued }}", strconv.Itoa(p.Env.BuildEnqueued), -1)
	bStr = strings.Replace(bStr, "{{ .BuildFinished }}", strconv.Itoa(p.Env.BuildFinished), -1)
	bStr = strings.Replace(bStr, "{{ .BuildNumber }}", strconv.Itoa(p.Env.BuildNumber), -1)
	bStr = strings.Replace(bStr, "{{ .BuildParent }}", strconv.Itoa(p.Env.BuildParent), -1)
	bStr = strings.Replace(bStr, "{{ .BuildStarted }}", strconv.Itoa(p.Env.BuildStarted), -1)
	bStr = strings.Replace(bStr, "{{ .RepositoryTimeout }}", strconv.Itoa(p.Env.RepositoryTimeout), -1)
	bytes = []byte(bStr)

	logrus.Info("pip")
	logrus.Info(bStr)
	logrus.Info("pip")
	// create a variable to hold our message
	var msg slack.WebhookMessage

	// cast bytes to go struct
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal json file: %w", err)
	}

	return msg.Attachments, err
}
