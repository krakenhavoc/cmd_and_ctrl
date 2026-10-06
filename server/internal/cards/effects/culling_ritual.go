package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Culling Ritual — Sorcery {2}{B}{G}:
//
//	"Destroy each nonland permanent with mana value 2 or less. Add {B}
//	 or {G} for each permanent destroyed this way."
//
// A sweep with a payoff: DestroyAllMatching reports how many
// permanents were actually destroyed (an indestructible one is not
// counted, CR 702.12b), and each of those is its own {B}-or-{G} slot,
// so the colour is chosen per mana rather than once for the lot.
// Tokens have mana value 0 and are swept and counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "15e5136e-ed15-49a8-b027-e09436673fb4",
		Name:     "Culling Ritual",
		// ADR 0126 §6: only mana value 2 or less.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: And(Nonland(), ManaValueLE(2)),
				Then: func(ctx *Context, _ []game.Card, destroyed int) error {
					if destroyed <= 0 {
						return nil
					}
					return AddMana{Player: ctx.Controller(), Produced: strings.Repeat("{B|G}", destroyed)}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
