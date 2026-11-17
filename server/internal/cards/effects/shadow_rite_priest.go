package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cf42a7e3-71ee-40b8-a6ca-0368c746e128",
		Name:         "Shadow-Rite Priest",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Cleric"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}{B}, {T}, Sacrifice another Cleric: Search your library for a black creature card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tutors: 1},
			Cost:    Plus(ManaCost("{3}{B}{B}"), TapCost(), SacrificeAnotherN(1, "another Cleric", OfCreatureType("Cleric"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				black := ctx.TextColor("B")
				return SearchLibrary{
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
