package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlegrace Angel — Creature — Angel {3}{W}{W}, 4/4 (Shards of Alara,
// #5588):
//
//	"Flying
//	 Exalted
//	 Whenever a creature you control attacks alone, it gains lifelink
//	 until end of turn."
//
// Rafiq of the Many's row with lifelink (CR 702.15b): the shared
// alaB1LoneAttackerGains trigger beside the exalted keyword (CR 702.83a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "40216d09-ae4d-4a41-b3f6-1a8035093045",
		Name:            "Battlegrace Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordExalted},
		Triggered: []game.TriggeredAbility{
			effects.AlaB1LoneAttackerGains("Battlegrace Angel", "lifelink"),
		},
	})
}
