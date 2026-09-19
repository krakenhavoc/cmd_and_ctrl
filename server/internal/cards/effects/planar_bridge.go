package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planar Bridge — Legendary Artifact {6} (EDHREC rank 4548):
//
//	"{8}, {T}: Search your library for a permanent card, put it onto
//	 the battlefield, then shuffle."
//
// Fourteen mana across two turns for one permanent, with no
// restriction on what the permanent is. The rate is absurd and the
// ceiling is why it gets played: in a deck with enough ramp the
// Bridge is a repeatable "put anything onto the battlefield", and
// the thing it puts there is usually an Eldrazi that ends the game.
//
// "Permanent card" is every artifact, creature, enchantment, land,
// planeswalker and battle in the library, which is wider than it
// sounds — the Bridge fetching a land is a real (bad) line, and the
// picker offers it because the card does.
//
// The permanent arrives UNTAPPED and is not sacrificed at end of
// turn: this is a tutor onto the battlefield, not an impulse. A
// creature put there this way is summoning sick like any other
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "853c6cc0-ed2b-4d65-add1-2449ade8cf68",
		Name:         "Planar Bridge",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{8}, {T}: Search your library for a permanent card, put it onto the battlefield, then shuffle.",
			Cost:  Plus(ManaCost("{8}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: isPermanentCard,
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Choose a permanent card to put onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
