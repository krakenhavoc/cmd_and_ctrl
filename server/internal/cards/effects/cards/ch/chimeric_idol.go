package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chimeric Idol — Artifact {3} (Prophecy):
//
//	"{0}: Tap all lands you control. This artifact becomes a 3/3 Turtle
//	 artifact creature until end of turn."
//
// Two instructions in one resolution, in the printed order: Nature's Will's tap
// of every untapped land its controller controls, then Selesnya Keyrune's
// animation (it stays an artifact and gains the Turtle creature type and a
// base 3/3 until end of turn, pinned to this object, CR 400.7). It is
// animated only while it is still the permanent the ability came from; the lands
// are tapped regardless. It can be activated at no cost any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78b364b7-d180-484c-a1d9-9d7a4e8abb57",
		Name:         "Chimeric Idol",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: Tap all lands you control. This artifact becomes a 3/3 Turtle artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{0}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.B24TapAllLandsControlledBy(ctx, ctx.Controller()); err != nil {
					return err
				}
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.BecomeCreatureUntilEOT{
					Target:       item.SourceCardID,
					Subtypes:     []string{"Turtle"},
					SetPower:     3,
					SetToughness: 3,
					Label:        "Chimeric Idol — a 3/3 Turtle artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
