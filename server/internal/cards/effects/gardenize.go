package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gardenize — Enchantment {1}{G}{G} (Reality Fracture):
//
//	"Whenever a creature you control dies, put a charge counter on this
//	 enchantment.
//	 At the beginning of your first main phase, add {G} for each charge
//	 counter on this enchantment."
//
// "First main phase" is the precombat main (EventBeginPrecombatMain,
// which the engine emits for the active player once per turn). The mana
// is added as the trigger resolves and empties at the end of the phase,
// as the rules say (CR 106.4). The counter count is read at resolution,
// so a creature that dies in response still counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e74a2a4c-ca3c-49e8-9b0f-b4d1bb91fb54",
		Name:         "Gardenize",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Gardenize — put a charge counter on this enchantment", func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: 1}.Apply(NewContext(g, item))
			}),
			AtYourPrecombatMain("Gardenize — add {G} for each charge counter on this enchantment", func(g *game.Game, item *game.StackItem) error {
				n := chargeCountersOn(g, item.Controller, item.SourceCardID)
				if n <= 0 {
					return nil
				}
				return AddMana{Player: item.Controller, Produced: strings.Repeat("{G}", n)}.Apply(NewContext(g, item))
			}),
		},
	})
}
