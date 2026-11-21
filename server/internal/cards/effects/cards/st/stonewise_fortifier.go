package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonewise Fortifier — Creature — Human Wizard {1}{W}:
//
//	"{4}{W}: Prevent all damage that would be dealt to this creature by target creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record whose
// source is the targeted creature, pinned as the ability resolves (CR
// 400.7), protecting the Fortifier itself. A Fortifier that has left the
// battlefield and come back is a new object, so nothing is shielded.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecfa6791-938a-4159-92bf-ff2b0ce69523",
		Name:         "Stonewise Fortifier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: Prevent all damage that would be dealt to this creature by target creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.ManaCost("{4}{W}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.PreventDamageFromSource{From: t.ID, Protect: effects.ShieldObject(item.SourceCardID)}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
