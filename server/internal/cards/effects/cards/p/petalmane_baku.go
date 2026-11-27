package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Petalmane Baku — Creature — Spirit {1}{G}, 1/2 (Betrayers of Kamigawa,
// #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, you may put a ki counter on
//	 this creature.
//	 {1}, Remove X ki counters from this creature: Add X mana of any one
//	 color."
//
// Quillmane Baku's ki counter trigger, and Molten Slagheap's counter-paid
// mana ability: a mana ability (no stack, CR 605.3b) whose cost is {1} plus
// the removal of X ki counters, no tap. The mana is X of ONE colour the player
// picks (ManaAmount.OneColor), X the counters actually removed; X may be 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f73f8eca-55fe-418f-a244-29ef82cfc47c",
		Name:         "Petalmane Baku",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.Or(effects.Subtype("Spirit"), effects.Subtype("Arcane")),
				"Petalmane Baku — put a ki counter on this creature", effects.PutACounterOnThis("ki")),
				"Petalmane Baku — put a ki counter on it?"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Mana:           "{1}",
				RemoveCounters: effects.RemoveCountersXFromThis("ki", 0).RemoveCounters,
			},
			ProducedForPaid: func(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
				return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"W", "U", "B", "R", "G"}, N: paid.CountersRemoved, OneColor: true}}
			},
			Label: "{1}, Remove X ki counters: Add X mana of any one color",
		}},
	})
}
