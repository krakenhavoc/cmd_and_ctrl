package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erratic Apparition — Creature — Spirit {2}{U}, 1/3:
//
//	"Flying, vigilance
//	Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, this creature gets +1/+1 until end of turn."
//
// Flying and vigilance ride PrintedKeywords; each trigger adds another +1/+1, and they stack.
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "23d5d43c-ec78-42dd-a53e-a1ee91c94b1d",
		Name:            "Erratic Apparition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Eerie("Erratic Apparition — +1/+1 until end of turn (eerie)",
				thisCreatureUntilEOT("Erratic Apparition — +1/+1 until end of turn", 1, 1)),
		},
	})
}
