package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freyalise Supplicant — Creature — Human Cleric {1}{G}, 1/1 (Ice Age, S82 OLD, #5636):
//
//	"{T}, Sacrifice a red or white creature: This creature deals damage to any target
//	 equal to half the sacrificed creature's power, rounded down."
//
// Bloodshot Cyclops's sacrifice-for-power with the red-or-white restriction on the cost
// and the amount halved. The power is the sacrificed creature's last-known power (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c28a7641-49e3-43b1-a366-8d6e9c8dab9c",
		Name:         "Freyalise Supplicant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a red or white creature: This creature deals damage to any target equal to half the sacrificed creature's power, rounded down.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a red or white creature", effects.Creature(), effects.Or(effects.OfColor("R"), effects.OfColor("W")))),
			Targets: effects.TargetAny(),
			Effect:  effects.S82OldB841HalfSacrificedPowerDamage,
		}},
	})
}
