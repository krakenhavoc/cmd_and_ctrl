package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID: "15e5136e-ed15-49a8-b027-e09436673fb4",
		Name:     "Culling Ritual",
		// ADR 0126 §6: only mana value 2 or less.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{
				Match:    effects.And(effects.Nonland(), effects.ManaValueLE(2)),
				ThenTail: ctx.Tail(cullingRitualMana),
			}.Apply(ctx)
		},
	})
}

// cullingRitualMana is "add {B} or {G} for each permanent destroyed this
// way", keyed (#4253): a table waiting on a leg's prompt is a restore
// point. An on-disk key: never renamed, never reused.
var cullingRitualMana = effects.MassTail("mass/culling-ritual-mana", func(ctx *effects.Context, _ []game.Card, destroyed int) error {
	if destroyed <= 0 {
		return nil
	}
	return effects.AddMana{Player: ctx.Controller(), Amount: effects.ManaN(destroyed, "B", "G")}.Apply(ctx)
})
