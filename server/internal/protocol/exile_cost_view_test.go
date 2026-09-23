package protocol

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_cost_view_test.go — #1283. A mana ability whose cost exiles a
// card from the activator's hand (Cadaverous Bloom) carries the discard
// triple's shape under its OWN names, because it is its own component.

func seatExileManaSource(g *game.Game, owner uuid.UUID, ec *game.ExileCost) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Exile Bloom",
		TypeLine:   "Enchantment",
		OracleID:   "00000000-0000-0000-0000-0000000000e3",
		Owner:      owner,
		Controller: owner,
		ManaAbilities: []game.ManaAbilityShape{{
			ExileCards: ec,
			Produced:   "{B2|G2}",
			Label:      "Exile a card from your hand: Add {B}{B} or {G}{G}",
		}},
	})
	return id
}

func TestManaAbilityViewCarriesExileCostAndOptions(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	src := seatExileManaSource(g, me.ID, &game.ExileCost{N: 1, Label: "a card"})
	a := seatHandCardFor(g, me, "Pitch A")
	b := seatHandCardFor(g, me, "Pitch B")
	g.BumpLayerVersionForTest()

	ma := vehicleView(t, g, src).ManaAbilities[0]
	if ma.ExileCostN != 1 || ma.ExileCostLabel != "a card" {
		t.Errorf("exile cost shape: n=%d label=%q", ma.ExileCostN, ma.ExileCostLabel)
	}
	got := append([]string(nil), ma.ExileCostOptions...)
	sort.Strings(got)
	want := []string{a.String(), b.String()}
	sort.Strings(want)
	if !sameStrings(got, want) {
		t.Errorf("exile_cost_options = %v, want the seat's whole hand", ma.ExileCostOptions)
	}
	// It is not a discard, and the view must not say it is.
	if ma.DiscardCostN != 0 || len(ma.DiscardCostOptions) != 0 {
		t.Errorf("an exile cost was published as a discard: n=%d options=%v", ma.DiscardCostN, ma.DiscardCostOptions)
	}
}
