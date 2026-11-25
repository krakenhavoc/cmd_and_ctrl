package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marauding Looter — Creature — Human Pirate {2}{U}{R}, 4/3 (Ixalan):
//
//	"Raid — At the beginning of your end step, if you attacked this turn, you
//	 may draw a card. If you do, discard a card."
//
// Perforating Artist's raid end-step trigger (an intervening "if", CR 603.4,
// checked as the step begins and again as it resolves) with Sage of the Falls's
// loot: the discard happens only if the draw was taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "427bf4eb-3dfd-4747-be92-4a4a3d96a228",
		Name:         "Marauding Looter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB223RaidAtYourEndStep("Marauding Looter — raid: you may draw a card; if you do, discard a card",
				effects.OldB223MayLoot("Marauding Looter — draw a card, then discard a card?")),
		},
	})
}
