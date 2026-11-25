package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperial Lancer — Creature — Human Knight {W}, 1/1 (Ixalan, #5532):
//
//	"This creature has double strike as long as you control a Dinosaur."
//
// Pterodon Knight's self grant with double strike: layer 6 (CR 613.1f),
// read off the board as it is, lost the moment the last Dinosaur leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0018eb9-6aae-4188-8f36-ab5fc69f5903",
		Name:         "Imperial Lancer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.Subtype("Dinosaur")), "double strike"),
		},
	})
}
