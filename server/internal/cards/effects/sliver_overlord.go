package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sliver Overlord — Legendary Creature — Sliver Mutant {W}{U}{B}{R}{G},
// 7/7 (Scourge):
//
//	"{3}: Search your library for a Sliver card, reveal that card, put it
//	 into your hand, then shuffle.
//	 {3}: Gain control of target Sliver. (This effect lasts indefinitely.)"
//
// The first ability is the shared reveal-and-shuffle tutor over cards with
// the Sliver subtype (a changeling card counts, CR 702.73a). The second is
// Beguiler of Wills' theft with no stated duration, which lasts until the
// game ends or the Sliver leaves (CR 611.2a, CR 400.7). Neither taps the
// Overlord, so both may be activated any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "762d864b-69bc-43ac-af2f-e18186010169",
		Name:         "Sliver Overlord",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}: Search your library for a Sliver card, reveal that card, put it into your hand, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
				Cost:    ManaCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					sliver := ctx.TextSubtype("Sliver")
					return b06TutorToHand("Sliver Overlord — a Sliver card",
						func(c game.Card) bool { return c.HasSubtype(sliver) })(item, ctx)
				},
			},
			{
				Label:   "{3}: Gain control of target Sliver.",
				Cost:    ManaCost("{3}"),
				Targets: TargetPermanent("target Sliver", HasSubtype("Sliver")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return GainControl{
						Target:     target,
						Controller: item.Controller,
						Duration:   game.IndefiniteDuration(),
						Label:      "Sliver Overlord — gain control (no stated duration)",
					}.Apply(ctx)
				},
			},
		},
	})
}
