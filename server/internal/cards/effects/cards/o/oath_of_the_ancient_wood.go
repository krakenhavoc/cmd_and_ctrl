package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of the Ancient Wood — Enchantment {2}{G} (Magic 2014, S82 old sets):
//
//	"Whenever this enchantment or another enchantment you control enters, you may
//	 put a +1/+1 counter on target creature."
//
// Daxos's Torment's constellation condition (its own entry, or another
// enchantment entering under its controller's control), with Mistblade Shinobi's
// optional targeted trigger: the creature is chosen as the trigger goes on the
// stack (CR 603.3d), any creature, and the "you may" is asked as it resolves
// (CR 603.5). A target that is no longer legal gets no counter (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "385d70ed-07ab-4e01-a473-4408f4bcf6d0",
		Name:         "Oath of the Ancient Wood",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B30SelfOrAnotherEnchantmentYouControlEntered(ev, source, g)
			}, "Oath of the Ancient Wood — put a +1/+1 counter on target creature", effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target creature")),
				"Oath of the Ancient Wood — put a +1/+1 counter on the target creature?"),
		},
	})
}
