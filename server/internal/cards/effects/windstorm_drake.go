package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windstorm Drake — Creature — Drake {4}{U}, 3/3 (Jumpstart 2020):
//
//	"Flying
//	 Other creatures you control with flying get +1/+0."
//
// Empyrean Eagle's lord at +1/+0: a layer 7c modify over the other
// creatures its controller controls that have flying when layer 7c
// runs, after every layer-6 grant or removal of flying.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:        "16977ebd-6384-4c28-b9c3-448563a07807",
		Name:            "Windstorm Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Other creatures you control with flying get +1/+0.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b29OtherCreaturesYouControlWithFlying,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
	})
}
