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
func (p *Proof) Update(u Update) (err error) {
	h := p.Leaf
	steps := p.Steps

	defer func() {
		if err == nil {
			p.Steps = steps
		}
	}()

	for i := 0; i <= len(steps); i++ {
		if len(u.u.roots) > i && u.u.roots[i] != nil && *u.u.roots[i] == h {
			steps = steps[:i]
			return nil
		}
		var step ProofStep
		if s, ok := u.updated[h]; ok {
			step = s
			steps = append(steps[:i], step)
		} else if i == len(steps) {
			break
		} else {
			step = steps[i]
		}
		if step.Left {
			h = u.u.hasher(step.H, h)
		} else {
			h = u.u.hasher(h, step.H)
		}
	}

	return ErrInvalid
}
