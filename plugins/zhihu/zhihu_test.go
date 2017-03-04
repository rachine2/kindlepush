package zhihu

import "testing"

func TestZhihuHandler(t *testing.T) {
	posts, err := zhihuDailyHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
