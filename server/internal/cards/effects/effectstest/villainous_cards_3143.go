package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// villainous_cards_3143_test.go — #3143: the six villainous-choice cards
// that each needed one piece more than the keyword action. The shape
// itself is villainous_choice_test.go.

const (
	VillainMaster  = "e5e2e4ef-1d78-41eb-b634-1a90d4cb7125"
	VillainGenesis = "71ff4603-9f7a-459e-b5b3-9b39583f1dc7"
	VillainHunted  = "7d178f1c-faf7-4efd-87a6-201a40c55ec8"
	VillainShuttle = "22fa2860-d746-4429-b7e3-3403bef27f7b"
	VillainMara    = "16090270-0ac5-404c-b0c8-a6c8a4b508a4"
	VillainMissy   = "905610cd-1461-4f5a-b9c8-19b193b9d513"
	SerraAngelID   = "4b7ac066-e5c7-43e6-9e7e-2739b24a905d"
)

func VillainDestroy(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

// ---- The Master, Gallifrey's End -------------------------------------------

func MasterBoard(t *testing.T) (g *game.Game, me *game.Player, robot uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me = g.Seats[0]
	B12Push(g, me.ID, "The Master, Gallifrey's End", "Legendary Creature — Time Lord Rogue", VillainMaster, 4, 3)
	robot = PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	AdvanceToMain(t, g)
	return g, me, robot
}
