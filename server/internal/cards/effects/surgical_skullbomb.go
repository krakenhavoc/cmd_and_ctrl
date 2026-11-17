package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surgical Skullbomb — Artifact {1} (Phyrexia: All Will Be One, #5479):
//
//	"{1}, Sacrifice this artifact: Draw a card.
//	 {2}{U}, Sacrifice this artifact: Return target creature to its
//	 owner's hand. Draw a card. Activate only as a sorcery."
//
// See Basilica Skullbomb.
//
// No simplification.
func init() {
	Register(s82OneSkullbomb("6d364283-617a-4661-865d-abd9c304b2f9", "Surgical Skullbomb",
		"{2}{U}, Sacrifice this artifact: Return target creature to its owner's hand. Draw a card. Activate only as a sorcery.",
		"{2}{U}", TargetCreature("target creature"),
		game.Purpose{},
		func(ctx *Context, id uuid.UUID) error {
			return BounceToHand{Target: id}.Apply(ctx)
		}))
}
