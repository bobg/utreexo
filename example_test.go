package utreexo_test

import (
	"crypto/sha256"
	"log"

	"github.com/bobg/utreexo"
)

func ExampleProof_Update() {
	hasher := sha256.New()

	hashfunc := func(a, b utreexo.Hash) utreexo.Hash {
		hasher.Reset()
		hasher.Write(a[:])
		hasher.Write(b[:])
		var result utreexo.Hash
		hasher.Sum(result[:0])
		return result
	}

	u := utreexo.New(hashfunc)

	// Some values to add to the utreexo.
	hashes := []utreexo.Hash{{0}, {1}, {2}}

	// No error from Update when only inserting.
	upd, _ := u.Update(nil, hashes)

	// Create proofs of inclusion for the newly added values.
	var proofs []*utreexo.Proof
	for _, hash := range hashes {
		proof := upd.Proof(hash)
		proofs = append(proofs, &proof)
	}

	// Remove the value in proofs[0] from the utreexo.
	upd, err := u.Update([]utreexo.Proof{*proofs[0]}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Update the remaining proofs to reflect the new utreexo structure.
	for _, proof := range proofs[1:] {
		proof.Update(upd)
	}
}
