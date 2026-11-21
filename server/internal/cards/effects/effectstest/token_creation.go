package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_creation_test.go — the catalog half of #762. The engine's own
// notion of a creation event is pinned in game/token_create_test.go;
// what is here is the printed cards, and the boards two of them make
// together.

const (
	ParallelLivesOracle     = "84dc94b2-95fb-4d53-aaa2-191cb645639f"
	AnointedProcessionOracl = "7246d45b-2185-4cdd-981b-5419b7d52bce"
	AcademyManufactorOracle = "f36d1d8b-8303-44a9-ab56-531931641ea2"
	MondrakOracle           = "fe83087d-c6c1-40be-9295-baaa1c6b2db1"
	PrimalVigorOracle       = "c665544f-557b-4631-a1dc-39571470ca2e"
)

// pushTokenReplacementCard puts a catalog permanent on the battlefield
// so its Replacements become live. The type line is the printed one so
// nothing downstream has to pretend.
func PushTokenReplacementCard(g *game.Game, oracleID, name, typeLine string, controller uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
}

// makeTokens runs one creation instruction and returns with the
// pipeline settled (or paused, which the caller then checks).
func MakeTokens(t *testing.T, g *game.Game, controller uuid.UUID, template game.Card, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(controller, template, n); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
}

func OnBattlefieldNamed(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return n
}
