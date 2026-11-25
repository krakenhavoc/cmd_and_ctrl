package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Squire's Devotion — Enchantment — Aura {2}{W} (Rivals of Ixalan, S82 OLD
// batch 217, #5531):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has lifelink.
//	 When this Aura enters, create a 1/1 white Vampire creature token with
//	 lifelink."
//
// The pump and the lifelink grant are the Aura's statics. The enters trigger is
// separate from the enchant: the token is created even when the enchanted
// creature has already gone, and the Aura's own entry puts it on the stack
// (CR 603.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4770ad4a-a346-4108-8f61-a077155d4bd1",
		Name:         "Squire's Devotion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Squire's Devotion — create a 1/1 white Vampire creature token with lifelink",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Vampire with lifelink"), N: 1})),
		},
	})
}
