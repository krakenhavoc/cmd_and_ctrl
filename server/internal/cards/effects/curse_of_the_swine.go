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
// #870: "for each creature exiled this way" is the batch's
// CONTINUATION, and it counts what ARRIVED in exile (CR 400.7). A
// Boar per exile call that returned no error paid out for a leg that
// had only PAUSED on the CR 903.9 prompt — a commander's controller
// got their 2/2 while the commander was still on the battlefield, and
// a second one if they then chose the command zone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5669ea7c-c4fc-494c-896b-4bce9b494817",
		Name:         "Curse of the Swine",
		Completeness: CompletenessFull,
		Targets:      targetsCountedByX(TargetCreature("X target creatures")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controllers := map[uuid.UUID]uuid.UUID{}
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				controller, ok := controllerOfTarget(ctx, t.ID)
				if !ok {
					continue
				}
				ids = append(ids, t.ID)
				controllers[t.ID] = controller
			}
			// The context is rebuilt inside the continuation from the
			// live *Game, the contract massEffect.apply explains.
			return ctx.Game.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
				ctx := NewContext(g, item)
				for _, id := range exiled {
					if err := (CreateToken{Controller: controllers[id], Template: TokenCard("2/2 green Boar"), N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
