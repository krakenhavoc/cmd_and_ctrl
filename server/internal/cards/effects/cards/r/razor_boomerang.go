package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razor Boomerang — Artifact — Equipment {3} (Worldwake, S82 OLD batch 490,
// #5582):
//
//	"Equipped creature has "{T}, Unattach Razor Boomerang: It deals 1 damage to
//	 any target. Return Razor Boomerang to its owner's hand."
//	 Equip {2}"
//
// Toralf's Hammer's shape: the equipped creature is granted an activated
// ability whose cost taps it and unattaches the Equipment (UnattachTheGrantor),
// and whose effect is one damage from the creature to any target, then the
// Equipment returns to its owner's hand if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2649e48-4252-43f3-bd1e-c4c6fcb0aa3a",
		Name:         "Razor Boomerang",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: razorBoomerangGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}, Unattach Razor Boomerang: It deals 1 damage to any target. Return Razor Boomerang to its owner's hand.",
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.TapCost(), effects.UnattachTheGrantor()),
				Targets: effects.TargetAny(),
				Effect:  effects.S82OldB490DamageToTargetThenReturnGrantor(1),
			}},
			Text: "{T}, Unattach Razor Boomerang: It deals 1 damage to any target. Return Razor Boomerang to its owner's hand.",
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(razorBoomerangGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

const razorBoomerangGrant = "razor-boomerang/throw"
