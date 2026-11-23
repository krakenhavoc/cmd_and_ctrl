package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Levitating Statue — Artifact {2} (The Brothers' War, #5482):
//
//	"Flying
//	 Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	 artifact.
//	 {2}: This artifact becomes a 1/1 Construct artifact creature until
//	 end of turn."
//
// Chimeric Staff's animation with a fixed 1/1. Flying is printed and
// does nothing until it is a creature. Counters go on whether or not it
// is a creature (CR 122.1); they add to the 1/1 base in layer 7c.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "259ca16d-7753-402b-bcd0-2425500dee99",
		Name:            "Levitating Statue",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Levitating Statue — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This artifact becomes a 1/1 Construct artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					"Levitating Statue — a 1/1 Construct artifact creature until end of turn", 1, 1, "Construct", nil)
			},
		}},
	})
}
