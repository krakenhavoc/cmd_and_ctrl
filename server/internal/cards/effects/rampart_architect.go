package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampart Architect — Creature — Elephant Advisor {3}{G}, 3/4 (Foundations
// Commander... Tarkir: Dragonstorm Commander):
//
//	"Whenever this creature enters or attacks, create a 1/3 white Wall
//	 creature token with defender.
//	 Whenever a creature you control with defender dies, you may search your
//	 library for a basic land card, put that card onto the battlefield tapped,
//	 then shuffle."
//
// The first is one printed ability with two conditions (WhenThisEntersOrAttacks)
// over the shared token row. The second watches creatures you control dying and
// reads "with defender" off the creature as it last existed on the battlefield
// (CR 603.10a: leftWithKeyword, so a Wall that lost defender to an effect is
// judged as it died); the Architect itself counts if it has defender. "You may"
// is the search's own optional prompt, and the library is shuffled either way
// it is searched (CR 701.23).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0edba6f-8f54-4dc4-b952-20b57a797d60",
		Name:         "Rampart Architect",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Rampart Architect — create a 1/3 white Wall creature token with defender",
				Do(CreateToken{Template: TokenCard("1/3 white Wall with defender"), N: 1})),
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !ACreatureYouControlDied(ev, source, lki, g) {
					return false
				}
				dead, _ := diedCreature(ev, g)
				return leftWithKeyword(ev, dead, "defender")
			}, "Rampart Architect — you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Optional:      true,
						Shuffle:       true,
						Source:        item.SourceCardID,
						Reason:        "Rampart Architect — a basic land card, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
