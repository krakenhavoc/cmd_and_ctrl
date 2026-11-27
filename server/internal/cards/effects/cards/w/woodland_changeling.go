package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodland Changeling — Creature — Shapeshifter {1}{G}, 2/2 (Lorwyn, S82
// OLD b548, #5593):
//
//	"Changeling (This card is every creature type.)"
//
// A vanilla changeling: the keyword is the whole card (CR 702.73a, a
// characteristic-defining ability, applied in layer 4 whatever the zone), so
// Lord-like tribal effects and "if you control a Goblin" see it as every type.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "338f0ab6-f744-4e5a-a358-05e71b93758d",
		Name:            "Woodland Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
	})
}
