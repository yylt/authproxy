package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/imroc/req/v3"
)

var csrfCook = http.Cookie{
	Name:     "_gorilla_csrf",
	Value:    "MTY2NjkxNzcwMnxJa0pwZHlzM2NsTktaRmM0YlVoRVRFOXdVR1ZSTm5oa0sxUXZTMEpNZG5kSGIxVkxZamhxTlc5RFRFVTlJZ289fJl42gPPeCFU1CWpY8fHIEngZeniXaRX1vzmkfU2mHJn",
	Path:     "/",
	Expires:  time.Now().Add(time.Hour),
	MaxAge:   43200,
	SameSite: http.SameSiteStrictMode,
	HttpOnly: true,
}

var csrfHead = map[string]string{
	"X-Harbor-CSRF-Token": "TTTEWKh2wYtI4O5Q+7azlj7YJZwzqGgSJTEy8EBomuJLGPq2HP+05G783J5fQSN9KaZqbrKGlBSEc6kCfgCSUw==",
}

func main() {
	address := flag.String("a", "", "address about harbor, must startwith http: or https:")
	user := flag.String("u", "admin", "user about harbor,default admin")
	pass := flag.String("pw", "", "user password")
	prj := flag.String("prj", "", "which project want to create, and it is public")
	flag.Parse()

	oriurl, err := url.Parse(*address)
	if err != nil {
		panic("novalid address")
	}
	var (
		scheme = "http://"
		host   = oriurl.Host
	)

	if oriurl.Scheme != "http" {
		scheme = "https://"
	}

	if *pass == "" || *prj == "" {
		panic("password is none or prj is none")
	}

	var (
		loginurl     = scheme + path.Join(host, "/c/login")
		createprjurl = scheme + path.Join(host, "api/v2.0/projects")
		payload      = map[string]string{
			"principal": *user,
			"password":  *pass,
		}
	)

	resp := req.C().Post(loginurl).
		SetFormData(payload).
		SetCookies(&csrfCook).
		SetHeaders(csrfHead).
		Do()
	if resp.Err != nil {
		panic(resp.Err)
	}
	if resp.Response.StatusCode != 200 {
		panic(fmt.Errorf("login failed, return code: %v, body:%s", resp.Response.StatusCode, string(resp.Bytes())))
	}

	respcs := append(resp.Response.Cookies(), &csrfCook)
	resp = req.C().Post(createprjurl).
		SetCookies(respcs...).
		SetHeaders(csrfHead).
		SetBodyJsonString(fmt.Sprintf(`{"project_name":"%s","metadata":{"public":"true"}"`, *prj)).
		Do()

	if resp.Err != nil {
		panic(resp.Err)
	}
	if resp.Response.StatusCode != 200 {
		panic(fmt.Errorf("login failed, return code: %v", resp.Response.StatusCode))
	}
	log.Printf("create public project %s success.", *prj)
	return
}

func validep(u *string) bool {
	if u == nil {
		panic("the endpoint is nil")
	}
	oriurl, err := url.Parse(*u)
	if err != nil {
		return false
	}
	if oriurl.Host == "" || oriurl.Scheme == "" {
		return false
	}
	return true
}
