package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sadistic Glee — Enchantment — Aura {B} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature
//	 Whenever a creature dies, put a +1/+1 counter on enchanted creature."
//
// Lumberknot's dies trigger (any creature's death, tokens and the enchanted
// creature's own included, read by the type it last had, CR 603.10a). The counter
// goes on the creature the Aura is attached to as the trigger resolves (CR 608.2h),
// and on nothing if the Aura is no longer attached to a creature (Level Up's
// shared enters body, LevelUpAuraEnters).
//
// No simplification.
func init() {
	const name = "Sadistic Glee"
	effects.Register(effects.Spec{
		OracleID:     "2af22962-8f66-4868-a175-ad5d7f4c9d58",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died
			}, name+" — put a +1/+1 counter on enchanted creature", effects.LevelUpAuraEnters),
		},
	})
}
