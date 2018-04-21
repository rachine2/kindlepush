package cltech

import "testing"

func TestCltechHandler(t *testing.T) {
	posts, err := cltechHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
