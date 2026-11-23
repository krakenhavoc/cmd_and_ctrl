package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bastion of Remembrance — Enchantment {2}{B}:
//
//	"When this enchantment enters, create a 1/1 white Human Soldier
//	creature token."
//	"Whenever a creature you control dies, each opponent loses 1 life
//	and you gain 1 life."
//
// A Zulaport Cutthroat that can't be killed by creature removal —
// which in an aristocrats deck is the point: the drain survives the
// board wipe that fuels it. It even brings its own first body.
//
// Two independent halves, both triggered abilities on the stack: the
// ETB token (moved off the direct AsEnters hook in #578, so it can be
// responded to) and a dies-trigger for the drain, identical in shape to
// the Cutthroat's.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7f33cea-2ec8-4081-9208-a5b1d86721b3",
		Name:         "Bastion of Remembrance",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bastion of Remembrance — create a 1/1 Human Soldier", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("1/1 white Human Soldier"),
				N:        1,
			})),
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Bastion of Remembrance — each opponent loses 1", effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
