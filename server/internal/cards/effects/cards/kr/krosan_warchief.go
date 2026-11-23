package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Warchief — Creature — Beast {2}{G}, 2/2 (Scourge):
//
//	"Beast spells you cast cost {1} less to cast.
//	 {1}{G}: Regenerate target Beast."
//
// Undead Warchief's discount for Beasts (one cost modifier, CR 601.2f,
// generic mana only) and Boneknitter's regenerate-a-type ability (CR
// 701.19a). The target is any Beast permanent, any player's, the Warchief
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88350861-244c-4634-9bc9-e9a79401bd02",
		Name:         "Krosan Warchief",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Beast spells you cast cost {1} less to cast.", effects.YourSpell(), effects.SpellOfCreatureType("Beast")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Regenerate target Beast.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetPermanent("target Beast", effects.HasSubtype("Beast")),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
