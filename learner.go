package hessfree

import (
	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/sgd"
)

const defaultDampingCoeff = 1

// A Learner has learnable parameters and can create
// Objectives based on a sample set and the current
// set of parameters.
type Learner interface {
	// Parameters returns the learnable parameters that may
	// be adjusted by Adjust().
	Parameters() []*autofunc.Variable

	// MakeObjective creates an objective whose approximation
	// is centered around the current underlying variables.
	//
	// This should be called once before every Adjust() call,
	// yielding a cycle of MakeObjective and Adjust.
	// This way, each Objective can be tweaked based on info
	// about the previous cycle (e.g. for damping).
	MakeObjective() Objective

	// Adjust updates the parameters after an Objective
	// (created by MakeObjective()) has been minimized.
	//
	// The provided SampleSet is the set of samples for which
	// the given delta supposedly to improves the cost.
	// Adjust may need this sample set to analyze the effects
	// of the delta, e.g. for damping purposes.
	Adjust(d ConstParamDelta, s sgd.SampleSet)
}

// A DampingLearner wraps a learner in the damping
// mechanism described in Martens (2010).
type DampingLearner struct {
	WrappedLearner Learner

	// DampingCoeff is the coefficient for the squared
	// deltas in the damping term.
	// It is adjusted during training using the heuristic
	// described in Martens (2010).
	// If DampingCoeff is 0, it will be set to a default
	// value during the first training iteration.
	//
	// During damping, this coefficient is multiplied by
	// the number of samples in each sample set, since it
	// is assumed that the total cost is the sum of the
	// costs for each sample.
	DampingCoeff float64

	lastObjective Objective
}

func (d *DampingLearner) Parameters() []*autofunc.Variable {
	return d.WrappedLearner.Parameters()
}

func (d *DampingLearner) MakeObjective() Objective {
	if d.DampingCoeff == 0 {
		d.DampingCoeff = defaultDampingCoeff
	}
	d.lastObjective = d.WrappedLearner.MakeObjective()
	return &dampedObjective{
		WrappedObjective: d.lastObjective,
		Coeff:            d.DampingCoeff,
	}
}

func (d *DampingLearner) Adjust(delta ConstParamDelta, s sgd.SampleSet) {
	quadOffset := d.lastObjective.Quad(delta, s)
	centerVal := d.lastObjective.Objective(ConstParamDelta{}, s)
	realOffset := d.lastObjective.Objective(ConstParamDelta{}, s)
	delta.AddToVars()

	trust := (realOffset - centerVal) / (quadOffset - centerVal)
	if trust < 0.25 {
		d.DampingCoeff *= 3.0 / 2.0
	} else if trust > 0.75 {
		d.DampingCoeff *= 2.0 / 3.0
	}
}

type dampedObjective struct {
	WrappedObjective Objective
	Coeff            float64
}

func (d *dampedObjective) Quad(delta ConstParamDelta, s sgd.SampleSet) float64 {
	res := d.WrappedObjective.Quad(delta, s)
	scaler := float64(s.Len())
	for _, subDelta := range delta {
		for _, x := range subDelta {
			res += scaler * x * x
		}
	}
	return res
}

func (d *dampedObjective) QuadGrad(delta ConstParamDelta, s sgd.SampleSet) ConstParamDelta {
	res := d.WrappedObjective.QuadGrad(delta, s)

	scaler := float64(2 * s.Len())
	for variable, subDelta := range delta {
		resVec := res[variable]
		for i, x := range subDelta {
			resVec[i] += scaler * x
		}
	}

	return res
}

func (d *dampedObjective) QuadHessian(delta ConstParamDelta, s sgd.SampleSet) ConstParamDelta {
	res := d.WrappedObjective.QuadHessian(delta, s)

	scaler := float64(2 * s.Len())
	for variable, subDelta := range delta {
		resVec := res[variable]
		for i, x := range subDelta {
			resVec[i] += scaler * x
		}
	}

	return res
}

func (d *dampedObjective) Objective(delta ConstParamDelta, s sgd.SampleSet) float64 {
	return d.WrappedObjective.Objective(delta, s)
}
