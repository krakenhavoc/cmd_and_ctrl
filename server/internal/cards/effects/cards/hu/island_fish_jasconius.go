package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Island Fish Jasconius — Creature — Fish {4}{U}{U}{U}, 6/8:
//
//	"This creature doesn't untap during your untap step.
//	 At the beginning of your upkeep, you may pay {U}{U}{U}. If you do,
//	 untap this creature.
//	 This creature can't attack unless defending player controls an
//	 Island.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a);
// the sacrifice is §1's CR 603.8 state trigger. Both read one
// PermanentQuery. "Doesn't untap during your untap step" is the
// catalog's untap-step restriction (CR 502.3), and the upkeep untap is a
// "you may pay" on the trigger's resolution (MayPay).
//
// No simplification.
func init() {
	q := effects.QuerySubtype("Island")
	effects.Register(effects.Spec{
		OracleID:              "bb217f12-532f-4833-a27a-99e290aa47d0",
		Name:                  "Island Fish Jasconius",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Static: []game.StaticAbility{
			effects.CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Island Fish Jasconius — you may pay {U}{U}{U} to untap it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{U}{U}{U}",
						Question: "Island Fish Jasconius — pay {U}{U}{U} to untap it?",
						Then:     islandFishJasconiusPaid,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.WhenYouControlNo(q, "Island Fish Jasconius — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}

// islandFishJasconiusPaid is the paid half, keyed (#4264).
var islandFishJasconiusPaid = effects.PayUnlessBranches("pay-unless/island-fish-jasconius-untap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.OnBattlefield(ctx.Game, ctx.Item.SourceCardID) {
			return nil
		}
		return effects.UntapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
	}, nil)
