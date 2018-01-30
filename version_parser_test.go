package main

import "testing"

func TestValidUrl(t *testing.T) {
	testcases := []struct {
		url      string
		expected string
	}{
		{"https://az764295.vo.msecnd.net/insider/08a27fe3b331c7947def0c08092c688240198caf/code-insider-1.20.0-1517206683_amd64.tar.gz", "1.20.0.1517206683"},
		{"https://az764295.vo.msecnd.net/stable/7c4205b5c6e52a53b81c69d2b2dc8a627abaa0ba/code-stable-code_1.19.3-1516876437_amd64.tar.gz", "1.19.3.1516876437"},
	}
	for _, tt := range testcases {
		v, err := parseVersion(tt.url)
		if err != nil {
			t.Errorf("got unexpected error: %v", err)
		}
		if v != tt.expected {
			t.Errorf("wrong version: got %v, expected %v", v, tt.expected)
		}
	}
}

func TestBadUrl(t *testing.T) {
	urls := []string{
		"hai",
		"https://az764295.vo.msecnd.net/insider/08a27fe3b331c7947def0c08092c688240198caf/code-insider-1.20.01517206683_amd64.tar.gz",
		"https://az764295.vo.msecnd.net/insider/08a27fe3b331c7947def0c08092c688240198caf/code-insider-1.20.0-1517206683-i386.tar.gz",
	}
	for _, url := range urls {
		_, err := parseVersion(url)
		if err == nil {
			t.Errorf("should fail with %v", url)
		}
	}
}
