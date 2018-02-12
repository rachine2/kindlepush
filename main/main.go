package main

import (
	"errors"
	"flag"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/zhengchun/kindlepush"
	_ "github.com/zhengchun/kindlepush/plugins/blogs/dotnet"
	_ "github.com/zhengchun/kindlepush/plugins/cnbeta"
	_ "github.com/zhengchun/kindlepush/plugins/engadget"
	_ "github.com/zhengchun/kindlepush/plugins/eth"
	_ "github.com/zhengchun/kindlepush/plugins/ftc"
	_ "github.com/zhengchun/kindlepush/plugins/nyt"
	_ "github.com/zhengchun/kindlepush/plugins/rfa"
	_ "github.com/zhengchun/kindlepush/plugins/t36kr"
	_ "github.com/zhengchun/kindlepush/plugins/techcrunch"
	_ "github.com/zhengchun/kindlepush/plugins/zhihu"
)

func config() kindlepush.Config {

	var (
		maxnum     int
		kindleAddr string
		subscribes string
		debug      bool

		from, username, password, smtp string
	)
	flag.IntVar(&maxnum, "max-number", 25, "the maximum number of items to fetch.")
	flag.StringVar(&kindleAddr, "kindle", "", "the kindle email address.")
	flag.StringVar(&subscribes, "subscribes", "", "the subscribed channel list.")
	flag.StringVar(&from, "email-from", "", "your email address to send.")
	flag.StringVar(&username, "email-username", "", "your email account name.")
	flag.StringVar(&password, "email-password", "", "your email account password.")
	flag.StringVar(&smtp, "email-smtp", "", "your email SMTP address.")
	flag.BoolVar(&debug, "debug", false, "")

	flag.Parse()

	config := kindlepush.Config{
		Debug:      debug,
		MaxNumber:  maxnum,
		KindleAddr: kindleAddr,
		Subscribes: strings.Split(subscribes, ","),
		Email: kindlepush.EmailConfig{
			From:     from,
			Username: username,
			Password: password,
			SMTP:     smtp,
		},
	}
	return config
}

func checkConfig(config kindlepush.Config) error {
	if len(config.Subscribes) == 0 {
		return errors.New("--subscribes is nil")
	}
	if config.KindleAddr == "" {
		return errors.New("--kindle is nil")
	}
	if config.Email.From == "" {
		return errors.New("--email-from is nil")
	}
	if config.Email.SMTP == "" {
		return errors.New("--email-smtp is nil")
	}
	return nil
}

func main() {
	config := config()
	if err := checkConfig(config); err != nil {
		logrus.Errorf("config error: %v", err)
		return
	}
	server := kindlepush.NewServer(config)
	server.Run()
}
