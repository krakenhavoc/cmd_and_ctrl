package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staff of Nin — Artifact {6} (Magic 2013, #5568):
//
//	"At the beginning of your upkeep, draw a card.
//	 {T}: This artifact deals 1 damage to any target."
//
// An upkeep draw and Granite Shard's ping with a bare {T} cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48d9f488-4ebb-4fdb-a568-6930fc2f5416",
		Name:         "Staff of Nin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Staff of Nin — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This artifact deals 1 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
