package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "da4bf807-93f3-41b7-8bfd-aa63a551754e",
		Name:          "Galadhrim Brigade",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{SquadMana("{1}{G}", 20)},
		Triggered:     []game.TriggeredAbility{Squad("Galadhrim Brigade")},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
