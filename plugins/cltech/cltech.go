package cltech

import (
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

const feed = "http://feeds.feedburner.com/cl-jstlq"

func cltechHandler() ([]*kindlepush.Post, error) {
	doc, err := xmlquery.LoadURL(feed)
	if err != nil {
		return nil, err
	}
	var posts []*kindlepush.Post
	elems := xmlquery.Find(doc, "//item")
	for _, elem := range elems {
		post := &kindlepush.Post{
			Title:       elem.SelectElement("title").InnerText(),
			Link:        elem.SelectElement("link").InnerText(),
			Date:        elem.SelectElement("pubDate").InnerText(),
			Description: elem.SelectElement("description").InnerText(),
			Author:      elem.SelectElement("author").InnerText(),
		}
		post.Body = post.Description
		posts = append(posts, post)
	}
	return posts, nil
}

func init() {
	kindlepush.RegisterPlugin("cltech", "1024 Tech Forum", kindlepush.PluginFunc(cltechHandler))
}
