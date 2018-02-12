package ftc

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/antchfx/xquery/html"
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

const (
	feed = "http://www.ftchinese.com/rss/feed"
)

func ftcHandler(feed string, feedburner bool) func() ([]*kindlepush.Post, error) {
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
				Description: elem.SelectElement("description").InnerText(),
			}
			// get full article from link
			if feedburner {
				wg.Add(1)

				post.Link = elem.SelectElement("link").InnerText()
				go func(post *kindlepush.Post) {
					defer wg.Done()

					post.Body = fetch(post.Link)
				}(post)
			} else {
				post.Body = elem.SelectElement("description").InnerText()
			}
			posts = append(posts, post)
		}
		wg.Wait()
		return posts, nil
	}
}

func fetch(link string) string {
	link += "?full=y"
	doc, err := htmlquery.LoadURL(link)
	if err != nil {
		logrus.Warnf("ftc fetching body %s got error: %v", link, err)
	} else {
		if elem := htmlquery.FindOne(doc, "//div[@class='story-body']"); elem != nil {
			// FIXME: mask for output html format error, lost </p>, make kindlegen fail
			// W29004: Forcefully closed opened Tag: <p>
			//return htmlquery.OutputHTML(elem, true)

			// ugly fix, output garbage characters, but it work
			return htmlquery.InnerText(elem)
		}
	}
	return ""
}

func init() {
	kindlepush.RegisterPlugin("ftc", "FT 中文版", kindlepush.PluginFunc(ftcHandler(feed, true)))
}
