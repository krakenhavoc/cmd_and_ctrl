package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Khalni Ambush // Khalni Territory — Modal DFC. Front: Instant
// {2}{G}: "Target creature you control fights target creature you don't
// control." Back: Land (mdfc_lands.go).
//
// Two target clauses, one per slot. The fight is the shared one
// (b10Fight, CR 701.12): each deals damage equal to its power to the
// other as one damage instance, and nothing happens if either has left
// the battlefield or either target became illegal. No simplification.
func init() {
	Register(Spec{
		OracleID:     "37a55560-6e32-4f54-b9a8-fd157aea6eb5",
		Name:         "Khalni Ambush",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
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
			return b10Fight(ctx, mine.ID, theirs.ID)
		},
	})
}
