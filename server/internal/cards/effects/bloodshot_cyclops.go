package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodshot Cyclops — Creature — Cyclops Giant {5}{R}, 4/4 (Urza's
// Destiny, #5621):
//
//	"{T}, Sacrifice a creature: This creature deals damage equal to the
//	 sacrificed creature's power to any target."
//
// Airdrop Condor's fling with {T} and any creature: the amount is the
// sacrificed creature's power as it last existed on the battlefield,
// floored at zero (CR 107.1b). The Cyclops may sacrifice itself and
// still deals the damage, as the creature it sacrificed. {T} needs it
// to have been under its controller's control since their most recent
// turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84d47913-53ac-4d4d-bac3-bc0306d43e12",
		Name:         "Bloodshot Cyclops",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: This creature deals damage equal to the sacrificed creature's power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeACreature()),
			Targets: TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageEqualToSacrificedPower(item, NewContext(g, item))
			},
		}},
	})
}
