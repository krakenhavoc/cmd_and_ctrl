package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// meld_test.go — ADR 0145, #2699: The Mightstone and Weakstone and
// Urza, Lord Protector meld into Urza, Planeswalker.

const (
	urzaLordProtectorOracle = "df2af646-3e5b-43a3-8f3e-50565889f456"
	UrzaPlaneswalkerOracle  = "759406d7-44ae-4260-9ef5-3bb2c92f751a"
	urzaPlaneswalkerPrint   = "40a01679-1a23-4b8f-9a4b-0e5d5e0f5a6c"
)

// urzaMeldPrint is the print the deck importer stamps on both halves.
func urzaMeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   UrzaPlaneswalkerOracle,
		ResultScryfallID: urzaPlaneswalkerPrint,
		Result: game.Face{
			Name:            "Urza, Planeswalker",
			TypeLine:        "Legendary Planeswalker — Urza",
			Colors:          []string{"W", "U"},
			StartingLoyalty: 7,
		},
		ResultNeedsEffect: true,
	}
}

// pushUrzaPair puts Urza, Lord Protector and The Mightstone and
// Weakstone onto the battlefield under owner, each carrying the meld
// print, and returns their IDs.
func PushUrzaPair(g *game.Game, owner uuid.UUID) (urza, stone uuid.UUID) {
	urza = PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Urza, Lord Protector", OracleID: urzaLordProtectorOracle,
		TypeLine: "Legendary Creature — Human Artificer", ManaCost: "{1}{W}{U}",
		Power: 2, Toughness: 4, Colors: []string{"W", "U"}, ColorIdentity: []string{"W", "U"},
		Owner: owner, Controller: owner, Layout: game.LayoutMeld, Meld: urzaMeldPrint(),
		ScryfallID: uuid.NewString(),
	})
	stone = PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Mightstone and Weakstone", OracleID: MightstoneOracle,
		TypeLine: "Legendary Artifact — Powerstone", ManaCost: "{5}",
		Owner: owner, Controller: owner, Layout: game.LayoutMeld, Meld: urzaMeldPrint(),
		ScryfallID: uuid.NewString(),
	})
	return urza, stone
}

// activateUrzaMeld activates Urza's {7} at sorcery speed and lets it
// resolve.
func ActivateUrzaMeld(t *testing.T, g *game.Game, me *game.Player, urza uuid.UUID) error {
	t.Helper()
	AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	for range 7 {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, urza, 0, game.ActivateAbilityParams{}); err != nil {
		return err
	}
	PassPriorityAroundTable(t, g)
	return nil
}

func MeldedUrza(t *testing.T, g *game.Game) game.Card {
	t.Helper()
	var found []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == UrzaPlaneswalkerOracle {
			found = append(found, c)
		}
	}
	if len(found) != 1 {
		t.Fatalf("want one Urza, Planeswalker on the battlefield, got %d: %+v", len(found), g.Battlefield.Cards)
	}
	return found[0]
}
