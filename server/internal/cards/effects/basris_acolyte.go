package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basri's Acolyte — Creature — Cat Cleric {2}{W}{W}, 2/3 (Jumpstart
// Selesnya deck request, #4893):
//
//	"Lifelink
//	 When this creature enters, put a +1/+1 counter on each of up to two
//	 other target creatures you control."
//
// The targets are chosen as the trigger goes on the stack: zero, one or two
// creatures you control, never the Acolyte itself ("other"). Each target
// still legal as it resolves gets one counter; one that left or changed
// hands in response gets none (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "92c824da-2351-43bd-9fa0-d09709ad8205",
		Name:            "Basri's Acolyte",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Basri's Acolyte — put a +1/+1 counter on each of up to two other target creatures you control",
					putPlusOneCounterOnEachLegalTarget),
				Another(TargetCreature("up to two other target creatures you control", YouControl())).WithCount(0, 2)),
		},
	})
}
