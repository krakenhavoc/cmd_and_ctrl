package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aerial Engineer — Creature — Human Artificer {2}{W}{U}, 2/4
// (Core Set 2019, #5527):
//
//	"As long as you control an artifact, this creature gets +2/+0 and has flying."
//
// Goblin Tomb Raider's shape: a layer-7c pump and a layer-6 keyword
// grant of flying, both while you control an artifact.
//
// No simplification.
func init() {
	whileArtifact := effects.WhileYouHave(effects.CloudsculptControlsAnArtifact, effects.SelfCreature)
	flying := effects.KeywordGrant(whileArtifact, "flying")
	flying.Label = "Has flying as long as you control an artifact."
	effects.Register(effects.Spec{
		OracleID:     "27f5969e-d613-4cbc-87af-98899bb5c58a",
		Name:         "Aerial Engineer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +2/+0 as long as you control an artifact.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileArtifact,
				Apply:     effects.PumpBy(2, 0),
			},
			flying,
		},
	})
}
