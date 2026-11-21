package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Thaumaturgist — Creature — Merfolk Wizard {2}{U}, 1/2:
//
//	"{T}: Switch target creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1220e11f-aab0-459b-ae3f-b095aad37294",
		Name:         "Merfolk Thaumaturgist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Switch target creature's power and toughness until end of turn.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SwitchEachTargetPT(effects.NewContext(g, item), "Merfolk Thaumaturgist — switch power and toughness")
				},
			},
		},
	})
}
