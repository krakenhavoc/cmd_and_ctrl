package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brazen Collector — Creature — Raccoon Rogue {1}{R}, 2/1 (Bloomburrow,
// #4467):
//
//	"First strike
//	 Whenever this creature attacks, add {R}. Until end of turn, you
//	 don't lose this mana as steps and phases end."
//
// Kessig Naturalist's attack trigger with a fixed {R}: the mana carries
// the keep-until-end-of-turn rider (KeepManaUntilEndOfTurn), so it
// survives the combat steps and can pay for a spell after combat. First
// strike rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f93dc9cc-3dd2-4e2b-b972-217ab9d89dfe",
		Name:            "Brazen Collector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Brazen Collector — add {R}", Do(AddMana{
				Produced: "{R}",
				Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
			})),
		},
	})
}
