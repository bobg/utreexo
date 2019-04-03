package utreexo

import "errors"

type (
	ProofStep struct {
		H    Hash
		Left bool
	}

	Proof struct {
		Leaf  Hash
		Steps []ProofStep // the steps progress from the leaf to the root
	}
)

var (
	ErrDeleted = errors.New("deleted")
	ErrInvalid = errors.New("invalid proof")
)

func (p *Proof) Update(u Update) error {
	if u.Deleted[p.Leaf] {
		return ErrDeleted
	}

	h := p.Leaf
	steps := p.Steps

	defer func() { p.Steps = steps }()

	for i := 0; i < len(steps); i++ {
		step := steps[i]
		if u.U.roots[i] != nil && *u.U.roots[i] == h {
			steps = steps[:i]
			return nil
		}
		if step.Left {
			h = u.U.hasher(step.H, h)
		} else {
			h = u.U.hasher(h, step.H)
		}
		if newStep, ok := u.Updated[h]; ok {
			steps = append(steps[:i], newStep)
		}
	}

	return ErrInvalid
}
