package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauthi Trapper — Creature — Dauthi Minion {2}{B}, 1/1 (Stronghold, S82 OLD
// batch 746, #5627):
//
//	"{T}: Target creature gains shadow until end of turn. (It can block or
//	 be blocked by only creatures with shadow.)"
//
// Shadow Rift's grant (a keyword until end of turn, CR 702.28) on a tap
// ability; the Trapper may target itself. Shadow is enforced in block
// legality, and the target is read again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3aa64915-8b1b-4414-b11e-d2078c1732a5",
		Name:         "Dauthi Trapper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains shadow until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "shadow")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"shadow"}, Label: "Dauthi Trapper — shadow until end of turn"}.Apply(ctx)
			},
		}},
	})
}
