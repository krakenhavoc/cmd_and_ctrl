package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rotting Rats — Creature — Zombie Rat {1}{B}, 1/1 (Conflux, #5587):
//
//	"When this creature enters, each player discards a card.
//	 Unearth {1}{B} ({1}{B}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next
//	 end step or if it would leave the battlefield. Unearth only as a
//	 sorcery.)"
//
// Cunning Lethemancer's "each player discards a card" on an enters
// trigger, so an unearthed Rats makes everyone discard again. Each
// player chooses their own card (CR 701.9a). The unearth ability is the
// shared one (unearth.go, CR 702.84a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f361ae9-98a2-40ff-a579-fa5b9f510df4",
		Name:         "Rotting Rats",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rotting Rats — each player discards a card", func(g *game.Game, item *game.StackItem) error {
				g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
					Source: item.SourceCardID, N: 1, Question: "Rotting Rats — discard a card",
				})
				return nil
			}),
		},
		Activated: []ActivatedAbility{Unearth("{1}{B}")},
	})
}
