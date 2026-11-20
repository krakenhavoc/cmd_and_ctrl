package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darkheart Sliver — Creature — Sliver {B}{G}, 2/2 (Planar Chaos, #5595):
//
//	"All Slivers have "Sacrifice this permanent: You gain 3 life.""
//
// Firewake Sliver's granted activated ability (ADR 0093) with Dedicated
// Martyr's body: each Sliver's controller may sacrifice that Sliver, any
// time they have priority, to gain 3 life.
//
// No simplification.
const darkheartSliverGrant = "darkheart-sliver/gain-life"

func init() {
	Register(Spec{
		OracleID:     "88e5b905-fe52-4fd6-8cdf-b7ede5166ebc",
		Name:         "Darkheart Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: darkheartSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "Sacrifice this permanent: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 3},
				Cost:    SacrificeThis(),
				Effect:  Do(GainLife{Amount: 3}),
			}},
			Text: "Sacrifice this permanent: You gain 3 life.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, darkheartSliverGrant),
		},
	})
}
