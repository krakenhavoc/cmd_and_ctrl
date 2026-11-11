package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pyreswipe Hawk — Creature — Elemental Bird {3}{R}{R}, 4/4 (Bloomburrow
// Commander, #5368):
//
//	"Flying, haste
//	 Whenever this creature attacks, it gets +X/+0 until end of turn,
//	 where X is the greatest mana value among artifacts you control.
//	 Whenever you expend 6, gain control of up to one target artifact for
//	 as long as you control this creature."
//
// X is read as the attack trigger resolves. The control change lasts while
// the same Hawk stays under its controller's control (CR 611.2b).
func init() {
	Register(Spec{
		OracleID:        "35a03b0f-9dac-476a-97e9-67767c1e453e",
		Name:            "Pyreswipe Hawk",
		Completeness:    CompletenessFull,
		ReadsManaSpent:  game.ReadsYourCasts,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pyreswipe Hawk — gets +X/+0 until end of turn, X the greatest mana value among artifacts you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := cmdr2S18GreatestArtifactManaValue(g, ctx.Controller())
					if x <= 0 {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Pyreswipe Hawk — +X/+0"}.Apply(ctx)
				}),
			Targeting(WheneverYouExpend(6, "Pyreswipe Hawk — gain control of up to one target artifact for as long as you control it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
					if !ok {
						return nil
					}
					return GainControl{Target: id, Duration: d, Label: "Pyreswipe Hawk — gain control of the artifact"}.Apply(ctx)
				}),
				TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1)),
		},
	})
}
