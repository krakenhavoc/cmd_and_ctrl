package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Groundswell — Instant {G}:
//
//	"Target creature gets +2/+2 until end of turn.
//	 Landfall — If you had a land enter the battlefield under your control
//	 this turn, that creature gets +4/+4 until end of turn instead."
//
// The landfall check reads the turn tally of lands that entered under the
// caster's control (PlayerTurnTally.LandsEntered, Smuggler's Share's
// count), at resolution: a land that entered any way this turn counts,
// including one that has since left. "Instead" means one boost, not two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddbacb74-1f98-4607-a92e-d14973b9d0ef",
		Name:         "Groundswell",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			n, label := 2, "Groundswell — +2/+2"
			if ctx.Game.TurnTallyFor(item.Controller).LandsEntered > 0 {
				n, label = 4, "Groundswell — landfall +4/+4"
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: n, Toughness: n, Label: label}.Apply(ctx)
		},
	})
}
