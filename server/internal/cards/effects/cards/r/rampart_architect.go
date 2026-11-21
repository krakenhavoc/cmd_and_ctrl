package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b0edba6f-8f54-4dc4-b952-20b57a797d60",
		Name:         "Rampart Architect",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Rampart Architect — create a 1/3 white Wall creature token with defender",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/3 white Wall with defender"), N: 1})),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !effects.ACreatureYouControlDied(ev, source, lki, g) {
					return false
				}
				dead, _ := effects.DiedCreature(ev, g)
				return effects.LeftWithKeyword(ev, dead, "defender")
			}, "Rampart Architect — you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Optional:      true,
						Shuffle:       true,
						Source:        item.SourceCardID,
						Reason:        "Rampart Architect — a basic land card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
