package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boon-Bringer Valkyrie — Creature — Angel Warrior {3}{W}{W}, 4/4 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following abilities until end of turn.)
//	 Flying, first strike, lifelink"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// flying, first strike, lifelink until end of turn; the card has them itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8652d88f-edca-4235-b942-6e7467608de8",
		Name:            "Boon-Bringer Valkyrie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Backup("Boon-Bringer Valkyrie", 1, "", "flying", "first strike", "lifelink"),
		},
	})
}
