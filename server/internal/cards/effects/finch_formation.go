package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "76c2e0b4-d4fa-4e98-b363-0a76eb481304",
		Name:            "Finch Formation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{Offspring("{3}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Finch Formation"),
			Targeting(WhenThisEnters("Finch Formation — target creature you control gains flying until end of turn",
				ebTargetCreatureGainsUntilEOT("flying", "Finch Formation — flying until end of turn")),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
