package cnbeta

import "testing"

func TestAaaHandler(t *testing.T) {
	posts, err := cnbetaDailyHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
