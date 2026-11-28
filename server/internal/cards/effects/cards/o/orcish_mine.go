package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Mine — Enchantment — Aura {1}{R}{R} (Homelands, S82 OLD b825, #5635):
//
//	"Enchant land
//	 This Aura enters with three ore counters on it.
//	 At the beginning of your upkeep and whenever enchanted land becomes
//	 tapped, remove an ore counter from this Aura.
//	 When the last ore counter is removed from this Aura, destroy enchanted land
//	 and this Aura deals 2 damage to that land's controller."
//
// Enters with the counters (CR 614.1c); the removal fires at your upkeep and
// whenever the enchanted land becomes tapped (Spreading Algae's tap trigger);
// the last removal (EventCounterPlaced with none left) destroys the land and
// the Aura deals 2 damage to the land's controller, read before the
// destruction.
//
// No simplification.
func init() {
	const removeLabel = "Orcish Mine — remove an ore counter from this Aura"
	effects.Register(effects.Spec{
		OracleID:     "a6bfa03e-6317-4550-960b-fbd2a30e521f",
		Name:         "Orcish Mine",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OreCounter, 3, "Orcish Mine enters with three ore counters on it")},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBeginUpkeep, game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventBeginUpkeep {
						return effects.ByYou(ev, source, ch, g)
					}
					return effects.EnchantedPermanentBecameTapped(ev, source, ch, g)
				}, removeLabel, effects.RemoveAnOreCounterFromThis),
			effects.On(game.EventCounterPlaced, effects.TheLastOreCounterWasRemovedFromThis,
				"Orcish Mine — destroy enchanted land and deal 2 damage to its controller",
				effects.DestroyEnchantedLandAndDealDamageToItsController(2)),
		},
	})
}
