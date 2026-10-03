package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carom — Instant {1}{W}:
//
//	"The next 1 damage that would be dealt to target creature this turn is
//	 dealt to another target creature instead.
//	 Draw a card."
//
// ADR 0108 §9 (#1905): a 1-point charged redirection (CR 615.7) from the
// first target to the second. The rulings: either target gone before the
// damage leaves it where it was, and the card is drawn as Carom resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "05775dea-d7f0-4e0b-af4d-ba8d320dc4c0",
		Name:         "Carom",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (RedirectDamage{Protect: ShieldClause(0), Amount: 1, To: RedirectToClause(1)}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
