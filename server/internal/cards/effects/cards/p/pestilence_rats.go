package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pestilence Rats — Creature — Rat {2}{B}, */3 (Ice Age, S82 OLD, #5636):
//
//	"Pestilence Rats's power is equal to the number of other Rats on the battlefield."
//
// A characteristic-defining ability (CR 604.3, layer 7a): power is the count of other Rats, whoever
// controls them, read live; toughness is the printed 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc6449c8-f0cd-4c07-8acf-436f97e8bbb2",
		Name:         "Pestilence Rats",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Pestilence Rats's power is equal to the number of other Rats on the battlefield.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply:     effects.S82OldB843PestilenceRatsPower,
		}},
	})
}
