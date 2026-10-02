package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unnatural Aggression — Instant {2}{G}, devoid:
//
//	"Devoid (This card has no color.)
//	 Target creature you control fights target creature an opponent
//	 controls. If the creature an opponent controls would die this turn,
//	 exile it instead."
//
// Devoid is colour data (the dump's colour list is empty). The two
// clauses are checked one by one at resolution (CR 608.2b). The fight
// needs both targets to be legal (b10Fight). The exile rider belongs to
// the spell, not the fight. A legal opponent's creature is marked even
// when your creature has gone and no fight happened (the 2015-08-25
// rulings), and even when the fight dealt it no damage.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "597d1dce-67e8-4c37-9781-eeaeb7c2d7b7",
		Name:         "Unnatural Aggression",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, okMine := ctx.ClauseTarget(0)
			theirs, okTheirs := ctx.ClauseTarget(1)
			okMine = okMine && mine.Kind == game.TargetCard
			okTheirs = okTheirs && theirs.Kind == game.TargetCard
			if okMine && okTheirs {
				if err := b10Fight(ctx, mine.ID, theirs.ID); err != nil {
					return err
				}
			}
			if !okTheirs {
				return nil
			}
			return ExileIfItWouldDieThisTurn{Target: theirs.ID}.Apply(ctx)
		},
	})
}
