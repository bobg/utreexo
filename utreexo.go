package utreexo

import "encoding/hex"

// Hash is the type of a value (and of internal nodes) in a Utreexo.
type Hash [32]byte

// This is for testing.
var hashnames = map[string]string{
	"319efef47197950dc90dbcf48b897f7cb8553030da7d18416f0eb163da0e84a2": "A",
	"6dfae426e23b27ae6c17c6b6fb6695306cf8895efc834527a75301f73276706a": "B",
	"58b9355a5a8b84a33bf1db57a7b44f8b1effc91b194a9d1dc530a0a68ae768c6": "C",
	"259dc6e788e97916aceb6ba1c3f8aa689f26e149289bbe64ce6e98f21c4016ae": "D",
	"1788814e4c75eb4b8a4c6b1a36f851633feb1b588f33e82244ca9655333cf929": "E",
	"149974a85fc19621ac87d4ba2511a8baf06675ee930f05b8f1f37e37acf76e65": "F",
	"d2d25124018d42d584b14780573523bd373c35db8ab7f7b0a1f54bd62edb1369": "G",
	"66687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f2925": "H",
	"2b32db6c2c0a6235fb1397e8225ea85e0f0e6e8c7b126d0016ccbde0e667151e": "I",
	"12771355e46cd47c71ed1721fd5319b383cca3a1f9fce3aa1c8cd3bd37af20d7": "J",
	"fe15c0d3ebe314fad720a08b839a004c2e6386f5aecc19ec74807d1920cb6aeb": "K",
	"376da11fe3ab3d0eaaddb418ccb49b5426d5c2504f526f7766580f6e45984e3b": "L",
	"4391a5c79ffdc79883036503ca551673c09deec28df432a8d88debc7fa2ec91e": "M",
	"5d1adcb5797c2eff1ba0460af9324ac6df5b6ffb66be6df2547872c2f29ba4c2": "N",
	"6a9b711ce5d3749ece29463110b6164dbb28dda28902586bf66e865e8c29c350": "O",
	"4e6e6acef5953a6a2087d8dd7d38a49b3ca0627d8ab339872ce56c5bd3b5a112": "P",
	"f13587bc89fe4882c7c889302511ffd738d136129b9f5be4c492cb4948a93a89": "Q",
	"e1ebe496bcd0c774e8c30b51aa624becf2c09cc742cb8149945ab92e771684a0": "R",
	"0000000000000000000000000000000000000000000000000000000000000000": "S",
}

func (h Hash) String() string {
	s := hex.EncodeToString(h[:])
	if s, ok := hashnames[s]; ok {
		return s
	}
	return s
}

// HashFunc is the type of a function that produces a parent hash from two child hashes.
type HashFunc func(Hash, Hash) Hash

// Utreexo is a forest of perfectly full Merkle trees,
// at most one of size 2^N for each N in 0..K
// (where the total number of leaves is bounded by 2^K).
type Utreexo struct {
	roots  []*Hash
	hasher HashFunc
}

// New produces an empty Utreexo with the given hash function.
func New(hasher HashFunc) *Utreexo {
	return &Utreexo{
		hasher: hasher,
	}
}

// Update is the output of Utreexo.Update.
// It contains information that can be used to update proofs
// (via Proof.Update)
// after the Utreexo changes.
type Update struct {
	u       *Utreexo
	updated map[Hash]ProofStep
}

// Update removes some values from a Utreexo and adds others.
// All deletions happen before any insertions.
// Each deletion is specified by a proof of inclusion.
// If any proof is invalid, no changes at all are made.
// All insertions are presumed to be unique.
// (Results are undefined if they're not.)
// The Update object that results should be used to update proofs
// (via Proof.Update)
// that may have been affected by the changes to the Utreexo.
// Note that this function never returns an error when len(deletions)==0.
func (u *Utreexo) Update(deletions []Proof, insertions []Hash) (Update, error) {
	newRoots := make([][]Hash, len(u.roots))
	for i, root := range u.roots {
		if root != nil {
			newRoots[i] = []Hash{*root}
		}
	}

	update := Update{
		u:       u,
		updated: make(map[Hash]ProofStep),
	}

	for _, d := range deletions {
		err := u.delHelper(d, &newRoots)
		if err != nil {
			return update, err
		}
	}

	if len(newRoots) == 0 {
		newRoots = [][]Hash{nil}
	}
	newRoots[0] = append(newRoots[0], insertions...)

	for i := 0; i < len(newRoots); i++ {
		for len(newRoots[i]) > 1 {
			a, b := newRoots[i][len(newRoots[i])-2], newRoots[i][len(newRoots[i])-1]
			newRoots[i] = newRoots[i][:len(newRoots[i])-2]
			h := u.hasher(a, b)
			if len(newRoots) <= i+1 {
				newRoots = append(newRoots, nil)
			}
			newRoots[i+1] = append(newRoots[i+1], h)
			update.updated[a] = ProofStep{H: b, Left: false}
			update.updated[b] = ProofStep{H: a, Left: true}
		}
	}

	for i := len(newRoots) - 1; i >= 0; i-- {
		if len(newRoots[i]) > 0 {
			break
		}
		newRoots = newRoots[:len(newRoots)-1]
	}

	for i, h := range newRoots {
		if len(u.roots) <= i {
			u.roots = append(u.roots, nil)
		}
		if len(h) == 0 {
			u.roots[i] = nil
		} else {
			u.roots[i] = &h[0]
		}
	}
	u.roots = u.roots[:len(newRoots)]

	return update, nil
}

func (u *Utreexo) delHelper(p Proof, newRoots *[][]Hash) error {
	if len(u.roots) <= len(p.Steps) || u.roots[len(p.Steps)] == nil {
		return ErrInvalid
	}

	height := 0
	hash := p.Leaf
	for {
		if height < len(*newRoots) {
			if index, ok := findRoot(hash, (*newRoots)[height]); ok {
				// Remove hash from newRoots.
				(*newRoots)[height] = append((*newRoots)[height][:index], (*newRoots)[height][index+1:]...)

				// If height < len(p.Steps),
				// then an earlier deletion "opened up" u.roots[len(p.Steps)]
				// and we just removed that subroot from newRoots.
				// Now verify the remainder of p.Steps against the unmodified tree.
				for {
					if height >= len(p.Steps) {
						if hash != *u.roots[height] {
							return ErrInvalid
						}
						return nil
					}
					s := p.Steps[height]
					hash = u.parent(hash, s)
					height++
				}
			}
		}
		if height >= len(p.Steps) {
			return ErrInvalid
		}
		for height >= len(*newRoots) {
			*newRoots = append(*newRoots, nil)
		}
		s := p.Steps[height]
		(*newRoots)[height] = append((*newRoots)[height], s.H)
		hash = u.parent(hash, s)
		height++
	}
}

func (u *Utreexo) parent(h Hash, s ProofStep) Hash {
	if s.Left {
		return u.hasher(s.H, h)
	}
	return u.hasher(h, s.H)
}

func findRoot(root Hash, roots []Hash) (int, bool) {
	for i, r := range roots {
		if root == r {
			return i, true
		}
	}
	return 0, false
}

// Proof produces the Proof for a newly added item after a call to Utreexo.Update.
// If item is not one of the items added in the call that produced this Update,
// the resulting Proof will probably be invalid,
// but there's a small chance it won't be.
func (u Update) Proof(item Hash) Proof {
	p := Proof{Leaf: item}
	for {
		s, ok := u.updated[item]
		if !ok {
			return p
		}
		p.Steps = append(p.Steps, s)
		item = u.u.parent(item, s)
	}
}
