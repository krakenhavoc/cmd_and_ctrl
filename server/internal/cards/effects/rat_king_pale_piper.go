package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rat King, Pale Piper — {3}{B} Legendary Creature — Rat Avatar 3/3
// (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Menace
//	 Whenever Rat King or another nontoken creature you control leaves
//	 the battlefield, create a 1/1 black Rat creature token.
//	 {2}, Sacrifice a token: Draw a card."
//
// The leave trigger is Pawn of Ulamog's widened to any exit (CR 603.10a
// looks back: the creature's type and controller are as it last existed
// on the battlefield). "Rat King or another nontoken": Rat King itself
// counts even as a token copy; any other creature must be a nontoken
// one. The activation takes any token its controller controls, a
// creature or not (Lord Dregg's cost); a Rat made by the trigger is one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "da359f85-7a11-4575-848d-642d1ae0ddbb",
		Name:            "Rat King, Pale Piper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				left, ok := tmcB2LeftCreature(ev, g)
				return ok && leftUnderControlOf(ev, left) == source.Controller &&
					(left.InstanceID == source.InstanceID || !IsToken(left))
			}, "Rat King — create a 1/1 black Rat creature token", Do(CreateToken{Template: TokenCard("1/1 black Rat"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice a token: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), b08SacrificeAToken()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
