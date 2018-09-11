package ygg

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/juliensalinas/torrengo/core"
	"golang.org/x/net/publicsuffix"
)

var loginURL = url.URL{
	Scheme: "https",
	Host:   baseURL,
	Path:   "user/login",
}

func authUser(userID string, userPass string) (*http.Client, error) {
	formData := url.Values{
		"id":   {userID},
		"pass": {userPass},
	}

	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &http.Client{
		Jar: cookieJar,
	}

	req, err := http.NewRequest("POST", loginURL.String(), strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could not build POST request to login url: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	req.Header.Set("User-Agent", core.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request to login url failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status code %d %s", resp.StatusCode, resp.Status)
	}

	return client, nil
}
