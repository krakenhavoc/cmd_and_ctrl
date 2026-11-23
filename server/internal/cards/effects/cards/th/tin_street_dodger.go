package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tin Street Dodger — Creature — Goblin Rogue {R}, 1/1 (Ravnica Remastered,
// #5948):
//
//	"Haste
//	 {R}: This creature can't be blocked this turn except by creatures with
//	 defender."
//
// Departed Deckhand's activated block rule: a scoped record pinned to the
// Dodger as the ability resolves (CR 611.2c), so a Dodger that left and came
// back is a new object (CR 400.7) and is not covered. Activating it again adds
// a second, identical rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fde3c8df-cc57-4e64-b78b-301e103d8487",
		Name:            "Tin Street Dodger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature can't be blocked this turn except by creatures with defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.CantBeBlockedThisTurnExceptBy{
					Target:   item.SourceCardID,
					Keywords: []string{"defender"},
					Text:     "creatures with defender",
					Label:    "Tin Street Dodger — can't be blocked this turn except by creatures with defender",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
