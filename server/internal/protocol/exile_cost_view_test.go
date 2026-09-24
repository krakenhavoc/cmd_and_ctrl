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

	// #1369: the controller's frame, where the hand list now lives.
	ma := controllerFrameCard(t, g, src).ManaAbilities[0]
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

// #1297: the CR 602 owner. An activated ability whose cost exiles cards
// from the activator's GRAVEYARD (Grim Lavamancer, Moorland Haunt)
// carries the same four fields, off the same walk: the clause's count
// and label, the matching cards in that graveyard (the source never
// among them, another seat's graveyard never read), and the pile.
func seatExileAbilitySource(g *game.Game, owner uuid.UUID, ec *game.ExileCost) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Exile Wizard",
		TypeLine:   "Creature — Human Wizard",
		OracleID:   "00000000-0000-0000-0000-0000000000e4",
		Power:      1, Toughness: 1,
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Exile cards from your graveyard: mark",
			Cost:   game.AbilityCost{ExileCards: ec},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	return id
}

func seatGraveyardCardFor(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func TestActivatedAbilityViewCarriesAGraveyardExileCost(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	me.Graveyard.Cards = nil
	src := seatExileAbilitySource(g, me.ID, &game.ExileCost{
		N: 1, Label: "a creature card", From: game.ZoneGraveyard,
		Match: func(c game.Card) bool { return c.IsCreature() },
	})
	bear := seatGraveyardCardFor(me, "Dead Bear", "Creature — Bear")
	seatGraveyardCardFor(me, "Spent Spell", "Sorcery")
	seatGraveyardCardFor(them, "Their Bear", "Creature — Bear")
	seatHandCardFor(g, me, "Held Creature")
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.ExileCostN != 1 || ab.ExileCostLabel != "a creature card" || ab.ExileCostZone != "graveyard" {
		t.Errorf("exile cost shape: n=%d label=%q zone=%q", ab.ExileCostN, ab.ExileCostLabel, ab.ExileCostZone)
	}
	if !sameStrings(ab.ExileCostOptions, []string{bear.String()}) {
		t.Errorf("exile_cost_options = %v, want only my graveyard's creature card %s", ab.ExileCostOptions, bear)
	}
	if ab.DiscardCostN != 0 || len(ab.DiscardCostOptions) != 0 {
		t.Errorf("an exile cost was published as a discard: n=%d options=%v", ab.DiscardCostN, ab.DiscardCostOptions)
	}
}

// The mana ability's #1283 clause names no pile and reads as the hand;
// the view says so, so the client resolves its options from the right
// zone. An ability with no exile clause ships none of the four fields.
func TestExileCostZoneDefaultsToTheHand(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	bloom := seatExileManaSource(g, me.ID, &game.ExileCost{N: 1, Label: "a card"})
	plain := seatExileAbilitySource(g, me.ID, nil)
	g.BumpLayerVersionForTest()

	if z := vehicleView(t, g, bloom).ManaAbilities[0].ExileCostZone; z != "hand" {
		t.Errorf("mana ability exile_cost_zone = %q, want hand", z)
	}
	ab := vehicleView(t, g, plain).ActivatedAbilities[0]
	if ab.ExileCostN != 0 || ab.ExileCostLabel != "" || len(ab.ExileCostOptions) != 0 || ab.ExileCostZone != "" {
		t.Errorf("an ability with no exile clause shipped one: %+v", ab)
	}
}
