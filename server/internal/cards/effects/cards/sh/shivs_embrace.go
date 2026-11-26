package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shiv's Embrace — Enchantment — Aura {2}{R}{R} (Magic 2014, S82 old sets):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying.
//	 {R}: Enchanted creature gets +1/+0 until end of turn."
//
// Spectral Flight's static pair, a layer-7c +2/+2 and a layer-6 flying grant,
// with Blessing's firebreathing: the Aura's own activated ability, each
// activation a separate boost to the creature it is attached to as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e41162bf-1a4f-4e69-870a-faf9836c56d5",
		Name:         "Shiv's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Enchanted creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.S82OldB399PumpEnchanted("Shiv's Embrace", 1, 0),
		}},
	})
}
