package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Woman, Stunning Savior — Legendary Creature — Spider Human Hero
// {1}{W/U}, 2/2 (Marvel's Spider-Man promos, #4460):
//
//	"Flying
//	 Venom Blast — Artifacts and creatures your opponents control enter
//	 tapped."
//
// Venom Blast is Blind Obedience's CR 614 replacement: an artifact or a
// creature entering under an opponent's control enters tapped, and one
// entering under Spider-Woman's controller's control does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be2b9c6d-4ecb-49ec-b276-4aa93c5dfc00",
		Name:            "Spider-Woman, Stunning Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{opponentsArtifactsAndCreaturesEnterTapped("Spider-Woman, Stunning Savior — Venom Blast")},
	})
}
