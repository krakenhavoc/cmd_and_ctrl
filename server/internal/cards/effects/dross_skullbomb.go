package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dross Skullbomb — Artifact {1} (Phyrexia: All Will Be One, #5479):
//
//	"{1}, Sacrifice this artifact: Draw a card.
//	 {2}{B}, Sacrifice this artifact: Return target creature card from
//	 your graveyard to your hand. Draw a card. Activate only as a
//	 sorcery."
//
// See Basilica Skullbomb. The target is a creature card in YOUR
// graveyard, re-checked on resolution.
//
// No simplification.
func init() {
	Register(s82OneSkullbomb("0c43e9b4-013a-47ee-9805-f41afb0a3bde", "Dross Skullbomb",
		"{2}{B}, Sacrifice this artifact: Return target creature card from your graveyard to your hand. Draw a card. Activate only as a sorcery.",
		"{2}{B}", TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		game.Purpose{},
		func(ctx *Context, id uuid.UUID) error {
			return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
		}))
}
