package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exploding Borders — Sorcery {2}{R}{G} (Conflux):
//
//	"Domain — Search your library for a basic land card, put that card onto
//	 the battlefield tapped, then shuffle. Exploding Borders deals X damage to
//	 target player or planeswalker, where X is the number of basic land types
//	 among lands you control."
//
// The search is Rampant Growth's. The damage is the tail of the search
// (so it waits for the caster's pick and for the land to be on the
// battlefield) and X is the domain count (DomainCount, #5666) read then, the
// fetched land included. With the target illegal as the spell resolves the
// whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a19fc9e-a822-49ca-ade0-d58054f3caf6",
		Name:         "Exploding Borders",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				TappedOnEntry: true,
				Limit:         1,
				Shuffle:       true,
				Reason:        "Exploding Borders — a basic land card, onto the battlefield tapped",
				ThenTail:      ctx.ItemTail(explodingBordersDamage),
			}.Apply(ctx)
		},
	})
}

// explodingBordersDamage is the damage once the land is in: X is the number of
// basic land types among the caster's lands now. An on-disk key: never renamed,
// never reused.
var explodingBordersDamage = effects.CardTail("tail/exploding-borders-damage", func(ctx *effects.Context, _ game.TailOutcome) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	x := effects.DomainCount(ctx.Game, ctx.Controller())
	if x == 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}.Apply(ctx)
})
