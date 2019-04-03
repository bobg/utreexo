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
		Leaf  Hash

		// Steps is a sequence of ProofSteps, from the leaf towards the root.
		Steps []ProofStep
	}
)

var (
	// ErrDeleted means a value has been deleted from the Utreexo and its proof cannot be updated.
	ErrDeleted = errors.New("deleted")

	// ErrInvalid means the proof is invalid.
	ErrInvalid = errors.New("invalid proof")
)

// Update updates the proof of inclusion for a value after the Utreexo has been updated.
func (p *Proof) Update(u Update) error {
	if u.Deleted[p.Leaf] {
		return ErrDeleted
	}

	h := p.Leaf
	steps := p.Steps

	defer func() { p.Steps = steps }()

	for i := 0; i <= len(steps); i++ {
		if len(u.U.roots) > i && u.U.roots[i] != nil && *u.U.roots[i] == h {
			steps = steps[:i]
			return nil
		}
		var step ProofStep
		if s, ok := u.Updated[h]; ok {
			step = s
			steps = append(steps[:i], step)
		} else if i == len(steps) {
			break
		} else {
			step = steps[i]
		}
		if step.Left {
			h = u.U.hasher(step.H, h)
		} else {
			h = u.U.hasher(h, step.H)
		}
	}

	return ErrInvalid
}
