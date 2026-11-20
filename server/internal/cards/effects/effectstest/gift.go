package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gift_test.go — CR 702.174 gift (ADR 0089, #1267): the keyword's
// shapes, one test per proof card, and the engine rules the keyword
// leans on.

const (
	GiftDawnsTruceOracle     = "37c06f89-db36-4937-9404-2b07cd22e1a6"
	GiftFloodMawOracle       = "8cb36a67-9206-4665-a03f-64f52ba559c4"
	GiftLongRiversOracle     = "f1993767-1d07-49c8-b8dc-04ec9840a999"
	GiftWearDownOracle       = "27905301-333e-4cdd-90cf-188159fcf8e9"
	GiftPeerlessOracle       = "938c03fc-8adf-4c7a-8ae1-eca8401f7a83"
	GiftScrapshooterOracle   = "235e3231-c5e5-4696-a867-705b2c4158fe"
	GiftValleyRallyOracle    = "5b919920-b1b6-499b-a7fe-c630778a7831"
	GiftTestLightningOracle  = LightningBoltOracle
	GiftTestFoodTokenKeyName = "Food"
)

// giftIndex is the position of the card's gift cost in its optional
// costs — the index a cast announces to promise the gift.
func GiftIndex(t *testing.T, oracle string) int {
	t.Helper()
	for i, oc := range game.OptionalCostsFor(oracle) {
		if oc.Key == game.GiftKey {
			return i
		}
	}
	t.Fatalf("%s offers no gift cost", oracle)
	return -1
}

// castWithGift puts the card in the active seat's hand and casts it at
// sorcery speed, promising the gift to `to` (uuid.Nil casts it
// unpromised). Returns the spell's ID and the announce error.
func CastWithGift(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef, to uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	AdvanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	id := PushCatalogHandCard(active, name, typeLine, oracle)
	params := game.CastSpellParams{Targets: targets}
	if to != uuid.Nil {
		params.OptionalCosts = []int{GiftIndex(t, oracle)}
		params.GiftOpponent = to
	}
	return id, g.CastSpell(active.ID, id, params)
}

func MustCastWithGift(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef, to uuid.UUID) uuid.UUID {
	t.Helper()
	id, err := CastWithGift(t, g, name, typeLine, oracle, targets, to)
	if err != nil {
		t.Fatalf("CastSpell %s (gift to %v): %v", name, to, err)
	}
	return id
}

func CardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// countTokensNamed counts `owner`'s battlefield tokens with this name,
// and how many of them are tapped.
func CountTokensNamed(g *game.Game, controller uuid.UUID, name string) (n, tapped int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && effects.IsToken(c) && c.Name == name {
			n++
			if c.Tapped {
				tapped++
			}
		}
	}
	return n, tapped
}

// --- Scrapshooter (the permanent half, CR 702.174b) ----------------

// resolveCreatureSpell passes priority until the creature has resolved
// or a prompt stops the table.
func ResolveCreatureSpell(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
}
