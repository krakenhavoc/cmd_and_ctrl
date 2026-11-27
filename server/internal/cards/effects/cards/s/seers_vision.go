package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seer's Vision (Invasion, S82 OLD batch 647, #5616):
//
//	Enchantment {2}{U}{B}:
//	"Your opponents play with their hands revealed.
//	Sacrifice this enchantment: Look at target player's hand and choose a card from it. That player discards that card. Activate only as a sorcery."
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f5276e83-ca3f-4755-9239-e44d91a446fd",
		Name:          "Seer's Vision",
		Completeness:  effects.CompletenessFull,
		HandsRevealed: game.HandRevealOpponents,
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice this enchantment: Look at target player's hand and choose a card from it. That player discards that card. Activate only as a sorcery.",
			Cost:         effects.SacrificeThis(),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target player"),
			Effect:       effects.S82OldB647LookAtHandAndDiscard,
		}},
	})
}
