package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slinking Giant — Creature — Giant Rogue {2}{R}{R}, 4/4 (Shadowmoor, #5591):
//
//	"Wither
//	 Whenever this creature blocks or becomes blocked, it gets -3/-0 until
//	 end of turn."
//
// Dromosaur's trigger with -3/-0: "blocks" triggers once however many
// attackers it blocks (CR 509.3a), "becomes blocked" once however many
// creatures block it (CR 509.3c). Wither rides PrintedKeywords.
//
// No simplification.
func init() {
	const label = "Slinking Giant — -3/-0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "6d86dcaa-2295-4f4d-ada7-216f96138d81",
		Name:            "Slinking Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"wither"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				label, effects.ThisGetsUntilEndOfTurn(-3, 0, label)),
		},
	})
}
