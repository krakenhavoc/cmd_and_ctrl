package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cartouche of Ambition — Enchantment — Aura Cartouche {2}{B} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant creature you control
//	 When this Aura enters, you may put a -1/-1 counter on target creature.
//	 Enchanted creature gets +1/+1 and has lifelink."
//
// An enters trigger with a "you may": the target is any creature, chosen as the
// trigger goes on the stack, and the counter is put on it only if the question
// is answered yes and it is still a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33fe5484-7cc7-45c7-a349-b15c0454ab08",
		Name:         "Cartouche of Ambition",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Cartouche of Ambition — you may put a -1/-1 counter on target creature",
						effects.R35NightmareSowerCounter),
					"Cartouche of Ambition — put a -1/-1 counter on target creature?"),
				effects.TargetCreature("target creature")),
		},
	})
}
