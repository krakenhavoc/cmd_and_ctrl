package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindlash Sliver — Creature — Sliver {B}, 1/1 (Time Spiral, #5596):
//
//	"All Slivers have "{1}, Sacrifice this permanent: Each player
//	 discards a card.""
//
// A granted activated ability (ADR 0093) on every Sliver at the table,
// whose cost sacrifices the Sliver that has it. Each player, its activator
// included, discards a card of their own choice.
//
// No simplification.
const mindlashSliverGrant = "mindlash-sliver/discard"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a08b371-1353-476d-ae72-379594f801da",
		Name:         "Mindlash Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: mindlashSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}, Sacrifice this permanent: Each player discards a card",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.TspB1EachPlayerDiscardsACard("Mindlash Sliver — discard a card"),
			}},
			Text: "{1}, Sacrifice this permanent: Each player discards a card.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, mindlashSliverGrant),
		},
	})
}
