package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oppressive Rays — Enchantment — Aura {W} (Magic 2015, #5556):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block unless its controller pays {3}.
//	 Activated abilities of enchanted creature cost {3} more to activate."
//
// Three clauses on the creature the Aura is attached to, each read live off
// the attachment: an attack tax and a block tax owed by its controller, the
// way Myr Prototype's self taxes are owed (CR 508.1h, 509.1d), and a cost
// increase on its activated abilities (CR 601.2f, 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a17fb46-70db-4c9a-ac0e-dfc774345673",
		Name:         "Oppressive Rays",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		AttackTaxes: []game.AttackTax{{
			Label:      "Enchanted creature can't attack unless its controller pays {3}.",
			OnAttacker: true,
			AppliesTo:  func(q game.AttackTaxQuery) bool { return q.Source.IsAttachedTo(q.Attacker.InstanceID) },
			ManaCost:   func(game.AttackTaxQuery) string { return "{3}" },
		}},
		BlockTaxes: []game.BlockTax{{
			Label:     "Enchanted creature can't block unless its controller pays {3}.",
			AppliesTo: func(q game.BlockTaxQuery) bool { return q.Source.IsAttachedTo(q.Blocker.InstanceID) },
			ManaCost:  func(game.BlockTaxQuery) string { return "{3}" },
		}},
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsMore(3,
				"Activated abilities of enchanted creature cost {3} more to activate.",
				func(q game.CostQuery) bool { return q.Ability != nil && q.Source.IsAttachedTo(q.Card.InstanceID) }),
		},
	})
}
