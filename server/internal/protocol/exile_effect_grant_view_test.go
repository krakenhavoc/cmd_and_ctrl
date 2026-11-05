package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Q, Stacking the Deck
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_effect_grant_view_test.go — the wire half of ADR 0219 (#4973). An
// ability a resolved effect gave an exiled card is a zone_abilities row like
// a zone grant's: no new field, labelled with who gave it, shown to the
// card's owner alone.

const exileGrantQOracle = "6866f0a4-5fe6-40d9-bbfc-6f116f4e415d"

func TestAnEffectGrantedExileAbilityIsAZoneAbilitiesRowForItsOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	q := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: q, Name: "Q, Stacking the Deck", TypeLine: "Legendary Creature — Q",
		OracleID: exileGrantQOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	card := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: card, Name: "Plain Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: "exile-grant-view-fixture", Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true, them.ID: true},
	})
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	var granted bool
	g.WithWriteLock(func() {
		granted = g.GrantAbilityToExiledCardForEffect(card, q, me.ID, "q-stacking-the-deck/copy-from-exile")
	})
	if !granted {
		t.Fatal("the grant did not land on the exiled card")
	}

	c := exileCardView(t, FilterViewFor(ViewOfGame(g), me.ID.String()), card)
	if len(c.ZoneAbilities) != 1 {
		t.Fatalf("zone_abilities = %+v, want the one granted row", c.ZoneAbilities)
	}
	ab := c.ZoneAbilities[0]
	if !strings.HasPrefix(ab.Label, "Exile a card from your hand: Create a copy of this card.") {
		t.Errorf("label = %q", ab.Label)
	}
	if ab.Ref != "grant:q-stacking-the-deck/copy-from-exile:0:0" {
		t.Errorf("ref = %q", ab.Ref)
	}
	if ab.GrantedBy == nil || ab.GrantedBy.ID != q.String() {
		t.Errorf("granted_by = %+v, want Q (%s)", ab.GrantedBy, q)
	}

	theirs := exileCardView(t, FilterViewFor(ViewOfGame(g), them.ID.String()), card)
	if len(theirs.ZoneAbilities) != 0 {
		t.Errorf("the opponent sees zone_abilities %+v on my exiled card", theirs.ZoneAbilities)
	}
}
