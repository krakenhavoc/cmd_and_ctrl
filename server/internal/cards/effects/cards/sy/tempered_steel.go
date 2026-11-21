package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempered Steel — Enchantment {1}{W}{W} (EDHREC rank 3760):
//
//	"Artifact creatures you control get +2/+2."
//
// The artifact-aggro anthem: a layer 7c static over every creature
// the controller controls that is also an artifact, read post-layer
// so a Thopter token, a crewed Vehicle and an artifact something
// else animated all count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e220138c-5fc6-487d-9fa9-f21b2fb1f12c",
		Name:         "Tempered Steel",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.B36ArtifactCreatureYouControl, 2, 2)},
	})
}
