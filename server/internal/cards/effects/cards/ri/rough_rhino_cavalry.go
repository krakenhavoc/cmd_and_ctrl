package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3ac94b28-4b0a-4121-99b9-9cd7dff7948c",
		Name:         "Rough Rhino Cavalry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Rough Rhino Cavalry — firebending 2", effects.Do(effects.NewFirebending(2))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exhaust — {8}: Put two +1/+1 counters on this creature. It gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    effects.ManaCost("{8}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := effects.PlusOneCountersOnThis(2)(g, item); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("trample", "Rough Rhino Cavalry — trample until end of turn")(g, item)
			},
		}},
	})
}
