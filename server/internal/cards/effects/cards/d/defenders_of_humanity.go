package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Defenders of Humanity — Enchantment {X}{2}{W} (Warhammer 40,000
// Commander):
//
//	"When this enchantment enters, create X 2/2 white Astartes Warrior
//	 creature tokens with vigilance.
//	 {X}{2}{W}, Exile this enchantment: Create X 2/2 white Astartes
//	 Warrior creature tokens with vigilance. Activate only if you control
//	 no creatures and only during your turn."
//
// The enters trigger reads the X the spell was cast with (CR 107.3m) and
// carries it on the trigger, so removing the enchantment in response
// still makes the tokens. The activated ability announces its own X
// (CR 107.3k), exiles the enchantment as a cost, and may only be
// activated on your turn while you control no creatures; both are
// checked at activation (CR 602.5b).
//
// No simplification.
func init() {
	warriors := func(ctx *effects.Context, x int) error {
		if x <= 0 {
			return nil
		}
		return effects.CreateToken{
			Controller: ctx.Controller(),
			Template:   effects.TokenCard("2/2 white Astartes Warrior with vigilance"),
			N:          x,
		}.Apply(ctx)
	}
	effects.Register(effects.Spec{
		OracleID:     "7662af07-1ebc-4545-bffd-7a2e0c513b71",
		Name:         "Defenders of Humanity",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       "Defenders of Humanity — create X Astartes Warriors",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Defenders of Humanity — create X Astartes Warriors")
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return warriors(effects.NewContext(g, item), item.Params.Amount)
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{2}{W}, Exile this enchantment: Create X 2/2 white Astartes Warrior creature tokens with vigilance. Activate only if you control no creatures and only during your turn.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{X}{2}{W}"), effects.ExileThis()),
			Condition: effects.AllConditions(effects.DuringYourTurn(), func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.CountControlled(g, controller, func(c game.Card) bool { return c.IsCreature() }) == 0
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return warriors(ctx, ctx.X())
			},
		}},
	})
}
