package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dominating Licid — Creature — Licid {1}{U}{U}, 1/1 (Exodus, S82 OLD b737,
// #5626):
//
//	"{1}{U}{U}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {U} to end this effect.
//	 You control enchanted creature."
//
// The licid seam (#5781): LicidAbility, LicidForm and the enchant clause.
// The control line is Enthralling Hold's, which refers to nothing while the
// licid is a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae2804c6-fb38-4845-ae21-7f41a766cbba",
		Name:         "Dominating Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.ControlAttachedBySource()),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{1}{U}{U}", "{U}")},
	})
}
