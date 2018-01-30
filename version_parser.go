package main

import (
	"fmt"
	"regexp"
	"strings"
)

var vscodeVersionRegex = regexp.MustCompile(`(\d+\.\d+\.\d+\-\d+)_amd64\.tar\.gz$`)

func parseVersion(URL string) (string, error) {
	matches := vscodeVersionRegex.FindStringSubmatch(URL)
	if len(matches) == 0 {
		return "", fmt.Errorf("could not find version substring")
	}
	return strings.Replace(matches[1], "-", ".", -1), nil
}
