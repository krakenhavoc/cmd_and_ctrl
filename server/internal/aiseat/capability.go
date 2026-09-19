package aiseat

// capability.go is how an OPTIONAL Policy extension survives being
// wrapped.
//
// # The defect this exists to prevent
//
// A Policy is one required interface (Name, Decide) and a growing set
// of optional ones — Conceder, Tracer, TargetOrderer, CostFuelPricer,
// Improviser, Spender. The runner and the enumerator find them by
// type assertion on the policy they were handed, and a type assertion
// that answers false is indistinguishable from a policy that has no
// opinion. There is no error, no log line and no failing test: the
// feature simply stops happening.
//
// Every shipped tier is a policy inside a WRAPPER — `heuristic` is a
// rules.Filter around the heuristic, `assisted` and `strong` are a
// model.Policy over it — and a wrapper implements the two required
// methods and, historically, whichever optional ones its author
// remembered. #1060 is what that costs: #687's threat ordering and
// #1013's fuel pricing were built, tested and dead on every seat the
// lobby could create, for a month, because neither wrapper forwarded
// them and the pin tests asserted against a bare heuristic no seat
// ever gets.
//
// # The mechanism
//
// A wrapper declares WHAT IT WRAPS, once, and never again has to know
// which optional interfaces exist:
//
//	func (f *Filter) Unwrap() aiseat.Policy { return f.Inner }
//
// and every lookup goes through Capability, which walks that chain
// outward-in and returns the OUTERMOST implementer. Outward-in is the
// whole of the semantics and it is the right rule in both directions:
// a wrapper that implements an extension itself means to override it
// (rules.Filter is a Tracer, and its own verdict is the one that must
// be reported), and a wrapper that does not implement it means to be
// transparent (rules.Filter is not a TargetOrderer, and the heuristic
// underneath it is).
//
// The alternative — a capability struct the factory fills in once —
// was rejected for one reason: it has to be edited every time an
// optional interface is added, which is exactly the omission being
// fixed. A chain covers the interface written next year without
// anybody touching a wrapper.

// maxUnwrapDepth bounds the walk. Three is the deepest chain this
// repository assembles (a test wrapper over model.Policy over the
// heuristic); the bound is here so that a policy that unwraps to
// itself is a lookup that fails rather than a bot seat that hangs.
const maxUnwrapDepth = 16

// Unwrapper is implemented by a Policy that WRAPS another policy. It
// is the whole of what a wrapper has to declare for every optional
// Policy extension to keep working through it — see the file comment.
//
// Named after errors.Unwrap, and for the same reason: the caller is
// asking "is there something underneath this", and the answer is a
// chain rather than a flag.
//
// Returning nil is legal and ends the chain, so a wrapper whose inner
// policy is optional does not need a second interface.
type Unwrapper interface {
	Unwrap() Policy
}

// Capability finds the outermost policy in p's wrapper chain that
// implements T, which is how the runner and the enumerator ask for
// every optional Policy extension.
//
// It is a plain type assertion for the common case — a policy that
// wraps nothing answers in one step — so the cost of the chain is
// paid only by the seats that have one.
//
//	orderer, ok := Capability[TargetOrderer](r.policy)
//
// A nil policy, or a chain with no implementer, answers the zero T
// and false, which is the same answer a bare type assertion gave and
// means the same thing: this seat has no opinion, keep the behaviour
// it had before the extension existed.
func Capability[T any](p Policy) (T, bool) {
	var zero T
	for depth := 0; p != nil && depth < maxUnwrapDepth; depth++ {
		if c, ok := any(p).(T); ok {
			return c, true
		}
		u, ok := any(p).(Unwrapper)
		if !ok {
			break
		}
		p = u.Unwrap()
	}
	return zero, false
}
