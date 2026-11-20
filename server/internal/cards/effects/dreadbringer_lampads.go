package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadbringer Lampads — Enchantment Creature — Nymph {4}{B}, 4/2 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, target creature gains intimidate until end of
//	 turn. (It can't be blocked except by artifact creatures and/or
//	 creatures that share a color with it.)"
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control.
// Intimidate is granted to the still-legal target; the engine enforces it
// when blocks are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e65fa865-3f60-41f2-8b14-b0333c897b93",
		Name:         "Dreadbringer Lampads",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, constellationTriggers, "Dreadbringer Lampads — target creature gains intimidate until end of turn (constellation)",
				pumpAndGrantFirstTarget(0, 0, "Dreadbringer Lampads — intimidate", "intimidate")), TargetCreature("target creature")),
		},
	})
}
