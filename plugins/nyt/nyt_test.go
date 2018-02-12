package nyt

import "testing"

func TestNytHandler(t *testing.T) {
	handler := nytDailyHandler(feed_cn)
	posts, err := handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
