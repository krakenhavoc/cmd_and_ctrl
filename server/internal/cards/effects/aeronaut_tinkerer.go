package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	flying := KeywordGrant(whileYouHave(cloudsculptControlsAnArtifact, selfCreature), "flying")
	flying.Label = "This creature has flying as long as you control an artifact."
	Register(Spec{
		OracleID:     "a15bb807-ea31-4bff-b573-c249c4c78dd6",
		Name:         "Aeronaut Tinkerer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}
