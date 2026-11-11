package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elenda's Hierophant — Creature — Vampire Cleric {2}{W}, 1/1 (Commander,
// #5368):
//
//	"Flying
//	 Whenever you gain life, put a +1/+1 counter on this creature.
//	 When this creature dies, create X 1/1 white Vampire creature tokens
//	 with lifelink, where X is its power."
//
// The counter trigger is Ajani's Pridemate's (once per gain event). The
// dies trigger reads the Hierophant's last-known power (CR 603.10a,
// ctx.SourcePermanent serves the stored record once it has left), so its
// counters and any pump count; a power of zero or less makes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb99c675-adc5-4dad-8be8-524f5f921213",
		Name:            "Elenda's Hierophant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			LifeGainPaying(
				WheneverYouGainLife("Elenda's Hierophant — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
			WhenThisDies("Elenda's Hierophant — create X 1/1 white Vampire creature tokens with lifelink, where X is its power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || info.Power <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 white Vampire with lifelink"), N: info.Power}.Apply(ctx)
				}),
		},
	})
}
