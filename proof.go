package utreexo

import "errors"

type (
	// ProofStep is one step in a Proof.
	ProofStep struct {
		H    Hash
		Left bool
	}

	// Proof is a proof of inclusion for a given value.
	Proof struct {
		// Leaf is the value whose inclusion is being proven.
		Leaf Hash

		// Steps is a sequence of ProofSteps, from the leaf towards the root.
		Steps []ProofStep
	}
)

// ErrInvalid means the proof is invalid.
var ErrInvalid = errors.New("invalid proof")

// Update updates the proof of inclusion for a value after the Utreexo has been updated.
// In case of error
// (i.e., the proof is invalid with respect to the Utreexo in u),
// the proof may be incompletely updated and should be discarded.
func (p *Proof) Update(u Update) error {
	h := p.Leaf

	for i := 0; i <= len(p.Steps); i++ {
		if len(u.u.roots) > i && u.u.roots[i] != nil && *u.u.roots[i] == h {
			p.Steps = p.Steps[:i]
			return nil
		}
		var step ProofStep
		if s, ok := u.updated[h]; ok {
			step = s
			p.Steps = append(p.Steps[:i], step)
		} else if i == len(p.Steps) {
			break
		} else {
			step = p.Steps[i]
		}
		h = u.u.parent(h, step)
	}

	return ErrInvalid
}
