package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/yylt/authproxy/pkg"
	"k8s.io/klog/v2"
)

func main() {

	address := flag.String("a", ":8089", "listen address")
	uep := flag.String("t", "", "token endpoint, like http://128.0.0.1:8443/api/token")
	lep := flag.String("l", "", "login endpoint, like http://128.0.0.1:8443/api/login")
	debug := flag.Bool("d", true, "debug or not, default true")
	flag.Parse()
	if err := validep(uep); err != nil {
		panic(fmt.Sprintf("token endpoint parse failed: %v", err))
	}
	if err := validep(lep); err != nil {
		panic(fmt.Sprintf("login endpoint parse failed: %v", err))
	}
	printVersion()
	klog.Infof("token endpoint: %s", *uep)
	klog.Infof("login endpoint: %s", *lep)

	engine := gin.Default()

	hand := pkg.NewProxyHandler(pkg.NewClient(*uep, *lep, *debug))

	engine.POST("/login", hand.LoginHandler)
	engine.POST("/tokenreview", gin.WrapH(hand.TokenHandler()))
	engine.Run(*address)
}

func validep(u *string) error {
	if u == nil {
		panic("the endpoint is nil")
	}
	oriurl, err := url.Parse(*u)
	if err != nil {
		return err
	}
	if oriurl.Host == "" {
		return fmt.Errorf("not found host in url %s", *u)
	}
	return nil
}
