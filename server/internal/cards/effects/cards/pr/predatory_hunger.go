package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Predatory Hunger — Enchantment — Aura {G} (Exodus, S82 OLD b739, #5626):
//
//	"Enchant creature
//	 Whenever an opponent casts a creature spell, put a +1/+1 counter on
//	 enchanted creature."
//
// "An opponent" is an opponent of the Aura's controller; the counter goes on
// whatever the Aura is attached to as the ability resolves (Primal Cocoon's
// read), nothing if it is attached to nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eed2606b-9a27-4af3-8120-c45a1d6b5c00",
		Name:         "Predatory Hunger",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.Creature()),
				"Predatory Hunger — put a +1/+1 counter on enchanted creature", effects.PlusOneCounterOnEnchantedCreature),
		},
	})
}
