package cnbeta

import (
	"github.com/antchfx/xquery/xml"
	"github.com/zhengchun/kindlepush"
)

// www.cnbeta.com
const feed = "http://rssdiy.com/u/2/cnbeta.xml"

func cnbetaDailyHandler() ([]*kindlepush.Post, error) {
	doc, err := xmlquery.LoadURL(feed)
	if err != nil {
		return nil, err
	}
	var posts []*kindlepush.Post

	titles := xmlquery.Find(doc, "//title")
	summarys := xmlquery.Find(doc, "//summary")

	dates := xmlquery.Find(doc, "//updated")

	for i, _ := range summarys {
		if dates[i] == nil {
			break
		}
		if summarys[i] == nil {
			break
		}
		post := &kindlepush.Post{
			Title: 	titles[i+1].InnerText(),
			Date:	dates[i].InnerText(),
			Description: summarys[i].InnerText(),
		}
		post.Body = post.Description
		posts = append(posts, post)

	}

	return posts, nil
}

func init() {
	kindlepush.RegisterPlugin("cnbeta", "Cnbeta IT news", kindlepush.PluginFunc(cnbetaDailyHandler))
}
