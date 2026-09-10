package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismatic Vista — Land:
//
//	"{T}, Pay 1 life, Sacrifice this land: Search your library for a
//	 basic land card, put it onto the battlefield, then shuffle."
//
// A fetchland that fetches basics of any type. Same cost as the ten
// Zendikar / Onslaught fetches (fetchlandCost) and the same untapped
// arrival — what it trades is the ability to grab a dual for the
// ability to grab any basic in a five-colour deck.
//
// Reuses the shared fetchland cost rather than restating it, and
// SearchLibrary's IsBasicLand predicate rather than enumerating the
// five types. It does NOT reuse fetchBasicTapped: that helper is
// Evolving Wilds, which fetches TAPPED and pays no life. The whole
// difference between the two families is those two clauses, so
// sharing the body would have quietly upgraded Evolving Wilds or
// downgraded this.
//
// Inherits the family's sandbox gap, stated once in
// fetchland_helpers.go: SearchLibrary takes the FIRST match in
// library order, so the player does not choose which basic. That
// bites less here than on a dual-fetching fetchland — the choice is
// between basics — but it is still a choice the card gives you.
func init() {
	Register(Spec{
		OracleID: "032b8a0d-491a-4a12-ab9f-689010054d5b",
		Name:     "Prismatic Vista",
		Activated: []ActivatedAbility{{
			Label: "{T}, Pay 1 life, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield, then shuffle.",
			Cost:  fetchlandCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
