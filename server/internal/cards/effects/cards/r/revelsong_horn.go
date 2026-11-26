package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revelsong Horn — Artifact {2}:
//
//	"{1}, {T}, Tap an untapped creature you control: Target creature gets
//	 +1/+1 until end of turn."
//
// The creature tapped as a cost is any untapped creature you control, chosen
// as the cost is paid, with no summoning-sickness check (it is not its own {T}
// symbol, CR 302.6); it does not target, so a hexproof creature can pay.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d80d87f5-af9e-404b-88e3-7a4b419dc32c",
		Name:         "Revelsong Horn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Tap an untapped creature you control: Target creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost: func() game.AbilityCost {
				c := effects.TapAnUntapped("an untapped creature you control", effects.Creature())
				c.Mana = "{1}"
				c.Tap = true
				return c
			}(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Revelsong Horn — +1/+1 until end of turn", 1, 1),
		}},
	})
}
