package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dread Fugue — Sorcery {B}:
//
//	"Cleave {2}{B} (You may cast this spell for its cleave cost. If you
//	 do, remove the words in square brackets.)
//	 Target player reveals their hand. You choose a nonland card from it
//	 [with mana value 2 or less]. That player discards that card."
//
// Cleave (CR 702.148a) is an alternative cost; paying it removes the
// bracketed words, which here touch only the pick's filter, not the
// target clause. So the pick (ADR 0116) is "nonland card with mana
// value 2 or less" for {B} and "nonland card" for {2}{B}. Mana value is
// read as the card is in the hand: X is 0 (CR 202.3e), a split card's
// halves are combined, and a modal double-faced card is its front face.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "efe4194a-b3c5-4308-abfe-553e035f5e64",
		Name:             "Dread Fugue",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Cleave("{2}{B}", nil)},
		Targets:          TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			pick := ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: And(Nonland(), ManaValueLE(2)),
				Label:  "nonland card with mana value 2 or less",
			}
			if ctx.PaidAltCost("cleave") {
				pick.Filter, pick.Label = Nonland(), "nonland card"
			}
			return pick.Apply(ctx)
		},
	})
}
