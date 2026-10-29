package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:     "aba04455-c13f-48f7-9674-56803b39a42b",
		Name:         "Neyith of the Dire Hunt",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYourCreaturesFightOrBecomeBlocked("Neyith of the Dire Hunt — draw a card", Do(DrawCards{N: 1})),
			Targeting(AtBeginningOfYourCombat("Neyith of the Dire Hunt — pay {2}{R/G} to double target creature's power", neyithCombat),
				TargetCreature("target creature")),
		},
	})
}

// The target is announced before opponents respond. Payment is optional
// during resolution; both the turn-long pump and combat-long requirement
// depend on paying, and the saved prompt names the same target object.
func neyithCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(id)
	if !ok {
		return nil
	}
	return MayPay{Chooser: ctx.Controller(), Cost: "{2}{R/G}", InThisStep: true,
		Question: "Neyith of the Dire Hunt — pay {2}{R/G} to double target creature's power?",
		Then:     neyithPaid, CarryObjects: []game.ObjectRef{ref},
	}.Apply(ctx)
}

var neyithPaid = PayUnlessBranches("pay-unless/neyith-of-the-dire-hunt-double-power",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) != 1 {
			return nil
		}
		ref := r.Objects[0]
		c, ok := ctx.Game.LookupCardForEffect(ref.ID)
		if !ok || !onBattlefield(ctx.Game, ref.ID) || c.ObjectEpoch != ref.Epoch {
			return nil
		}
		if err := DoublePowerUntilEOT(ctx, ref.ID, "Neyith of the Dire Hunt — double power"); err != nil {
			return err
		}
		d, ok := DurationUntilEndOfCombat(ctx)
		if !ok {
			return nil
		}
		return ScopedEffectFor{Target: ref.ID, Duration: d,
			Mods:  []game.Mod{game.AddBlockRequirementMod(game.BlockRequirementMustBeBlocked)},
			Label: "Neyith of the Dire Hunt — must be blocked this combat if able",
		}.Apply(ctx)
	}, nil)
