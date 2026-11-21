package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spire Mechcycle — Artifact — Vehicle {4}{R}, 5/4 (Aetherdrift, #4464):
//
//	"Haste
//	 Exhaust — Tap another untapped Mount or Vehicle you control: This
//	 Vehicle becomes an artifact creature. Put a +1/+1 counter on it for
//	 each Mount and/or Vehicle you control other than this Vehicle.
//	 (Activate each exhaust ability only once.)
//	 Crew 2"
//
// Rocketeer Boostbuggy's exhaust shape with TapAnotherUntapped as the
// whole cost (not the {T} symbol, so a Mount that arrived this turn may
// pay it, CR 302.6). The counters are counted as the ability resolves,
// over the Mounts and Vehicles its controller controls other than the
// Mechcycle, the one tapped for the cost included. "Mount" is a word of
// the card's text (CR 612.2); Vehicle is an artifact type, which no text
// change reaches. The Mechcycle stays an artifact creature until it
// leaves (CR 301.7b), and one that left and came back in response is a
// new object and gets nothing (CR 400.7). Crew is the engine's
// (CR 702.122a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d347d59-0cd5-4207-86cb-5d6244c7c70a",
		Name:            "Spire Mechcycle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Exhaust — Tap another untapped Mount or Vehicle you control: This Vehicle becomes an artifact creature. Put a +1/+1 counter on it for each Mount and/or Vehicle you control other than this Vehicle.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
				Exhaust: true,
				Cost:    effects.TapAnotherUntapped("another untapped Mount or Vehicle you control", effects.Or(effects.Subtype("Mount"), effects.HasSubtype("Vehicle"))),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if effects.SourceIsNewObject(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.BecomeArtifactCreature{
						Label: "Spire Mechcycle — becomes an artifact creature",
					}).Apply(ctx); err != nil {
						return err
					}
					mount := ctx.TextSubtype("Mount")
					n := effects.CountControlled(g, ctx.Controller(), func(c game.Card) bool {
						return c.InstanceID != item.SourceCardID && (c.HasSubtype(mount) || c.HasSubtype("Vehicle"))
					})
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				},
			},
			{
				Label:   "Crew 2",
				Cost:    effects.CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Spire Mechcycle"),
			},
		},
	})
}
