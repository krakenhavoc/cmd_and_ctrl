package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sun-Crested Pterodon — Creature — Dinosaur {4}{W}, 2/5 (Rivals of
// Ixalan, #5531):
//
//	"Flying
//	 This creature has vigilance as long as you control another Dinosaur."
//
// Flying is printed data; the vigilance is Stampeding Horncrest's
// "another Dinosaur" grant, read through the layers on every pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "747dbf0c-dcea-49d1-ab6c-d74016d580a4",
		Name:            "Sun-Crested Pterodon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.BnsB29SelfWhileAnother("Dinosaur"), "vigilance"),
		},
	})
}
