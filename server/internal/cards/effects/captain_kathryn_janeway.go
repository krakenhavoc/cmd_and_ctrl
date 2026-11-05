package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Kathryn Janeway — Legendary Creature — Human Officer {3}{G}, 2/3
// (Star Trek, #4973):
//
//	"You may play an additional land on each of your turns.
//	 Whenever Captain Janeway or another creature you control enters, that
//	 creature explores."
//
// The extra land drop is Oracle of Mul Daya's AdditionalLandPlays (CR
// 305.2a counts it). "Captain Janeway or another creature you control" is
// every creature that enters under her controller, herself included, and
// the creature that entered explores (CR 701.44a) through Path of
// Discovery's body (thatCreatureExplores); one that has left by then
// explores by its last known information (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "d77d1895-ce02-4302-87c9-649e852fd0ea",
		Name:                "Captain Kathryn Janeway",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl, "Captain Kathryn Janeway — that creature explores", thatCreatureExplores),
		},
	})
}
