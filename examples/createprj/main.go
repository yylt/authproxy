package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"path"

	"github.com/imroc/req/v3"
)

var client = req.C()

type meta struct {
	Public string `json:"public"`
}
type project struct {
	Prj   string `json:"project_name"`
	Meta  *meta  `json:"metadata"`
	Limit int    `json:"storage_limit"`
}

const harborCsrfHead = "X-Harbor-Csrf-Token"

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

	client.EnableDumpAll()

	resp := login(loginurl, payload)

	csrft := resp.GetHeader(harborCsrfHead)
	jsonb := &project{
		Prj: *prj,
		Meta: &meta{
			Public: "true",
		},
		Limit: -1,
	}

	resp = client.Post(createprjurl).
		SetHeader(harborCsrfHead, csrft).
		SetBodyJsonMarshal(jsonb).
		Do()

	if resp.Err != nil {
		panic(resp.Err)
	}
	if resp.Response.StatusCode/100 != 2 {
		panic(fmt.Errorf("create project failed, return code: %v", resp.Response.StatusCode))
	}
	log.Printf("create public project %s success.", *prj)
	return
}

func login(loginurl string, data map[string]string) *req.Response {
	resp := client.Post(loginurl).
		SetFormData(data).
		Do()
	if resp.Err != nil {
		panic(resp.Err)
	}
	if resp.Response.StatusCode/100 == 4 {
		csrft := resp.GetHeader(harborCsrfHead)
		resp = client.Post(loginurl).
			SetFormData(data).
			SetHeader(harborCsrfHead, csrft).
			Do()
	}
	if resp.Response.StatusCode != 200 {
		panic(fmt.Errorf("login failed, return code: %v, body:%s", resp.Response.StatusCode, string(resp.Bytes())))
	}

	return resp
}
