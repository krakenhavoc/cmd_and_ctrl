package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// For the Family — Instant {G} (Streets of New Capenna, #5491):
//
//	"Target creature gets +2/+2 until end of turn. If you control four or
//	 more creatures, that creature gets +4/+4 until end of turn instead."
//
// The creatures are counted as the spell resolves (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6abd7e5-8c52-4b36-98c5-a4df8b5b6863",
		Name:         "For the Family",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			n := 2
			if len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl()))) >= 4 {
				n = 4
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: n, Toughness: n, Label: "For the Family — +N/+N"}.Apply(ctx)
		},
	})
}
