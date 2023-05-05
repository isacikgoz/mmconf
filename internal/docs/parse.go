package docs

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"
)

const (
	docsFile = "https://raw.githubusercontent.com/mattermost/docs/master/source/configure/site-configuration-settings.rst"
)

var rg = regexp.MustCompile(`:configjson:\s(.+)`)

func ParseDocs() ([]string, error) {
	resp, err := http.Get(docsFile)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve documentation file from GitHub: %w", err)
	}
	defer resp.Body.Close()

	confs := make([]string, 0)
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Text()
		if !rg.MatchString(line) {
			continue
		}

		sm := rg.FindStringSubmatch(line)[1]
		if sm == "N/A" {
			continue
		}

		confs = append(confs, sm[1:])
	}

	return confs, nil
}
