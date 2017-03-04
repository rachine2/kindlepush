package dotnet

import (
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

const feed = "https://blogs.msdn.microsoft.com/dotnet/feed/"

func dotnetblogHandler() ([]*kindlepush.Post, error) {
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
			Author:      elem.SelectElement("dc:creator").InnerText(),
			Description: elem.SelectElement("description").InnerText(),
			Body:        elem.SelectElement("content:encoded").InnerText(),
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func init() {
	kindlepush.RegisterPlugin("blog_msdn_dotnet", "Microsoft Dotnet Blog", kindlepush.PluginFunc(dotnetblogHandler))
}
