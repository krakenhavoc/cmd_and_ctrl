package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain's Maneuver — Instant {X}{R}{W}:
//
//	"The next X damage that would be dealt to target creature,
//	 planeswalker, or player this turn is dealt to another target
//	 creature, planeswalker, or player instead."
//
// ADR 0108 §9 (#1905): a charged redirection (CR 615.7) from the first
// target to the second, X fixed as it resolves. The rulings: a second
// target that can't be dealt damage when the damage would be redirected
// leaves the damage where it was; a first target that has gone is dealt
// nothing to redirect.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a8b93d4d-bb67-4063-ac6d-7775be1b1f10",
		Name:         "Captain's Maneuver",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets: Clauses(
			targetCreaturePlaneswalkerOrPlayer("target creature, planeswalker, or player"),
			Distinct(targetCreaturePlaneswalkerOrPlayer("another target creature, planeswalker, or player")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.X() < 1 {
				return nil
			}
			return RedirectDamage{Protect: ShieldClause(0), Amount: ctx.X(), To: RedirectToClause(1)}.Apply(ctx)
		},
	})
}
