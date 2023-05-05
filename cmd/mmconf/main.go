package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/isacikgoz/mmconf/internal/clients"
	"github.com/isacikgoz/mmconf/internal/config"
	"github.com/isacikgoz/mmconf/internal/docs"
	"github.com/isacikgoz/prompt"
	"github.com/mattermost/mattermost-server/server/v8/model"
)

const (
	openaiDocsAddress = "https://platform.openai.com/account/api-keys"
	questionFormat    = "what is %s setting in mattermost server configuration?"
	basicAuth         = "Basic(username/password)"
	tokenAuth         = "Token(uses MM_AUTHTOKEN variable)"
)

func main() {
	tok := os.Getenv("OPENAI_APIKEY")
	if tok == "" {
		fmt.Printf("It appears that you didn't set the OpenAI API key.\n"+
			"The key is required to acceess ChatGPT API, please follow the instructions from %s\n"+
			"Also please set it via an environment variable %s\n", openaiDocsAddress, "OPENAI_APIKEY")
		os.Exit(1)
	}

	faint := color.New(color.Faint)

	configurations, err := docs.ParseDocs()
	exitIfErr(err)

	serverAddr := "http://localhost:8065"
	faint.Print("What is your current mattermost URL? ")
	fmt.Scanln(&serverAddr)

	sel, err := prompt.NewSelection("How do you want to authentication you would like to use?",
		[]string{basicAuth, tokenAuth}, "", 2)
	exitIfErr(err)

	answer, err := sel.Run()
	exitIfErr(err)

	client := model.NewAPIv4Client(serverAddr)
	if answer == basicAuth {
		username := ""
		faint.Print("What is your username? ")
		fmt.Scanln(&username)
		password := ""
		faint.Print("What is your password? ")
		fmt.Scanln(&password)
		_, _, err = client.Login(username, password)
		exitIfErr(err)
	} else {
		client.HTTPHeader = map[string]string{"User-Agent": "mmconf"}

		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		client.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
				Proxy:           http.ProxyFromEnvironment,
			},
		}
		if os.Getenv("MM_AUTHTOKEN") == "" {
			fmt.Printf("It appears that you didn't set the Mattermost auth token.\n"+
				"Please set it via an environment variable %s\n", "MM_AUTHTOKEN")
			os.Exit(1)
		}
		client.AuthType = model.HeaderBearer
		client.AuthToken = os.Getenv("MM_AUTHTOKEN")

		_, _, err = client.GetMe("")
		exitIfErr(err)
	}

	faint.Println("Authentication successful!")

	for _, conf := range configurations {
		fmt.Println(color.CyanString(conf))
		answer, err := clients.AskChatGPT(context.Background(), tok, fmt.Sprintf(questionFormat, conf))
		exitIfErr(err)
		fmt.Printf("%s\n%s\n%s\n", strings.Repeat("-", 60), answer, strings.Repeat("-", 60))
		setting := ""
		faint.Printf("Would you like to change %s (leave empty to skip)? ", conf)
		fmt.Scanln(&setting)
		if setting != "" {
			cfg, _, err := client.GetConfig()
			exitIfErr(err)

			if cErr := config.SetConfigValue(strings.Split(conf, "."), cfg, []string{setting}); cErr != nil {
				exitIfErr(err)
			}

			_, _, err = client.PatchConfig(cfg)
			exitIfErr(err)
		}
	}
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err)
		os.Exit(1)
	}
}
