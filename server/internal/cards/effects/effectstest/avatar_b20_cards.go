package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b20_cards_test.go — Avatar: The Last Airbender, batches 20 and 21
// (#4459). Each test plays the real catalog entry and asserts what its text
// says.

const (
	AvB20PlanetariumOracle   = "f89d093d-c815-4df7-a348-2005aa227586"
	AvB20BoomerangOracle     = "3c032e74-69c6-48ca-95d4-51aa3d894c14"
	AvB20EngineRoomOracle    = "01ed9139-293d-4477-a145-3c6fe59266c7"
	AvB20RealmOfKohOracle    = "bb9ce416-eef1-49e8-89a0-2b6837505070"
	AvB20RumbleArenaOracle   = "83f2470e-2114-4cd9-ab1a-21b76e9cdf63"
	AvB20SecretTunnelOracle  = "632e2979-d88a-482e-9bb8-57b683c5310f"
	AvB20WhiteLotusOracle    = "2cdbfda3-98fc-4108-b551-c7049168924e"
	AvB20AzulaOracle         = "7c9e351d-2e6c-4050-9c3d-b15bb9255a4d"
	AvB20DaiLiAgentsOracle   = "8c05e716-b6eb-4f95-866b-7e43a898ebe2"
	AvB20OzaiOracle          = "7472af8e-2e93-43b2-bc32-a3741f6d7502"
	AvB20LessonSpellTypeLine = "Sorcery — Lesson"
)

// avB20LibraryTop puts a card on top of p's library.
func AvB20LibraryTop(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Library.PushTop(c) })
	return c.InstanceID
}
