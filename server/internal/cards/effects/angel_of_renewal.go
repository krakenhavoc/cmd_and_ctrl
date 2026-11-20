package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Renewal — Creature — Angel Ally {5}{W}, 4/4 (Battle for
// Zendikar, #5547):
//
//	"Flying
//	 When this creature enters, you gain 1 life for each creature you
//	 control."
//
// Conclave Phalanx's entry trigger on a flyer: the creatures are
// counted as it resolves (CR 608.2h), the Angel itself included while
// it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c604fda4-164c-4966-94f8-1e3a96d5ef9d",
		Name:            "Angel of Renewal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Angel of Renewal — you gain 1 life for each creature you control", cmdr2S37GainLifePerCreature),
		},
	})
}
