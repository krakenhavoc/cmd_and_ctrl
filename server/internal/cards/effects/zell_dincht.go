package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zell Dincht — Legendary Creature — Human Monk {2}{R}, 0/3 (Final
// Fantasy, #4462):
//
//	"You may play an additional land on each of your turns.
//	 Zell Dincht gets +1/+0 for each land you control.
//	 At the beginning of your end step, return a land you control to its
//	 owner's hand."
//
// Aesi's extra land play, Blackblade Reforged's land count as a layer
// 7c self pump, and the karoos' resolution-time choice for the end-step
// return: the land is chosen as the trigger resolves, not targeted, and
// with no land the trigger does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "f207ce73-e4cd-4dfc-b3b9-5704f5148959",
		Name:                "Zell Dincht",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Static: []game.StaticAbility{{
			Label:     "Zell Dincht gets +1/+0 for each land you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsLand() })
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Zell Dincht — return a land you control to its owner's hand",
				Do(ReturnOneYouControl{
					Match:    MatchLand,
					Question: "Zell Dincht — return a land you control to its owner's hand",
				})),
		},
	})
}
