package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darkslick Drake — Creature — Phyrexian Drake {2}{U}{U}, 2/4 (Scars of Mirrodin, #5577):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// Kingfisher's shape: printed flying and a dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cef11518-ecb2-4250-996c-a03fa9dfb09e",
		Name:            "Darkslick Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Darkslick Drake — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
