package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starling, Aerial Ally — Legendary Creature — Human Hero {4}{W}, 3/4 (Marvel
// Super Heroes Commander, #4455):
//
//	"Flying
//	 When Starling enters, another target creature you control gains flying
//	 until end of turn."
//
// Hanweir Battlements' keyword-until-end-of-turn body on an enters trigger
// that excludes the source. With no other creature there is no trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13fd7ca6-c278-409f-a9c4-57f16051ffce",
		Name:            "Starling, Aerial Ally",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Starling, Aerial Ally — another target creature you control gains flying until end of turn",
				effects.EbTargetCreatureGainsUntilEOT("flying", "Starling, Aerial Ally — flying until end of turn")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
