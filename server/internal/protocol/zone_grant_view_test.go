package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Satoru Umezawa
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zone_grant_view_test.go — the wire half of #3233 (ADR 0093's
// amendment of 2026-10-13). A hand card's granted ability is a
// zone_abilities row like any printed one, labelled with who gave it
// (ADR 0093 Decision 8), shown to the card's owner alone.

const zoneGrantSatoruOracle = "ed1d21ba-ddfe-4f82-9e37-93f9e2c7ba33"

func TestAGrantedHandAbilityIsAZoneAbilitiesRowForItsOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	card := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: card, Name: "Plain Ninja", TypeLine: "Creature — Human Ninja",
		OracleID: "zone-grant-view-fixture", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	satoru := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: satoru, Name: "Satoru Umezawa", TypeLine: "Legendary Creature — Human Ninja",
		OracleID: zoneGrantSatoruOracle, Power: 2, Toughness: 4,
		Owner: me.ID, Controller: me.ID,
	})
	g.BumpLayerVersionForTest()

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := handCardView(t, v, 0, card)
	if len(c.ZoneAbilities) != 1 {
		t.Fatalf("zone_abilities = %+v, want the one granted ninjutsu row", c.ZoneAbilities)
	}
	ab := c.ZoneAbilities[0]
	if !strings.HasPrefix(ab.Label, "Ninjutsu {2}{U}{B}") || ab.ManaCost != "{2}{U}{B}" {
		t.Errorf("row = %q (%s), want Ninjutsu {2}{U}{B}", ab.Label, ab.ManaCost)
	}
	if ab.Ref != "grant:satoru-umezawa/ninjutsu:0:0" {
		t.Errorf("ref = %q, want grant:satoru-umezawa/ninjutsu:0:0", ab.Ref)
	}
	if ab.GrantedBy == nil || ab.GrantedBy.ID != satoru.String() || ab.GrantedBy.Name != "Satoru Umezawa" {
		t.Errorf("granted_by = %+v, want Satoru Umezawa (%s)", ab.GrantedBy, satoru)
	}

	// The opponent sees neither the card nor a row on it.
	theirs := FilterViewFor(ViewOfGame(g), them.ID.String())
	for _, hc := range theirs.Seats[0].Hand.Cards {
		if len(hc.ZoneAbilities) != 0 {
			t.Errorf("the opponent sees zone_abilities %+v on my hand card", hc.ZoneAbilities)
		}
	}
}
