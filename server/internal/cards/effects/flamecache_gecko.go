package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flamecache Gecko — Creature — Lizard Warlock {1}{R}, 2/2 (Bloomburrow,
// #4467):
//
//	"When this creature enters, if an opponent lost life this turn, add
//	 {B}{R}.
//	 {1}{R}, Discard a card: Draw a card."
//
// The entry trigger is Bloodtithe Collector's intervening "if" (CR
// 603.4), checked as it would trigger and again as it resolves, on Priest
// of Gix's mana: a triggered ability that adds mana but is not a mana
// ability (CR 605.1b), so it uses the stack and the mana lands as it
// resolves. The rummage is Charging Strifeknight's discard cost (CR
// 601.2h via 602.2b) with a mana cost in place of the tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d767463b-0fc9-43fc-a85f-e60f030f9b96",
		Name:         "Flamecache Gecko",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Flamecache Gecko — add {B}{R} (an opponent lost life this turn)",
				Do(If{
					Cond: avB34CondAnOpponentLostLife{},
					Then: []Applier{AddMana{Produced: "{B}{R}"}},
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Discard a card: Draw a card.",
			Cost:    Plus(ManaCost("{1}{R}"), DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
