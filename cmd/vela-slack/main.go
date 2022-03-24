// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"time"

	"github.com/go-vela/vela-slack/version"
	"github.com/slack-go/slack"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"crypto/tls"
	"crypto/x509"

	"github.com/go-ldap/ldap/v3"

	_ "github.com/joho/godotenv/autoload"
)

type GitHubUser struct {
	Email string
}

// nolint: funlen // ignore length for main
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
	app := cli.NewApp()

	// Plugin Information

	app.Name = "vela-slack"
	app.HelpName = "vela-slack"
	app.Usage = "Vela Slack plugin for sending data to a Slack channel"
	app.Copyright = "Copyright (c) 2022 Target Brands, Inc. All rights reserved."
	app.Authors = []*cli.Author{
		{
			Name:  "Vela Admins",
			Email: "vela@target.com",
		},
	}

	// Plugin Metadata

	app.Action = run
	app.Compiled = time.Now()
	app.Version = v.Semantic()

	// Plugin Flags

	app.Flags = []cli.Flag{

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LOG_LEVEL", "SLACK_LOG_LEVEL"},
			FilePath: "/vela/parameters/slack/log_level,/vela/secrets/slack/log_level",
			Name:     "log.level",
			Usage:    "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
			Value:    "info",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_SSL_CERT_FILE", "SSL_CERT_FILE"},
			FilePath: "/vela/parameters/sslcert/filepath,/vela/secrets/sslcert/filepath",
			Name:     "sslcert.path",
			Usage:    "path to ssl cert file",
		},

		// Config Flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_FILEPATH", "SLACK_FILEPATH"},
			FilePath: "/vela/parameters/slack/filepath,/vela/secrets/slack/filepath",
			Name:     "filepath",
			Usage:    "file path field for setting a path to a message file",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_WEBHOOK", "SLACK_WEBHOOK"},
			FilePath: "/vela/parameters/slack/webhook,/vela/secrets/slack/webhook",
			Name:     "webhook",
			Usage:    "slack webhook used to post log messages to channel",
		},

		// Webhook Flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_USERNAME", "SLACK_USERNAME"},
			FilePath: "/vela/parameters/slack/username,/vela/secrets/slack/username",
			Name:     "slack-username",
			Usage:    "webhook message field for setting the username",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ICON_EMOJI", "SLACK_ICON_EMOJI"},
			FilePath: "/vela/parameters/slack/icon_emoji,/vela/secrets/slack/icon_emoji",
			Name:     "icon-emoji",
			Usage:    "webhook message field for setting the icon emoji",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ICON_URL", "SLACK_ICON_URL"},
			FilePath: "/vela/parameters/slack/icon_url,/vela/secrets/slack/icon_url",
			Name:     "icon-url",
			Usage:    "webhook message field for setting the icon url",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_CHANNEL", "SLACK_CHANNEL"},
			FilePath: "/vela/parameters/slack/channel,/vela/secrets/slack/channel",
			Name:     "channel",
			Usage:    "webhook message field for setting channel",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_THREAD_TS", "SLACK_THREAD_TS"},
			FilePath: "/vela/parameters/slack/thread_ts,/vela/secrets/slack/thread_ts",
			Name:     "thread-ts",
			Usage:    "webhook message field for setting the thread timestamp",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_TEXT", "SLACK_TEXT"},
			FilePath: "/vela/parameters/slack/text,/vela/secrets/slack/text",
			Name:     "text",
			Usage:    "webhook message field for setting text",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_PARSE", "SLACK_PARSE"},
			FilePath: "/vela/parameters/slack/parse,/vela/secrets/slack/parse",
			Name:     "parse",
			Usage:    "webhook message field for setting parse options",
		},

		// Build Environment Variable Flags

		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_AUTHOR", "BUILD_AUTHOR"},
			Name:    "build-author",
			Usage:   "environment variable reference for reading in build author",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_AUTHOR_EMAIL", "BUILD_AUTHOR_EMAIL"},
			Name:    "build-author-email",
			Usage:   "environment variable reference for reading in build author-email",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_BRANCH", "BUILD_BRANCH"},
			Name:    "build-branch",
			Usage:   "environment variable reference for reading in build branch",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_CHANNEL", "BUILD_CHANNEL"},
			Name:    "build-channel",
			Usage:   "environment variable reference for reading in build channel",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_COMMIT", "BUILD_COMMIT"},
			Name:    "build-commit",
			Usage:   "environment variable reference for reading in build commit",
		},
		&cli.IntFlag{
			EnvVars: []string{"VELA_BUILD_CREATED", "BUILD_CREATED"},
			Name:    "build-created",
			Usage:   "environment variable reference for reading in build created",
		},
		&cli.IntFlag{
			EnvVars: []string{"VELA_BUILD_ENQUEUED", "BUILD_ENQUEUED"},
			Name:    "build-enqueued",
			Usage:   "environment variable reference for reading in build enqueued",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_EVENT", "BUILD_EVENT"},
			Name:    "build-event",
			Usage:   "environment variable reference for reading in build event",
		},
		&cli.IntFlag{
			EnvVars: []string{"VELA_BUILD_FINISHED", "BUILD_FINISHED"},
			Name:    "build-finished",
			Usage:   "environment variable reference for reading in build finished",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_HOST", "BUILD_HOST"},
			Name:    "build-host",
			Usage:   "environment variable reference for reading in build host",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_LINK", "BUILD_LINK"},
			Name:    "build-link",
			Usage:   "environment variable reference for reading in build link",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_MESSAGE", "BUILD_MESSAGE"},
			Name:    "build-message",
			Usage:   "environment variable reference for reading in build message",
		},
		&cli.IntFlag{
			EnvVars: []string{"VELA_BUILD_NUMBER", "BUILD_NUMBER"},
			Name:    "build-number",
			Usage:   "environment variable reference for reading in build number",
		},
		&cli.IntFlag{
			EnvVars: []string{"VELA_BUILD_PARENT", "BUILD_PARENT"},
			Name:    "build-parent",
			Usage:   "environment variable reference for reading in build parent",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_REF", "BUILD_REF"},
			Name:    "build-ref",
			Usage:   "environment variable reference for reading in build ref",
		},
		&cli.IntFlag{
			EnvVars: []string{"VELA_BUILD_STARTED", "BUILD_STARTED"},
			Name:    "build-started",
			Usage:   "environment variable reference for reading in build started",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_SOURCE", "BUILD_SOURCE"},
			Name:    "build-source",
			Usage:   "environment variable reference for reading in build source",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_TAG", "BUILD_TAG"},
			Name:    "build-tag",
			Usage:   "environment variable reference for reading in build tag",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_TITLE", "BUILD_TITLE"},
			Name:    "build-title",
			Usage:   "environment variable reference for reading in build title",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_BUILD_WORKSPACE", "BUILD_WORKSPACE"},
			Name:    "build-workspace",
			Usage:   "environment variable reference for reading in build workspace",
		},

		// Repository Environment Variable Flags

		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_BRANCH", "REPOSITORY_BRANCH"},
			Name:    "repo-branch",
			Usage:   "environment variable reference for reading in repository branch",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_CLONE", "REPOSITORY_CLONE"},
			Name:    "repo-clone",
			Usage:   "environment variable reference for reading in repository clone",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_FULL_NAME", "REPOSITORY_FULL_NAME"},
			Name:    "repo-full-name",
			Usage:   "environment variable reference for reading in repository full name",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_LINK", "REPOSITORY_LINK"},
			Name:    "repo-link",
			Usage:   "environment variable reference for reading in repository link",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_NAME", "REPOSITORY_NAME"},
			Name:    "repo-name",
			Usage:   "environment variable reference for reading in repository name",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_ORG", "REPOSITORY_ORG"},
			Name:    "repo-org",
			Usage:   "environment variable reference for reading in repository org",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_PRIVATE", "REPOSITORY_PRIVATE"},
			Name:    "repo-private",
			Usage:   "environment variable reference for reading in repository private",
		},
		&cli.IntFlag{
			EnvVars: []string{"VELA_REPO_TIMEOUT", "REPOSITORY_TIMEOUT"},
			Name:    "repo-timeout",
			Usage:   "environment variable reference for reading in repository timeout",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_REPO_TRUSTED", "REPOSITORY_TRUSTED"},
			Name:    "repo-trusted",
			Usage:   "environment variable reference for reading in repository trusted",
		},

		// Optional LDAP config flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LDAP_USERNAME", "LDAP_USERNAME"},
			FilePath: string("/vela/parameters/ldap/username,/vela/secrets/ldap/username"),
			Name:     "ldap-username",
			Usage:    "environment variable for LDAP username",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LDAP_PASSWORD", "LDAP_PASSWORD"},
			FilePath: string("/vela/parameters/ldap/password,/vela/secrets/ldap/password"),
			Name:     "ldap-password",
			Usage:    "environment variable for LDAP password",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LDAP_SERVER", "LDAP_SERVER"},
			FilePath: string("/vela/parameters/ldap/server,/vela/secrets/ldap/server"),
			Name:     "ldap-server",
			Usage:    "environment variable for enterprise LDAP server",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LDAP_PORT", "LDAP_PORT"},
			FilePath: string("/vela/parameters/ldap/port,/vela/secrets/ldap/port"),
			Name:     "ldap-port",
			Usage:    "environment variable for enterprise LDAP port",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LDAP_SEARCH_BASE", "LDAP_SEARCH_BASE"},
			FilePath: string("/vela/parameters/ldap/searchbase,/vela/secrets/ldap/searchbase"),
			Name:     "ldap-search-base",
			Usage:    "environment variable for enterprise LDAP search base",
		},

		// Optional GitHub API config flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_GITHUB_ACCESS_TOKEN", "GITHUB_ACCESS_TOKEN"},
			FilePath: string("/vela/parameters/github/token,/vela/secrets/github/token"),
			Name:     "github-access-token",
			Usage:    "environment variable for GitHub access token",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_GITHUB_USERNAME", "GITHUB_USERNAME"},
			FilePath: string("/vela/parameters/github/username,/vela/secrets/github/username"),
			Name:     "github-username",
			Usage:    "environment variable for GitHub username",
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

// run executes the plugin based off the configuration provided.
func run(c *cli.Context) error {
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
		"docs":     "https://go-vela.github.io/docs/plugins/registry/slack",
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
		Env: &Env{
			BuildAuthor:               c.String("build-author"),
			BuildAuthorEmail:          c.String("build-author-email"),
			BuildAuthorSAMAccountName: getSAMAccountName(c),
			BuildBranch:               c.String("build-branch"),
			BuildChannel:              c.String("build-channel"),
			BuildCommit:               c.String("build-commit"),
			BuildCreated:              c.Int("build-created"),
			BuildEnqueued:             c.Int("build-enqueued"),
			BuildEvent:                c.String("build-event"),
			BuildFinished:             c.Int("build-finished"),
			BuildHost:                 c.String("build-host"),
			BuildLink:                 c.String("build-link"),
			BuildMessage:              c.String("build-message"),
			BuildNumber:               c.Int("build-number"),
			BuildParent:               c.Int("build-parent"),
			BuildRef:                  c.String("build-ref"),
			BuildStarted:              c.Int("build-started"),
			BuildSource:               c.String("build-source"),
			BuildTag:                  c.String("build-tag"),
			BuildTitle:                c.String("build-title"),
			BuildWorkspace:            c.String("build-workspace"),
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
			RepositoryTimeout:         c.Int("repo-timeout"),
			RepoTimeout:               c.Int("repo-timeout"),
			RepositoryTrusted:         c.String("repo-trusted"),
			RepoTrusted:               c.String("repo-trusted"),
		},
	}

	// validate the plugin
	err := p.Validate()
	if err != nil {
		return err
	}

	// execute the plugin
	return p.Exec()
}

// Retrieves sAMAccountName from LDAP server using build author's email.
func getSAMAccountName(c *cli.Context) string {
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

	// look up the build author's email address if it wasn't provided in the GitHub event
	if email == "" {
		email = getUserEmail(c)
	}

	// create LDAP client
	roots := x509.NewCertPool()
	caCerts, err := ioutil.ReadFile(c.String("sslcert.path"))

	if err != nil {
		logrus.Errorf("%s", err)
		return ""
	}

	roots.AppendCertsFromPEM(caCerts)

	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: ldapServer,
		RootCAs:    roots,
	}

	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%s", ldapServer, ldapPort), config)
	if err != nil {
		logrus.Errorf("%s", err)
		return ""
	}

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

func getUserEmail(c *cli.Context) string {
	buildSource := c.String("build-source")
	buildAuthor := c.String("build-author")
	githubUsername := c.String("github-username")
	githubAccessToken := c.String("github-access-token")

	if githubUsername == "" || githubAccessToken == "" {
		return ""
	}

	parsedUrl, err := url.Parse(buildSource)
	if err != nil {
		logrus.Errorf("unable to parse build source as URL: %s", err)
		return ""
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/api/v3/users/%s", parsedUrl.Scheme, parsedUrl.Host, buildAuthor), nil)
	if err != nil {
		logrus.Errorf("unable to create GitHub API request: %s", err)
		return ""
	}
	req.SetBasicAuth(githubUsername, githubAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("unable to fetch email address from GitHub: %s", err)
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("unable to read GitHub API response: %s", err)
		return ""
	}

	var user GitHubUser
	err = json.Unmarshal([]byte(body), &user)
	if err != nil {
		logrus.Errorf("unable to unmarshal GitHub API response: %s", err)
		return ""
	}

	return user.Email
}
