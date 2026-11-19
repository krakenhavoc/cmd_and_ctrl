package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serenity — Enchantment {1}{W}:
//
//	"At the beginning of your upkeep, destroy all artifacts and
//	 enchantments. They can't be regenerated."
//
// A mandatory upkeep trigger over DestroyAllMatching with CantBeRegenerated
// (CR 701.19c). Serenity is an enchantment, so it destroys itself with the
// rest; indestructible permanents survive (the sweep's narrowing).
func init() {
	Register(Spec{
		OracleID:     "f4f751c0-02dc-45a2-a4b0-980592509944",
		Name:         "Serenity",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Serenity — destroy all artifacts and enchantments; they can't be regenerated",
				Do(DestroyAllMatching{Match: Or(Artifact(), Enchantment()), CantBeRegenerated: true})),
		},
	})
}
