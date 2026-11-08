package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intimidation Tactics — Sorcery {B} (Aetherdrift, #4464):
//
//	"Target opponent reveals their hand. You choose an artifact or
//	 creature card from it. Exile that card.
//	 Cycling {3} ({3}, Discard this card: Draw a card.)"
//
// Appetite for Brains with the filter an artifact or creature card
// (ADR 0116's Exile variant). The hand is revealed to every player
// (CR 701.20a); exiling is not discarding, so madness and "whenever a
// player discards" never see it (CR 701.9a). A hand with no artifact or
// creature card is revealed and nothing is exiled (CR 609.3). Cycling is
// the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "121bd1af-da21-4937-a911-377c5b430796",
		Name:         "Intimidation Tactics",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Or(Artifact(), Creature()),
				Label:  "artifact or creature card",
				Exile:  true,
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
