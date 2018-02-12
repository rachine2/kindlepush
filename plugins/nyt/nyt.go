package nyt

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/antchfx/xquery/html"
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

const feed_cn = "https://cn.nytimes.com/rss.html"
const feed_wld = "http://www.nytimes.com/services/xml/rss/nyt/World.xml"
const feed_us = "http://www.nytimes.com/services/xml/rss/nyt/US.xml"
const feed_bs = "http://rss.nytimes.com/services/xml/rss/nyt/Business.xml"

func nytDailyHandler(feed string) func() ([]*kindlepush.Post, error) {
	return func() ([]*kindlepush.Post, error) {
		doc, err := xmlquery.LoadURL(feed)
		if err != nil {
			return nil, err
		}
		var posts []*kindlepush.Post
		var wg sync.WaitGroup

		elems := xmlquery.Find(doc, "//channel/item")
		for _, elem := range elems {
			post := &kindlepush.Post{
				Title:       elem.SelectElement("title").InnerText(),
				Link:        elem.SelectElement("link").InnerText(),
				Date:        elem.SelectElement("pubDate").InnerText(),
				Description: elem.SelectElement("description").InnerText(),
			}

			wg.Add(1)
			go func(post *kindlepush.Post) {
				defer wg.Done()

				post.Body = fetch(post.Link)
				//post.Description = post.Body
			}(post)

			posts = append(posts, post)
		}
		wg.Wait()
		return posts, nil
	}
}

func fetch(link string) string {
	doc, err := htmlquery.LoadURL(link)
	if err != nil {
		logrus.Warnf("nyt fetching body %s got error: %v", link, err)
	} else {
		if elem := htmlquery.FindOne(doc, "//section[@class='article-body']"); elem != nil {
			return htmlquery.InnerText(elem)
		}
	}
	return ""
}

func init() {
	kindlepush.RegisterPlugin("nyt_cn", "new york times cn", kindlepush.PluginFunc(nytDailyHandler(feed_cn)))
	kindlepush.RegisterPlugin("nyt_wld", "new york times world", kindlepush.PluginFunc(nytDailyHandler(feed_wld)))
	kindlepush.RegisterPlugin("nyt_bs", "new york times business", kindlepush.PluginFunc(nytDailyHandler(feed_bs)))
	kindlepush.RegisterPlugin("nyt_us", "new york times us", kindlepush.PluginFunc(nytDailyHandler(feed_us)))
}
