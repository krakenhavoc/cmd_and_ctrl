package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_zones_view_test.go — S29. The client's cast buttons come
// straight off this projection, so the two facts to pin are the two
// that turn into a wrong button:
//
//   - `castable_here` marks exactly the graveyard cards whose text
//     opens the zone, and never a hand card (the hand is always a
//     cast surface; a bit that was always true there would be noise
//     the client had to ignore).
//   - the offers partition by zone. Overload belongs to the hand
//     copy and flashback to the graveyard copy, and an unfiltered
//     stamp would render a flashback button the server rejects.

func TestCastableHereStampedOnGraveyardCards(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-flashback"

	prevZones := game.CatalogCastableZones
	game.CatalogCastableZones = func(id string) []game.ZoneKind {
		if id == oracle {
			return []game.ZoneKind{game.ZoneGraveyard}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastableZones = prevZones })

	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{
			{Key: "overload", Label: "Overload {4}{R}", ManaCost: "{4}{R}", ClearsTargets: true},
			{Key: "flashback", Label: "Flashback {2}{R}", ManaCost: "{2}{R}", FromZone: game.ZoneGraveyard},
		}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })

	// Every seeded card is marked known to its owner: an unknown
	// card is redacted wholesale on the way out, which would make
	// this test pass for the wrong reason.
	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}

	inYard := game.NewCard("Test Looting", me.ID)
	inYard.TypeLine = "Sorcery"
	inYard.OracleID = oracle
	inYard.KnownBy = seen
	me.Graveyard.PushTop(inYard)

	plain := game.NewCard("Plain Sorcery", me.ID)
	plain.TypeLine = "Sorcery"
	plain.KnownBy = seen
	me.Graveyard.PushTop(plain)

	inHand := game.NewCard("Test Looting", me.ID)
	inHand.TypeLine = "Sorcery"
	inHand.OracleID = oracle
	inHand.KnownBy = seen
	me.Hand.PushTop(inHand)

	v := ViewOfGameFor(g, me.ID.String())
	yard := v.Seats[0].Graveyard
	var flashbackCard, plainCard *CardView
	for i := range yard.Cards {
		switch yard.Cards[i].InstanceID {
		case inYard.InstanceID.String():
			flashbackCard = &yard.Cards[i]
		case plain.InstanceID.String():
			plainCard = &yard.Cards[i]
		}
	}
	if flashbackCard == nil || plainCard == nil {
		t.Fatalf("graveyard view missing a seeded card")
	}
	if !flashbackCard.CastableHere {
		t.Errorf("declared card in the graveyard is not marked castable_here")
	}
	if len(flashbackCard.AlternativeCosts) != 1 || flashbackCard.AlternativeCosts[0].Key != "flashback" {
		t.Errorf("graveyard offers = %+v, want flashback only", flashbackCard.AlternativeCosts)
	}
	if plainCard.CastableHere {
		t.Errorf("undeclared graveyard card is marked castable_here")
	}
	if plainCard.AlternativeCosts != nil {
		t.Errorf("undeclared graveyard card carries offers: %+v", plainCard.AlternativeCosts)
	}

	var handCard *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == inHand.InstanceID.String() {
			handCard = &v.Seats[0].Hand.Cards[i]
		}
	}
	if handCard == nil {
		t.Fatalf("hand view missing the seeded card")
	}
	if handCard.CastableHere {
		t.Errorf("hand card carries castable_here — the hand is always a cast surface")
	}
	if len(handCard.AlternativeCosts) != 1 || handCard.AlternativeCosts[0].Key != "overload" {
		t.Errorf("hand offers = %+v, want overload only", handCard.AlternativeCosts)
	}
}
