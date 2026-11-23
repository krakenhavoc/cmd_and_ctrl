package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mass Polymorph — Sorcery {5}{U}:
//
//	"Exile all creatures you control, then reveal cards from the top
//	 of your library until you reveal that many creature cards. Put all
//	 creature cards revealed this way onto the battlefield, then
//	 shuffle the rest of the revealed cards into your library."
//
// The number is the count of creatures that actually reached exile (a
// commander sent to the command zone by its owner's CR 903.9 answer is
// not counted), which is why the reveal runs from ExileAllMatching's
// continuation. The reveal and the entry are rfReprintAMassPolymorphReveal.
//
// It wipes only the caster's own board, so it declares its Sweep as a
// partial creature exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b133129d-1ffb-4b78-9f5b-852d163cc9b7",
		Name:         "Mass Polymorph",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileAllMatching{Match: effects.And(effects.Creature(), effects.YouControl()), ThenTail: ctx.ItemTail(massPolymorphReveal)}.Apply(ctx)
		},
	})
}

// massPolymorphReveal is "reveal cards until that many creature cards",
// keyed (#4253 slice 10a).
var massPolymorphReveal = effects.MassTail("mass/mass-polymorph-reveal", func(ctx *effects.Context, _ []game.Card, exiled int) error {
	return effects.RfReprintAMassPolymorphReveal(ctx, exiled)
})
