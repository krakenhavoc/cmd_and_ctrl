package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swat Away — Instant {2}{U}{U} (Lorwyn Eclipsed, #4458):
//
//	"This spell costs {2} less to cast if a creature is attacking you.
//	 The owner of target spell or creature puts it on their choice of the
//	 top or bottom of their library."
//
// The discount is a self cost modifier read as the cost is determined
// (CR 601.2f): a creature attacking you yourself, not your planeswalker or
// battle. The placement is the OWNER's choice, not the caster's (Dire
// Downdraft's ownerTucksFirstTargetTopOrBottom). A spell target is not
// countered: it is put into the library from the stack, so a spell that
// can't be countered still goes (CR 701.6 never applies).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79090b00-d920-4134-a838-fb434cf3ee4d",
		Name:         "Swat Away",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpellOrPermanent("target spell or creature", nil, Creature()),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if a creature is attacking you.", stdB63CreatureAttackingYou),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return stdB63OwnerTucksTargetTopOrBottom(ctx, "Swat Away — put it on your choice of the top or bottom of your library")
		},
	})
}
