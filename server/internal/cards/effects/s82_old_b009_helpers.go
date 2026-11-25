package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b009_helpers.go — shared pieces of the Secret Lair Drop batch 9
// cards (#5462).

// ReturnAllExiledWithThisTogether is "put the exiled cards onto the
// battlefield under their owners' control": every card still in exile linked
// to the source object that left (its last-known record, the leave trigger's
// own source object), returned as one entry. A token exiled this way is gone
// for good, and a card that has since left exile has lost the link
// (Savior of Ollenbock, Parallax Tide).
func ReturnAllExiledWithThisTogether(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	return ReturnFromExileTogether{Targets: ids}.Apply(ctx)
}
