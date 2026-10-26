package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghostly Flicker — Instant {2}{U}:
//
//	"Exile two target artifacts, creatures, and/or lands you control,
//	 then return those cards to the battlefield under your control."
//
// Both cards leave together and come back as one entry, so each sees
// the other enter (CR 603.6a), and each returns as a new object
// (CR 400.7): fresh summoning sickness, no counters, every ETB
// re-triggered. Both targets are required; a target that left in
// response is skipped (CR 608.2b), and a token exiled this way
// ceases to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad0070db-6454-41cb-861f-8f5b8fc2a3b8",
		Name:         "Ghostly Flicker",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("two target artifacts, creatures, and/or lands you control",
			Or(Artifact(), Creature(), Land()), YouControl()).WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard && t.ID != uuid.Nil {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(ghostlyFlickerReturn))
		},
	})
}

// ghostlyFlickerReturn returns what reached exile under your control (#4253 slice 9).
var ghostlyFlickerReturn = CardTail("ghostly-flicker/return-under-your-control", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled, controller := ctx.Game, ctx.Item, o.IDs, ctx.Controller()

	return ReturnFromExileTogether{Targets: exiled, Controller: controller}.Apply(NewContext(g, item))
})
