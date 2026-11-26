package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Living Totem — Creature — Plant Elemental {3}{G}, 2/3 (Magic 2015, #5556):
//
//	"Convoke
//	 When this creature enters, you may put a +1/+1 counter on another target
//	 creature."
//
// Aerial Boost's convoke and an optional enters trigger whose target is any
// creature but the Totem (Another). The "may" is asked as the ability
// resolves, and a target that has gone gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a2e2d97-1348-4fa6-ad94-4da23ad8f43b",
		Name:         "Living Totem",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WhenThisEnters("Living Totem — you may put a +1/+1 counter on another target creature",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							id := effects.FirstLegalBattlefieldTarget(ctx)
							if id == uuid.Nil {
								return nil
							}
							return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
						}),
					effects.Another(effects.TargetCreature("another target creature"))),
				"Living Totem — put a +1/+1 counter on another target creature?"),
		},
	})
}
