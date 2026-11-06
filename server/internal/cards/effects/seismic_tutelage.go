package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seismic Tutelage — Enchantment — Aura {3}{G} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Enchant creature
//	 When this Aura enters, put a +1/+1 counter on enchanted creature.
//	 Whenever enchanted creature attacks, double the number of +1/+1
//	 counters on it."
//
// The enters trigger is Level Up's: the creature the Aura enchants as the
// trigger resolves gets the counter, and a host that has left gets
// nothing (CR 608.2b). The attack trigger is the Aura's own (CR 109.5),
// and it doubles the attacking creature's +1/+1 counters by putting as
// many more on it as it has as the trigger resolves (CR 701.10e), so a
// Hardened Scales applies; a creature with none, or one that has left,
// gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ece47b1-0b4b-40ac-a90f-4573a8e3b8bd",
		Name:         "Seismic Tutelage",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Seismic Tutelage — put a +1/+1 counter on enchanted creature", levelUpAuraEnters),
			WheneverEnchantedCreatureAttacks("Seismic Tutelage — double the +1/+1 counters on enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					return b23DoublePlusOneCountersOn(NewContext(g, item), item.Trigger.Event.CardID)
				}),
		},
	})
}
