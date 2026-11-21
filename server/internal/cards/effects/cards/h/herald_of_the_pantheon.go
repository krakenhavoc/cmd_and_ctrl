package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of the Pantheon — Creature — Centaur Shaman {1}{G}, 2/2
// (Commander Masters):
//
//	"Enchantment spells you cast cost {1} less to cast.
//	 Whenever you cast an enchantment spell, you gain 1 life."
//
// The discount is Jukai Naturalist's (CR 601.2f), the life gain is a cast
// trigger on the stack above the spell (Argothian Enchantress's shape), so
// it happens even if the spell is countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ef24a79-fade-4c38-b63c-22983206b4ad",
		Name:         "Herald of the Pantheon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 1},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Enchantment spells you cast cost {1} less to cast.", effects.YourSpell(), effects.SocB13EnchantmentSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.EnchantmentSpellCastByYou, "Herald of the Pantheon — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
