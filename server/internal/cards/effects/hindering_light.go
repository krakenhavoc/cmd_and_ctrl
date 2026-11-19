package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hindering Light — Instant {W}{U}:
//
//	"Counter target spell that targets you or a permanent you control.
//	 Draw a card."
//
// The clause reads the candidate spell's announced targets: the caster,
// or a permanent the caster controls on the battlefield. If the spell
// has gone by resolution the target is illegal and nothing happens, so
// no card is drawn (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "945ac15c-780b-44a9-8a8e-011cf1c30876",
		Name:         "Hindering Light",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      TargetSpell("target spell that targets you or a permanent you control", hinderingLightTargetsYouOrYourPermanent),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
