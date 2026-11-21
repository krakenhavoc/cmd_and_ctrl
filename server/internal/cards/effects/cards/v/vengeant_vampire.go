package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeant Vampire — Creature — Vampire {4}{B}{B}, 4/4 (M19 Gift Pack,
// #5524):
//
//	"Lifelink
//	 When this creature dies, destroy target creature an opponent
//	 controls and you gain 4 life."
//
// Mudbutton Cursetosser's death trigger with Fear of Lost Teeth's life
// after it. The life is part of the same targeted ability, so with no
// creature an opponent controls to target the trigger is removed
// (CR 603.3d), and with its target gone or no longer legal by resolution
// it does not resolve (CR 608.2b): no life either way. A target that is
// legal but indestructible survives, and the life is still gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9f4d669-d915-4984-bd3c-12202c366e9b",
		Name:            "Vengeant Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Vengeant Vampire — destroy target creature an opponent controls and you gain 4 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.DestroyFirstLegalTarget(g, item); err != nil {
						return err
					}
					return effects.GainLife{Amount: 4}.Apply(effects.NewContext(g, item))
				}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
