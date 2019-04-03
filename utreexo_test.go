package utreexo

import (
	"crypto/sha256"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestUtreexo(t *testing.T) {
	hasher := sha256.New()

	var items [12]Hash
	for i := 1; i < len(items); i++ {
		hasher.Reset()
		hasher.Write(items[i-1][:])
		copy(items[i][:], hasher.Sum(nil))
	}

	u := New(func(h1, h2 Hash) Hash {
		hasher.Reset()
		hasher.Write([]byte{0})
		hasher.Write(h1[:])
		hasher.Write([]byte{1})
		hasher.Write(h2[:])
		var result Hash
		copy(result[:], hasher.Sum(nil))
		return result
	})

	upd, err := u.Update(nil, items[:11])
	if err != nil {
		t.Fatal(err)
	}

	var proofs [11]Proof
	for i := 0; i < len(proofs); i++ {
		proofs[i] = upd.Proof(items[i])
	}

	t.Logf("tree:\n%s", spew.Sdump(u))
	t.Logf("proofs:\n%s", spew.Sdump(proofs))

	upd, err = u.Update([]Proof{proofs[10]}, nil)
	if err != nil {
		t.Fatalf("updating tree: %s", err)
	}

	t.Logf("update:\n%s", spew.Sdump(upd))

	for i := 0; i < 10; i++ {
		err = proofs[i].Update(upd)
		if err != nil {
			t.Fatalf("updating proof %d (for %s): %s", i, proofs[i].Leaf, err)
		}
	}
	t.Logf("after deletion of %s, tree:\n%s", proofs[10].Leaf, spew.Sdump(u))
	t.Logf("proofs:\n%s", spew.Sdump(proofs[:10]))

	_, err = u.Update([]Proof{proofs[10]}, nil)
	if err != ErrInvalid {
		t.Errorf("got error %v, want %s", err, ErrInvalid)
	}
}
