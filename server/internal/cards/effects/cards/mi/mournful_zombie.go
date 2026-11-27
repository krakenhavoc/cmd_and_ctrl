package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mournful Zombie — Creature — Zombie {2}{B}, 2/1 (Apocalypse, #5614):
//
//	"{W}, {T}: Target player gains 1 life."
//
// Orzhov Guildmage's life gain behind a mana-and-tap cost; the target is
// checked again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0ed2cf1-7bd4-4586-a682-fd588632cb3d",
		Name:         "Mournful Zombie",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Target player gains 1 life.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 1}),
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TargetPlayerGainsLife(1)(item, effects.NewContext(g, item))
			},
		}},
	})
}
