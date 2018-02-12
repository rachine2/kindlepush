package t36kr

import (
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

const feed = "http://www.36kr.com/feed"

func t36krHandler() ([]*kindlepush.Post, error) {
	doc, err := xmlquery.LoadURL(feed)
	if err != nil {
		return nil, err
	}
	var posts []*kindlepush.Post
	elems := xmlquery.Find(doc, "//channel/item")
	for _, elem := range elems {
		post := &kindlepush.Post{
			Title:       elem.SelectElement("title").InnerText(),
			Link:        elem.SelectElement("link").InnerText(),
			Date:        elem.SelectElement("pubDate").InnerText(),
			Description: elem.SelectElement("description").InnerText(),
		}
		post.Body = post.Description
		if elem := elem.SelectElement("dc:creator"); elem != nil {
			post.Author = elem.InnerText()
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func init() {
	kindlepush.RegisterPlugin("36kr", "36氪", kindlepush.PluginFunc(t36krHandler))
}
