package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charix, the Raging Isle — Legendary Creature — Leviathan Crab {2}{U}{U},
// 0/17 (Secret Lair Drop):
//
//	"Spells your opponents cast that target Charix cost {2} more to cast.
//	 {3}: Charix gets +X/-X until end of turn, where X is the number of
//	 Islands you control."
//
// The tax is Kopala's target-reading cost modifier, aimed at this creature
// itself and limited to spells (as printed: not abilities). X is counted as
// the pump resolves, over the Islands its controller controls, and a Charix
// with no Island gets nothing.
//
// No simplification.
func init() {
	spells := effects.CostsMore(2, "Spells your opponents cast that target Charix cost {2} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	spells.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:      "c38cc5fe-936a-4d61-a816-9be773ae1d02",
		Name:          "Charix, the Raging Isle",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{spells},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Charix gets +X/-X until end of turn, where X is the number of Islands you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.S82OldB022PlusXMinusXForIslands,
		}},
	})
}
