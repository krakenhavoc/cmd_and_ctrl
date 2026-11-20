package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oakheart Dryads — Enchantment Creature — Nymph Dryad {2}{G}, 2/3 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, target creature gets +1/+1 until end of turn."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control.
// Guardian of Pilgrims' pump on the still-legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60bac6e8-0c72-4665-9b01-d07c99dd1716",
		Name:         "Oakheart Dryads",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, constellationTriggers, "Oakheart Dryads — target creature gets +1/+1 until end of turn (constellation)",
				pumpAndGrantFirstTarget(1, 1, "Oakheart Dryads — +1/+1")), TargetCreature("target creature")),
		},
	})
}
