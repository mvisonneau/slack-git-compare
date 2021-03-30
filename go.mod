module github.com/mvisonneau/slack-git-compare

go 1.16

require (
	github.com/creasty/defaults v1.5.1
	github.com/felixge/httpsnoop v1.0.1
	github.com/go-playground/validator/v10 v10.4.1
	github.com/google/go-github/v33 v33.0.0
	github.com/gorilla/mux v1.8.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/lithammer/fuzzysearch v1.1.1
	github.com/mvisonneau/go-helpers v0.0.1
	github.com/openlyinc/pointy v1.1.2
	github.com/prometheus/client_golang v1.10.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.8.2
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmihailenco/taskq/v3 v3.2.3
	github.com/xanzy/go-gitlab v0.48.0
	github.com/xeonx/timeago v1.0.0-rc4
	golang.org/x/oauth2 v0.0.0-20210323180902-22b0adad7558
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/vmihailenco/taskq/v3 => github.com/mvisonneau/taskq/v3 v3.2.4-0.20201127170227-fddacd1811f5
