package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dr. Beverly Crusher — {2}{W} Legendary Creature — Human Doctor 2/2
// (Star Trek, #4973):
//
//	"If you gaining life causes a triggered ability of a permanent you
//	 control to trigger, that ability triggers an additional time.
//	 Whenever you gain life, put a +1/+1 counter on Dr. Crusher."
//
// The first ability is a CR 603.2d trigger doubler (trekB2DoublesYourLifeGain):
// the cause is a life-gain event of Crusher's controller, and the doubled
// ability belongs to a permanent that player controls, Crusher included, so
// one life gain puts two counters on her. An opponent gaining life, an
// emblem's ability and a delayed trigger are not doubled. The second is
// Ajani's Pridemate's trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5fe6f4ae-3fb1-4df6-8760-88c8666068df",
		Name:            "Dr. Beverly Crusher",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{trekB2DoublesYourLifeGain("Dr. Beverly Crusher")},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Dr. Beverly Crusher — put a +1/+1 counter on Dr. Crusher", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
