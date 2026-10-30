package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent Bishop, Man in Black — {2}{W} Legendary Creature — Human Soldier 1/2
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on each
//	 of up to two target creatures."
//
// Rishkar's "each of up to two target creatures" clause on a
// beginning-of-combat trigger: any creatures, either player's, chosen as the
// trigger goes on the stack (CR 603.3d), and one that left in response is
// skipped at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f33217cc-175f-45e8-bc69-801f41cb4495",
		Name:         "Agent Bishop, Man in Black",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Agent Bishop — put a +1/+1 counter on each of up to two target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
		},
	})
}
