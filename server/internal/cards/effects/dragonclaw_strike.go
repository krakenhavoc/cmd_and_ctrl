package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonclaw Strike — Sorcery {2/G}{2/U}{2/R} (Tarkir: Dragonstorm,
// #4463):
//
//	"Double the power and toughness of target creature you control until
//	 end of turn. Then it fights up to one target creature an opponent
//	 controls. (Each deals damage equal to its power to the other.)"
//
// Two target clauses (slot 0 yours, slot 1 an opponent's, "up to one").
// The double is Zopandrel's (doubling.go, CR 701.10b): +X/+Y with X and Y
// the power and toughness as it resolves. It happens whenever the first
// target is still legal; "then it fights" reads the doubled creature, so
// the fight (Ulvenwald Tracker's slotZeroFightsSlotOne, CR 701.14b) uses
// its new power, and is skipped with no second target or once either has
// left or become illegal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69e0d9f1-f9dd-4614-ad9f-13e7153710ce",
		Name:         "Dragonclaw Strike",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := DoublePowerAndToughnessOfEachUntilEOT(ctx, []uuid.UUID{mine.ID}, "Dragonclaw Strike — double power and toughness"); err != nil {
				return err
			}
			return slotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
