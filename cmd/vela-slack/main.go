// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/mail"
	"os"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/urfave/cli/v3"

	_ "github.com/joho/godotenv/autoload"

	"github.com/go-vela/vela-slack/version"
)

//nolint:funlen // ignore length for main
func main() {
	// capture application version information
	v := version.New()

	// serialize the version information as pretty JSON
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logrus.Fatal(err)
	}

	// output the version information to stdout
	fmt.Fprintf(os.Stdout, "%s\n", string(bytes))

	// create new CLI application
	cmd := &cli.Command{
		Name:      "vela-slack",
		Usage:     "Vela Slack plugin for sending data to a Slack channel",
		Copyright: "Copyright 2021 Target Brands, Inc. All rights reserved.",
		Authors: []any{
			&mail.Address{
				Name:    "Vela Admins",
				Address: "vela@target.com",
			},
		},
		// Plugin Metadata
		Version: v.Semantic(),
		Action:  run,
	}

	// Plugin Flags
	cmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "log.level",
			Value: "info",
			Usage: "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_LOG_LEVEL"),
				cli.EnvVar("SLACK_LOG_LEVEL"),
				cli.File("/vela/parameters/slack/log_level"),
				cli.File("/vela/secrets/slack/log_level"),
			),
		},
		&cli.StringFlag{
			Name:  "sslcert.path",
			Usage: "path to ssl cert file",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SSL_CERT_FILE"),
				cli.EnvVar("SSL_CERT_FILE"),
				cli.File("/vela/parameters/ssl_cert_file/filepath"),
				cli.File("/vela/secrets/ssl_cert_file/filepath"),
			),
		},

		// Config Flags

		&cli.StringFlag{
			Name:  "filepath",
			Usage: "file path field for setting a path to a message file",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_FILEPATH"),
				cli.EnvVar("SLACK_FILEPATH"),
				cli.File("/vela/parameters/slack/filepath"),
				cli.File("/vela/secrets/slack/filepath"),
			),
		},
		&cli.StringFlag{
			Name:  "webhook",
			Usage: "slack webhook used to post log messages to channel",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_WEBHOOK"),
				cli.EnvVar("SLACK_WEBHOOK"),
				cli.File("/vela/parameters/slack/webhook"),
				cli.File("/vela/secrets/slack/webhook"),
			),
		},
		&cli.BoolFlag{
			Name:  "remote",
			Usage: "if filepath is remote or not",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_REMOTE"),
				cli.EnvVar("SLACK_REMOTE"),
				cli.File("/vela/parameters/slack/remote"),
				cli.File("/vela/secrets/slack/remote"),
			),
		},

		// Webhook Flags

		&cli.StringFlag{
			Name:  "slack-username",
			Usage: "webhook message field for setting the username",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_USERNAME"),
				cli.EnvVar("SLACK_USERNAME"),
				cli.File("/vela/parameters/slack/username"),
				cli.File("/vela/secrets/slack/username"),
			),
		},
		&cli.StringFlag{
			Name:  "icon-emoji",
			Usage: "webhook message field for setting the icon emoji",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_ICON_EMOJI"),
				cli.EnvVar("SLACK_ICON_EMOJI"),
				cli.File("/vela/parameters/slack/icon_emoji"),
				cli.File("/vela/secrets/slack/icon_emoji"),
			),
		},
		&cli.StringFlag{
			Name:  "icon-url",
			Usage: "webhook message field for setting the icon url",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_ICON_URL"),
				cli.EnvVar("SLACK_ICON_URL"),
				cli.File("/vela/parameters/slack/icon_url"),
				cli.File("/vela/secrets/slack/icon_url"),
			),
		},
		&cli.StringFlag{
			Name:  "channel",
			Usage: "webhook message field for setting channel",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_CHANNEL"),
				cli.EnvVar("SLACK_CHANNEL"),
				cli.File("/vela/parameters/slack/channel"),
				cli.File("/vela/secrets/slack/channel"),
			),
		},
		&cli.StringFlag{
			Name:  "thread-ts",
			Usage: "webhook message field for setting the thread timestamp",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_THREAD_TS"),
				cli.EnvVar("SLACK_THREAD_TS"),
				cli.File("/vela/parameters/slack/thread_ts"),
				cli.File("/vela/secrets/slack/thread_ts"),
			),
		},
		&cli.StringFlag{
			Name:  "text",
			Usage: "webhook message field for setting text",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_TEXT"),
				cli.EnvVar("SLACK_TEXT"),
				cli.File("/vela/parameters/slack/text"),
				cli.File("/vela/secrets/slack/text"),
			),
		},
		&cli.StringFlag{
			Name:  "parse",
			Usage: "webhook message field for setting parse options",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_PARSE"),
				cli.EnvVar("SLACK_PARSE"),
				cli.File("/vela/parameters/slack/parse"),
				cli.File("/vela/secrets/slack/parse"),
			),
		},

		// Build Environment Variable Flags

		&cli.StringFlag{
			Name:    "build-author",
			Usage:   "environment variable reference for reading in build author",
			Sources: cli.EnvVars("VELA_BUILD_AUTHOR", "BUILD_AUTHOR"),
		},
		&cli.StringFlag{
			Name:    "build-author-email",
			Usage:   "environment variable reference for reading in build author-email",
			Sources: cli.EnvVars("VELA_BUILD_AUTHOR_EMAIL", "BUILD_AUTHOR_EMAIL"),
		},
		&cli.StringFlag{
			Name:    "build-branch",
			Usage:   "environment variable reference for reading in build branch",
			Sources: cli.EnvVars("VELA_BUILD_BRANCH", "BUILD_BRANCH"),
		},
		&cli.StringFlag{
			Name:    "build-channel",
			Usage:   "environment variable reference for reading in build channel",
			Sources: cli.EnvVars("VELA_BUILD_CHANNEL", "BUILD_CHANNEL"),
		},
		&cli.StringFlag{
			Name:    "build-commit",
			Usage:   "environment variable reference for reading in build commit",
			Sources: cli.EnvVars("VELA_BUILD_COMMIT", "BUILD_COMMIT"),
		},
		&cli.IntFlag{
			Name:    "build-created",
			Usage:   "environment variable reference for reading in build created",
			Sources: cli.EnvVars("VELA_BUILD_CREATED", "BUILD_CREATED"),
		},
		&cli.IntFlag{
			Name:    "build-enqueued",
			Usage:   "environment variable reference for reading in build enqueued",
			Sources: cli.EnvVars("VELA_BUILD_ENQUEUED", "BUILD_ENQUEUED"),
		},
		&cli.StringFlag{
			Name:    "build-event",
			Usage:   "environment variable reference for reading in build event",
			Sources: cli.EnvVars("VELA_BUILD_EVENT", "BUILD_EVENT"),
		},
		&cli.IntFlag{
			Name:    "build-finished",
			Usage:   "environment variable reference for reading in build finished",
			Sources: cli.EnvVars("VELA_BUILD_FINISHED", "BUILD_FINISHED"),
		},
		&cli.StringFlag{
			Name:    "build-host",
			Usage:   "environment variable reference for reading in build host",
			Sources: cli.EnvVars("VELA_BUILD_HOST", "BUILD_HOST"),
		},
		&cli.StringFlag{
			Name:    "build-link",
			Usage:   "environment variable reference for reading in build link",
			Sources: cli.EnvVars("VELA_BUILD_LINK", "BUILD_LINK"),
		},
		&cli.StringFlag{
			Name:    "build-message",
			Usage:   "environment variable reference for reading in build message",
			Sources: cli.EnvVars("VELA_BUILD_MESSAGE", "BUILD_MESSAGE"),
		},
		&cli.IntFlag{
			Name:    "build-number",
			Usage:   "environment variable reference for reading in build number",
			Sources: cli.EnvVars("VELA_BUILD_NUMBER", "BUILD_NUMBER"),
		},
		&cli.IntFlag{
			Name:    "build-parent",
			Usage:   "environment variable reference for reading in build parent",
			Sources: cli.EnvVars("VELA_BUILD_PARENT", "BUILD_PARENT"),
		},
		&cli.StringFlag{
			Name:    "build-ref",
			Usage:   "environment variable reference for reading in build ref",
			Sources: cli.EnvVars("VELA_BUILD_REF", "BUILD_REF"),
		},
		&cli.StringFlag{
			Name:    "build-sender",
			Usage:   "environment variable reference for reading in build sender",
			Sources: cli.EnvVars("VELA_BUILD_SENDER", "BUILD_SENDER"),
		},
		&cli.IntFlag{
			Name:    "build-started",
			Usage:   "environment variable reference for reading in build started",
			Sources: cli.EnvVars("VELA_BUILD_STARTED", "BUILD_STARTED"),
		},
		&cli.StringFlag{
			Name:    "build-source",
			Usage:   "environment variable reference for reading in build source",
			Sources: cli.EnvVars("VELA_BUILD_SOURCE", "BUILD_SOURCE"),
		},
		&cli.StringFlag{
			Name:    "build-tag",
			Usage:   "environment variable reference for reading in build tag",
			Sources: cli.EnvVars("VELA_BUILD_TAG", "BUILD_TAG"),
		},
		&cli.StringFlag{
			Name:    "build-title",
			Usage:   "environment variable reference for reading in build title",
			Sources: cli.EnvVars("VELA_BUILD_TITLE", "BUILD_TITLE"),
		},
		&cli.StringFlag{
			Name:    "build-workspace",
			Usage:   "environment variable reference for reading in build workspace",
			Sources: cli.EnvVars("VELA_BUILD_WORKSPACE", "BUILD_WORKSPACE"),
		},

		// Repository Environment Variable Flags

		&cli.StringFlag{
			Name:    "repo-branch",
			Usage:   "environment variable reference for reading in repository branch",
			Sources: cli.EnvVars("VELA_REPO_BRANCH", "REPOSITORY_BRANCH"),
		},
		&cli.StringFlag{
			Name:    "repo-clone",
			Usage:   "environment variable reference for reading in repository clone",
			Sources: cli.EnvVars("VELA_REPO_CLONE", "REPOSITORY_CLONE"),
		},
		&cli.StringFlag{
			Name:    "repo-full-name",
			Usage:   "environment variable reference for reading in repository full name",
			Sources: cli.EnvVars("VELA_REPO_FULL_NAME", "REPOSITORY_FULL_NAME"),
		},
		&cli.StringFlag{
			Name:    "repo-link",
			Usage:   "environment variable reference for reading in repository link",
			Sources: cli.EnvVars("VELA_REPO_LINK", "REPOSITORY_LINK"),
		},
		&cli.StringFlag{
			Name:    "repo-name",
			Usage:   "environment variable reference for reading in repository name",
			Sources: cli.EnvVars("VELA_REPO_NAME", "REPOSITORY_NAME"),
		},
		&cli.StringFlag{
			Name:    "repo-org",
			Usage:   "environment variable reference for reading in repository org",
			Sources: cli.EnvVars("VELA_REPO_ORG", "REPOSITORY_ORG"),
		},
		&cli.StringFlag{
			Name:    "repo-private",
			Usage:   "environment variable reference for reading in repository private",
			Sources: cli.EnvVars("VELA_REPO_PRIVATE", "REPOSITORY_PRIVATE"),
		},
		&cli.IntFlag{
			Name:    "repo-timeout",
			Usage:   "environment variable reference for reading in repository timeout",
			Sources: cli.EnvVars("VELA_REPO_TIMEOUT", "REPOSITORY_TIMEOUT"),
		},
		&cli.StringFlag{
			Name:    "repo-trusted",
			Usage:   "environment variable reference for reading in repository trusted",
			Sources: cli.EnvVars("VELA_REPO_TRUSTED", "REPOSITORY_TRUSTED"),
		},

		// Registry

		&cli.StringFlag{
			Name:  "registry-url",
			Usage: "registry url",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_REGISTRY"),
				cli.EnvVar("REGISTRY_URL"),
				cli.File("/vela/parameters/slack/registry_url"),
				cli.File("/vela/secrets/slack/registry_url"),
			),
		},

		// Optional LDAP config flags

		&cli.StringFlag{
			Name:  "ldap-username",
			Usage: "environment variable for LDAP username",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_LDAP_USERNAME"),
				cli.EnvVar("LDAP_USERNAME"),
				cli.File("/vela/parameters/slack/ldap_username"),
				cli.File("/vela/secrets/slack/ldap_username"),
			),
		},
		&cli.StringFlag{
			Name:  "ldap-password",
			Usage: "environment variable for LDAP password",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_LDAP_PASSWORD"),
				cli.EnvVar("LDAP_PASSWORD"),
				cli.File("/vela/parameters/slack/ldap_password"),
				cli.File("/vela/secrets/slack/ldap_password"),
			),
		},
		&cli.StringFlag{
			Name:  "ldap-server",
			Usage: "environment variable for enterprise LDAP server",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_LDAP_SERVER"),
				cli.EnvVar("LDAP_SERVER"),
				cli.File("/vela/parameters/slack/ldap_server"),
				cli.File("/vela/secrets/slack/ldap_server"),
			),
		},
		&cli.StringFlag{
			Name:  "ldap-port",
			Usage: "environment variable for enterprise LDAP port",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_LDAP_PORT"),
				cli.EnvVar("LDAP_PORT"),
				cli.File("/vela/parameters/slack/ldap_port"),
				cli.File("/vela/secrets/slack/ldap_port"),
			),
		},
		&cli.StringFlag{
			Name:  "ldap-search-base",
			Usage: "environment variable for enterprise LDAP search base",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_LDAP_SEARCH_BASE"),
				cli.EnvVar("LDAP_SEARCH_BASE"),
				cli.File("/vela/parameters/slack/ldap_search_base"),
				cli.File("/vela/secrets/slack/ldap_search_base"),
			),
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "github token from user",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_TOKEN"),
				cli.EnvVar("GITHUB_TOKEN"),
				cli.File("/vela/parameters/slack/token"),
				cli.File("/vela/secrets/slack/token"),
			),
		},
	}

	err = cmd.Run(context.Background(), os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

// run executes the plugin based off the configuration provided.
func run(ctx context.Context, c *cli.Command) error {
	// set the log level for the plugin
	switch c.String("log.level") {
	case "t", "trace", "Trace", "TRACE":
		logrus.SetLevel(logrus.TraceLevel)
	case "d", "debug", "Debug", "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "w", "warn", "Warn", "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "e", "error", "Error", "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "f", "fatal", "Fatal", "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "p", "panic", "Panic", "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	case "i", "info", "Info", "INFO":
		fallthrough
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.WithFields(logrus.Fields{
		"code":     "https://github.com/go-vela/vela-slack",
		"docs":     "https://go-vela.github.io/docs/plugins/registry/pipeline/slack",
		"registry": "https://hub.docker.com/r/target/vela-slack",
	}).Info("Vela Slack Plugin")

	// create the plugin
	p := &Plugin{
		Webhook: c.String("webhook"),
		Path:    c.String("filepath"),
		WebhookMsg: &slack.WebhookMessage{
			Username:        c.String("slack-username"),
			IconEmoji:       c.String("icon-emoji"),
			IconURL:         c.String("icon-url"),
			Channel:         c.String("channel"),
			ThreadTimestamp: c.String("thread-ts"),
			Text:            c.String("text"),
			Parse:           c.String("parse"),
		},
		Remote: c.Bool("remote"),
		Env: &Env{
			BuildAuthor:               c.String("build-author"),
			BuildAuthorEmail:          c.String("build-author-email"),
			BuildAuthorSAMAccountName: getSAMAccountName(c),
			BuildBranch:               c.String("build-branch"),
			BuildChannel:              c.String("build-channel"),
			BuildCommit:               c.String("build-commit"),
			BuildCreated:              int64(c.Int("build-created")),
			BuildEnqueued:             int64(c.Int("build-enqueued")),
			BuildEvent:                c.String("build-event"),
			BuildFinished:             int64(c.Int("build-finished")),
			BuildHost:                 c.String("build-host"),
			BuildLink:                 c.String("build-link"),
			BuildMessage:              c.String("build-message"),
			BuildNumber:               int64(c.Int("build-number")),
			BuildParent:               int64(c.Int("build-parent")),
			BuildRef:                  c.String("build-ref"),
			BuildSender:               c.String("build-sender"),
			BuildStarted:              int64(c.Int("build-started")),
			BuildSource:               c.String("build-source"),
			BuildTag:                  c.String("build-tag"),
			BuildTitle:                c.String("build-title"),
			BuildWorkspace:            c.String("build-workspace"),
			RegistryURL:               c.String("registry-url"),
			RepositoryBranch:          c.String("repo-branch"),
			RepoBranch:                c.String("repo-branch"),
			RepositoryClone:           c.String("repo-clone"),
			RepoClone:                 c.String("repo-clone"),
			RepositoryFullName:        c.String("repo-full-name"),
			RepoFullName:              c.String("repo-full-name"),
			RepositoryLink:            c.String("repo-link"),
			RepoLink:                  c.String("repo-link"),
			RepositoryName:            c.String("repo-name"),
			RepoName:                  c.String("repo-name"),
			RepositoryOrg:             c.String("repo-org"),
			RepoOrg:                   c.String("repo-org"),
			RepositoryPrivate:         c.String("repo-private"),
			RepoPrivate:               c.String("repo-private"),
			RepositoryTimeout:         int64(c.Int("repo-timeout")),
			RepoTimeout:               int64(c.Int("repo-timeout")),
			RepositoryTrusted:         c.String("repo-trusted"),
			RepoTrusted:               c.String("repo-trusted"),
			Token:                     c.String("token"),
		},
	}

	// validate the plugin
	err := p.Validate()
	if err != nil {
		return err
	}

	// execute the plugin
	return p.Exec(ctx)
}

// Retrieves sAMAccountName from LDAP server using build author's email.
func getSAMAccountName(c *cli.Command) string {
	// LDAP environment variables
	email := c.String("build-author-email")
	username := c.String("ldap-username")
	password := c.String("ldap-password")
	ldapServer := c.String("ldap-server")
	ldapPort := c.String("ldap-port")
	ldapSearchBase := c.String("ldap-search-base")

	// return if LDAP info not provided
	if username == "" || password == "" {
		return ""
	}

	// create LDAP client
	roots := x509.NewCertPool()

	caCerts, err := os.ReadFile(c.String("sslcert.path"))
	if err != nil {
		logrus.Errorf("%s", err)
		return ""
	}

	roots.AppendCertsFromPEM(caCerts)

	configTLS := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: ldapServer,
		RootCAs:    roots,
	}

	// forcing LDAPS scheme
	// TODO: allow to define scheme as a plugin option
	serverFQDN := fmt.Sprintf("ldaps://%s:%s", ldapServer, ldapPort)

	l, err := ldap.DialURL(serverFQDN, ldap.DialWithTLSConfig(configTLS))
	if err != nil {
		logrus.Errorf("%s", err)
		return ""
	}
	defer l.Close()

	err = l.Bind(username, password)
	if err != nil {
		logrus.Errorf("%s", err)
		return ""
	}

	filter := fmt.Sprintf("mail=%s", email)
	req := ldap.NewSearchRequest(
		ldapSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(%s)", filter),
		[]string{"dn", "displayName", "sAMAccountName", "mail"},
		nil,
	)

	// search for records
	sr, err := l.Search(req)
	if err != nil {
		logrus.Errorf("%s", err)
		return ""
	}

	if len(sr.Entries) != 1 {
		logrus.Errorf("user does not exist or too many entries returned: %d", len(sr.Entries))
		return ""
	}

	// return sAMAccountName
	sAMAccountName := sr.Entries[0].GetAttributeValue("sAMAccountName")

	return sAMAccountName
}
