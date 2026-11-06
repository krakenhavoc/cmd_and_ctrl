package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rough Rhino Cavalry — Creature — Human Mercenary {4}{R}, 5/5:
//
//	"Firebending 2 (Whenever this creature attacks, add {R}{R}. This
//	 mana lasts until end of combat.)
//	 Exhaust — {8}: Put two +1/+1 counters on this creature. It gains
//	 trample until end of turn. (Activate each exhaust ability only
//	 once.)"
//
// Firebending 2 is CR 702.189a; exhaust is CR 702.177a. The exhaust
// ability does nothing to a new object (CR 400.7) if the creature left
// and came back in response, as Bristling Hydra's ability.
func init() {
	Register(Spec{
		OracleID:     "3ac94b28-4b0a-4121-99b9-9cd7dff7948c",
		Name:         "Rough Rhino Cavalry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Rough Rhino Cavalry — firebending 2", Do(firebending(2))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {8}: Put two +1/+1 counters on this creature. It gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    ManaCost("{8}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := plusOneCountersOnThis(2)(g, item); err != nil {
					return err
				}
				return thisGainsKeywordUntilEndOfTurn("trample", "Rough Rhino Cavalry — trample until end of turn")(g, item)
			},
		}},
	})
}
