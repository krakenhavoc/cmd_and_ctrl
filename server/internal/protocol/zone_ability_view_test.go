package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zone_ability_view_test.go — the wire half of #1221. `zone_abilities`
// now reaches a card in a PUBLIC pile, so the test that matters is the
// per-viewer one: the graveyard's own seat sees the row and nobody
// else does, even though every seat can read the card itself.
//
// Through #660 the field rode the hand's own secrecy — FilterViewFor
// blanks another seat's hand wholesale — and a graveyard would have
// handed one seat's `legal_targets` to the whole table (#1055). The
// rows ride castOffers instead; these tests are what pins that.

// seatUnearthCard seeds a card with a graveyard ability, marked known
// to EVERY seat — which is what a card in a public pile is, and what
// makes "the row is scoped" a different assertion from "the card is
// hidden".
func seatUnearthCard(g *game.Game, zone *game.Zone, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	known := make(map[uuid.UUID]bool, len(g.Seats))
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	zone.PushTop(game.Card{
		InstanceID: id,
		Name:       "Dregscape Zombie",
		TypeLine:   "Creature — Zombie",
		OracleID:   "00000000-0000-0000-0000-0000000000dz",
		Power:      2, Toughness: 1,
		Owner:      owner,
		Controller: owner,
		KnownBy:    known,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:        "Unearth {B}",
			Cost:         game.AbilityCost{Mana: "{B}"},
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
		}},
	})
	return id
}

func graveyardCardView(t *testing.T, v GameView, seat int, id uuid.UUID) CardView {
	t.Helper()
	for _, c := range v.Seats[seat].Graveyard.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("card %s not in seat %d's graveyard view", id, seat)
	return CardView{}
}

func exileCardView(t *testing.T, v GameView, id uuid.UUID) CardView {
	t.Helper()
	for _, c := range v.Exile.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("card %s not in the exile view", id)
	return CardView{}
}

// The graveyard's own seat gets the unearth row.
func TestGraveyardAbilitiesAreProjectedForTheOwner(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := seatUnearthCard(g, me.Graveyard, me.ID)

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := graveyardCardView(t, v, 0, id)
	if len(c.ZoneAbilities) != 1 {
		t.Fatalf("zone_abilities = %+v, want the one unearth row", c.ZoneAbilities)
	}
	row := c.ZoneAbilities[0]
	if row.Index != 0 || row.Label != "Unearth {B}" || row.ManaCost != "{B}" || !row.SorcerySpeed {
		t.Errorf("row = %+v, want index 0 / Unearth {B} / {B} / sorcery_speed", row)
	}
	// A graveyard card is not a permanent, so the battlefield list
	// stays empty — the two are disjoint by construction.
	if len(c.ActivatedAbilities) != 0 {
		t.Errorf("activated_abilities = %+v, want none off the battlefield", c.ActivatedAbilities)
	}
}

// And nobody else does. A graveyard is PUBLIC — the card, its name and
// its printed text all survive redaction for every seat — so this is
// the assertion that the row is scoped and not merely hidden by a
// zone that happens to be secret.
func TestGraveyardAbilitiesAreNotProjectedForOtherSeats(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := seatUnearthCard(g, me.Graveyard, me.ID)

	v := FilterViewFor(ViewOfGame(g), them.ID.String())
	c := graveyardCardView(t, v, 0, id)
	if len(c.ZoneAbilities) != 0 {
		t.Fatalf("zone_abilities = %+v on another seat's frame, want none", c.ZoneAbilities)
	}
	// The card itself is still readable: the scoping is of the
	// ANSWER, not of the pile.
	if c.Name != "Dregscape Zombie" {
		t.Errorf("name = %q, want the card still readable in a public pile", c.Name)
	}
}

// A spectator — no seat — gets no seat's answer, for the reason
// applyCastStampsFor gives about legal target sets.
func TestGraveyardAbilitiesAreNotProjectedForASpectator(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := seatUnearthCard(g, me.Graveyard, me.ID)

	v := FilterViewFor(ViewOfGame(g), "")
	c := graveyardCardView(t, v, 0, id)
	if len(c.ZoneAbilities) != 0 {
		t.Fatalf("zone_abilities = %+v on a spectator's frame, want none", c.ZoneAbilities)
	}
}

// Exile is the shared pile with no seat to hang a per-seat walk off,
// so its "you" is read off each card's owner (CR 108.4).
func TestExileAbilitiesAreProjectedForTheCardsOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id,
		Name:       "Exile Test",
		TypeLine:   "Creature — Test",
		OracleID:   "00000000-0000-0000-0000-0000000000ex",
		Owner:      me.ID,
		Controller: me.ID,
		KnownBy:    map[uuid.UUID]bool{me.ID: true, them.ID: true},
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{1}: do nothing",
			Cost:  game.AbilityCost{Mana: "{1}"},
			Zones: []game.ZoneKind{game.ZoneExile},
		}},
	})

	mine := exileCardView(t, FilterViewFor(ViewOfGame(g), me.ID.String()), id)
	if len(mine.ZoneAbilities) != 1 {
		t.Errorf("owner's frame: zone_abilities = %+v, want the one row", mine.ZoneAbilities)
	}
	theirs := exileCardView(t, FilterViewFor(ViewOfGame(g), them.ID.String()), id)
	if len(theirs.ZoneAbilities) != 0 {
		t.Errorf("another seat's frame: zone_abilities = %+v, want none", theirs.ZoneAbilities)
	}
}

// CR 113.6 again, on the wire: the same card on the battlefield
// offers no graveyard row, and a battlefield ability on a graveyard
// card offers none either.
func TestZoneRowsAreFilteredByTheZoneTheCardIsIn(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	onBoard := seatUnearthCard(g, g.Battlefield, me.ID)
	inGrave := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: inGrave,
		Name:       "Sol Ring",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000000sr",
		Owner:      me.ID,
		Controller: me.ID,
		KnownBy:    map[uuid.UUID]bool{me.ID: true},
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{2}: do nothing",
			Cost:  game.AbilityCost{Mana: "{2}"},
		}},
	})

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != onBoard.String() {
			continue
		}
		if len(c.ZoneAbilities) != 0 {
			t.Errorf("a graveyard ability on a permanent: %+v", c.ZoneAbilities)
		}
		if len(c.ActivatedAbilities) != 0 {
			t.Errorf("a graveyard ability in activated_abilities: %+v", c.ActivatedAbilities)
		}
	}
	c := graveyardCardView(t, v, 0, inGrave)
	if len(c.ZoneAbilities) != 0 {
		t.Errorf("a battlefield ability on a graveyard card: %+v", c.ZoneAbilities)
	}
}
