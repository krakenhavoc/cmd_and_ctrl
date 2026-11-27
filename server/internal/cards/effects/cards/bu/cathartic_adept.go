package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathartic Adept — Creature — Human Wizard {U}, 1/1 (Shards of Alara, S82
// OLD batch 510, #5588):
//
//	"{T}: Target player mills a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44d30168-8203-4a8e-a11f-b4a068231e55",
		Name:         "Cathartic Adept",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target player mills a card.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.B31MillChosenPlayer(1),
		}},
	})
}
