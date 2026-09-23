package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Willbender — 1/2 Human Wizard for {1}{U}:
//
//	"Morph {1}{U}. When this creature is turned face up, change the
//	 target of target spell or ability with a single target."
//
// The card morph was printed for, and half of it is here. The morph is
// whole: cast face down for {3} it is a nameless 2/2 neither opponent
// can price, and {1}{U} at any time they have to respond to — in
// response to the removal spell pointed at it, which is the point —
// turns it face up (CR 708.6, CR 702.37c). Willbender is also the
// printed reason the turn-face-up EVENT exists at all (CR 708.8,
// ADR 0082 decision 7): a permanent turning over has to be observable
// or this trigger has nothing to watch.
//
// DECLARED SIMPLIFICATION — the retarget is not modelled, and the
// trigger is NOT declared. "Change the target of target spell or
// ability with a single target" needs a primitive that rewrites a
// StackItem's chosen target and re-checks it under CR 115.7b (the new
// target must be legal, and the change simply does not happen if it is
// not). Nothing in the engine rewrites a stack item's targets, and
// building it here would be building a different seam inside this one;
// it is somebody's open work.
//
// A trigger that went on the stack and changed nothing was the other
// option and is the worse one: it would put a real response window in
// front of the table for an effect that is not there, and a player
// would hold up mana for it. Silence plus a caveat the catalog page
// shows before the card is sleeved is the honest shape (#259). When
// the retarget primitive lands, this file grows one
// `WhenThisIsTurnedFaceUp` and drops its caveat.
func init() {
	Register(Spec{
		OracleID:     "0aae277e-e58e-4115-b5fd-0459451e17ec",
		Name:         "Willbender",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Turning Willbender face up does not change the target of anything — retargeting a spell or ability is not implemented yet. The morph itself works.",
		},
		AlternativeCosts: []game.AlternativeCost{
			Morph("{1}{U}"),
		},
	})
}
