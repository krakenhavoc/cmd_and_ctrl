package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kura, the Boundless Sky — {3}{G}{G} Legendary Creature — Dragon Spirit
// 4/4 (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"Flying, deathtouch
//	 When Kura dies, choose one —
//	 • Search your library for up to three land cards, reveal them, put
//	   them into your hand, then shuffle.
//	 • Create an X/X green Spirit creature token, where X is the number of
//	   lands you control."
//
// The mode is chosen as the ability goes on the stack (CR 700.2); X is read
// as the ability resolves, counting the lands Kura's controller has then
// (the token's size is set as it is created, CR 107.3 does not apply: it
// is not an {X} cost).
//
// No simplification.
func init() {
	dies := effects.WhenThisDies("Kura, the Boundless Sky — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	dies.Modes = effects.ChooseOne(
		effects.ModeWithPurpose(effects.ModeDoing("Search your library for up to three land cards, reveal them, put them into your hand, then shuffle.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsLand() },
					Dest:      game.ZoneHand,
					Limit:     3,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Kura, the Boundless Sky — up to three land cards, to your hand",
				}.Apply(ctx)
			}), game.Purpose{Tutors: 3}),
		effects.ModeWithPurpose(effects.ModeDoing("Create an X/X green Spirit creature token, where X is the number of lands you control.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				x := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsLand() {
						x++
					}
				}
				tok := ctx.TextToken(effects.TokenCard("0/0 green Spirit"))
				tok.Power, tok.Toughness = x, x
				return ctx.Game.CreateTokenForEffect(item.Controller, tok, 1)
			}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:        "6426e73a-7e11-48ab-ac5c-15647c6cf9ba",
		Name:            "Kura, the Boundless Sky",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered:       []game.TriggeredAbility{dies},
	})
}
