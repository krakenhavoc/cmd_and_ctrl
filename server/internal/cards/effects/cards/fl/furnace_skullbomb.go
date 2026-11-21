package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Furnace Skullbomb — Artifact {1} (Phyrexia: All Will Be One, #5479):
//
//	"{1}, Sacrifice this artifact: Draw a card.
//	 {1}{R}, Sacrifice this artifact: Put two oil counters on target
//	 artifact or creature you control. Draw a card. Activate only as a
//	 sorcery."
//
// See Basilica Skullbomb.
//
// No simplification.
func init() {
	effects.Register(effects.S82OneSkullbomb("77e0c121-724c-43a0-84e3-1800e2d94a34", "Furnace Skullbomb",
		"{1}{R}, Sacrifice this artifact: Put two oil counters on target artifact or creature you control. Draw a card. Activate only as a sorcery.",
		"{1}{R}", effects.S82OneYourArtifactOrCreature(),
		game.Purpose{},
		func(ctx *effects.Context, id uuid.UUID) error {
			return effects.AddCounter{Target: id, Kind: effects.S82OneOilCounter, N: 2}.Apply(ctx)
		}))
}
