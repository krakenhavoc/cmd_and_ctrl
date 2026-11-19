package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cynette, Jelly Drover — Legendary Creature — Human Wizard {3}{U}, 2/2
// (Foundations Jumpstart):
//
//	"When Cynette enters or dies, create a 1/1 blue Jellyfish creature
//	 token with flying.
//	 Creatures you control with flying get +1/+1."
//
// One enters-or-dies trigger (WhenThisEntersOrDies) and Favorable
// Winds' flying anthem, read through the layers; it pumps its own
// Jellyfish and itself when it has flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8170ddd6-c054-4851-bdc2-4d319b60ba41",
		Name:         "Cynette, Jelly Drover",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Cynette, Jelly Drover — create a 1/1 blue Jellyfish with flying",
				Do(CreateToken{Template: TokenCard("1/1 blue Jellyfish with flying"), N: 1})),
		},
		Static: []game.StaticAbility{creaturesYouControlWithFlyingGetPlusOne()},
	})
}
