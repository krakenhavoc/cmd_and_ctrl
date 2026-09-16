package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of the Swine — Sorcery {X}{U}{U} (EDHREC rank 839):
//
//	"Exile X target creatures. For each creature exiled this way, its
//	 controller creates a 2/2 green Boar creature token."
//
// Blue's scalable exile-based removal — Pongify for X creatures, and
// the token goes to each creature's controller, so pointing it at your
// own board is a legal (if odd) play.
//
// The target count is the announced X (Waterbender's Restoration's
// targetsCountedByX), so casting for X=3 demands exactly three
// targets at announce. At resolution every still-legal target is
// exiled first and the Boars are created after, in printed order;
// the controllers are read BEFORE the exile, because afterwards the
// card is in exile and its controller field is stale for a creature
// that had changed hands (Beast Within's note).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5669ea7c-c4fc-494c-896b-4bce9b494817",
		Name:         "Curse of the Swine",
		Completeness: CompletenessFull,
		Targets:      targetsCountedByX(TargetCreature("X target creatures")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			type victim struct {
				id         uuid.UUID
				controller uuid.UUID
			}
			var victims []victim
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				controller, ok := controllerOfTarget(ctx, t.ID)
				if !ok {
					continue
				}
				victims = append(victims, victim{t.ID, controller})
			}
			for _, v := range victims {
				if err := (ExileTarget{Target: v.id}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, v := range victims {
				if err := (CreateToken{Controller: v.controller, Template: TokenCard("2/2 green Boar"), N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
