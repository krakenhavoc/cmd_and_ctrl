package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Finch Formation — Creature — Bird Scout {2}{U}, 2/2 (Bloomburrow,
// #4467):
//
//	"Offspring {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, when this creature enters, create a 1/1 token copy of it.)
//	 Flying
//	 When this creature enters, target creature you control gains flying
//	 until end of turn."
//
// Offspring is the shared keyword (the optional cost and its token-copy
// trigger). The entry trigger is Starling, Aerial Ally's flying grant
// without "another", so the Formation may name itself. The token copy
// is a Finch Formation too, so it brings its own entry trigger. A target
// gone by resolution gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "76c2e0b4-d4fa-4e98-b363-0a76eb481304",
		Name:            "Finch Formation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Offspring("{3}")},
		Triggered: []game.TriggeredAbility{
			effects.OffspringToken("Finch Formation"),
			effects.Targeting(effects.WhenThisEnters("Finch Formation — target creature you control gains flying until end of turn",
				effects.EbTargetCreatureGainsUntilEOT("flying", "Finch Formation — flying until end of turn")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
