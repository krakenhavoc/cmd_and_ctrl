package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "aba04455-c13f-48f7-9674-56803b39a42b",
		Name:         "Neyith of the Dire Hunt",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYourCreaturesFightOrBecomeBlocked("Neyith of the Dire Hunt — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.Targeting(effects.AtBeginningOfYourCombat("Neyith of the Dire Hunt — pay {2}{R/G} to double target creature's power", neyithCombat),
				effects.TargetCreature("target creature")),
		},
	})
}

// The target is announced before opponents respond. Payment is optional
// during resolution; both the turn-long pump and combat-long requirement
// depend on paying, and the saved prompt names the same target object.
func neyithCombat(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(id)
	if !ok {
		return nil
	}
	return effects.MayPay{Chooser: ctx.Controller(), Cost: "{2}{R/G}", InThisStep: true,
		Question: "Neyith of the Dire Hunt — pay {2}{R/G} to double target creature's power?",
		Then:     neyithPaid, CarryObjects: []game.ObjectRef{ref},
	}.Apply(ctx)
}

var neyithPaid = effects.PayUnlessBranches("pay-unless/neyith-of-the-dire-hunt-double-power",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) != 1 {
			return nil
		}
		ref := r.Objects[0]
		c, ok := ctx.Game.LookupCardForEffect(ref.ID)
		if !ok || !effects.OnBattlefield(ctx.Game, ref.ID) || c.ObjectEpoch != ref.Epoch {
			return nil
		}
		if err := effects.DoublePowerUntilEOT(ctx, ref.ID, "Neyith of the Dire Hunt — double power"); err != nil {
			return err
		}
		d, ok := effects.DurationUntilEndOfCombat(ctx)
		if !ok {
			return nil
		}
		return effects.ScopedEffectFor{Target: ref.ID, Duration: d,
			Mods:  []game.Mod{game.AddBlockRequirementMod(game.BlockRequirementMustBeBlocked)},
			Label: "Neyith of the Dire Hunt — must be blocked this combat if able",
		}.Apply(ctx)
	}, nil)
