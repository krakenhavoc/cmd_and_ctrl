package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avatar of Fury — Creature — Avatar {6}{R}{R}, 6/6 (Commander 2011, #5574):
//
//	"If an opponent controls seven or more lands, this spell costs {6} less to
//	 cast.
//	 Flying
//	 {R}: This creature gets +1/+0 until end of turn."
//
// The reduction is a cost modifier on the spell itself (CR 601.2f): it applies
// when ONE opponent controls at least seven lands as the cost is determined,
// and takes {6} of the generic part off, leaving {R}{R}. Flying is the printed
// keyword and the activation is the usual firebreathing with no tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4854f836-2247-43c3-adbc-1d1ee5eaf120",
		Name:            "Avatar of Fury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(6, "This spell costs {6} less to cast if an opponent controls seven or more lands.",
				effects.S82OldB457AnOpponentControlsLandsAtLeast(7)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Avatar of Fury — +1/+0"),
		}},
	})
}
