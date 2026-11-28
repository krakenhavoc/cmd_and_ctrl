package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veldrane of Sengir — Legendary Creature — Human Rogue {5}{B}{B}, 5/5
// (Homelands, S82 OLD b823, #5635):
//
//	"{1}{B}{B}: Veldrane gets -3/-0 and gains forestwalk until end of turn."
//
// The creature's own pump and keyword grant until the cleanup step (CR 514.2),
// both from the one ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39322103-c40e-4d05-8e87-a92d1cbc7b68",
		Name:         "Veldrane of Sengir",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{B}: Veldrane gets -3/-0 and gains forestwalk until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{B}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: -3, Label: "Veldrane of Sengir — -3/-0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"forestwalk"}, Label: "Veldrane of Sengir — forestwalk"}.Apply(ctx)
			},
		}},
	})
}
