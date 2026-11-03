package game

import "github.com/google/uuid"

// ExileTopFaceDownUnseenKeyedForEffect exiles the top n cards of the
// player's library face down so that NOBODY may look at them (CR 406.3),
// the owner included, until a cast permission later names them: they land
// as FaceDownPermitted with no permission yet, so a grant made over them
// afterwards is what makes its holder their one viewer
// (stampPermittedViewersLocked), and they turn face up as they leave exile.
// Abstract Performance's face-down pile is the shape (#4834): an opponent
// chooses a pile without seeing it, and its controller looks at it only if
// it is the pile left to them.
//
// It differs from ExileTopFaceDownKeyedForEffect (Necropotence's `exiled`
// kind, which no later grant makes readable) and from
// ExileTopFaceDownWithPermissionForEffect (which grants on the way in, so
// its holder sees the cards at once). `then` is told the cards that landed
// in exile, top card first, once the batch is done; it also runs, told
// nothing, when there was nothing to exile.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileTopFaceDownUnseenKeyedForEffect(playerID uuid.UUID, n int, then KeyedTail) error {
	plan, err := g.exileTopFaceDownPlanLocked(playerID, n)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		return then.Run(g, nil)
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: playerID, FaceDown: FaceDownPermitted}, plan, then)
}
