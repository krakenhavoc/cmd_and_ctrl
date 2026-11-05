package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cha'DIch Investigator — Creature — Klingon Warrior {4}{B}, 4/4 (Star Trek):
//
//	When this creature enters, create a 2/1 red Klingon Warrior creature
//	token with haste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfdef803-6f7a-4055-8d51-8e8c60953928",
		Name:         "Cha'DIch Investigator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cha'DIch Investigator — create a 2/1 red Klingon Warrior token with haste",
				Do(CreateToken{Template: TokenCard("2/1 red Klingon Warrior with haste"), N: 1})),
		},
	})
}
