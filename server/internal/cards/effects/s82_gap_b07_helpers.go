package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b07_helpers.go — shared pieces of S82 GAP batch 7 (Innistrad
// artifacts, #5948).

// gapB07SacrificeTheFormerHost is "whenever this Equipment becomes unattached
// from a permanent, sacrifice that permanent" (Stitcher's Graft): the host named
// by the unattach event, or the host the Equipment last had when it left the
// battlefield while attached, sacrificed if it is still on the battlefield.
func GapB07SacrificeTheFormerHost(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var host uuid.UUID
	if tc := ctx.Trigger(); tc.Event.Kind == game.EventUnattach {
		host = tc.Event.Target
	} else if info, ok := ctx.TriggeringPermanent(); ok && info.AttachedTo.Kind == game.TargetCard {
		host = info.AttachedTo.ID
	}
	if host == uuid.Nil {
		return nil
	}
	if z := g.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return SacrificePermanent{Target: host}.Apply(ctx)
}
