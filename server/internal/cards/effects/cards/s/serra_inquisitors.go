package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra Inquisitors — Creature — Human Cleric {4}{W}, 3/3 (Homelands, S82 OLD
// b823, #5635):
//
//	"Whenever this creature blocks or becomes blocked by one or more black
//	creatures, this creature gets +2/+0 until end of turn."
//
// Amphibious Kavu's trigger with black as the colour: one trigger per block
// (CR 509.3a, 509.3c) however many black creatures are in it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c6a32e8-0c2a-45cb-8b79-c7f022fe5129",
		Name:         "Serra Inquisitors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.OldB823BlocksOrIsBlockedByBlack(true),
				"Serra Inquisitors — this creature gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Serra Inquisitors — +2/+0 until end of turn")),
		},
	})
}
