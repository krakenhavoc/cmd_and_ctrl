package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weapon Rack — Artifact {4} (Dominaria Remastered):
//
//	"This artifact enters with three +1/+1 counters on it.
//	 {T}: Move a +1/+1 counter from this artifact onto target creature.
//	 Activate only as a sorcery."
//
// Explorer's Cache's move (simicFluxmageMove), with three counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10d366cb-b92f-401c-a88e-60b267bd8ae5",
		Name:         "Weapon Rack",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 3, "Weapon Rack: enters with three +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Move a +1/+1 counter from this artifact onto target creature. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect:       effects.SimicFluxmageMove,
		}},
	})
}
