package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "88350861-244c-4634-9bc9-e9a79401bd02",
		Name:         "Krosan Warchief",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Beast spells you cast cost {1} less to cast.", YourSpell(), SpellOfCreatureType("Beast")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate target Beast.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{G}"),
			Targets: TargetPermanent("target Beast", HasSubtype("Beast")),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
