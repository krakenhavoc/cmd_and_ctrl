package effects

import "github.com/google/uuid"

// dst_b2_helpers.go — shared bodies for Darksteel, batch 2 (#5605).

// dstB2TapAllArtifacts taps every artifact on the battlefield — Metal
// Fatigue's spell and Roaring Slagwurm's attack trigger. The set is read
// once, as the effect starts (CR 608.2h); an artifact already tapped
// stays tapped.
func DstB2TapAllArtifacts(ctx *Context) error {
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, Artifact()) {
		ids = append(ids, c.InstanceID)
	}
	return TapEach(ctx, ids)
}
