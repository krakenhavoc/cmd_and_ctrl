package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow of the Grave — {1}{B} Instant (Amonkhet, S82 old batch 247):
//
//	"Return to your hand all cards in your graveyard that you cycled or
//	 discarded this turn."
//
// Cycling's cost is to discard the card (CR 702.29d), so every card cycled or
// discarded is one EventDiscardCard by this player; each such card still in
// the graveyard as the spell resolves returns, and one that has already left
// (or came back and went elsewhere, a new object, CR 400.7) does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f33d1132-064d-42f1-8814-95d27fde8bf8",
		Name:         "Shadow of the Grave",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			you := ctx.Controller()
			seen := map[string]bool{}
			for _, ev := range ctx.Game.EventsThisTurn() {
				if ev.Kind != game.EventDiscardCard || ev.Actor != you || seen[ev.CardID.String()] {
					continue
				}
				seen[ev.CardID.String()] = true
				if z := ctx.Game.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneGraveyard {
					continue
				}
				if err := (effects.ReturnFromGraveyard{Target: ev.CardID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
