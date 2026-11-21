package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Village Rites — Instant {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards."
//
// One mana for two cards is a rate no other black instant matches,
// and the "drawback" is the point: aristocrats decks want a free
// sacrifice outlet, and this one is instant-speed and cantrips twice
// over. It is also the standard answer to targeted removal — sacrifice
// the creature in response and the removal spell fizzles.
//
// The sacrifice is a COST, which is what makes the card play the way
// it does: the creature dies with Village Rites still on the stack, so
// a Blood Artist trigger drains before the cards are drawn, and
// countering the spell does not hand the creature back.
//
// No simplification. Reviewed for the Shirei deck request (#3173).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "365548fb-5acc-4a8a-b20b-26d28b7d029f",
		Name:           "Village Rites",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 2},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
