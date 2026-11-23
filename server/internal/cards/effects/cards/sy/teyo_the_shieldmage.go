package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teyo, the Shieldmage — Legendary Planeswalker — Teyo {2}{W}, loyalty 5 (War of the Spark, #5520):
//
//	"You have hexproof. (You can't be the target of spells or abilities your opponents control.)
//	 −2: Create a 0/3 white Wall creature token with defender."
//
// Starting loyalty is printed data (ADR 0032). "You have hexproof" is a player keyword
// (Metropolis Reformer's); the −2 makes a 0/3 white Wall with defender.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "cda93cf5-da51-4117-b461-f4448aa2d16c",
		Name:           "Teyo, the Shieldmage",
		Completeness:   effects.CompletenessFull,
		PlayerKeywords: []string{effects.KeywordHexproof},
		Activated: []effects.ActivatedAbility{{
			Label:  "−2: Create a 0/3 white Wall creature token with defender.",
			Cost:   effects.LoyaltyCost(-2),
			Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("0/3 white Wall with defender"), N: 1}),
		}},
	})
}
