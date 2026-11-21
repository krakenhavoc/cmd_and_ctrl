package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reaver Titan — Artifact — Vehicle {7}, 10/10:
//
//	"Void Shields — Protection from mana value 3 or less
//	 Gatling Blaster — Whenever this Vehicle attacks, it deals 5
//	 damage to each opponent.
//	 Crew 4"
//
// The protection is the grammar's mana value quality (#2181, CR
// 702.16a): the source's mana value is compared at every protection
// check, so a cheap creature can't block it, a cheap spell can't
// target it or damage it, and a cheap Aura or Equipment can't attach
// to it. A spell counts the X it was cast with (CR 202.3e) and a token
// that is no copy is mana value 0.
//
// Protection applies while it is a creature as well as while it is
// not: the Vehicle carries the keyword on the battlefield either way,
// and the damage and targeting checks read it on any permanent.
//
// The attack trigger is "it deals 5 damage to each opponent", one
// damage event per opponent from the Vehicle itself. Crew 4 is the
// standard crew cost (CR 702.122).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2951671b-ac01-4e65-857f-0b7b7c7478ce",
		Name:            "Reaver Titan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from mana value 3 or less"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			}, "Reaver Titan — 5 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 5)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    effects.CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Reaver Titan"),
		}},
	})
}
