package techcrunch

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/antchfx/xquery/html"
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

const (
	feed_cn = "http://techcrunch.cn/feed/"
	feed_en = "http://feeds.feedburner.com/TechCrunch/"
)

func techcrunchHandler(feed string, feedburner bool) func() ([]*kindlepush.Post, error) {
	return func() ([]*kindlepush.Post, error) {
		doc, err := xmlquery.LoadURL(feed)
		if err != nil {
			return nil, err
		}
		elems := xmlquery.Find(doc, "//channel/item")
		var posts []*kindlepush.Post
		var wg sync.WaitGroup

		for _, elem := range elems {
			post := &kindlepush.Post{
				Title:       elem.SelectElement("title").InnerText(),
				Link:        elem.SelectElement("link").InnerText(),
				Date:        elem.SelectElement("pubDate").InnerText(),
				Author:      elem.SelectElement("dc:creator").InnerText(),
				Description: elem.SelectElement("description").InnerText(),
			}
			// if the RSS feeds provider via feedburner
			if feedburner {
				wg.Add(1)

				post.Link = elem.SelectElement("feedburner:origLink").InnerText()
				go func(post *kindlepush.Post) {
					defer wg.Done()

					post.Body = fetch(post.Link)
				}(post)
			} else {
				post.Body = elem.SelectElement("content:encoded").InnerText()
			}
			posts = append(posts, post)
		}
		wg.Wait()
		return posts, nil
	}
}

func fetch(link string) string {
	doc, err := htmlquery.LoadURL(link)
	if err != nil {
		logrus.Warnf("techcrunch fetching body %s got error: %v", link, err)
	} else {
		if elem := htmlquery.FindOne(doc, "//div[@class='article-entry text']"); elem != nil {
			return htmlquery.OutputHTML(elem, true)
		}
	}
	return ""
}

func init() {
	kindlepush.RegisterPlugin("techcrunch_cn", "Techcrunch 中文版", kindlepush.PluginFunc(techcrunchHandler(feed_cn, false)))
	kindlepush.RegisterPlugin("techcrunch_en", "Techcrunch", kindlepush.PluginFunc(techcrunchHandler(feed_en, true)))
}
