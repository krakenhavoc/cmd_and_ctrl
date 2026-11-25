package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ojutai, Soul of Winter — Legendary Creature — Dragon {5}{W}{U}, 5/6
// (Commander 2017, #5533):
//
//	"Flying, vigilance
//	 Whenever a Dragon you control attacks, tap target nonland permanent an
//	 opponent controls. That permanent doesn't untap during its controller's
//	 next untap step."
//
// Utvara Hellkite's "a Dragon you control attacks" condition (one trigger per
// attacking Dragon, CR 508.3a, a changeling counts) with Berg Strider's tap
// and freeze (TapAndFreeze): the permanent is tapped, then it skips its
// controller's next untap step. The target is chosen as the trigger goes on
// the stack (CR 603.3d) and a trigger with no legal target does nothing. A
// target that is already tapped stays tapped and still does not untap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27d7f737-59a4-4948-9b2e-599ae848e054",
		Name:            "Ojutai, Soul of Winter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B14DragonYouControlAttacked(ev, source, g)
				}, "Ojutai — tap target nonland permanent an opponent controls; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Ojutai — it doesn't untap during its controller's next untap step"}.Apply(ctx)
					}),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Not(effects.Land()), effects.OpponentControls())),
		},
	})
}
