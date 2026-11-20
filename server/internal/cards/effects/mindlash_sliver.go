package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7a08b371-1353-476d-ae72-379594f801da",
		Name:         "Mindlash Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: mindlashSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}, Sacrifice this permanent: Each player discards a card",
				Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  tspB1EachPlayerDiscardsACard("Mindlash Sliver — discard a card"),
			}},
			Text: "{1}, Sacrifice this permanent: Each player discards a card.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, mindlashSliverGrant),
		},
	})
}
