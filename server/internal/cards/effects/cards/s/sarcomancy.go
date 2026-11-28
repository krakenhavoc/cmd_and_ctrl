package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarcomancy — {B} Enchantment (Tempest, S82 OLD batch 769, #5628):
//
//	"When this enchantment enters, create a 2/2 black Zombie creature token.
//	 At the beginning of your upkeep, if there are no Zombies on the battlefield,
//	 this enchantment deals 1 damage to you."
//
// Two triggers. The enter trigger makes one Zombie. The upkeep trigger has an
// intervening "if" (CR 603.4): no Zombie of any kind, token or card, any
// controller's, on the battlefield, checked as the upkeep begins and again as
// the ability resolves; the damage is the enchantment's, to its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd1e93da-7524-4bdd-9738-accd76bda7f7",
		Name:         "Sarcomancy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sarcomancy — create a 2/2 black Zombie creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1})),
			effects.InterveningIfYou("there are no Zombies on the battlefield", effects.S82OldB769NoPermanentOfSubtype("Zombie", false),
				effects.AtYourUpkeep("Sarcomancy — it deals 1 damage to you", effects.Do(effects.DamageController{N: 1}))),
		},
	})
}
