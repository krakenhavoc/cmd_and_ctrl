package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minister of Inquiries — Creature — Vedalken Advisor {U}, 1/2:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}: Target player mills three cards."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99894900-771f-4fca-bebf-9f23ec4654c2",
		Name:         "Minister of Inquiries",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Minister of Inquiries", 2),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}: Target player mills three cards.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(1)),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.MillCards{Player: t.ID, N: 3}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
