package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Backslide — Instant for {1}{U}:
//
//	"Turn target creature with a morph ability face down.
//	 Cycling {U}"
//
// The card the "turn a permanent face down" primitive was printed for
// (#1209), and the one that shows why turning face down is NOT the
// mirror of turning face up. Backslide re-hides a morph, and the
// morph can then be turned face up AGAIN for its morph cost — so
// everything its "when this is turned face up" trigger does, it does
// twice. That is not Backslide's gift: CR 702.37e is "any time you
// have priority, you may turn a face-down permanent you control with
// a morph ability face up", which keys on the CARD having morph and
// not on how the permanent came to be face down. CR 708.7's "the
// ability or rules that allow a permanent to be face down MAY ALSO
// allow … turn it face up" is the same fact from the other side, and
// it is why the engine records `turned` as a face-down kind of its
// own instead of reusing `morphed`.
//
// The clause is `WithMorphAbility()` and not "any creature": megamorph
// is a morph ability (CR 702.37b) and disguise is not (CR 702.168),
// and a face-down permanent is not a legal target either, because
// CR 708.2a leaves it with no declaration to read — which costs
// nothing, since CR 708.2b says turning it face down again would do
// nothing anyway.
//
// No simplification. Cycling is the shared keyword ability.
func init() {
	Register(Spec{
		OracleID:     "6d9746bc-0b72-4ac4-b48a-742b63b1c41b",
		Name:         "Backslide",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with a morph ability", WithMorphAbility()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return turnSingleTargetFaceDown(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{U}")},
	})
}
