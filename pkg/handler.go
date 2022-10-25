package pkg

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/authentication/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/authentication"
)

type respBody struct {
	Sid string `json:"session_id"`
}

type AuthProxy struct {
	becli *Client
}

func NewProxyHandler(becli *Client) *AuthProxy {
	return &AuthProxy{
		becli: becli,
	}
}

// handler /login
func (a *AuthProxy) LoginHandler(ctx *gin.Context) {
	user, password, ok := ctx.Request.BasicAuth()
	if !ok {
		ctx.AbortWithError(401, fmt.Errorf("user or password error"))
		return
	}
	if user == JWTTOKEN {
		klog.Info("login with jwt-token")
		_, err := a.becli.VerifyToken(password)
		if err != nil {
			ctx.AbortWithError(401, err)
		} else {
			ctx.JSON(200, &respBody{Sid: password})
		}
	} else {
		klog.Infof("login with user %s", user)
		token, err := a.becli.VerifyLogin(user, password)
		if err != nil {
			ctx.AbortWithError(401, err)
		} else {
			ctx.JSON(200, &respBody{Sid: token})
		}
	}
}

// handler /tokenreview
func (a *AuthProxy) TokenHandler() http.Handler {
	return &authentication.Webhook{
		Handler: newAuthenticator(a.becli),
	}
}

// authenticator validates tokenreviews
type authenticator struct {
	becli *Client
}

func newAuthenticator(becli *Client) *authenticator {
	return &authenticator{
		becli: becli,
	}
}

// authenticator admits a request by the token.
func (a *authenticator) Handle(ctx context.Context, req authentication.Request) authentication.Response {
	if req.Spec.Token == "invalid" {
		return authentication.Unauthenticated("invalid is an invalid token", v1.UserInfo{})
	}
	user, err := a.becli.VerifyToken(req.Spec.Token)
	if err != nil {
		return authentication.Unauthenticated("verify token failed", v1.UserInfo{})
	}
	return authentication.Authenticated("", v1.UserInfo{
		Username: user.Username,
		UID:      user.Userid,
		Groups:   []string{"1"},
	})
}
