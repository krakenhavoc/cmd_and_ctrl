package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "259ca16d-7753-402b-bcd0-2425500dee99",
		Name:            "Levitating Statue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, b10NoncreatureSpellCastByYou,
				"Levitating Statue — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This artifact becomes a 1/1 Construct artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return cmdr2B13Animate(NewContext(g, item), item.SourceCardID,
					"Levitating Statue — a 1/1 Construct artifact creature until end of turn", 1, 1, "Construct", nil)
			},
		}},
	})
}
