package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treva, the Renewer — Legendary Creature — Dragon {3}{G}{W}{U}, 6/6 (Invasion, S82 #5616):
//
//	"Flying
//	 Whenever Treva deals combat damage to a player, you may pay {2}{W}. If you do,
//	 choose a color, then you gain 1 life for each permanent of that color."
//
// Rith, the Awakener's shape: every permanent of the chosen colour counts,
// whoever controls it, counted as the tail runs after the colour is named.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "72524f3c-674b-4ebf-9526-18737b354010",
		Name:            "Treva, the Renewer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Treva, the Renewer — you may pay {2}{W}; if you do, choose a color and gain 1 life for each permanent of that color",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{W}",
						Question: "Treva, the Renewer — pay {2}{W}?",
						Then:     trevaPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var trevaPaid = effects.PayUnlessBranches("pay-unless/treva-the-renewer-life",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		effects.ChooseColorTail(game.ColorForBenefit, ctx.Game, ctx.Controller(), ctx.Source(),
			"Treva, the Renewer — choose a color", ctx.ItemTail(trevaLife))
		return nil
	}, nil)

var trevaLife = effects.CardTail("then/treva-the-renewer-life", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.HasColor(color) {
			n++
		}
	}
	return effects.GainLife{Amount: n}.Apply(ctx)
})
