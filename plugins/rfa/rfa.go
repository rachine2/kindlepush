package rfa

import (
	"github.com/Sirupsen/logrus"
	"github.com/antchfx/xquery/html"
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

const feed = "https://www.rfa.org/mandarin/rss"

func rfaHandler() ([]*kindlepush.Post, error) {
	doc, err := xmlquery.LoadURL(feed)
	if err != nil {
		return nil, err
	}
	var posts []*kindlepush.Post

	titles := xmlquery.Find(doc, "//title")
	links := xmlquery.Find(doc, "//link")
	dates := xmlquery.Find(doc, "//dc:date")

	for i, _ := range titles {
		if i == 0 {
			continue
		}
		if titles[i] == nil || links[i] == nil || dates[i-1] == nil {
			break
		}
		post := &kindlepush.Post{
			Title: titles[i].InnerText(),
			Link:  links[i].InnerText(),
			Date:  dates[i-1].InnerText(),
		}
		post.Body = fetch(post.Link)
		posts = append(posts, post)
	}

	return posts, nil
}

func fetch(link string) string {
	doc, err := htmlquery.LoadURL(link)
	if err != nil {
		logrus.Warnf("rfa fetching body %s got error: %v", link, err)
	} else {
		if elem := htmlquery.FindOne(doc, "//div[@id='storytext']"); elem != nil {
			return htmlquery.OutputHTML(elem, true)
			//return htmlquery.InnerText(elem)
		}
	}
	return ""
}

func init() {
	kindlepush.RegisterPlugin("rfa", "Radio Free Asia Mandarin", kindlepush.PluginFunc(rfaHandler))
}
