package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pursuit of Knowledge — Enchantment {3}{W} (Stronghold, S82 OLD batch 748,
// #5627):
//
//	"If you would draw a card, you may put a study counter on this enchantment
//	 instead.
//	 Remove three study counters from this enchantment, Sacrifice this
//	 enchantment: Draw seven cards."
//
// Dredge's optional draw replacement (CR 614.1, 121.6): asked for each card
// drawn, so "draw three" asks three times. A "yes" cancels that draw and puts a
// study counter on the enchantment, and the rest of a multi-card draw goes on
// (CR 121.6b). The activated ability's cost is the three counters and the
// sacrifice (CR 118.3), paid as it is announced, so it cannot be activated with
// fewer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f5124a4-9ffd-475a-8b77-3ff59a9871a1",
		Name:         "Pursuit of Knowledge",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventDrawCard},
			Optional: true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
			},
			DrawInstead: game.RegisterDrawInstead("pursuit-of-knowledge/study-counter-instead",
				func(g *game.Game, _, source uuid.UUID, done game.DrawDone) error {
					if err := g.AddCounterForEffect(source, "study", 1); err != nil && err != game.ErrCardNotFound {
						return err
					}
					return done.Run(g)
				}),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			PromptQuestion: "Pursuit of Knowledge — put a study counter on it instead of drawing a card?",
			Label:          "Pursuit of Knowledge: put a study counter instead of drawing",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three study counters from this enchantment, Sacrifice this enchantment: Draw seven cards.",
			Purpose: game.Purpose{Draws: 7, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("study", 3), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 7}),
		}},
	})
}
