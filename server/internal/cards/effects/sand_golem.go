package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sand Golem — Artifact Creature — Golem {5}, 3/3:
//
//	"When a spell or ability an opponent controls causes you to discard
//	 this card, return this card from your graveyard to the battlefield
//	 with a +1/+1 counter on it at the beginning of the next end step."
//
// The family's "this card" trigger (ADR 0153 §4), watching from
// wherever the card landed (WhereverDiscarded, #3859) on a discard an opponent's spell or ability caused (the stack
// item's controller, CR 113.8). Unlike the Baloth it is not a
// replacement: the card reaches the graveyard first, then a delayed
// trigger at the next end step returns it under its owner's control with
// the counter as a CR 614.1c enters-with clause. The card is pinned as
// an object (CR 400.7): one that has left the graveyard in the meantime
// is not returned.
//
// No simplification: the discard ability is watched from wherever the
// discard left the card (WhereverDiscarded, #3859), so madness or Rest in
// Peace sending it to exile instead of the graveyard does not stop it.
func init() {
	Register(Spec{
		OracleID:          "0f9b78ca-1556-4157-a91c-55276fd78127",
		Name:              "Sand Golem",
		Completeness:      CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Triggered: []game.TriggeredAbility{
			WhereverDiscarded(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Sand Golem — discarded: return it to the battlefield with a +1/+1 counter at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return scheduleReturnOfThisCard(g, item,
						"Sand Golem — return it from your graveyard to the battlefield with a +1/+1 counter",
						returnDiscardedWithCounterBody)
				})),
		},
	})
}
