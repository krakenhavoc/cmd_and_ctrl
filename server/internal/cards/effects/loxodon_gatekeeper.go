package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Gatekeeper — Creature — Elephant Soldier {2}{W}{W}, 2/3 (Ravnica:
// City of Guilds, #5600):
//
//	"Artifacts, creatures, and lands your opponents control enter tapped."
//
// Kismet's CR 614 replacement, shared with it: an artifact, creature or
// land entering under an opponent's control enters tapped. Tokens run the
// same entry pipeline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "861f0b1c-3b37-4d38-9232-e42565247076",
		Name:         "Loxodon Gatekeeper",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{opponentsArtifactsCreaturesAndLandsEnterTapped("Loxodon Gatekeeper")},
	})
}
