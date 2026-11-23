package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow-Rite Priest — Creature — Human Cleric {1}{B}, 2/2 (Dominaria
// United):
//
//	"Other Clerics you control get +1/+1.
//	 {3}{B}{B}, {T}, Sacrifice another Cleric: Search your library for a
//	 black creature card, put it onto the battlefield, then shuffle."
//
// A Layer 7c lord over the other Clerics its controller controls, and a
// tutor to the battlefield whose cost is mana, the tap and another Cleric
// (the Priest can't pay it with itself, CR 701.17 / "another"). The tutor
// is the controller's choice among the black creature cards, and may find
// nothing (CR 701.23b); the library is shuffled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf42a7e3-71ee-40b8-a6ca-0368c746e128",
		Name:         "Shadow-Rite Priest",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Cleric"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{B}, {T}, Sacrifice another Cleric: Search your library for a black creature card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}{B}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another Cleric", effects.OfCreatureType("Cleric"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				black := ctx.TextColor("B")
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasColor(black) },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
				}.Apply(ctx)
			},
		}},
	})
}
