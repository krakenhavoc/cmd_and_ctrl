package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodhill Bastion — Plane — Equilor (#4870):
//
//	"Whenever a creature enters, it gains double strike and haste until
//	 end of turn.
//	 Whenever chaos ensues, exile target nontoken creature you control,
//	 then return it to the battlefield under your control."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability triggers for a creature under any player's control and
// gives it the keywords while it is still the object that entered
// (CR 400.7). The chaos ability (CR 311.7) blinks the target and it returns
// under the planar controller's control, so a creature the planar
// controller only borrowed stays theirs (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce8139b9-31f3-464d-86ea-7efbdbfce478",
		Name:         "Bloodhill Bastion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEntered, "Bloodhill Bastion — it gains double strike and haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					return GrantKeywordUntilEOT{Target: ctx.Trigger().Object.Ref().ID, Keywords: []string{"double strike", "haste"},
						Label: "Bloodhill Bastion — double strike and haste until end of turn"}.Apply(ctx)
				}),
			Targeting(WheneverChaosEnsues("Bloodhill Bastion — exile target nontoken creature you control, then return it under your control",
				flickerFirstLegalTarget),
				TargetCreature("target nontoken creature you control", YouControl(), Not(IsTokenPredicate()))),
		},
	})
}
