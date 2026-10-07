package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minister of Inquiries — Creature — Vedalken Advisor {U}, 1/2:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}: Target player mills three cards."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99894900-771f-4fca-bebf-9f23ec4654c2",
		Name:         "Minister of Inquiries",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Minister of Inquiries", 2),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay {E}: Target player mills three cards.",
			Cost:    Plus(TapCost(), PayEnergy(1)),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return MillCards{Player: t.ID, N: 3}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
