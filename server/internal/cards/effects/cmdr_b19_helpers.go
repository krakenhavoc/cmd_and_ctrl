package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b19_helpers.go — shared pieces of Commander deck cards, batches 19
// and 20 (#5368).

// cmdrB19DefenderProduced is the mana ability's X for Axebane Guardian:
// the number of creatures with defender its controller controls.
func cmdrB19DefenderProduced(g *game.Game, controller, _ uuid.UUID) int {
	return cmdrB17DefendersControlled(g, controller)
}

// cmdrB19ReturnListedToCommandZoneBody is "return it to the command zone"
// at a delayed trigger (Hellkite Courser): every listed card that is still
// on the battlefield goes home. A card that went elsewhere in the meantime
// stays where it is.
var cmdrB19ReturnListedToCommandZoneBody = game.SimpleDelayedBody("commander/return-listed-to-command-zone",
	func(g *game.Game, item *game.StackItem) error {
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := g.ReturnToCommandZoneForEffect(t.ID); err != nil {
				return err
			}
		}
		return nil
	})
