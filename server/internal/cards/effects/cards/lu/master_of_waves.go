package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master of Waves — Creature — Merfolk Wizard {3}{U}, 2/1 (Secret Lair Drop):
//
//	"Protection from red
//	 Elemental creatures you control get +1/+1.
//	 When this creature enters, create a number of 1/0 blue Elemental
//	 creature tokens equal to your devotion to blue."
//
// The anthem is the tribal lord (TribalAnthem; Master is not an Elemental
// itself, so "other" makes no difference). Devotion counts every {U} in the
// mana costs of permanents the controller has (CR 700.5), read as the
// trigger resolves. The tokens are 1/0 until the anthem lifts them to 2/1, so
// a Master that has left leaves them to die as state-based actions (CR 704.5f).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "07d40173-53e7-4328-a3b0-6214d93855b6",
		Name:            "Master of Waves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Elemental"}, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Master of Waves — create a number of 1/0 blue Elemental creature tokens equal to your devotion to blue",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("1/0 blue Elemental"),
					Name:     "your devotion to blue",
					Count: func(ctx *effects.Context) int {
						return effects.DevotionTo(ctx.Game, ctx.Controller(), ctx.TextColor("U"))
					},
				})),
		},
	})
}
