package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eldrazi Obligator — {2}{R} Creature — Eldrazi 3/1 (Oath of the Gatewatch,
// S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, you may pay {1}{C}. If you do, gain control of
//	 target creature until end of turn, untap that creature, and it gains haste
//	 until end of turn.
//	 Haste"
//
// A cast trigger (CR 601.2i) that resolves above the spell, so it works even if
// the Obligator is countered. The target is chosen as the trigger goes on the
// stack (CR 603.3d); the {1}{C} is asked as it resolves (CR 118.12) and only
// colorless mana pays the {C}. Paid, control, untap and haste are Flayer of
// Loyalties' threaten.
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Eldrazi Obligator — you may pay {1}{C}; if you do, gain control of target creature until end of turn, untap it, and it gains haste",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}{C}",
				Question: "Eldrazi Obligator — pay {1}{C} to gain control of the target creature until end of turn?",
				Then:     eldraziObligatorPaid,
			}.Apply(effects.NewContext(g, item))
		})
	cast.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:        "d09402cb-cf64-493e-a82d-244b31700710",
		Name:            "Eldrazi Obligator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "haste"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}

// eldraziObligatorPaid is the paid branch. An on-disk key: never renamed,
// never reused.
var eldraziObligatorPaid = effects.PayUnlessBranches("pay-unless/eldrazi-obligator-threaten",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard && t.ID != uuid.Nil {
				return effects.ThreatenAndGrant(ctx, t.ID, "Eldrazi Obligator", "haste")
			}
		}
		return nil
	}, nil)
