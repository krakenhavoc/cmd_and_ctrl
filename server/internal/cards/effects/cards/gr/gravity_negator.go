package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gravity Negator — {3}{U} Creature — Eldrazi Drone 2/3 (Oath of the
// Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 Whenever this creature attacks, you may pay {C}. If you do, another
//	 target creature gains flying until end of turn."
//
// The attack trigger aims at another creature as it goes on the stack (CR
// 603.3d); the {C} is asked as it resolves (CR 118.12) and only colorless mana
// pays it. Paid, the target — if still legal (CR 608.2b) — gains flying.
//
// No simplification.
func init() {
	const label = "Gravity Negator — you may pay {C}; if you do, another target creature gains flying until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "18dcf3db-d6c6-4780-a931-546e71a3ae3f",
		Name:            "Gravity Negator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{C}",
					Question: "Gravity Negator — pay {C} to give another target creature flying until end of turn?",
					Then:     gravityNegatorPaid,
				}.Apply(effects.NewContext(g, item))
			}), effects.Another(effects.TargetCreature("another target creature"))),
		},
	})
}

// gravityNegatorPaid is the paid branch: the target gains flying. An on-disk
// key: never renamed, never reused.
var gravityNegatorPaid = effects.PayUnlessBranches("pay-unless/gravity-negator-flying",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Gravity Negator — flying"}.Apply(ctx)
			}
		}
		return nil
	}, nil)
