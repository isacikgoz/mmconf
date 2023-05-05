# mmconf

A CLI configuration management tool for mattermost-server.

## Usage

```sh
mmconf
```

You will need to set `OPENAI_APIKEY` to use ChatGPT. Also the tool expects you to have `MM_AUTHTOKEN` if you'd like to authenticate with your server via token. 

The tool will parse mattermost site configuration settings [page](https://docs.mattermost.com/configure/site-configuration-settings.html) and leverage ChatGPT to explain how to configure your server. It will prompt you to provide desired configuration settings and apply them via mattermost client.

This is an experimental project to let mattermost administartors configure their service with AI assistance.
