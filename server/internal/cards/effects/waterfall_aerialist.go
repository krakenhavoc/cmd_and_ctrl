package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterfall Aerialist — Creature — Djinn Wizard {3}{U}, 3/1 (Strixhaven:
// School of Mages, #5502):
//
//	"Flying
//	 Ward {2} (Whenever this creature becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays
//	 {2}.)"
//
// Flying is the printed keyword; ward is the shared trigger (Aegis
// Sculptor's), asking the opponent's spell or ability's controller to pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "736026b2-84d4-4452-ae0d-5d1b0af0a400",
		Name:            "Waterfall Aerialist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Waterfall Aerialist — ward {2}"),
		},
	})
}
