package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elvish Clancaller — Creature — Elf Druid {G}{G}, 1/1 (Core Set 2019,
// #5527):
//
//	"Other Elves you control get +1/+1.
//	 {4}{G}{G}, {T}: Search your library for a card named Elvish
//	 Clancaller, put it onto the battlefield, then shuffle."
//
// Morcant's Loyalist's Elf anthem (others, yours only) and Whisper
// Squad's search for a card of its own name, put onto the battlefield
// untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8f5440e-5d4b-420a-8cfd-27a3c72be0d2",
		Name:         "Elvish Clancaller",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{G}, {T}: Search your library for a card named Elvish Clancaller, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}{G}{G}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return game.CardNameMatches(c, "Elvish Clancaller") },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Elvish Clancaller — a card named Elvish Clancaller",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
