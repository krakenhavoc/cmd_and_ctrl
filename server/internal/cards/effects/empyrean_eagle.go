package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empyrean Eagle — Creature — Bird Spirit {1}{W}{U}, 2/3 (EDHREC
// rank 3077):
//
//	"Flying
//	 Other creatures you control with flying get +1/+1."
//
// The fliers lord. Flying rides PrintedKeywords; the anthem is a
// layer 7c modify over the controller's OTHER creatures that have
// flying when layer 7c runs — printed, granted by a lord, or
// granted until end of turn, since layer 6 has already been applied
// to every permanent by then (b29OtherCreaturesYouControlWithFlying).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "270d14b2-07bc-46bc-918f-658102265ccf",
		Name:            "Empyrean Eagle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b29OtherCreaturesYouControlWithFlying,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}
