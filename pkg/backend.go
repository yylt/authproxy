package pkg

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/imroc/req/v3"
	"k8s.io/klog/v2"
)

const JWTTOKEN = "jwt_token"

type loginReqBody struct {
	Codeid int `json:"codeId"`
	// always is -1
	Type int    `json:"type"`
	User string `json:"username"`
	Pass string `json:"password"`
}

type user struct {
	Userid   int    `json:"userId"`
	Username string `json:"realName"`
}

type tokenRespBody struct {
	Data user `json:"data"`
}

type Client struct {
	cli     *req.Client
	tokenep string
	loginep string

	adminHash string
}

var tokenReg = regexp.MustCompile((`^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}$`))

func NewClient(tokenep, loginep string, debug bool) *Client {
	cli := req.C()
	if debug {
		cli.EnableDumpAll()
	}

	return &Client{
		cli:     cli,
		tokenep: tokenep,
		loginep: loginep,
	}
}

func (c *Client) VerifyToken(token string) (*user, error) {
	if token == adminHash() {
		return &user{
			Userid:   1,
			Username: "admin",
		}, nil
	}

	var tokencookie = &http.Cookie{
		Name:  JWTTOKEN,
		Value: token,
	}
	respd := tokenRespBody{}
	resp := c.cli.Get(c.tokenep).
		SetCookies(tokencookie).
		SetResult(&respd).
		Do()
	if resp.Err != nil {
		klog.Errorf("verify token failed: %v", resp.Err)
		return nil, resp.Err
	}
	if resp.Response.StatusCode/100 != 2 {
		klog.Errorf("verify token failed, resp code:%d", resp.Response.StatusCode)
		return nil, fmt.Errorf("resp code is not 2xx")
	}

	klog.Infof("verify token success, resp userid: %v", respd.Data.Userid)

	return &respd.Data, nil
}

// return token ,error
func (c *Client) VerifyLogin(user, password string) (string, error) {
	if user == "admin" && password == adminPassword() {
		klog.Infof("password and user match harbor admin")
		return adminHash(), nil
	}
	if !tokenReg.Match([]byte(password)) {
		klog.Infof("password not match token regexp")
		//TODO, verify in loginurl
	}
	return password, nil
}

func adminPassword() string {
	return os.Getenv("HARBOR_ADMIN_PASSWORD")
}

func adminHash() string {
	passwd := adminPassword()
	if passwd == "" {
		return ""
	}
	md := md5.New()
	md.Write([]byte("admin"))
	md.Write([]byte(passwd))
	return base64.StdEncoding.EncodeToString(md.Sum(nil))
}
