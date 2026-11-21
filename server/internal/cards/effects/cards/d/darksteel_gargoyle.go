package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darksteel Gargoyle — Artifact Creature — Gargoyle {7} (Darksteel, #5605):
//
//	"Flying
//	 Indestructible (Damage and effects that say "destroy" don't destroy
//	 this creature.)"
//
// Two printed keywords, both engine-native (CR 702.9, CR 702.12).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73010421-374f-458e-aa88-248ef8ae4f8b",
		Name:            "Darksteel Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "indestructible"},
	})
}
