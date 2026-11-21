package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Berg Strider — Snow Creature — Giant Wizard {4}{U}, 4/4 (Jumpstart 2022,
// #5948):
//
//	"When this creature enters, tap target artifact or creature an opponent
//	 controls. If {S} was spent to cast this spell, that permanent doesn't
//	 untap during its controller's next untap step."
//
// The snow mana is read off the cast the Strider entered from
// (ManaSpentToCastThis, WantsManaFrom snow, CR 400.7d); the tap is
// unconditional and the freeze rides it only when snow was spent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "9bdd68e7-7787-4579-a128-09e78fdbf0e2",
		Name:           "Berg Strider",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceSnow,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Berg Strider — tap target artifact or creature an opponent controls; if {S} was spent to cast this spell, it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if ctx.ManaSpentToCastThis().Snow() {
							return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Berg Strider — it doesn't untap during its controller's next untap step"}.Apply(ctx)
						}
						return g.TapTargetForEffect(id)
					}),
				effects.TargetPermanent("target artifact or creature an opponent controls", effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls())),
		},
	})
}
