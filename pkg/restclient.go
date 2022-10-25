package pkg

import (
	"net/http"

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
	Userid   string `json:"userId"`
	Username string `json:"userName"`
}

type tokenRespBody struct {
	Data user `json:"data"`
}

type Client struct {
	cli     *req.Client
	tokenep string
	loginep string
}

func NewClient(tokenep, loginep string, debug bool) *Client {
	cli := req.C()
	if debug {
		cli.EnableDumpAll()
	}
	return &Client{
		cli: cli,
	}
}

func (c *Client) VerifyToken(token string) (*user, error) {
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
	klog.InfoS("verify token succ, resp body: %v", respd.Data)

	return &respd.Data, nil
}

// return token ,error
func (c *Client) VerifyLogin(user, password string) (string, error) {
	return "", nil
}
