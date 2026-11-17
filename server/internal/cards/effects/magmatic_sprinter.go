package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmatic Sprinter — Creature — Phyrexian Warrior {2}{R}, 3/2 (Phyrexia: All Will Be One):
//
//		"Haste
//		 When this creature enters, put two oil counters on target artifact or creature
//		 you control.
//		 At the beginning of your end step, return this creature to its owner's hand
//		 unless you remove two oil counters from it."
//
//	  - The enters trigger targets an artifact or creature its controller controls (the
//	    Sprinter itself is one), and the counters land only if that target is still legal.
//	  - The end-step trigger is "unless you remove": the controller is asked only when the
//	    Sprinter has two oil counters to remove (CR 118.3 — a cost can't be paid without
//	    the means); with fewer it returns at once. Saying no returns it. Both branches are
//	    read as the Sprinter stands when the trigger resolves, and a Sprinter that has left
//	    the battlefield is not returned (CR 400.7).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Magmatic Sprinter — put two oil counters on target artifact or creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return AddCounter{Target: id, Kind: s82OneB09Oil, N: 2}.Apply(ctx)
		})
	enters.Targets = TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl())
	Register(Spec{
		OracleID:        "85d025d9-158d-4bf1-b3b2-361f3fe59663",
		Name:            "Magmatic Sprinter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			enters,
			AtYourEndStep("Magmatic Sprinter — return it to its owner's hand unless you remove two oil counters from it", magmaticSprinterEndStep),
		},
	})
}

func magmaticSprinterEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src := ctx.Source()
	if !onBattlefield(g, src) {
		return nil
	}
	c, ok := g.LookupCardForEffect(src)
	if !ok {
		return nil
	}
	if s82OneB09OilCount(&c) < 2 {
		return BounceToHand{Target: src}.Apply(ctx)
	}
	return UpkeepPayUnless{
		Chooser:  ctx.Controller(),
		Cost:     "Remove two oil counters from it",
		Action:   &game.PayAction{Kind: game.PayActionNothing},
		Question: "Magmatic Sprinter — remove two oil counters, or return it to its owner's hand?",
		Else:     magmaticSprinterBranches,
	}.Apply(ctx)
}

var magmaticSprinterBranches = PayUnlessBranches("pay-unless/magmatic-sprinter-oil",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return AddCounter{Target: ctx.Source(), Kind: s82OneB09Oil, N: -2}.Apply(ctx)
	},
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if !onBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return BounceToHand{Target: ctx.Source()}.Apply(ctx)
	})
