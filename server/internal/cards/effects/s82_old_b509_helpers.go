package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB509ColoredCreature is one kind of a "Sacrifice a red creature, a
// green creature, and a white creature" cost (the Heralds of Shards of
// Alara): a creature with the color, read through the layered view.
func S82OldB509ColoredCreature(label, color string) game.SacrificeKind {
	return game.SacrificeKind{Label: label, CardTypes: []string{"creature"}, Colors: []string{color}}
}

// S82OldB509Herald is the shared body of Behemoth's Herald, Angel's Herald
// and Demon's Herald: "{2}{C}, {T}, Sacrifice a <c1> creature, a <c2>
// creature, and a <c3> creature: Search your library for a card named
// <name>, put it onto the battlefield, then shuffle." Each kind must be
// filled by its own creature (a multicolored creature fills one).
func S82OldB509Herald(oracleID, name, manaCost, fetched string, kinds ...game.SacrificeKind) Spec {
	labels := kinds[0].Label + ", " + kinds[1].Label + ", and " + kinds[2].Label
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   manaCost + ", {T}, Sacrifice " + labels + ": Search your library for a card named " + fetched + ", put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost(manaCost), TapCost(), SacrificeEach(labels, kinds...)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return game.CardNameMatches(c, fetched) },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Source:    item.SourceCardID,
					Reason:    name + " — a card named " + fetched,
				}.Apply(NewContext(g, item))
			},
		}},
	}
}
