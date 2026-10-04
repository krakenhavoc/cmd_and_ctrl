package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_alternative_cost_view_test.go — ADR 0118 §3, #2163: the view
// stamps a granted offer into `alternative_costs` exactly like a
// printed one, which is the whole of the client's half (the cost picker
// opens on a non-empty list and shows each label). No new wire field.

func TestViewStampsAGrantedOffer(t *testing.T) {
	const jodah = "test-2163-view-jodah"
	prev := game.CatalogGrantedAlternativeCosts
	game.CatalogGrantedAlternativeCosts = func(key string) []game.GrantedAlternativeCost {
		if key != jodah {
			return nil
		}
		return []game.GrantedAlternativeCost{{Offer: game.AlternativeCost{
			Key: "granted-wubrg", Label: "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost",
			ManaCost: "{W}{U}{B}{R}{G}",
		}}}
	}
	t.Cleanup(func() { game.CatalogGrantedAlternativeCosts = prev })

	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me, them := g.Seats[0], g.Seats[1]
	seen := map[uuid.UUID]bool{me.ID: true, them.ID: true}

	src := game.NewCard("Test Jodah", me.ID)
	src.TypeLine = "Legendary Creature — Human Wizard"
	src.OracleID = jodah
	src.Controller = me.ID
	src.KnownBy = seen
	g.Battlefield.PushTop(src)

	spell := game.NewCard("Test Wurm", me.ID)
	spell.TypeLine = "Creature — Wurm"
	spell.ManaCost = "{4}{G}{G}"
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	v := ViewOfGameFor(g, me.ID.String())
	var card *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == spell.InstanceID.String() {
			card = &v.Seats[0].Hand.Cards[i]
		}
	}
	if card == nil {
		t.Fatalf("hand view missing the seeded card")
	}
	if len(card.AlternativeCosts) != 1 {
		t.Fatalf("hand offers = %+v, want the granted one", card.AlternativeCosts)
	}
	got := card.AlternativeCosts[0]
	if got.Key != "granted-wubrg" || got.ManaCost != "{W}{U}{B}{R}{G}" ||
		got.Label != "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost (Test Jodah)" {
		t.Errorf("granted offer on the wire = %+v", got)
	}
	if card.AlternativeCostRequired {
		t.Errorf("the printed cost is still claimable from hand, but the card says an alternative cost is required")
	}
}
