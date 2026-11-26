package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Shielddare — Creature — Kithkin Soldier {1}{W} 1/1:
//
//	"{W}, {T}: Target blocking creature gets +2/+2 until end of turn."
//
// A {T} cost, so summoning sickness applies (CR 302.6); the target must be a
// blocking creature (any player's) both when chosen and as the ability resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0734272d-5cce-476d-a663-fbf0bd7cd3a7",
		Name:         "Kithkin Shielddare",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Target blocking creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    game.AbilityCost{Mana: "{W}", Tap: true},
			Targets: effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			Effect:  effects.BoostTargetUntilEOT("Kithkin Shielddare — +2/+2 until end of turn", 2, 2),
		}},
	})
}
