package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hyalopterous Lemure — Creature — Spirit {4}{B}, 4/3 (Dominaria
// Remastered):
//
//	"{0}: This creature gets -1/-0 and gains flying until end of turn."
//
// A free activated ability: each activation is its own stack item, and the
// -1/-0 stacks. Both halves are pinned to this object for the turn
// (BoostUntilEOT and GrantKeywordUntilEOT on the source), so a Lemure
// that left and came back in response is a new object and is not
// affected (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc3576fb-3c08-456e-b3b6-9eb01b2366d6",
		Name:         "Hyalopterous Lemure",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: This creature gets -1/-0 and gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{0}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: -1, Toughness: 0, Label: "Hyalopterous Lemure — -1/-0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"}, Label: "Hyalopterous Lemure — flying until end of turn"}.Apply(ctx)
			},
		}},
	})
}
