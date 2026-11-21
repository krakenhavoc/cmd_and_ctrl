package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aeronaut Tinkerer — Creature — Human Artificer {2}{U}, 2/3 (Magic 2015,
// #5556):
//
//	"This creature has flying as long as you control an artifact."
//
// Gravblade Heavy's conditional keyword grant: a layer-6 flying on the
// Tinkerer alone while its controller controls any artifact
// (cloudsculptControlsAnArtifact), so it is lost the moment the last one
// leaves.
//
// No simplification.
func init() {
	flying := effects.KeywordGrant(effects.WhileYouHave(effects.CloudsculptControlsAnArtifact, effects.SelfCreature), "flying")
	flying.Label = "This creature has flying as long as you control an artifact."
	effects.Register(effects.Spec{
		OracleID:     "a15bb807-ea31-4bff-b573-c249c4c78dd6",
		Name:         "Aeronaut Tinkerer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}
