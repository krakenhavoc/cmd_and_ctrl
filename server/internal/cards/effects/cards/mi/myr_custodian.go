package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Custodian — Artifact Creature — Myr {3}, 2/3:
//
//	"When this creature enters, scry 2. Then each opponent may scry 1."
//
// You scry first; once that is done each opponent in turn order is asked
// "scry 1?" and a yes scries before the next is asked (Model of Unity's
// ask-in-turn, with this card's own keys).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43923f24-7e95-4934-9a5c-8d9a5e06ddea",
		Name:         "Myr Custodian",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Myr Custodian — scry 2, then each opponent may scry 1", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				next := ctx.ItemTail(effects.MyrCustodianAskOpponents)
				next.Carry = ctx.Opponents()
				return effects.Scry{N: 2, ThenTail: next}.Apply(ctx)
			}),
		},
	})
}
