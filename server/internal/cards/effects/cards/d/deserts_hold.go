package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desert's Hold — {2}{W} Enchantment — Aura (Hour of Devastation):
//
//	"Enchant creature
//	 When this Aura enters, if you control a Desert or there is a Desert card
//	 in your graveyard, you gain 3 life.
//	 Enchanted creature can't attack or block, and its activated abilities can't
//	 be activated."
//
// The "if" is an intervening condition (CR 603.4). Its activated abilities
// include its mana abilities (Lawmage's Binding's wording).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3639381-e401-409e-8f07-0a3bc736c50d",
		Name:         "Desert's Hold",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Desert or there is a Desert card in your graveyard", effects.OldB239YouHaveADesert,
				effects.WhenThisEnters("Desert's Hold — you gain 3 life", effects.Do(effects.GainLife{Amount: 3}))),
		},
	})
}
