package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wicker Witch — Artifact Creature — Scarecrow {3}, 3/1
// (Shadows over Innistrad, #5543):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc2f945f-9048-4978-b61f-49eefb3d94d7",
		Name:         "Wicker Witch",
		Completeness: effects.CompletenessFull,
	})
}
