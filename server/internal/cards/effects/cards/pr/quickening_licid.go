package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quickening Licid — Creature — Licid {1}{W}, 1/1 (Tempest, S82 OLD b752,
// #5628):
//
//	"{1}{W}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {W} to end this effect.
//	 Enchanted creature has first strike."
//
// The licid seam (#5781, CR 116.2c): LicidAbility is the printed ability (activated
// row 0), LicidForm is the Aura form it takes away while it lasts, and the Aura's
// own text is the ordinary "enchanted creature has first strike" grant, which finds
// nothing while the licid is a creature. Paying {W} to end the effect is the
// engine's end_effect special action, derived from the permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bdd9ca3-542a-43c0-9831-c6add23890d6",
		Name:         "Quickening Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.GrantToAttached("first strike")),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{1}{W}", "{W}")},
	})
}
