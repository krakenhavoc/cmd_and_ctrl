package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "441cdd7d-6c4c-4a91-a118-7eaa35276660",
		Name:         "Priest of the Blood Rite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Priest of the Blood Rite — create a 5/5 black Demon creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 black Demon with flying"), N: 1})),
				game.Purpose{Tokens: 1}),
			effects.AtYourUpkeep("Priest of the Blood Rite — you lose 2 life", effects.Do(effects.GainLife{Amount: -2})),
		},
	})
}
