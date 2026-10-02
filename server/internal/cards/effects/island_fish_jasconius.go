package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:              "bb217f12-532f-4833-a27a-99e290aa47d0",
		Name:                  "Island Fish Jasconius",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Island Fish Jasconius — you may pay {U}{U}{U} to untap it",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{U}{U}{U}",
						Question: "Island Fish Jasconius — pay {U}{U}{U} to untap it?",
						OnPay: func(ctx *Context) error {
							if !onBattlefield(ctx.Game, ctx.Item.SourceCardID) {
								return nil
							}
							return UntapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				}),
			WhenYouControlNo(q, "Island Fish Jasconius — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
