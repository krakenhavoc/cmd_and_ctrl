package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prophesied End — Instant {1}{W} (Reality Fracture, tracker #2795):
//
//	"Destroy target creature. If it wasn't attacking, its controller
//	 draws a card."
//
// "Wasn't attacking" is read as the spell begins to resolve, before the
// destroy removes the creature from combat, and the controller is read
// at the same moment, so the draw goes to whoever controlled it even
// though the creature is gone. An indestructible creature survives and
// its controller still draws: the draw is not conditional on the
// destruction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45f4d057-5134-40a2-9590-ddba73a65582",
		Name:         "Prophesied End",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				wasAttacking := c.AttackingTarget != uuid.Nil
				owner := c.Controller
				// The draw is not conditional on the destruction, so it
				// rides the destroy's continuation rather than the line
				// after it: a replacement that pauses or redirects the
				// destruction cannot make it draw early.
				return ctx.Game.DestroyPermanentsThenForEffect([]uuid.UUID{t.ID}, func(g *game.Game, _ []uuid.UUID) error {
					if wasAttacking {
						return nil
					}
					return DrawCards{Player: owner, N: 1}.Apply(NewContext(g, item))
				})
			}
			return nil
		},
	})
}
