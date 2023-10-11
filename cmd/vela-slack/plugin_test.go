// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slack-go/slack"
)

func TestSlack_Plugin_Validate(t *testing.T) {
	// setup types
	p := &Plugin{
		Webhook: "webhook_url",
		Env:     &Env{},
		Path:    "",
		WebhookMsg: &slack.WebhookMessage{
			Text: "hello",
		},
	}

	err := p.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestSlack_Plugin_Validate_Missing_Webhook(t *testing.T) {
	// setup types
	p := &Plugin{
		Webhook:    "",
		Env:        &Env{},
		Path:       "",
		WebhookMsg: &slack.WebhookMessage{},
	}

	err := p.Validate()
	if err == nil {
		t.Error("Validate should return err due to missing webhook")
	}
}

func TestSlack_Plugin_Validate_Missing_Text_And_Path(t *testing.T) {
	// setup types
	p := &Plugin{
		Webhook:    "webhook_url",
		Env:        &Env{},
		Path:       "",
		WebhookMsg: &slack.WebhookMessage{},
	}

	err := p.Validate()
	if err == nil {
		t.Error("Validate should return err due to missing text and filepath")
	}
}

func TestSlack_Plugin_Exec(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook: ts.URL,
		Env:     &Env{},
		Path:    "",
		WebhookMsg: &slack.WebhookMessage{
			Text: "hello",
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Exec returned err: %v", err)
	}
}

func TestSlack_Plugin_Exec_Attachment(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook: ts.URL,
		Env:     &Env{},
		Path:    "testdata/slack_attachment.json",
		WebhookMsg: &slack.WebhookMessage{
			Text: "hello",
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Exec returned err: %v", err)
	}
}

func TestSlack_Plugin_Exec_Bad_Attachment(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook:    ts.URL,
		Env:        &Env{},
		Path:       "testdata/slack_attachment_bad.json",
		WebhookMsg: &slack.WebhookMessage{},
	}

	err := p.Exec()
	if err == nil {
		t.Error("Exec should return err due to invalid JSON file")
	}
}

func TestSlack_Plugin_Exec_Bad_File_Ref(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook:    ts.URL,
		Env:        &Env{},
		Path:       "testdata/slack_attachment_404.json",
		WebhookMsg: &slack.WebhookMessage{},
	}

	err := p.Exec()
	if err == nil {
		t.Error("Exec should return err due to file not existing")
	}
}

func TestSlack_Plugin_Exec_Newline(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook: ts.URL,
		Env: &Env{
			BuildMessage: `Testing
Newlines`,
		},
		Path:       "testdata/slack_attachment.json",
		WebhookMsg: &slack.WebhookMessage{},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Exec returned err: %v", err)
	}
}

func TestSlack_Plugin_Exec_Newline_Embedded(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook: ts.URL,
		Env: &Env{
			BuildMessage: `Testing
Newlines`,
		},
		Path: "",
		WebhookMsg: &slack.WebhookMessage{
			Text: "Build Message: {{ .BuildMessage }}",
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Exec returned err: %v", err)
	}
}

func TestSlack_Plugin_Exec_Sprig_Text(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook: ts.URL,
		Env:     &Env{},
		Path:    "",
		WebhookMsg: &slack.WebhookMessage{
			Text: "{{ .BuildAuthorEmail | lower }}",
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Exec returned err: %v", err)
	}
}

func TestSlack_Plugin_Exec_Remove_Escape_Chara(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook: ts.URL,
		Env:     &Env{},
		Path:    "",
		WebhookMsg: &slack.WebhookMessage{
			Text: "{{ trimAll \"@company.com\" .BuildAuthorEmail }}",
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Exec returned err: %v", err)
	}
}

func TestSlack_Plugin_Exec_Do_Not_Remove_Escape_Chara(t *testing.T) {
	// setup types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	p := &Plugin{
		Webhook: ts.URL,
		Env:     &Env{},
		Path:    "",
		WebhookMsg: &slack.WebhookMessage{
			Text: "{\"hello\": \"world\", \"hello_world\": true, \"urls\": {\"url_one\": \"https://github.com\", \"url_two\": \"https://github.com/octocat\"}}",
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Exec returned err: %v", err)
	}
}
