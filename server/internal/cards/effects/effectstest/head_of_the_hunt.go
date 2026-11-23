package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// head_of_the_hunt_test.go — #4517: "If a creature an opponent controls
// would die, exile it instead. When you do, create a 2/2 green Wolf
// creature token." A reflexive trigger whose creator is the replacement
// (CR 603.12, CR 614.6), driven through the real destroy, wipe, CR 616
// and restore paths.

const HeadOfTheHuntOracle = "0523e271-5e0b-4e4b-9eac-86f915d18952"

func SeedHeadOfTheHunt(g *game.Game, controller uuid.UUID) uuid.UUID {
	return PushCatalogPermanent(g, controller, "Head of the Hunt", "Creature — Wolf", HeadOfTheHuntOracle, false)
}
