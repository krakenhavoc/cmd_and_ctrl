package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seal of Doom — Enchantment {2}{B}
// (Nemesis, #5618):
//
//	"Sacrifice this enchantment: Destroy target nonblack creature. It can't be regenerated."
//
// Seal of Removal's shape on Snuff Out's body: the sacrifice is a cost,
// so the Seal is gone before anyone can respond, and the destroyed
// creature can't be regenerated (CR 701.19c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71d54ba3-91fe-4a14-8b99-c284b4c75013",
		Name:         "Seal of Doom",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Destroy target nonblack creature. It can't be regenerated.",
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target nonblack creature", NonBlack()),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return destroyTheTargetPermanentNoRegen(item, NewContext(g, item))
			},
		}},
	})
}
