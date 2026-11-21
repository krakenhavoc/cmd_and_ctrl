package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shanna, Purifying Blade — Legendary Creature — Human Warrior {G}{W}{U},
// 3/3 (Dominaria United):
//
//	"Lifelink
//	 At the beginning of your end step, you may pay {X}. If you do, draw X
//	 cards. X can't be greater than the amount of life you gained this turn."
//
// The trigger is MayPayX with Symbol "{1}" and a cap: X is chosen as the
// trigger resolves (CR 608.2d), from 0 to the least of the mana the
// controller can produce and the life they gained this turn, read at
// resolution. A turn with no life gained asks nothing, since X can only be
// 0. The draw is the keyed paid half, so a restore point holds it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7677c7ce-1db3-41f5-8f86-e68718ea2d11",
		Name:            "Shanna, Purifying Blade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Shanna, Purifying Blade — you may pay {X}; draw X cards (X at most the life you gained this turn)", shannaPay),
		},
	})
}

func shannaPay(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	gained := effects.B15LifeGainedThisTurn(g, ctx.Controller())
	if gained <= 0 {
		return nil
	}
	return effects.MayPayX{
		Symbol: "{1}",
		Max:    gained,
		Label:  "Shanna, Purifying Blade",
		Buys:   "draw X cards",
		Goal:   effects.XAsHighAsYouCan,
		Unit:   game.PayAmountCards,
		Then:   shannaPaid,
	}.Apply(ctx)
}

var shannaPaid = effects.PayUnlessBranches("pay-unless/shanna-draw",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.DrawCards{N: r.Number}.Apply(ctx)
	}, nil)
