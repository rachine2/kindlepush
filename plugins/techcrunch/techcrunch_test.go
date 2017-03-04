package techcrunch

import "testing"

func TestTechcrunchCNHandler(t *testing.T) {
	handler := techcrunchHandler(feed_cn, false)
	posts, err := handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}

func TestTechcrunchENHandler(t *testing.T) {
	handler := techcrunchHandler(feed_en, false)
	posts, err := handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
