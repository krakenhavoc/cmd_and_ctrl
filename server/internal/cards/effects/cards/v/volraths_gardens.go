package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volrath's Gardens — Enchantment {1}{G} (Stronghold, S82 OLD b750, #5627):
//
//	"{2}, Tap an untapped creature you control: You gain 2 life. Activate only
//	 as a sorcery."
//
// Field Surgeon's cost (tapping an untapped creature you control is not the {T}
// symbol, CR 302.6, so a creature that arrived this turn may pay it), with
// sorcery timing (CR 117.1a) and a fixed 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff99337e-410c-43af-9475-7995d690bcef",
		Name:         "Volrath's Gardens",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, Tap an untapped creature you control: You gain 2 life. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapAnUntapped("an untapped creature you control", effects.Creature())),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
