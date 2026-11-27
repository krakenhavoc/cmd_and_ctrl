package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soltari Visionary — Creature — Soltari Cleric {1}{W}{W}, 2/2 (Exodus, S82 OLD batch 744, #5626):
//
//	"Shadow
//	 Whenever this creature deals damage to a player, destroy target
//	 enchantment that player controls."
//
// Damage of any kind to a player triggers it, one trigger per damage event. The
// target is chosen as the trigger goes on the stack, among the enchantments THAT
// player controls (Alela's ControlledByThatPlayer, read off the damage event);
// with none the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7971f5c2-97b7-441f-ad77-954f43ede0a1",
		Name:            "Soltari Visionary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDealDamage, effects.ThisDealtDamageToAPlayer,
					"Soltari Visionary — destroy target enchantment that player controls", effects.S82OldB744DestroyTheTargetCreature),
				effects.ControlledByThatPlayer(effects.TargetPermanent("target enchantment that player controls", effects.Enchantment()))),
		},
	})
}
