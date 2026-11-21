package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chief of the Foundry — Artifact Creature — Construct {3}, 2/3
// (EDHREC rank 2885):
//
//	"Other artifact creatures you control get +1/+1."
//
// The colourless artifact lord. One layer-7c anthem (b16Anthem) over
// the controller's OTHER artifact creatures — Master of Etherium's
// second half on its own — reading post-layer types, so a Thopter,
// a Construct token and an animated Karn all get it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fdfa41a-75a5-4b36-8c9e-083e26d137e2",
		Name:         "Chief of the Foundry",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.B27OtherArtifactCreaturesYouControl, 1, 1),
		},
	})
}
