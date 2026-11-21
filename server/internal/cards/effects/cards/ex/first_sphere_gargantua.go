package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// First-Sphere Gargantua — Creature — Phyrexian Horror {4}{B}{B}, 5/4:
//
//	"When this creature enters, you draw a card and you lose 1 life.
//	 Unearth {2}{B}"
//
// The enters trigger is Gray Merchant-style draw then life loss as one
// trigger; unearth is Artificer's Dragon's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e101edcd-34d8-4bcd-8ba6-d5941584f9fa",
		Name:         "First-Sphere Gargantua",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("First-Sphere Gargantua — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1})),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{2}{B}")},
	})
}
