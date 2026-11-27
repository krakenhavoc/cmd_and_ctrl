package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Close Quarters — Enchantment {2}{R}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Whenever a creature you control becomes blocked, this enchantment deals
//	 1 damage to any target."
//
// "Becomes blocked" is once per blocked attacker, however many creatures block it
// (CR 509.3c). The creature the event names must be one the enchantment's
// controller controls as it becomes blocked; the target is chosen as the trigger
// goes on the stack and the damage is the enchantment's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3f841c4-6e12-46fc-802a-3025182e7ff4",
		Name:         "Close Quarters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBecomesBlocked, closeQuartersYourCreatureBlocked,
				"Close Quarters — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetAny()),
		},
	})
}

// closeQuartersYourCreatureBlocked is "a creature you control becomes blocked".
func closeQuartersYourCreatureBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesBlocked {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsCreature() && c.Controller == source.Controller
}
