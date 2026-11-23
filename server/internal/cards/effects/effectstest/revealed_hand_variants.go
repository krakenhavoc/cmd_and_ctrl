package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2115, ADR 0116's 2026-10-05 amendment: the revealed-hand pick's
// variants, through the cards that ship on them. One test per card's
// own clause; the engine's rules are pinned in
// game/revealed_hand_pick_test.go.

const (
	AppetiteForBrainsOracle      = "4ffc16e0-1778-4aba-ade3-58146c800f5f"
	AggressiveNegotiationsOracle = "c469133e-174d-476b-b135-bbf15e415e72"
	AgonizingRemorseOracle       = "1267dfda-eb1a-4963-9fe3-fa619d924d7a"
	NightsnareOracle             = "b9efd1bd-d82a-43ad-aba7-df172b9751ce"
	TourachsCanticleOracle       = "738d2be9-c40b-4a9b-95ed-18d659080918"
	TalarasBaneOracle            = "2a573cc3-0a2f-4f06-bbac-c762adb9bae4"
	ReckonerShakedownOracle      = "15159091-84ec-4b8c-8d84-8a8bb8baf27d"
	BindingNegotiationOracle     = "86dc3a41-4fb9-4b1f-8279-2b5edbc7e170"
	ExtractTheTruthOracle        = "18207fe8-41e4-418d-b5f9-e1239c527e67"
	AddleOracle                  = "3745aea4-4455-4859-9772-bcfd14e4067b"
	LastRitesOracle              = "3196e1f4-7f94-4eb6-ae2b-ed11ba392552"
	MegrimOracle                 = "633ad9e2-9f55-4a1c-9248-661ad4b0e1dc"
)

// openVariant returns the one revealed_hand_pick prompt, or fails.
func OpenVariant(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	var found *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRevealedHandPick {
			if found != nil {
				t.Fatal("more than one revealed-hand pick is open")
			}
			found = c
		}
	}
	if found == nil {
		t.Fatal("no revealed-hand pick is open")
	}
	return found
}

func NoVariant(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && (c.Kind == game.PendingChoiceRevealedHandPick || c.Kind == game.PendingChoiceDiscardFromHand) {
			t.Fatalf("a revealed-hand pick is open: %+v", c)
		}
	}
}

// castAt casts a catalog spell at the given targets and resolves it.
func CastAt(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) {
	t.Helper()
	CastCatalogSpell(t, g, name, typeLine, oracle, targets)
	PassPriorityAroundTable(t, g)
}

func RvDiscards(g *game.Game, card uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventDiscardCard && ev.CardID == card {
			n++
		}
	}
	return n
}

func RvBattlefieldCard(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func RvBig() game.Card {
	return game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{4}{G}{G}{G}", Colors: []string{"G"}}
}
