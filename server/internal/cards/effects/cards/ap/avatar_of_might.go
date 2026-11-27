package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avatar of Might — Creature — Avatar {6}{G}{G}, 8/8 (Prophecy, S82 OLD b670,
// #5617):
//
//	"If an opponent controls at least four more creatures than you, this spell
//	 costs {6} less to cast.
//	 Trample"
//
// Avatar of Fury's self cost reduction with the condition on creature counts:
// ONE opponent must control at least four more creatures than the caster, as
// the cost is determined (CR 601.2f), leaving {G}{G}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68d84ffa-5bd0-4ed9-a987-6cef13095438",
		Name:            "Avatar of Might",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(6, "This spell costs {6} less to cast if an opponent controls at least four more creatures than you.",
				effects.OldB670AnOpponentControlsCreaturesMoreThanYou(4)),
		},
	})
}
