package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Lifechanter — Creature — Elephant Cleric {5}{W}, 4/6 (Core Set 2020):
//
//	"When this creature enters, you may have your life total become the total
//	 toughness of creatures you control.
//	 {5}{W}: This creature gets +X/+X until end of turn, where X is your life
//	 total."
//
// The enters trigger is Magister Sphinx's "becomes" (CR 119.5): asked as it
// resolves, it gains or loses the difference between your life total and the
// total toughness of the creatures you control then (the Lifechanter itself
// among them, with its counters and any anthem). The pump reads your life
// total as the ability resolves and applies it to this permanent, only if it
// is still the same object (CR 400.7).
//
// No simplification.
func init() {
	const name = "Loxodon Lifechanter"
	effects.Register(effects.Spec{
		OracleID:     "24adb147-95b2-4854-86a5-17d43f904601",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters(name+" — your life total becomes the total toughness of creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.B31LifeBecomes(ctx, ctx.Controller(), effects.B30TotalToughnessControlled(g, ctx.Controller()))
				}),
				name+" — have your life total become the total toughness of creatures you control?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{W}: This creature gets +X/+X until end of turn, where X is your life total.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				p := ctx.PlayerByID(ctx.Controller())
				if p == nil {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: p.Life, Toughness: p.Life, Label: name + " — +X/+X"}.Apply(ctx)
			},
		}},
	})
}
