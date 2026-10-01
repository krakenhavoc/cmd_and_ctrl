package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tidal Influence — Enchantment {2}{U}:
//
//	"Cast this spell only if no permanents named Tidal Influence are on
//	 the battlefield.
//	 This enchantment enters with a tide counter on it.
//	 At the beginning of your upkeep, put a tide counter on this
//	 enchantment.
//	 As long as there is exactly one tide counter on this enchantment,
//	 all blue creatures get -2/-0.
//	 As long as there are exactly three tide counters on this
//	 enchantment, all blue creatures get +2/+0.
//	 Whenever there are four or more tide counters on this enchantment,
//	 remove all tide counters from it."
//
// ADR 0107 §1 (#1858). Homarid's tide cycle on an enchantment, applied
// to every blue creature on the battlefield, anyone's. The cast
// restriction is checked as the spell is cast (CR 601.2) and reads each
// permanent's current name, so a copy of Tidal Influence counts, and an
// Influence put onto the battlefield without being cast is not stopped.
// The reset is a CR 603.8 state trigger.
//
// No simplification.
func init() {
	const tide = "tide"
	Register(Spec{
		OracleID:     "b1393388-b8b3-4c54-a086-f5e6d9908972",
		Name:         "Tidal Influence",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Effective().Name == "Tidal Influence" {
					return false
				}
			}
			return true
		},
		CastConditionLabel: "Cast this spell only if no permanents named Tidal Influence are on the battlefield.",
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(tide, 1, "Tidal Influence: enters with a tide counter"),
		},
		Static: []game.StaticAbility{
			whileExactlyTideCounters(1, blueCreatures, -2, 0),
			whileExactlyTideCounters(3, blueCreatures, 2, 0),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Tidal Influence — put a tide counter", putACounterOnThis(tide)),
			WhenThisHasAtLeast(tide, 4, "Tidal Influence — remove all tide counters", removeAllCountersFromThis(tide)),
		},
	})
}
