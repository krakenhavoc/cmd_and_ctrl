package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const esiorOracle = "ddf38a32-302a-4658-9eaf-b813ebe275c9"

// esiorBoard puts Esior and a commander on the first seat's battlefield,
// with a plain creature beside them and a commander on the second seat's.
func esiorBoard(t *testing.T) (g *game.Game, me, opp *game.Player, cmdr, bear, theirCmdr uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Esior, Wardwing Familiar", "Legendary Creature — Bird", esiorOracle, false)
	cmdr = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	bear = effectstest.SeedBear(g, me.ID)
	theirCmdr = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Elf",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, IsCommander: true,
	})
	return g, me, opp, cmdr, bear, theirCmdr
}

func esiorAt(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }

// CR 601.2f: an opponent's spell that targets a commander Esior's
// controller controls costs {3} more; any other spell does not.
func TestEsiorTaxesOpponentsSpellsAtYourCommanders(t *testing.T) {
	g, me, opp, cmdr, bear, theirCmdr := esiorBoard(t)
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, esiorAt(cmdr)); got != 4 {
		t.Errorf("an opponent's spell at my commander costs %d, want 4", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, esiorAt(bear)); got != 1 {
		t.Errorf("an opponent's spell at my non-commander costs %d, want 1", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, esiorAt(theirCmdr)); got != 1 {
		t.Errorf("an opponent's spell at their own commander costs %d, want 1", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, me, esiorAt(cmdr)); got != 1 {
		t.Errorf("my own spell at my commander costs %d, want 1", got)
	}
	// "One or more": two commander targets are still {3}.
	two := append(esiorAt(cmdr), game.TargetRef{Kind: game.TargetCard, ID: bear})
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, two); got != 4 {
		t.Errorf("an opponent's spell at my commander and a bear costs %d, want 4", got)
	}
}

// The issue's acceptance: an opponent's Murder at your commander costs
// {4}{B}{B}, at a non-commander {1}{B}{B}, through the cast path.
func TestEsiorMurderAtYourCommanderCostsThreeMore(t *testing.T) {
	g, _, opp, cmdr, bear, _ := esiorBoard(t)
	murder := game.Card{
		InstanceID: uuid.New(), Name: "Murder", TypeLine: "Instant", ManaCost: "{1}{B}{B}",
		Owner: opp.ID, Controller: opp.ID,
	}
	for _, tc := range []struct {
		target uuid.UUID
		want   string
	}{{cmdr, "{4}{B}{B}"}, {bear, "{1}{B}{B}"}} {
		price, err := g.PriceCast(opp.ID, murder, game.CastSpellParams{Targets: esiorAt(tc.target)})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		if got := price.Total.String(); got != tc.want {
			t.Errorf("Murder costs %s, want %s", got, tc.want)
		}
	}
}
