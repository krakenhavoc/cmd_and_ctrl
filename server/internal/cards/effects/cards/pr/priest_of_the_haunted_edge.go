package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Priest of the Haunted Edge — Snow Creature — Zombie Cleric {1}{B}, 0/4:
//
//	"{T}, Sacrifice this creature: Target creature gets -X/-X until end of
//	 turn, where X is the number of snow lands you control. Activate only as a
//	 sorcery."
//
// The tap and the sacrifice are costs, so the Priest can't be activated the
// turn it arrives (CR 302.6) and is gone whatever happens to the target. X is
// the snow lands its controller controls as the ability resolves (CR 608.2h),
// the Priest not being one; a target that has left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb711b7d-bd7f-48e3-a6de-a12f94ef55d2",
		Name:         "Priest of the Haunted Edge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is the number of snow lands you control. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost:         effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.CountYouControl(ctx, effects.And(effects.Land(), effects.Snow()))
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x,
							Label: "Priest of the Haunted Edge — -X/-X until end of turn"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
