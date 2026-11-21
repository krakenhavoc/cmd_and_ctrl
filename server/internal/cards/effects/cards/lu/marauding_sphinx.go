package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marauding Sphinx — Creature — Sphinx Rogue {3}{U}{U}, 3/5 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flying, vigilance, ward {2}
//	 Whenever you commit a crime, surveil 2. This ability triggers only
//	 once each turn. (Targeting opponents, anything they control, and/or
//	 cards in their graveyards is a crime.)"
//
// Ward {2} is the ward trigger (CR 702.21a). The crime trigger is Deepmuck
// Desperado's: the once-per-announcement crime (CR 700.13), once each
// turn, then surveil 2 (CR 701.25a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "70bea459-f6e9-411a-bd33-339bf95ea47a",
		Name:            "Marauding Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Marauding Sphinx — ward {2}"),
			effects.WheneverYouCommitACrimeOnceEachTurn("Marauding Sphinx — surveil 2 (crime)", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
