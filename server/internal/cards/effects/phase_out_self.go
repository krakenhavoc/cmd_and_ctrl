package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phase_out_self.go — "{cost}: This creature phases out." (Rainbow
// Efreet, Teferi's Honor Guard; Visions, #5631). Append-only.

// thisPhasesOut is the whole Effect of "This creature phases out": the
// ability's source phases out with everything attached to it (PhaseOut,
// CR 702.26b) and phases back in during its controller's next untap step
// (CR 702.26a). PhaseOut skips a source that has left the battlefield or
// come back as a new object (CR 400.7).
func thisPhasesOut(g *game.Game, item *game.StackItem) error {
	return PhaseOut{Targets: []uuid.UUID{item.SourceCardID}}.Apply(NewContext(g, item))
}
