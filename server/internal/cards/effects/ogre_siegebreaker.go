package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ogre Siegebreaker — Creature — Ogre Berserker {2}{B}{R}, 4/3 (Core Set 2020, #5518):
//
//	"{2}{B}{R}: Destroy target creature that was dealt damage this turn."
//
// Fatal Blow's target clause on an activated ability, judged at
// activation and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1432db6a-9466-48f6-9011-60728c437608",
		Name:         "Ogre Siegebreaker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}{R}: Destroy target creature that was dealt damage this turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    ManaCost("{2}{B}{R}"),
			Targets: TargetCreature("target creature that was dealt damage this turn", DealtDamageThisTurn()),
			Effect:  destroyChosenPermanent,
		}},
	})
}
