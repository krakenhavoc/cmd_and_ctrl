package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conscripted Infantry — Creature — Human Soldier {2}{R}, 3/1 (The
// Brothers' War, #5482):
//
//	"When this creature dies, create a 1/1 colorless Soldier artifact
//	 creature token."
//
// The token is an artifact creature, so it feeds every artifact count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf0f3aee-78b6-4a98-8e09-8861b73c2e74",
		Name:         "Conscripted Infantry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Conscripted Infantry — a 1/1 Soldier artifact creature", Do(CreateToken{Template: TokenCard(broB05SoldierToken), N: 1})),
		},
	})
}
