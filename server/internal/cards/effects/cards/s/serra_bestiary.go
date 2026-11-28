package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra Bestiary — Enchantment — Aura {W}{W} (Homelands, S82 OLD b825, #5635):
//
//	"Enchant creature
//	 At the beginning of your upkeep, sacrifice this Aura unless you pay {W}{W}.
//	 Enchanted creature can't attack or block, and its activated abilities with
//	 {T} in their costs can't be activated."
//
// The upkeep payment is the shared "sacrifice unless you pay"; can't attack or
// block is Pacifism's restriction; the activation lock is an Aura-contributed
// ActivationRestriction on abilities whose cost has {T} (CR 602.5), mana
// abilities included (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "320d8f2b-005d-4769-80d2-f46d96afffe7",
		Name:         "Serra Bestiary",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttackOrBlock)},
		ActivationRestrictions: []game.ActivationRestriction{
			effects.EnchantedCreatureTapAbilitiesCantActivate("Serra Bestiary — enchanted creature's activated abilities with {T} in their costs can't be activated."),
		},
		Triggered: []game.TriggeredAbility{effects.UpkeepSacrificeUnlessYouPay("Serra Bestiary", "{W}{W}")},
	})
}
