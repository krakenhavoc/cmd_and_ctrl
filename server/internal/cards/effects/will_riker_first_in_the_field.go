package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Will Riker, First in the Field — Legendary Creature — Human Officer
// {3}{R}, 4/3 (Star Trek, #4973):
//
//	"At the beginning of combat on your turn, reveal cards from the top
//	 of your library until you reveal two land cards. Put those land
//	 cards onto the battlefield and the rest on the bottom of your
//	 library in a random order. Sacrifice those lands at the beginning
//	 of your next end step."
//
// One beginning-of-combat trigger (AtBeginningOfYourCombat). Its body is
// Open the Way's reveal with two lands, untapped (trekB18WillRikerReveal):
// a library with fewer than two land cards reveals itself whole and puts
// the lands it has. The put's keyed tail sends the rest to the bottom in
// a random order and schedules one delayed trigger at the beginning of
// the controller's next end step that sacrifices exactly those lands —
// each only while it is still the same object on the battlefield under
// that player's control (CR 400.7, 603.7c, 701.21a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c1fd7c6-2f3d-496f-9f09-660a13ef923f",
		Name:         "Will Riker, First in the Field",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Will Riker, First in the Field — reveal until two lands; put them onto the battlefield",
				trekB18WillRikerReveal),
		},
	})
}
