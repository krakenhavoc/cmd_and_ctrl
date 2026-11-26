package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Efreet Weaponmaster — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Efreet Monk {3}{U}{R}{W}, 4/3: First strike. When this creature enters
//	or is turned face up, another target creature you control gets +3/+0 until
//	end of turn. Morph {2}{U}{R}{W}

func init() {
	effects.Register(effects.Spec{
		OracleID:         "882a9dc9-b1bb-4be8-850f-e28bad8f5840",
		Name:             "Efreet Weaponmaster",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"first strike"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{U}{R}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, effects.Self,
					"Efreet Weaponmaster — another target creature you control gets +3/+0 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.BoostUntilEOT{Target: id, Power: 3, Label: "Efreet Weaponmaster — +3/+0"}.Apply(ctx)
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
