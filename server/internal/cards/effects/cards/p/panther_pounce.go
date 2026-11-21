package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Panther Pounce — {W} Instant (Marvel Super Heroes, #4455):
//
//	"Target player investigates. Target creature gets +1/+0 and gains
//	 flying until end of turn. Untap it."
//
// Two target clauses, each rechecked on its own (CR 608.2b): a player who
// is gone makes no Clue, a creature that is gone gets no pump and no untap,
// and neither half stops the other. The Clue is made under the targeted
// player's control, who is the one investigating (CR 701.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ead0820e-fc6b-4ab2-9110-048d6d53985d",
		Name:         "Panther Pounce",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Clauses(effects.TargetPlayer("target player"), effects.TargetCreature("target creature")),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(1, 1, 0, "flying")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if who, ok := ctx.ClauseTarget(0); ok {
				if err := (effects.Investigate{Player: who.ID, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			creature, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: creature.ID, Power: 1, Label: "Panther Pounce — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: creature.ID, Keywords: []string{"flying"}, Label: "Panther Pounce — flying"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: creature.ID}.Apply(ctx)
		},
	})
}
