package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Presence of the Wise — Sorcery {2}{W}{W} (Saviors of Kamigawa, #5601):
//
//	"You gain 2 life for each card in your hand."
//
// Gerrard's Wisdom's text: the hand is counted as the spell resolves
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6728064e-1d4e-4d4f-a0ab-127c23478f3b",
		Name:         "Presence of the Wise",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Player: ctx.Controller(), Amount: 2 * effects.SokB1CardsInYourHand(ctx)}.Apply(ctx)
		},
	})
}
