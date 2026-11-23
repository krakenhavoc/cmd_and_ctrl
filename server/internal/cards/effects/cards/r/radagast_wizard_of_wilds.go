package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radagast, Wizard of Wilds — Legendary Creature — Avatar Wizard
// {2}{G}{U}, 3/5 (Tales of Middle-earth Commander):
//
//	"Ward {1}
//	 Beasts and Birds you control have ward {1}.
//	 Whenever you cast a spell with mana value 5 or greater, choose one —
//	 • Create a 3/3 green Beast creature token.
//	 • Create a 2/2 blue Bird creature token with flying."
//
// Hexing Squelcher's granted ward (every Beast and Bird you control, read
// through the layered view); the cast trigger reads the spell's mana value
// on the stack, X included (CR 202.3e), and is Gandalf the Grey's modal
// cast trigger with token bullets.
//
// No simplification.
func init() {
	const label = "Radagast, Wizard of Wilds — whenever you cast a spell with mana value 5 or greater, choose one"
	cast := effects.WheneverYouCast(effects.ManaValueGE(5), label, func(_ *game.Game, _ *game.StackItem) error { return nil })
	cast.Modes = effects.ChooseOne(
		effects.ModeDoing("Create a 3/3 green Beast creature token.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Create a 2/2 blue Bird creature token with flying.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("2/2 blue Bird with flying"), N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "50f144f8-29eb-4af7-9a50-eded3b9c83c2",
		Name:         "Radagast, Wizard of Wilds",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Radagast, Wizard of Wilds — ward {1}"),
			effects.WardGranted(effects.WardMana("{1}"), "Radagast, Wizard of Wilds — Beasts and Birds you control have ward {1}",
				effects.TribeFilter{Tribes: []string{"Beast", "Bird"}, YoursOnly: true}.Matches),
			cast,
		},
	})
}
