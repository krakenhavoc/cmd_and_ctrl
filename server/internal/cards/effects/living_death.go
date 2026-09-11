package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Living Death — Sorcery {3}{B}{B} (EDHREC rank 460):
//
//	"Each player exiles all creature cards from their graveyard, then
//	 sacrifices all creatures they control, then puts all cards they
//	 exiled this way onto the battlefield."
//
// The reanimator deck's board swap: everything in every graveyard
// trades places with everything on every battlefield. Three passes,
// each over a snapshot taken before anything moves, in the printed
// order across the whole table:
//
//  1. every seat's graveyard creature cards → exile (ExileTarget);
//  2. every creature on the battlefield → sacrificed (SacrificePermanent
//     — a sacrifice, not a destroy, so indestructible doesn't save it
//     and "whenever you sacrifice" payoffs fire);
//  3. the cards exiled in step 1 → the battlefield under their
//     OWNER's control (ReturnFromExile with Controller zero).
//
// The creatures sacrificed in step 2 land in graveyards and stay
// there — they were not exiled "this way" — which is the whole
// trick of the card. Everything that comes back is a new object
// (fresh InstanceID, ETB triggers fire), as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "9e6a3df4-67a3-452e-a6ef-f04dbadb21ef",
		Name:     "Living Death",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var exiled []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.IsCreature() {
						exiled = append(exiled, c.InstanceID)
					}
				}
			}
			for _, id := range exiled {
				if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, id := range ctx.CreatureIDs() {
				if err := (SacrificePermanent{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, id := range exiled {
				if err := (ReturnFromExile{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
