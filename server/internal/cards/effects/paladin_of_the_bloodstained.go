package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paladin of the Bloodstained — Creature — Vampire Knight {3}{W}, 3/2
// (Foundations Jumpstart):
//
//	"When this creature enters, create a 1/1 white Vampire creature token
//	 with lifelink."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f340018d-8b68-45d5-a72e-a922be68365d",
		Name:         "Paladin of the Bloodstained",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Paladin of the Bloodstained — create a 1/1 white Vampire creature token with lifelink",
				Do(CreateToken{Template: TokenCard("1/1 white Vampire with lifelink"), N: 1})),
		},
	})
}
