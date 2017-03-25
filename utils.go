package main

import (
	"github.com/dpapathanasiou/go-recaptcha"
	"net"
)

func checkReCaptcha(recaptchaResponse string, remoteAddr string) (result bool) {
	if conf.IgnoreCaptcha {
		return true
	}
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return
	}
	result = recaptcha.Confirm(ip, recaptchaResponse)
	return
}
