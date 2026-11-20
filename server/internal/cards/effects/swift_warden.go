package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swift Warden — Creature — Merfolk Warrior {1}{G}{G}, 3/3 (Rivals of
// Ixalan, #5531):
//
//	"Flash
//	 When this creature enters, target Merfolk you control gains hexproof
//	 until end of turn. (It can't be the target of spells or abilities
//	 your opponents control.)"
//
// Flash is printed data; the entry's target is chosen as the trigger goes
// on the stack and may be the Warden itself, and it gains hexproof as the
// trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0928ddda-20ea-462e-9f00-6c10ef1cfb4b",
		Name:            "Swift Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Swift Warden — target Merfolk you control gains hexproof until end of turn",
				jmpB7FirstLegalTargetGains("Swift Warden — hexproof", "hexproof")),
				TargetCreature("target Merfolk you control", YouControl(), Subtype("Merfolk"))),
		},
	})
}
