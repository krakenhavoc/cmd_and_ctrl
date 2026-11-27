package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Misthollow Griffin — {2}{U}{U} Creature — Griffin 3/3: "Flying. You may cast this
// card from exile."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:             "a990e744-bfc4-435d-aef6-6256ac10f5a6",
		Name:                 "Misthollow Griffin",
		Completeness:         effects.CompletenessFull,
		PrintedKeywords:      []string{"flying"},
		CastableZones:        []game.ZoneKind{game.ZoneExile},
		CastsItselfFromExile: true,
	})
}
