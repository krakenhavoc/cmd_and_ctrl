package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Protector — Creature — Angel {3}{W}, 2/2 (Jumpstart 2022, #5948):
//
//	"Flying
//	 Whenever this creature becomes the target of a spell or ability, this
//	 creature gets +0/+3 until end of turn."
//
// Phantasmal Image's "becomes the target" row (EventBecomesTarget on the
// source): one trigger per spell or ability that targets it, resolving
// above the targeting spell, so the toughness is there before the spell
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cff434e1-e567-4a80-8dc5-593d70b0f348",
		Name:            "Angelic Protector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self,
				"Angelic Protector — this creature gets +0/+3 until end of turn",
				effects.ThisGetsUntilEndOfTurn(0, 3, "Angelic Protector — +0/+3")),
		},
	})
}
