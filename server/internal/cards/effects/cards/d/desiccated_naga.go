package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desiccated Naga — {2}{B} Creature — Zombie Snake 3/2 (Amonkhet, #5535):
//
//	"{3}{B}: Target opponent loses 2 life and you gain 2 life. Activate only
//	 if you control a Liliana planeswalker."
//
// The gate is read as the ability is activated (CR 602.5b): a planeswalker
// with the subtype Liliana, whatever its name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab172e3e-e71e-4627-9b9a-a9295b891533",
		Name:         "Desiccated Naga",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}{B}: Target opponent loses 2 life and you gain 2 life. Activate only if you control a Liliana planeswalker.",
			Purpose:   game.Purpose{LifeGain: 2},
			Cost:      effects.ManaCost("{3}{B}"),
			Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
			Condition: effects.ControlsAtLeast(1, func(c game.Card) bool { return c.IsPlaneswalker() && c.HasSubtypeWord("Liliana") }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TargetOpponentLosesAndYouGain(g, item, 2)
			},
		}},
	})
}
