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
// The monstrous line is a gated static (#1700): reach switches on with
// the designation and off again only when the Giant leaves the
// battlefield.
//
// Simplification: the engine's block declaration maps each blocker to
// ONE attacker (block_rules.go), so there is no shape for "can block
// an additional N creatures" — Brave the Sands carries the same
// caveat. Weaker than printed, never stronger: a monstrous
// Hundred-Handed One still blocks one creature, flier or not.
func init() {
	Register(Spec{
		OracleID:        "954aaaa7-c3c1-4696-9ace-bcb5359b1709",
		Name:            "Hundred-Handed One",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Once monstrous it can still block only one creature each combat, not a hundred."},
		PrintedKeywords: []string{"vigilance"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{3}{W}{W}{W}"), 3)},
		Static:          []game.StaticAbility{MonstrousKeywords("reach")},
	})
}
