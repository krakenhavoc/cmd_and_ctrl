package effects

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
	Register(Spec{
		OracleID:       "cda93cf5-da51-4117-b461-f4448aa2d16c",
		Name:           "Teyo, the Shieldmage",
		Completeness:   CompletenessFull,
		PlayerKeywords: []string{KeywordHexproof},
		Activated: []ActivatedAbility{{
			Label:  "−2: Create a 0/3 white Wall creature token with defender.",
			Cost:   LoyaltyCost(-2),
			Effect: Do(CreateToken{Template: TokenCard("0/3 white Wall with defender"), N: 1}),
		}},
	})
}
