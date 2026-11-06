package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flopsie, Bumi's Buddy — Legendary Creature — Ape Goat {4}{G}{G}, 4/4
// (Avatar: The Last Airbender, #4459):
//
//	"When Flopsie enters, put a +1/+1 counter on each creature you
//	 control.
//	 Each creature you control with power 4 or greater can't be blocked
//	 by more than one creature."
//
// The entry is Abandoned Air Temple's counter on each creature you
// control, Flopsie included, one placement per creature, read as it
// resolves. The block limit is Outland Colossus's MaxBlockers over each
// creature its controller controls with power 4 or more, judged as
// blockers are declared (CR 509.1b): raising a creature's power after
// it is blocked by two doesn't undo the block (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8966245-501c-47e4-8ce0-849dcb532eb2",
		Name:         "Flopsie, Bumi's Buddy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Flopsie, Bumi's Buddy — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
		BlockRules: []game.BlockRule{
			MaxBlockers(OnMatching(And(Creature(), YouControl(), PowerGE(4))), 1),
		},
	})
}
