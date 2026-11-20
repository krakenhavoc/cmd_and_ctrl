package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bringer of the Green Dawn — Creature — Bringer {7}{G}{G}, 5/5 (Fifth Dawn, #5604):
//
//	"You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana
//	 cost.
//	 Trample
//	 At the beginning of your upkeep, you may create a 3/3 green Beast
//	 creature token."
//
// The shared Bringer alternative cost (bringer_alternative_cost.go), and an
// optional upkeep trigger that creates the token table's 3/3 green Beast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f0548af1-d1cc-44d1-b5a0-f013563f29de",
		Name:             "Bringer of the Green Dawn",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{bringerAlternativeCost()},
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Bringer of the Green Dawn — create a 3/3 green Beast",
				Do(CreateToken{Template: TokenCard("3/3 green Beast"), N: 1})),
				"Bringer of the Green Dawn — create a 3/3 green Beast creature token?"),
		},
	})
}
