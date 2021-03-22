# Quickly get started using slack-git-compare

## Requirements

- **~10 min of your time**

- A personal access token on whether or both:
  - [github.com](https://github.com/settings/tokens) (or your own instance) with the `repo` scope
  - [gitlab.com](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) (or your own instance) with the `read_api` scope

- [git](https://git-scm.com/) & [docker-compose](https://docs.docker.com/compose/)
- You will also need network connectivity from Slack public endpoints towards the local process you are about to start. If you are attempting this from your laptop, you will probably need something like [ngrok](https://ngrok.com/) or equivalent to be able to do so.

## 🚀

### Start ngrok to get an externally available endpoint for the process

You can skip this step if you can manage to get external access onto the app via other means.

Otherwise, if you do not have it already, [download & install ngrok](https://ngrok.com/download).

```bash
~$ ngrok http 8080 --region eu
ngrok by @inconshreveable                                      (Ctrl+C to quit)
                                                                               
Session Status                online
Version                       2.3.35                                           
Region                        Europe (eu)                                      
Web Interface                 http://127.0.0.1:4040                            
Forwarding                    http://92fb34b5f2ee.eu.ngrok.io -> http://localho
Forwarding                    https://92fb34b5f2ee.eu.ngrok.io -> http://localh
                                                                               
Connections                   ttl     opn     rt1     rt5     p50     p90      
                              0       0       0.00    0.00    0.00    0.00
```

You will be able to use `https://92fb34b5f2ee.eu.ngrok.io`

### Create and configure the Slack app

1. Create the Slack app at https://api.slack.com/apps

![create-slack-app](/docs/images/create-slack-app.png)

2. Configure Interactivity & Shortcuts pane

Set both:
- _"Interactivity > Request URL"_ with _endpoint_**/slack/modal**
- _"Select Menus > Options Load URL"_ with _endpoint_**/slack/select**


![interactivity-and-shortcuts](/docs/images/interactivity-and-shortcuts.png)

3. In the _"Slash Commands"_ pane, create a new slash command

As the URL, use _endpoint_**/slack/slash**

![create-new-command](/docs/images/create-new-command.png)

4. Configure oauth2 scopes and install the app in your workspace!

Set all of the following scopes:
- `chat:write`
- `chat:write:public`
- `commands`
- `users:read`
- `users:read:email`

![oauth-scopes](/docs/images/oauth-scopes.png)

5. Fetch your app **token** and **signing secret**

![slack-token](/docs/images/slack-token.png)

![signing-secret](/docs/images/signing-secret.png)


### Start the app locally

```bash
# Clone this repository
~$ git clone https://github.com/mvisonneau/slack-git-compare.git
~$ cd slack-git-compare/examples/quickstart

# Configure the container according to your needs, you need to provide
# the Slack token and signing-secret as well as at least one git provider
# token and org/group you want to work onto from Slack, eg:

```yaml
# Edit docker-compose.yml
version: '3.8'
services:
  slack-git-compare:
    image: docker.io/mvisonneau/slack-git-compare:latest
    ports:
      - 8080:8080
    environment:
      SGC_SLACK_TOKEN: xoxb-123456789-xxx-xxx
      SGC_SLACK_SIGNING_SECRET: 123456789xxxxxx
      SGC_GITHUB_TOKEN: xxxxx
      SGC_GITHUB_ORG: cilium,foo,bar
      SGC_GITLAB_TOKEN: xxxx
      SGC_GITLAB_GROUP: gitlab-org,foo,bar
```

```bash
# Start slack-git-compare container !
~$ docker-compose up -d        
Creating network "quickstart_default" with driver "bridge"
Creating quickstart_slack-git-compare_1 ... done

# Logs should be looking like this:
~$ docker-compose logs -f
docker-compose logs -f
Attaching to quickstart_slack-git-compare_1
slack-git-compare_1  | time="2021-03-22T15:44:47Z" level=info msg="fetched repositories from provider" count=89 provider=github
slack-git-compare_1  | time="2021-03-22T15:44:48Z" level=info msg="fetched repositories from provider" count=123 provider=gitlab
slack-git-compare_1  | time="2021-03-22T15:44:48Z" level=info msg="http server started" listen-address=":8080"

```

### Try it out from Slack!

You should now be all set and able to trigger the `/compare` command from the Slack workspace in which
you have installed your application.

## Troubleshooting

### Slackbot /compare failed with the error "dispatch_failed"

This error means that Slack was not able to access your public endpoints. Please verify that the values
which are set on the app configuration regarding those endpoints are correct. In a nominal state, if you
are leveraging ngrok you should be able to see the requests coming through, eg:

```
HTTP Requests                                                                  
-------------                                                                  
                                                                               
POST /slack/modal              200 OK                                          
POST /slack/modal              200 OK                                          
POST /slack/modal              200 OK                                          
POST /slack/select             200 OK                                          
POST /slack/modal              200 OK                                          
POST /slack/select             200 OK                                          
POST /slack/modal              200 OK                                          
POST /slack/select             200 OK                                          
POST /slack/slash              200 OK 
```

### Attempt to reach your container from the public address

After a few seconds, you should be able to query the URL you got from `ngrok`:

```bash
~$ curl -i https://92fb34b5f2ee.eu.ngrok.io/health/ready
HTTP/2 200 
content-type: application/json; charset=utf-8
date: Mon, 22 Mar 2021 16:03:56 GMT
content-length: 3

{}
```

## Cleanup

- SIGINT ngrok
- `docker-compose down`
- delete the app in your slack workspace
