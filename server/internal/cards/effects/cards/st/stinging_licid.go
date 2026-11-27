package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stinging Licid — Creature — Licid {1}{U}, 1/1 (Tempest, S82 OLD batch
// 758, #5628):
//
//	"{1}{U}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You
//	 may pay {U} to end this effect.
//	 Whenever enchanted creature becomes tapped, this creature deals 2 damage
//	 to that creature's controller."
//
// A licid (#5781) with Insolence's trigger: the damage is dealt to the
// controller of the creature as it became tapped, by the licid (now an
// Aura). Nothing is enchanted while it is a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6e48eaf-02f8-43b3-90bf-970d4e412fa6",
		Name:         "Stinging Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       effects.LicidForm(0),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{1}{U}", "{U}")},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Stinging Licid — 2 damage to that creature's controller",
			Effect:    effects.S82OldB636InsolenceDamage,
		}},
	})
}
