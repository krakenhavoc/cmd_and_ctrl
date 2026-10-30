package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Collision — Sorcery {3}{G} (Marvel Super Heroes Commander,
// #4455):
//
//	"Put a +1/+1 counter on target creature you control. Then that
//	 creature deals damage equal to its power to target creature an
//	 opponent controls.
//	 Basic landcycling {2}"
//
// Archdruid's Charm's second mode as a whole spell: two target clauses
// read positionally (slot 0 is the creature you control, slot 1 the
// opponent's), each with its own predicate, and the shared counterThenBite
// lands the counter first so the bite reads the grown power. Basic
// landcycling is the cycling family's BasicLandcycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2f7f39b-0569-4c32-bdbf-7fb0a0efd5cf",
		Name:         "Colossal Collision",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			return counterThenBite(ctx, mine.ID, theirs.ID)
		},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
