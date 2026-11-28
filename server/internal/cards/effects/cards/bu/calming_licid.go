package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Calming Licid — Creature — Licid {2}{W}, 2/2 (Stronghold, S82 OLD batch 746, #5627):
//
//	"{W}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {W} to end this effect.
//	 Enchanted creature can't attack."
//
// The licid seam (#5781, CR 116.2c): LicidAbility is the printed ability,
// LicidForm what it turns the permanent into, and ReturnedAura the "enchant
// creature" clause. "Enchanted creature can't attack" is Forced Worship's restriction. Paying to end the effect is the engine's end_effect
// special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93d20859-249f-47ae-8f6d-095228e7b6e9",
		Name:         "Calming Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.RestrictAttached(game.CantAttack)),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{W}", "{W}")},
	})
}
