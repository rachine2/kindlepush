package engadgetcn

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/antchfx/xquery/html"
	"github.com/antchfx/xquery/xml"

	"github.com/zhengchun/kindlepush"
)

const feed_en = "http://www.engadget.com/rss.xml"
const feed_cn = "http://cn.engadget.com/rss.xml"

func engadgetHandler(feed string) func() ([]*kindlepush.Post, error) {
	return func() ([]*kindlepush.Post, error) {
		root, err := xmlquery.LoadURL(feed)
		if err != nil {
			return nil, err
		}
		var posts []*kindlepush.Post
		var wg sync.WaitGroup

		elems := xmlquery.Find(root, "//channel/item")
		for _, elem := range elems {
			post := &kindlepush.Post{
				Title:       elem.SelectElement("title").InnerText(),
				Link:        elem.SelectElement("link").InnerText(),
				Description: elem.SelectElement("description").InnerText(),
				Author:      elem.SelectElement("dc:creator").InnerText(),
				Date:        elem.SelectElement("pubDate").InnerText(),
			}
			wg.Add(1)
			// fetching article's body via HTTP.
			go func(post *kindlepush.Post) {
				defer wg.Done()
				post.Body = fetchBody(post.Link)
			}(post)

			posts = append(posts, post)
		}
		wg.Wait()
		return posts, nil
	}
}

func fetchBody(link string) string {
	doc, err := htmlquery.LoadURL(link)
	if err != nil {
		logrus.Warnf("engadget_cn fetching body %s got error: %v", link, err)
	} else {
		if elem := htmlquery.FindOne(doc, "//div[@class='flush-top flush-bottom']"); elem != nil {
			return htmlquery.OutputHTML(elem)
		}
	}
	return ""
}

func init() {
	kindlepush.RegisterPlugin("engadget_cn", "Engadget 中文版", kindlepush.PluginFunc(engadgetHandler(feed_cn)))
	kindlepush.RegisterPlugin("engadget_en", "Engadget", kindlepush.PluginFunc(engadgetHandler(feed_en)))
}
