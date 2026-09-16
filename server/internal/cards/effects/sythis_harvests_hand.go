package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sythis, Harvest's Hand — Legendary Enchantment Creature — Nymph,
// {G}{W}, 1/2 (EDHREC rank 982):
//
//	"Whenever you cast an enchantment spell, you gain 1 life and draw
//	 a card."
//
// The enchantress commander: Mesa Enchantress's trigger, mandatory,
// with a life on top. Fires on cast — the spell is on the stack, so
// the card is drawn before the enchantment resolves, as printed —
// and not on Sythis's own cast, since she is not on the battlefield
// to see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fc64fd6-f057-4056-9dca-47accb7ff036",
		Name:         "Sythis, Harvest's Hand",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, b08EnchantmentSpellCastByYou, "Sythis, Harvest's Hand — gain 1 life and draw a card", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GainLife{Player: item.Controller, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}
