package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zone_mana_ability_view_test.go — the wire half of #1228.
// `zone_mana_abilities` is `zone_abilities`' twin one ability kind
// over: a card in hand whose MANA ability functions there (a Spirit
// Guide) publishes its row on the per-seat carrier, the exported
// `mana_abilities` field keeps meaning "what does this permanent do",
// and neither list ever carries the other's rows.

func seatSpiritGuideHandCard(g *game.Game, owner uuid.UUID, hand *game.Zone) uuid.UUID {
	id := uuid.New()
	hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Simian Spirit Guide",
		TypeLine:   "Creature — Ape Spirit",
		OracleID:   "00000000-0000-0000-0000-0000000000sg",
		ManaCost:   "{2}{R}",
		Power:      2, Toughness: 2,
		Owner:      owner,
		Controller: owner,
		KnownBy:    map[uuid.UUID]bool{owner: true},
		ManaAbilities: []game.ManaAbilityShape{{
			Zones:     []game.ZoneKind{game.ZoneHand},
			ExileSelf: true,
			Produced:  "{R}",
			Label:     "Exile this card from your hand: Add {R}",
		}},
	})
	return id
}

// The owner gets the row, with the cost flag and the produced string
// the client renders the button from.
func TestZoneManaAbilitiesAreProjectedForTheOwner(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := seatSpiritGuideHandCard(g, me.ID, me.Hand)

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := handCardView(t, v, 0, id)
	if len(c.ZoneManaAbilities) != 1 {
		t.Fatalf("zone_mana_abilities = %+v, want the one Spirit Guide row", c.ZoneManaAbilities)
	}
	row := c.ZoneManaAbilities[0]
	if row.Index != 0 || row.Produced != "{R}" || !row.ExileSelf || row.TapCost {
		t.Errorf("row = %+v, want index 0 / {R} / exile_self / no tap cost", row)
	}
	// The battlefield list stays empty on a hand card whose ability
	// does not function there, so a client reading only
	// `mana_abilities` offers nothing rather than something wrong.
	if len(c.ManaAbilities) != 0 {
		t.Errorf("mana_abilities = %+v on a Spirit Guide in hand, want none", c.ManaAbilities)
	}
}

// CR 113.6 the other way: the same card on the battlefield publishes
// neither list. A Spirit Guide that got cast is a 2/2 Ape.
func TestSpiritGuideOnTheBattlefieldPublishesNoManaRow(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := seatSpiritGuideHandCard(g, me.ID, g.Battlefield)

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	var found bool
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != id.String() {
			continue
		}
		found = true
		if len(c.ManaAbilities) != 0 {
			t.Errorf("mana_abilities = %+v on the battlefield, want none", c.ManaAbilities)
		}
		if len(c.ZoneManaAbilities) != 0 {
			t.Errorf("zone_mana_abilities = %+v on the battlefield, want none", c.ZoneManaAbilities)
		}
	}
	if !found {
		t.Fatal("the card is not in the battlefield view")
	}
}

// An ordinary "{T}: Add {G}" still publishes on the exported field
// from the battlefield — the half that has always worked and must
// keep working, since the exported field now filters by zone.
func TestBattlefieldManaAbilitiesStillPublishPublicly(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	known := make(map[uuid.UUID]bool, len(g.Seats))
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      me.ID,
		Controller: me.ID,
		KnownBy:    known,
	})

	v := FilterViewFor(ViewOfGame(g), g.Seats[1].ID.String())
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != id.String() {
			continue
		}
		if len(c.ManaAbilities) == 0 {
			t.Fatal("a Forest on the battlefield publishes no mana_abilities")
		}
		return
	}
	t.Fatal("the Forest is not in the battlefield view")
}

// The scoping claim. A hand is hidden wholesale, so this is
// belt-and-braces today — but the rows ride castOffers, and that is
// what will still be true the day a public pile joins
// supportedManaAbilityZones.
func TestZoneManaAbilitiesDoNotReachAnotherSeat(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := seatSpiritGuideHandCard(g, me.ID, me.Hand)

	v := FilterViewFor(ViewOfGame(g), them.ID.String())
	for _, c := range v.Seats[0].Hand.Cards {
		if c.InstanceID == id.String() && len(c.ZoneManaAbilities) != 0 {
			t.Fatalf("another seat reads %+v off my hand card", c.ZoneManaAbilities)
		}
	}
}
