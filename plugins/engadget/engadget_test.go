package engadgetcn

import (
	"testing"
)

func TestEngadgetCNHandler(t *testing.T) {
	handler := engadgetHandler(feed_cn)
	posts, err := handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}

func TestEngadgetENHandler(t *testing.T) {
	handler := engadgetHandler(feed_en)
	posts, err := handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
