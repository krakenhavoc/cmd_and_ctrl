package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const twistsAndTurnsOracleID = "740aa9d9-91a9-431e-8bf9-1344e5273e27"

// Twists and Turns — Enchantment {G}, the front of a transforming card whose
// back is Mycoid Maze (The Lost Caverns of Ixalan, #4470):
//
//	"If a creature you control would explore, instead you scry 1, then that
//	 creature explores.
//	 When this enchantment enters, target creature you control explores.
//	 When a land you control enters, if you control seven or more lands,
//	 transform this enchantment."
//
// The first line is a CR 614 replacement of the explore action
// (game.RegisterExploreScryKey): it scries 1 and then the creature explores,
// every explore of a creature you control, the one this card's own enters
// ability makes included. Each copy applies once, to what the one before
// it made (CR 614.5), so two copies are two scries. The land trigger's
// intervening "if" (CR 603.4) is read as a land enters under your control
// (the land itself counts) and again as it resolves.
//
// Mycoid Maze:
//
//	"{T}: Add {G}.
//	 {3}{G}, {T}: Look at the top four cards of your library. You may
//	 reveal a creature card from among them and put that card into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// No simplification.
func init() {
	game.RegisterExploreScryKey(twistsAndTurnsOracleID)
	Register(Spec{
		OracleID:     twistsAndTurnsOracleID,
		Name:         "Twists and Turns",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Twists and Turns — target creature you control explores", targetCreatureYouControlExplores),
				targetCreatureYouControl()),
			InterveningIf(twistsAndTurnsSevenLands, On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand() && b02CountLandsControlledBy(g, source.Controller) >= 7
			}, "Twists and Turns — transform it if you control seven or more lands", Do(Instruction(func(ctx *Context) error {
				return TransformThis{}.Apply(ctx)
			})))),
		},
	})
	Register(Spec{
		OracleID:     twistsAndTurnsOracleID + "#1",
		Name:         "Mycoid Maze",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "{T}: Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, {T}: Look at the top four cards of your library. You may reveal a creature card from among them and put that card into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:    Plus(ManaCost("{3}{G}"), TapCost()),
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Effect:  LookAtTopThenMayTakeToHand(4, Creature(), 1, "Mycoid Maze — reveal a creature card and put it into your hand"),
		}},
	})
}

// twistsAndTurnsSevenLands is "if you control seven or more lands".
var twistsAndTurnsSevenLands = CondBoard{
	Name: "you control seven or more lands",
	Test: func(ctx *Context) bool { return b02CountLandsControlledBy(ctx.Game, ctx.Controller()) >= 7 },
}
