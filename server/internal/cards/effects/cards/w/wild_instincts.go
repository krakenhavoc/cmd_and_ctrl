package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Instincts — Sorcery {3}{G} (Magic Origins):
//
//	"Target creature you control gets +2/+2 until end of turn. It fights
//	 target creature an opponent controls."
//
// Blizzard Brawl's two required clauses with the pump unconditional:
// the +2/+2 is applied first, so the fight reads the bigger power
// (CR 701.14). If the opponent's creature is gone, the pump still
// applies and no fight happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99d9d928-a426-431a-85ee-cdec86a283cd",
		Name:         "Wild Instincts",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 2, Toughness: 2, Label: "Wild Instincts — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			return effects.B10Fight(ctx, mine.ID, theirs.ID)
		},
	})
}
