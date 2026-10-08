package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fateful Absence — Instant {1}{W}:
//
//	"Destroy target creature or planeswalker. Its controller
//	 investigates. (They create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// Pongify's shape: read the controller before the destroy, then hand
// the victim the compensation. Investigating is creating a Clue, and
// the engine has no "whenever you investigate" watcher to tell the two
// apart, so the Clue is the whole instruction (Tireless Tracker's
// shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "35bba442-1aec-4d33-b502-4c580d61644b",
		Name:         "Fateful Absence",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			controller, ok := controllerOfTarget(ctx, t.ID)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: controller, Template: ClueToken(), N: 1}.Apply(ctx)
		},
	})
}
