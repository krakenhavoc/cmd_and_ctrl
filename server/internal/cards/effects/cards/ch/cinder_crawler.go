package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cinder Crawler — Creature — Salamander {1}{R}, 1/2 (Exodus, S82 OLD b740,
// #5626):
//
//	"{R}: This creature gets +1/+0 until end of turn. Activate only if this
//	 creature is blocked."
//
// The gate is read on activation only: this creature is a blocked attacker
// (CR 509.1h), which it stays even if every creature blocking it leaves combat.
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +1/+0 until end of turn. Activate only if this creature is blocked."
	effects.Register(effects.Spec{
		OracleID:     "3850479f-1c1e-4a72-ae47-6da2997202e2",
		Name:         "Cinder Crawler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:      effects.ManaCost("{R}"),
			Condition: func(g *game.Game, _, source uuid.UUID) bool { return g.BlockedAttackerForEffect(source) },
			Effect:    effects.ThisGetsUntilEndOfTurn(1, 0, "Cinder Crawler — +1/+0"),
		}},
	})
}
