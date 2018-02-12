package rfa

import "testing"

func TestRfaHandler(t *testing.T) {
	posts, err := rfaHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
