package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// In Garruk's Wake — Sorcery {7}{B}{B}:
//
//	"Destroy all creatures you don't control and all planeswalkers
//	 you don't control."
//
// Nine mana for a one-sided wrath, which sounds unplayable and is
// one of the most-played finishers in the format anyway: at nine
// mana the game is decided by whoever still has a board, and this
// guarantees that it is you. Your creatures attack the turn it
// resolves.
//
// # The exclusion, and why it is not a target restriction
//
// "You don't control" is the S23 non-X exclusion in its simplest
// form: OpponentControls() as a predicate, applied at RESOLUTION to
// whatever is on the battlefield then. That matters in two ways a
// targeted removal spell does not share:
//
//   - Nothing is targeted, so hexproof, shroud and ward are all
//     irrelevant — the sweep is not a target and cannot be fizzled.
//   - Control is read when the spell resolves. A creature you
//     donated to an opponent in response dies; one you stole is
//     yours and survives.
//
// The two halves of the printed text — creatures and planeswalkers —
// are Or'd rather than swept separately, because the card destroys
// them at the same time and a permanent that is somehow both
// (a Layer-4 effect) should be destroyed once.
func init() {
	Register(Spec{
		OracleID:     "a6899b94-427d-4851-a474-4087e0a0918a",
		Name:         "In Garruk's Wake",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: And(OpponentControls(), Or(Creature(), Planeswalker())),
			}.Apply(ctx)
		},
	})
}
