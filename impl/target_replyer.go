package impl

import "github.com/egon12/pgswitcher/ctrl"

var TargetReplyer ctrl.TargetReplyer = &targetReplyer{}

type targetReplyer struct{}

func (t *targetReplyer) Cancel(s ctrl.Source) {
	s.Write([]byte("you should be error"))
}
