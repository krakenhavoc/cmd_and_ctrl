package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living Destiny — Instant {3}{G} (Rise of the Eldrazi, S82 OLD batch 481,
// #5581):
//
//	"As an additional cost to cast this spell, reveal a creature card from
//	 your hand.
//	 You gain life equal to the revealed card's mana value."
//
// Induce Despair's cost: the revealed card is chosen as the spell is cast
// and its mana value is read from that choice as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba7be3e9-d240-4fed-943d-a405409b13b5",
		Name:         "Living Destiny",
		Completeness: effects.CompletenessFull,
		AdditionalCost: &game.AdditionalCost{
			Reveal: &game.RevealCost{CardType: "creature", Chosen: true},
			Label:  "Reveal a creature card from your hand",
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x, ok := ctx.Game.ChosenManaValueForEffect(item)
			if !ok {
				return nil
			}
			return effects.GainLife{Amount: x}.Apply(ctx)
		},
	})
}
