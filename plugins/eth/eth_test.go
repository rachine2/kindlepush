package eth

import "testing"

func TestEthHandler(t *testing.T) {
	posts, err := ethHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		t.Log(post)
	}
}
