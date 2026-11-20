package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goldenhide Ox — Enchantment Creature — Ox {5}{G}, 5/4 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, target creature must be blocked this turn if able."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control.
// Irresistible Prey's block requirement: the target must be blocked this
// turn if able, enforced when blocks are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2934233a-b90c-45df-b604-ce462480bbb7",
		Name:         "Goldenhide Ox",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, constellationTriggers, "Goldenhide Ox — target creature must be blocked this turn if able (constellation)",
				Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return BlockRequirementUntilEOT{Target: id, Kind: game.BlockRequirementMustBeBlocked,
						Label: "Goldenhide Ox — must be blocked this turn"}.Apply(ctx)
				}))), TargetCreature("target creature")),
		},
	})
}
