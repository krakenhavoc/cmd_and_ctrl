package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rufus Shinra — Legendary Creature — Human Noble {1}{W}{B}, 2/4
// (Final Fantasy, #4462):
//
//	"Whenever Rufus Shinra attacks, if you don't control a creature
//	 named Darkstar, create Darkstar, a legendary 2/2 white and black Dog
//	 creature token."
//
// An intervening "if" (CR 603.4): checked as the attack is declared and
// again as the trigger resolves, by name and through the layers, so a
// Darkstar that died or changed hands in response is replaced or not as
// the board then stands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0caeb396-e574-4e80-a1f6-79988fe9ba19",
		Name:         "Rufus Shinra",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(rufusNoDarkstar,
				On(game.EventAttack, AllOf(ThisAttacked, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return youControlNoCreatureNamed(g, source.Controller, "Darkstar")
				}),
					"Rufus Shinra — if you don't control a creature named Darkstar, create Darkstar, a legendary 2/2 white and black Dog",
					Do(CreateToken{Template: TokenCard("2/2 white and black Legendary Dog named Darkstar"), N: 1}))),
		},
	})
}

var rufusNoDarkstar = CondBoard{
	Name: "you don't control a creature named Darkstar",
	Test: func(ctx *Context) bool { return youControlNoCreatureNamed(ctx.Game, ctx.Controller(), "Darkstar") },
}

// youControlNoCreatureNamed is "you don't control a creature named N".
func youControlNoCreatureNamed(g *game.Game, you uuid.UUID, name string) bool {
	return countControlled(g, you, func(c game.Card) bool { return c.IsCreature() && c.Name == name }) == 0
}
