package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Setback — Instant {2}{U}{U} (Murders at Karlov Manor, #4469):
//
//	"The owner of target spell or nonland permanent puts it on their
//	 choice of the top or bottom of their library."
//
// Unlucky Drop's placement (the choice is the OWNER's, so the
// top-or-bottom prompt goes to the owner) over a target that may be a spell
// or a nonland permanent (TargetSpellOrPermanent). A permanent is tucked
// from the battlefield; a spell leaves the stack for the library WITHOUT
// being countered (Game.PutSpellInLibraryKeyedForEffect): a spell that
// can't be countered still moves. A target that has gone makes the spell do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "565ba85c-bad8-4cbd-a2ff-b4384b886308",
		Name:         "Sudden Setback",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrPermanent("target spell or nonland permanent", nil, Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var id uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					id = t.ID
					break
				}
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || c.Owner == uuid.Nil {
				return nil
			}
			from := game.ZoneBattlefield
			if ctx.Game.StackItemForEffect(id) != nil {
				from = game.ZoneStack
			}
			return PutInLibraryInAnyOrder{
				Chooser:   c.Owner,
				Cards:     []uuid.UUID{id},
				From:      from,
				Placement: game.LibraryPlaceTopOrBottom,
				Label:     "Sudden Setback — put it on your choice of the top or bottom of your library",
			}.Apply(ctx)
		},
	})
}
