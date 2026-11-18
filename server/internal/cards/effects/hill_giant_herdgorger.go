package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hill Giant Herdgorger — Creature — Giant {4}{G}{G}, 7/6 (Adventures in the Forgotten Realms):
//
//	When this creature enters, you gain 3 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05c5752a-6ccb-45ff-bb85-5f89efd5944b",
		Name:         "Hill Giant Herdgorger",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hill Giant Herdgorger — gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
