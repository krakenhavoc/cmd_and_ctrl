package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of the Dawn — Creature — Angel {4}{W}, 3/3 (Core Set 2019, #5527):
//
//	"Flying
//	When this creature enters, creatures you control get +1/+1 and gain vigilance until end of turn."
//
// Flying, and Devoted Paladin's enters trigger: the creatures you
// control as it resolves get +1/+1 and vigilance until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f63ea0e5-7769-4a22-bbb3-658281141d8c",
		Name:            "Angel of the Dawn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Angel of the Dawn — creatures you control get +1/+1 and gain vigilance until end of turn",
				Do(
					BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "Angel of the Dawn — +1/+1"},
					GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"vigilance"}, Label: "Angel of the Dawn — vigilance"},
				)),
		},
	})
}
