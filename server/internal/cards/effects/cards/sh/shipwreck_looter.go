package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shipwreck Looter — Creature — Human Pirate {1}{U}, 2/1 (Ixalan):
//
//	"Raid — When this creature enters, if you attacked this turn, you may
//	 draw a card. If you do, discard a card."
//
// Wingmate Roc's raid trigger (an intervening "if", CR 603.4, read as the
// creature enters and again as it resolves) with Sage of the Falls's loot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ead1accc-b90e-424a-a55f-2f4771726ce1",
		Name:         "Shipwreck Looter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB223RaidWhenThisEnters("Shipwreck Looter — raid: you may draw a card; if you do, discard a card",
				effects.OldB223MayLoot("Shipwreck Looter — draw a card, then discard a card?")),
		},
	})
}
