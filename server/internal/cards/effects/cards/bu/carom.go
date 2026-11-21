package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "05775dea-d7f0-4e0b-af4d-ba8d320dc4c0",
		Name:         "Carom",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.RedirectDamage{Protect: effects.ShieldClause(0), Amount: 1, To: effects.RedirectToClause(1)}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
