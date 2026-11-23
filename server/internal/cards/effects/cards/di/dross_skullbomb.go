package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.S82OneSkullbomb("0c43e9b4-013a-47ee-9805-f41afb0a3bde", "Dross Skullbomb",
		"{2}{B}, Sacrifice this artifact: Return target creature card from your graveyard to your hand. Draw a card. Activate only as a sorcery.",
		"{2}{B}", effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
		game.Purpose{},
		func(ctx *effects.Context, id uuid.UUID) error {
			return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
		}))
}
