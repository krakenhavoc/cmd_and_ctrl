package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ageless Entity — Creature — Elemental {3}{G}{G}, 4/4 (Darksteel, #5605):
//
//	"Whenever you gain life, put that many +1/+1 counters on this
//	 creature."
//
// It fires once per life-gain event (every gain path, lifelink included,
// emits a positive EventChangeLife) and puts as many counters as the life
// gained, read from the event. The body is Selfless Squire's "put that
// many +1/+1 counters on this creature": nothing happens if the creature
// has left the battlefield, or come back as a new object, by the time
// the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a93236d-f182-4826-8112-4048bb3cbaf9",
		Name:         "Ageless Entity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Ageless Entity — put that many +1/+1 counters on this creature", selflessSquireCounters),
		},
	})
}
