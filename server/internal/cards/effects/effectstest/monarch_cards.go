package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// monarch_cards_test.go pins #1722: a card effect can make a player
// the monarch (CR 725), and the twelve catalog cards built on it.

const (
	CourtOfGraceOracle          = "f63c2438-27d4-449a-828f-f0a2ea86ff16"
	CourtOfAmbitionOracle       = "652e71a9-e46f-41b3-8695-76b3606b1955"
	CourtOfCunningOracle        = "2ab88c99-aaa0-4a91-9225-0bbfba04b6bc"
	CourtOfBountyOracle         = "e55d9377-f89a-41e5-a094-730d6f24caf0"
	CourtOfIreOracle            = "13f58292-9b78-4cd1-a16e-b1779a170d33"
	QueenMarchesaOracle         = "d7ac4be1-dcca-49b6-8ddb-d1b0e6cf2dcf"
	PalaceJailerOracle          = "180eda7c-fca2-403b-85cd-8ffebaf9f408"
	EntourageOfTrestOracle      = "27b0bd3f-2c11-474a-8512-fe035a514905"
	CustodiLichOracle           = "0d95ee14-9ff0-4d7f-be56-361a500ce36f"
	KnightsOfTheBlackRoseOracle = "9dbfa01b-ce7e-4dd1-9257-0c89b831f477"
	RegalBehemothOracle         = "797c2b1c-c373-4735-b397-f561559c7c61"
	ThroneOfTheHighCityOracle   = "9684447a-5955-4bc7-8ad0-8bb8b316873b"
)

// monSettle resolves everything on and headed for the stack, answering
// any trigger-ordering prompt in the order it offers. It stops at any
// other prompt, which the caller answers.
func MonSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if StackFullyEmpty(g) || len(g.PendingChoices) > 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("the stack did not settle")
}

// monCrown sets the monarch by hand (the sandbox action), for setup.
func MonCrown(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	if err := g.SetMonarch(player); err != nil {
		t.Fatalf("SetMonarch: %v", err)
	}
}

// monTokens counts `controller`'s battlefield tokens named `name`.
func MonTokens(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && effects.IsToken(c) {
			n++
		}
	}
	return n
}

// --- Palace Jailer ---------------------------------------------------

// jailerExile casts Palace Jailer for seat 0 and exiles `victim`.
func JailerExile(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	jailer := CastAndResolveCreature(t, g, "Palace Jailer", "Creature — Human Soldier", PalaceJailerOracle)
	PickCard(t, g, me.ID, victim)
	MonSettle(t, g)
	if z := E2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature is in %q, want exile", z)
	}
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want the Jailer's controller", g.Monarch)
	}
	return jailer
}
