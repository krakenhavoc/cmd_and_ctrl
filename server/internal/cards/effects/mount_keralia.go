package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mount Keralia — Plane — Regatha (Planechase Anthology, #4870):
//
//	"At the beginning of your end step, put a pressure counter on Mount
//	 Keralia.
//	 When you planeswalk away from Mount Keralia, it deals damage equal to
//	 the number of pressure counters on it to each creature and each
//	 planeswalker.
//	 Whenever chaos ensues, prevent all damage that planes named Mount
//	 Keralia would deal this game to permanents you control."
//
// A plane (ADR 0216 §6). The counters sit on the face-up card in the command
// zone; a planeswalk away makes it a new object with none (CR 901.7a), so the
// away trigger reads them from the planeswalk event, which tells the count the
// plane had as it left (EventPlaneswalkFrom.Amount, last-known information).
// The damage comes from the plane as it last existed. The chaos ability is a
// shield for the rest of the game (ruling), read as damage would be dealt: its
// sources are the planes named Mount Keralia whoever owns them
// (DamageSourceFilter.Names) and what it protects is the permanents its
// controller controls at that moment, "you" being the planar controller when
// the ability resolved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0dc9986-1099-4189-bfff-8f41bab9019d",
		Name:         "Mount Keralia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Mount Keralia — put a pressure counter on it",
				func(g *game.Game, item *game.StackItem) error {
					g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, mountKeraliaPressure, 1)
					return nil
				}),
			WhenYouPlaneswalkAwayFromThis("Mount Keralia — it deals damage equal to the number of pressure counters on it to each creature and each planeswalker",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := ctx.Trigger().Event.Amount
					if n <= 0 {
						return nil
					}
					return g.DamageInstanceForEffect(func() error {
						return damageEachMatching(ctx, Or(Creature(), Planeswalker()), n)
					})
				}),
			WheneverChaosEnsues("Mount Keralia — prevent all damage that planes named Mount Keralia would deal this game to permanents you control",
				Do(PreventDamageFromSource{
					Filter:  game.DamageSourceFilter{Names: []string{"Mount Keralia"}},
					Protect: ShieldPermanentsYouControl(),
					Lasts:   ShieldRestOfTheGame,
					Label:   "Mount Keralia — prevent all damage that planes named Mount Keralia would deal to permanents you control",
				})),
		},
	})
}

// mountKeraliaPressure is the pressure counter's kind.
const mountKeraliaPressure = "pressure"
