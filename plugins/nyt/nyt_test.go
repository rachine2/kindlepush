package nyt

import "testing"

func TestNytHandler(t *testing.T) {
	posts, err := nytDailyHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
