package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stolen Strategy — Enchantment {4}{R} (Battlebond, #5529):
//
//	"At the beginning of your upkeep, exile the top card of each
//	 opponent's library. Until end of turn, you may cast spells from among
//	 those exiled cards, and you may spend mana as though it were mana of
//	 any color to cast those spells."
//
// Brainstealer Dragon's exile with a one-turn window: each opponent's top
// card in turn, face up, through the impulse exile (ExileTopWithPermission),
// CastOnly ("cast spells", so a land among them is stranded), AnyColor
// ("as though it were mana of any color", colour only, not type), and the
// default window, until end of turn. The permission is the player's, so it
// outlives the enchantment. Timing rules still apply to each cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6807167a-e290-4259-940e-3cbbfa75d0a1",
		Name:         "Stolen Strategy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Stolen Strategy — exile the top card of each opponent's library; you may cast spells from among them this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return exileTopEachPlayerInTurn(ctx, ctx.Opponents(),
						ExileTopWithPermission{GrantTo: item.Controller, N: 1, CastOnly: true, AnyColor: true}, game.KeyedTail{})
				}),
		},
	})
}
