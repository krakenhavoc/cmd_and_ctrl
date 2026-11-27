package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kinsbaile Balloonist — Creature — Kithkin Soldier {3}{W}, 2/2 (Lorwyn):
//
//	"Flying
//	 Whenever this creature attacks, you may have target creature gain flying
//	 until end of turn."
//
// Bishop of Rebirth's shape: the target is chosen as the trigger goes on the
// stack (CR 603.3d) and may be any creature, the Balloonist itself or an
// opponent's; the "may" is asked as the trigger resolves (CR 603.5), and a
// target that is gone or illegal by then is left alone (CR 608.2b). Flying is
// the engine's.
//
// No simplification.
func init() {
	t := effects.Targeting(
		effects.WheneverThisAttacks("Kinsbaile Balloonist — target creature gains flying until end of turn",
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Kinsbaile Balloonist — flying"}.Apply(ctx)
			}),
		effects.TargetCreature("target creature"))
	t = effects.Optional(t, "Kinsbaile Balloonist — have target creature gain flying until end of turn?")
	effects.Register(effects.Spec{
		OracleID:        "4ca5d2c7-621b-4d5e-9bf4-ff3ae271ba44",
		Name:            "Kinsbaile Balloonist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.TriggerWithPurpose(t, effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")))},
	})
}
