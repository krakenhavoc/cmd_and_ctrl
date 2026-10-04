package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fist of Suns — Artifact {3}:
//
//	"You may pay {W}{U}{B}{R}{G} rather than pay the mana cost for
//	 spells you cast."
//
// The granted-alternative-cost seam (ADR 0118 §3, #2163): the offer is
// derived from the battlefield on every cast query, so it lasts exactly
// as long as Fist of Suns is under its controller. CR 118.9 and 601.2b.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                "6fd5e591-fae8-4128-a4d3-a848c8a8ffda",
		Name:                    "Fist of Suns",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayWUBRGForSpellsYouCast()},
	})
}
