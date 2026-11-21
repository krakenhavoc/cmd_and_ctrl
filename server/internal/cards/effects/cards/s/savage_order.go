package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Order — Sorcery {2}{G}{G} (Jurassic World Collection, #5948):
//
//	"As an additional cost to cast this spell, sacrifice a creature with
//	 power 4 or greater.
//	 Search your library for a Dinosaur creature card, put it onto the
//	 battlefield, then shuffle. It gains indestructible until your next
//	 turn."
//
// The sacrifice is the spell's additional cost, a creature judged through
// the layered view. The search is a battlefield search; the rest of the
// sentence is the search's keyed tail, told the card actually found (nothing
// found, nothing gains it). Indestructible is granted to that object until
// the beginning of the caster's next turn (CR 611.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "8dd5b7d5-3d92-40e2-9572-bf70ae591bed",
		Name:           "Savage Order",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Tutors: 1},
		AdditionalCost: effects.SacrificeCost("a creature with power 4 or greater", effects.Creature(), effects.PowerGE(4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			dinosaur := ctx.TextSubtype("Dinosaur")
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(dinosaur) },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Savage Order — a Dinosaur creature card",
				ThenTail:  ctx.Tail(effects.GapB82SavageOrderTail),
			}.Apply(ctx)
		},
	})
}
