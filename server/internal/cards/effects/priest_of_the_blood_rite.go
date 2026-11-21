package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Priest of the Blood Rite — Creature — Human Cleric {3}{B}{B}, 2/2
// (Commander Masters):
//
//	"When this creature enters, create a 5/5 black Demon creature token with
//	 flying.
//	 At the beginning of your upkeep, you lose 2 life."
//
// Two printed triggers, two registered ones: the Demon token (a printed 5/5
// black flier) and the upkeep life loss, which is the Priest's controller's
// as the upkeep begins, whether or not the Priest still lives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "441cdd7d-6c4c-4a91-a118-7eaa35276660",
		Name:         "Priest of the Blood Rite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisEnters("Priest of the Blood Rite — create a 5/5 black Demon creature token with flying",
				Do(CreateToken{Template: TokenCard("5/5 black Demon with flying"), N: 1})),
				game.Purpose{Tokens: 1}),
			AtYourUpkeep("Priest of the Blood Rite — you lose 2 life", Do(GainLife{Amount: -2})),
		},
	})
}
