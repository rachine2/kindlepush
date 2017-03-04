package kindlepush

import (
	"regexp"
	"strings"
	"time"
)

// The subscribed channel and newest posts.
type channel struct {
	name  string
	posts []*Post
}

// About An article information and body.
type Post struct {
	Title       string
	Body        string
	Date        string
	Author      string
	Link        string
	Description string
}

func (p *Post) FormatDate() string {
	if len(p.Date) > 0 {
		if t, err := time.Parse(time.RFC1123Z, p.Date); err == nil {
			return t.Local().Format("02/01/2006")
		}
	}
	return p.Date
}

func (p *Post) FormatDescription(n int) string {
	s := removeHtmlTags(strings.TrimSpace(p.Description))
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

var (
	htmlTagRegexp = regexp.MustCompile(`</?[^>]+>`)
	imgRegexp     = regexp.MustCompile(`<img[^>]*src="([^"]*)"[^>]*/>`)
)

func removeHtmlTags(s string) string {
	return htmlTagRegexp.ReplaceAllString(s, "")
}
