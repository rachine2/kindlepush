package ftc

import "testing"

func TestFTCHandler(t *testing.T) {
	handler := ftcHandler(feed, false)
	posts, err := handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}

