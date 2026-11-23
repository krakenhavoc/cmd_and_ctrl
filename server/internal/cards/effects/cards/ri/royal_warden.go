package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Royal Warden — Artifact Creature — Necron {3}{B}{B}, 3/2 (Warhammer
// 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Phalanx Commander — When this creature enters, create two tapped 2/2
//	 black Necron Warrior artifact creature tokens.
//	 Unearth {3}{B}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0192ba05-6144-4fec-a803-462ecdf64195",
		Name:         "Royal Warden",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Unearth("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Royal Warden — create two tapped 2/2 black Necron Warrior artifact creature tokens",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("2/2 black Necron Warrior artifact")).EntersTapped(), N: 2})),
		},
	})
}
