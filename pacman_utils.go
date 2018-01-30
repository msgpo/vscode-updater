package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

func compareVersions(a, b string) (int, error) {
	cmd := exec.Command("vercmp", a, b)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	o, err := strconv.Atoi(strings.TrimSpace(out.String()))
	if err != nil {
		return 0, err
	}
	return o, nil
}
