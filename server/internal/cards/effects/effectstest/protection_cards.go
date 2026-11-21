package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// protection_cards_test.go — #662, the catalog half. One assertion
// per card that lost its protection caveat, and each one is about the
// thing that caveat said was missing rather than about the Spec
// struct: a test that only read PrintedKeywords back would have
// passed before the engine change too.

const (
	BaneslayerOracle    = "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe"
	SwordFireIceOracle  = "2ccdc60a-49a9-44b9-a7af-0ebf18b26785"
	MotherOfRunesOracle = "60433b48-d27f-413c-905c-43839b1943f1"
	SwordBodyMindOracle = "fac42229-4f5f-4d04-85dd-5031d4e435aa"
)

// protectionsOn reads the engine's one quality reader off a
// battlefield permanent, as printed strings.
func ProtectionsOn(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != id {
				continue
			}
			for _, q := range game.ProtectionQualities(c) {
				out = append(out, q.Printed)
			}
		}
	})
	return out
}

func HasString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func FindStackCard(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == id {
			return &g.Stack.Cards[i]
		}
	}
	return nil
}
