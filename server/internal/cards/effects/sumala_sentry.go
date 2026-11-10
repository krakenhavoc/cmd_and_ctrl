package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sumala Sentry — Creature — Elf Archer {G}{W} (Murders at Karlov Manor,
// #4469):
//
//	"Reach
//	 Whenever a face-down permanent you control is turned face up, put a
//	 +1/+1 counter on it and a +1/+1 counter on this creature."
//
// The permanent is face up by the time the event is emitted (CR 708.8),
// and it may be the Sentry itself (it has the ability once face up), which
// then gets both counters. Any permanent counts, not only a creature: a
// disguised land is turned up too. A counter goes on each permanent that
// is still on the battlefield as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6b869c33-a77c-4eb7-9df9-9815d70f7444",
		Name:            "Sumala Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventTurnedFaceUp, APermanentYouControlWasTurnedFaceUp,
				"Sumala Sentry — a +1/+1 counter on it and on this creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					for _, id := range []uuid.UUID{item.Trigger.Event.CardID, item.SourceCardID} {
						if onBattlefield(g, id) {
							if err := g.AddCounterForEffect(id, "+1/+1", 1); err != nil {
								return err
							}
						}
					}
					return nil
				}),
		},
	})
}
