package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Investigator's Journal — Artifact — Book Clue {2}, (Innistrad: Crimson
// Vow, #5495):
//
//	"This artifact enters with a number of suspect counters on it equal to
//	 the greatest number of creatures a player controls.
//	 {2}, {T}, Remove a suspect counter from this artifact: Draw a card.
//	 {2}, Sacrifice this artifact: Draw a card."
//
// The counters are a self "enters with" replacement counted as it enters
// (Blacksnag Buzzard's b19EntersWithCountersCounted): the largest creature
// count among every player, the controller's included, with the Journal
// itself not a creature. A "suspect" counter is only a name (it is not the
// suspect designation of CR 701.60). The first ability's cost is {2}, {T}
// and removing one counter, so it can't be paid with none left; the
// second, printed as a Clue's, is sacrificed for {2} regardless of counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22e0e822-c4aa-4ec1-b7e4-0c1c869ef71b",
		Name:         "Investigator's Journal",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted("suspect", func(g *game.Game, _ *game.Card) int {
				most := 0
				per := map[uuid.UUID]int{}
				for _, c := range g.BattlefieldCardsForEffect() {
					if !c.IsCreature() {
						continue
					}
					per[c.Controller]++
					most = max(most, per[c.Controller])
				}
				return most
			}, "Investigator's Journal enters with a suspect counter for the greatest number of creatures a player controls"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}, Remove a suspect counter from this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.RemoveCountersFromThis("suspect", 1)),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{2}, Sacrifice this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
		},
	})
}
