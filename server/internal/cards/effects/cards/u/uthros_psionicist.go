package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uthros Psionicist — Creature — Jellyfish Scientist {2}{U}, 2/4 (Edge of
// Eternities, #4461):
//
//	"The second spell you cast each turn costs {2} less to cast."
//
// Alisaie Leveilleur's dualcast discount: the turn's cast tally does not
// hold the spell being priced yet, so a tally of exactly one means "this
// is the second spell". The Psionicist itself counts when it was the
// first spell cast this turn, as the ruling says. It takes {2} off the
// generic part of the cost only (CR 601.2f), and the spell's mana value
// is unchanged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f51aede-fb86-4b7e-9027-2ba8be62b567",
		Name:         "Uthros Psionicist",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "The second spell you cast each turn costs {2} less to cast.",
				effects.YourSpell(), effects.SecondSpellYouCastThisTurn()),
		},
	})
}
