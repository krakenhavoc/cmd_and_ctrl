package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Great Gilded Boat — Artifact — Vehicle {2}{U}, 4/4:
//
//	"Whenever you attack, recruit. (Draw a card, then discard a card. If
//	 you discarded a nonland card, create a 1/1 white Human Soldier
//	 creature token.)
//	 Crew 2 (Tap any number of creatures you control with total power 2
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// "Whenever you attack" is one trigger per attack declaration by its
// controller (OncePerBatch, Adeline's shape), whether or not the Boat is
// crewed or attacking itself.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:     "d4607026-80b2-4fa6-99d5-694cd3b43868",
		Name:         "Great Gilded Boat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Great Gilded Boat — recruit", Do(Recruit{}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Great Gilded Boat"),
		}},
	})
}
