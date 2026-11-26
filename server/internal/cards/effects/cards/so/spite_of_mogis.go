package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spite of Mogis — Sorcery {R} (Journey into Nyx, S82 OLD, #5558):
//
//	"Spite of Mogis deals damage to target creature equal to the number of
//	 instant and sorcery cards in your graveyard. Scry 1."
//
// Counted as the spell resolves, when it is on the stack and so never counts
// itself (CR 608.2h). The scry happens even when the target is gone or the
// graveyard holds none, because the spell resolves while any target is legal
// and the scry is its own instruction; if the only target is illegal the spell
// is countered on resolution and does nothing (CR 608.2b).
//
// No simplification.
func init() {
	count := effects.G2CardsInYourGraveyard(func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
	effects.Register(effects.Spec{
		OracleID:     "aab5f76b-1eb0-4fbc-ab01-593161083c16",
		Name:         "Spite of Mogis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := effects.FirstLegalTarget(ctx); ok {
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: count(item, ctx)}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
