package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_choice_cards_test.go — #826 / ADR 0070 through the real
// catalog: Winter Orb, Static Orb, Winter Moon (caps) and Rust Tick /
// Amber Prison (opt-outs).

const (
	OracleWinterOrb   = "1dcbd583-3388-4b34-a7cd-131648aa6abd"
	OracleStaticOrb   = "0004ebd0-dfd6-4276-b4a6-de0003e94237"
	OracleWinterMoon  = "b922f057-1c91-43eb-b74a-8a933b1be2d2"
	OracleRustTick    = "c7f20899-2625-4b3a-8bd8-0bcee07ed86e"
	OracleAmberPrison = "1c69fdcf-ba87-480a-88df-70aa4ec9fff0"
)

// pushTappedForTest parks a tapped permanent on the battlefield.
func PushTappedForTest(g *game.Game, owner uuid.UUID, name, oracleID, typeLine string) uuid.UUID {
	id := PushPermanentForTest(g, owner, name, oracleID, typeLine)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	return id
}

// advanceToUntapChoiceOf walks the cursor until the named seat's untap
// step stops for its CR 502.3 prompt. AdvanceStep is refused once the
// prompt is open (the gate), so the loop stops on the prompt itself.
func AdvanceToUntapChoiceOf(t *testing.T, g *game.Game, seat int) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 300; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceUntapChoice {
				if g.Turn.ActiveSeat != seat {
					t.Fatalf("untap prompt opened on seat %d, want %d", g.Turn.ActiveSeat, seat)
				}
				return c
			}
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward seat %d's untap choice: %v", seat, err)
		}
	}
	t.Fatalf("seat %d's untap step never asked", seat)
	return nil
}

func TappedForTest(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return c.Tapped
}
