package hessfree

import (
	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/num-analysis/linalg"
)

// A ParamDelta is a displacement vector (t - t0) where
// t contains new variable values and t0 contains the
// current value of those variables.
type ParamDelta map[*autofunc.Variable]autofunc.RResult

// outputRVector returns an RVector mapping variables in
// p to the corresponding RResults' outputs.
func (p ParamDelta) outputRVector() autofunc.RVector {
	res := autofunc.RVector{}
	for variable, r := range p {
		res[variable] = r.Output()
	}
	return res
}

// rOutputRVector is like outputRVector, but it uses the
// RResults' ROutput instead of their Output.
func (p ParamDelta) rOutputRVector() autofunc.RVector {
	res := autofunc.RVector{}
	for variable, r := range p {
		res[variable] = r.ROutput()
	}
	return res
}

// zeroGradient produces an autofunc.Gradient with zero
// vectors for all of the delta's variables.
func (p ParamDelta) zeroGradient() autofunc.Gradient {
	res := autofunc.Gradient{}
	for variable := range p {
		res[variable] = make(linalg.Vector, len(variable.Vector))
	}
	return res
}