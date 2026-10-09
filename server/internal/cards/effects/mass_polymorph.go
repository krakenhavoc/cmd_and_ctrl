package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b133129d-1ffb-4b78-9f5b-852d163cc9b7",
		Name:         "Mass Polymorph",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ExileAllMatching{
				Match: And(Creature(), YouControl()),
				Then: func(ctx *Context, _ []game.Card, exiled int) error {
					return rfReprintAMassPolymorphReveal(ctx, exiled)
				},
			}.Apply(ctx)
		},
	})
}
