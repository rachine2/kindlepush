package t36kr

import "testing"

func Test36KrHandler(t *testing.T) {
	posts, err := t36krHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
