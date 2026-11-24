package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Make Your Mark — Instant {R/W} (Strixhaven: School of Mages, #5502):
//
//	"Target creature gets +1/+0 until end of turn. When that creature dies this
//	 turn, create a 3/2 red and white Spirit creature token."
//
// Blessed Defiance's shape: the pump, then a delayed triggered ability bound
// to that one object (CR 603.7a) and controlled by the spell's controller
// (CR 603.7d), so the Spirit is the caster's whoever owned the creature. The
// target may be any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c7ca399-7d85-452b-97c0-5d7f6c0eeb12",
		Name:         "Make Your Mark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Make Your Mark — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.WhenThatCreatureDiesThisTurn(ctx, t.ID, "Make Your Mark — create a 3/2 red and white Spirit", makeYourMarkSpirit)
		},
	})
}

// makeYourMarkSpirit is the delayed body: a 3/2 red and white Spirit for the
// controller of the delayed trigger. An on-disk key.
var makeYourMarkSpirit = game.SimpleDelayedBody("s82-old-b046/make-your-mark-spirit", func(g *game.Game, item *game.StackItem) error {
	return effects.CreateToken{Template: effects.TokenCard("3/2 red and white Spirit"), N: 1}.Apply(effects.NewContext(g, item))
})
