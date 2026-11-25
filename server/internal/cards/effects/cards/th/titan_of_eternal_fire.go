package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const titanOfEternalFireGrant = "titan-of-eternal-fire/ping"

// Titan of Eternal Fire — Creature — Giant {5}{R}, 5/6 (Commander 2020, S82
// OLD, #5513):
//
//	"Each Human creature you control has "{R}, {T}: This creature deals 1
//	 damage to any target.""
//
// Zombie Master's granted activated ability (ADR 0093) over Viridian
// Longbow's ping: "this creature" is the Human that has it, only its
// controller can pay for it, and its {T} follows summoning sickness
// (CR 302.6). The Titan itself is a Giant, not a Human, so it gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "490df8a0-182a-40b8-a36e-f2666fa7c7fa",
		Name:         "Titan of Eternal Fire",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: titanOfEternalFireGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{R}, {T}: This creature deals 1 damage to any target.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(1),
			}},
			Text: "{R}, {T}: This creature deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Human"}, YoursOnly: true}, titanOfEternalFireGrant),
		},
	})
}
