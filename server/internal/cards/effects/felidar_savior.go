package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Felidar Savior — Creature — Cat Beast {3}{W}, 2/3 (Foundations, #4465):
//
//	"Lifelink
//	 When this creature enters, put a +1/+1 counter on each of up to two
//	 other target creatures you control."
//
// Basri's Acolyte's text on another body. The targets are chosen as the
// trigger goes on the stack: zero, one or two creatures you control, never
// the Savior itself ("other"). Each target still legal as it resolves gets
// one counter (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2bee2508-cd5d-4c62-a205-12d8bbb3143d",
		Name:            "Felidar Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Felidar Savior — put a +1/+1 counter on each of up to two other target creatures you control",
					putPlusOneCounterOnEachLegalTarget),
				Another(TargetCreature("up to two other target creatures you control", YouControl())).WithCount(0, 2)),
		},
	})
}
