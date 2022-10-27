package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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

	klog.Infof("login with user %s, password %s", user, password)
	token, err := a.becli.VerifyLogin(user, password)
	if err != nil {
		ctx.AbortWithError(401, err)
	} else {
		ctx.JSON(200, &respBody{Sid: token})
	}
}

func (a *AuthProxy) TokenReviewHandler(ctx *gin.Context) {
	var (
		req    = v1.TokenReview{}
		err    error
		status = v1.TokenReviewStatus{}
		resp   = v1.TokenReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "TokenReview",
				APIVersion: "authentication.k8s.io/v1beta1",
			},
		}
	)

	err = ctx.Bind(&req)
	if err != nil {
		klog.Errorf("body is not tokenreview type, failed: %v", err)
		status.Error = err.Error()
	}
	klog.Infof("check token: %s", req.Spec.Token)
	user, err := a.becli.VerifyToken(req.Spec.Token)
	if err != nil {
		status.Error = err.Error()
	} else {
		status.Authenticated = true
		uid := fmt.Sprintf("%d", user.Userid)
		status.User = v1.UserInfo{
			Username: user.Username,
			UID:      uid,
			Groups:   []string{user.Username, uid},
		}
	}
	resp.Status = status
	klog.Infof("resp status: %v", status)
	ctx.JSON(200, resp)
}
