package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "7677c7ce-1db3-41f5-8f86-e68718ea2d11",
		Name:            "Shanna, Purifying Blade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Shanna, Purifying Blade — you may pay {X}; draw X cards (X at most the life you gained this turn)", shannaPay),
		},
	})
}

func shannaPay(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	gained := b15LifeGainedThisTurn(g, ctx.Controller())
	if gained <= 0 {
		return nil
	}
	return MayPayX{
		Symbol: "{1}",
		Max:    gained,
		Label:  "Shanna, Purifying Blade",
		Buys:   "draw X cards",
		Goal:   XAsHighAsYouCan,
		Unit:   game.PayAmountCards,
		Then:   shannaPaid,
	}.Apply(ctx)
}

var shannaPaid = PayUnlessBranches("pay-unless/shanna-draw",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return DrawCards{N: r.Number}.Apply(ctx)
	}, nil)
