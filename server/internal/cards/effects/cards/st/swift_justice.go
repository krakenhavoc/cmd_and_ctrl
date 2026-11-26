package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swift Justice — Instant {W} (Duel Decks: Speed vs. Cunning, #5555):
//
//	"Until end of turn, target creature gets +1/+0 and gains first strike and
//	 lifelink."
//
// A pump and a keyword grant on the one target, both until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d363e135-0a76-4d0e-b3b7-3293944f392c",
		Name:         "Swift Justice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			id := targets[0].ID
			if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Swift Justice — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike", "lifelink"}, Label: "Swift Justice — first strike and lifelink"}.Apply(ctx)
		},
	})
}
