package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hundred-Handed One — 3/5 Giant for {2}{W}{W}:
//
//	"Vigilance
//	 {3}{W}{W}{W}: Monstrosity 3. (If this creature isn't monstrous,
//	 put three +1/+1 counters on it and it becomes monstrous.)
//	 As long as this creature is monstrous, it has reach and can block
//	 an additional ninety-nine creatures each combat."
//
// The monstrous line is two gated statics (#1700): reach, and — since
// #1706 — CanBlockAdditional(99), so a monstrous Giant blocks up to a
// hundred attackers and divides its damage among them (CR 510.1d).
// Both switch on with the designation and off again only when the
// Giant leaves the battlefield.
func init() {
	blocks := CanBlockAdditional(selfOnly, 99)
	blocks.ActiveWhen = Monstrous()
	Register(Spec{
		OracleID:        "954aaaa7-c3c1-4696-9ace-bcb5359b1709",
		Name:            "Hundred-Handed One",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{3}{W}{W}{W}"), 3)},
		Static:          []game.StaticAbility{MonstrousKeywords("reach"), blocks},
	})
}
