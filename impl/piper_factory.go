package impl

import "github.com/egon12/pgswitcher/ctrl"

type piperFactor struct{}

var PiperFactory = &piperFactor{}

func (p *piperFactor) New() ctrl.Piper {
	return &Piper{
		b:  &buffer{},
		tr: &TargetReader{},
	}
}
