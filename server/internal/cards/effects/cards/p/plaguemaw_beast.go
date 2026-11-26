package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plaguemaw Beast — Creature — Phyrexian Beast {3}{G}{G}, 4/3 (Mirrodin
// Besieged, S82 OLD b466, #5576):
//
//	"{T}, Sacrifice a creature: Proliferate."
//
// The cost is the {T} symbol (so summoning sickness applies) and a creature
// the controller picks at announce, the Beast itself included; the
// proliferate is the shared primitive, which asks which permanents and
// players to add a counter to as the ability resolves (CR 701.34).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22252186-15d0-468f-8cab-c74db9b71439",
		Name:         "Plaguemaw Beast",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a creature", effects.Creature())),
			Effect:  effects.Do(effects.Proliferate{}),
		}},
	})
}
