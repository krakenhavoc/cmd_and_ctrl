package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galadhrim Brigade — Creature — Elf Soldier {2}{G}, 2/2 (Tales of
// Middle-earth Commander, #5948):
//
//	"Squad {1}{G} (As an additional cost to cast this spell, you may pay
//	 {1}{G} any number of times. When this creature enters, create that many
//	 tokens that are copies of it.)
//	 Other Elves you control get +1/+1."
//
// Powder Ganger's squad (SquadMana, CR 702.157a, the engine's cap of 20
// payments on one announcement) and an Elvish Champion-style anthem (a lord's
// +1/+1, Others, YoursOnly) over the other Elves you control, so each token
// copy pumps the others and the original too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "da4bf807-93f3-41b7-8bfd-aa63a551754e",
		Name:          "Galadhrim Brigade",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.SquadMana("{1}{G}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Squad("Galadhrim Brigade")},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
