package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

var genCreditsErr = errors.New("can't derive repository URL from a package not hosted on github.com")

func GenCredits(pkg string, deps map[string]struct{}) error {
	uniqueUrls := map[string]struct{}{}

	for dep := range deps {
		depParts := strings.SplitN(dep, "/", 4)

		if strings.Contains(depParts[0], ".") {
			if depParts[0] == "github.com" {
				uniqueUrls["https://"+strings.Join(depParts[:3], "/")] = struct{}{}
			} else {
				return genCreditsErr
			}
		}
	}

	urls := make([]string, len(uniqueUrls))
	urlIdx := 0

	for url := range uniqueUrls {
		urls[urlIdx] = url
		urlIdx++
	}

	sort.Strings(urls)

	return ioutil.WriteFile(
		"GithubcomMasif_upgraderCommon.go",
		[]byte(fmt.Sprintf("package %s\nvar GithubcomMasif_upgraderCommon = %#v", pkg, urls)),
		0666,
	)
}
