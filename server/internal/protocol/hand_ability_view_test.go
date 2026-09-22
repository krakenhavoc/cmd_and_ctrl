package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_ability_view_test.go — the wire half of #660 (ADR 0062
// Decision 5). A hand card's abilities ride `zone_abilities`, a
// permanent's ride `activated_abilities`, neither list leaks the
// other's rows, and the whole thing is stripped from a hand the
// viewer cannot see.

func seatCyclingHandCard(g *game.Game, owner uuid.UUID, hand *game.Zone) uuid.UUID {
	id := uuid.New()
	hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Ketria Triome",
		TypeLine:   "Land — Forest Island Mountain",
		OracleID:   "00000000-0000-0000-0000-0000000000ce",
		Owner:      owner,
		Controller: owner,
		KnownBy:    map[uuid.UUID]bool{owner: true},
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:   "Cycling {2}",
			Cost:    game.AbilityCost{Mana: "{2}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Cycling: true,
		}},
	})
	return id
}

func handCardView(t *testing.T, v GameView, seat int, id uuid.UUID) CardView {
	t.Helper()
	for _, c := range v.Seats[seat].Hand.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("card %s not in seat %d's hand view", id, seat)
	return CardView{}
}

func TestZoneAbilitiesAreProjectedForTheOwner(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := seatCyclingHandCard(g, me.ID, me.Hand)

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := handCardView(t, v, 0, id)
	if len(c.ZoneAbilities) != 1 {
		t.Fatalf("zone_abilities = %+v, want the one cycling row", c.ZoneAbilities)
	}
	ha := c.ZoneAbilities[0]
	if ha.Index != 0 || ha.Label != "Cycling {2}" || ha.ManaCost != "{2}" || !ha.DiscardSelf {
		t.Errorf("hand ability = %+v, want index 0 / Cycling {2} / {2} / discard_self", ha)
	}
	// The battlefield list stays empty on a hand card, so a client
	// that reads only that one offers nothing rather than something
	// wrong.
	if len(c.ActivatedAbilities) != 0 {
		t.Errorf("activated_abilities = %+v on a hand card, want none", c.ActivatedAbilities)
	}
}

// CR 113.6 both ways: the same card sitting on the battlefield offers
// no cycling row, because the ability does not function there.
func TestCyclingIsNotProjectedOnABattlefieldPermanent(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Ketria Triome",
		TypeLine:   "Land — Forest Island Mountain",
		OracleID:   "00000000-0000-0000-0000-0000000000cf",
		Owner:      me.ID,
		Controller: me.ID,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Cycling {2}",
			Cost:  game.AbilityCost{Mana: "{2}", DiscardSelf: true},
			Zones: []game.ZoneKind{game.ZoneHand},
		}},
	})

	for _, c := range ViewOfGame(g).Battlefield.Cards {
		if c.InstanceID != id.String() {
			continue
		}
		if len(c.ActivatedAbilities) != 0 {
			t.Errorf("activated_abilities = %+v on the battlefield, want none", c.ActivatedAbilities)
		}
		if len(c.ZoneAbilities) != 0 {
			t.Errorf("zone_abilities = %+v on the battlefield, want none", c.ZoneAbilities)
		}
		return
	}
	t.Fatal("the permanent is missing from the battlefield view")
}

// A hand is not public. An opponent's view of the same card carries no
// hand abilities — the row would quote the card as loudly as its mana
// cost does.
func TestZoneAbilitiesAreStrippedForOtherSeats(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := seatCyclingHandCard(g, me.ID, me.Hand)

	v := FilterViewFor(ViewOfGame(g), them.ID.String())
	for _, c := range v.Seats[0].Hand.Cards {
		if c.InstanceID != id.String() {
			continue
		}
		if len(c.ZoneAbilities) != 0 {
			t.Errorf("an opponent sees zone_abilities %+v", c.ZoneAbilities)
		}
	}
}

// The general discard component ships its clause and the cards that
// could pay it, so the client can skip its picker when the options
// number exactly N.
func TestActivatedAbilityViewCarriesDiscardCostOptions(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	bear := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Bolt", TypeLine: "Instant",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	src := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: src, Name: "Fauna Shaman", TypeLine: "Creature — Elf Shaman",
		OracleID: "00000000-0000-0000-0000-0000000000d0",
		Owner:    me.ID, Controller: me.ID,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{G}, {T}, Discard a creature card: tutor",
			Cost: game.AbilityCost{
				Mana: "{G}", Tap: true,
				DiscardCards: &game.DiscardCost{
					N: 1, Label: "a creature card",
					Match: func(c game.Card) bool { return c.IsCreature() },
				},
			},
		}},
	})

	for _, c := range ViewOfGame(g).Battlefield.Cards {
		if c.InstanceID != src.String() {
			continue
		}
		if len(c.ActivatedAbilities) != 1 {
			t.Fatalf("activated_abilities = %+v, want one", c.ActivatedAbilities)
		}
		a := c.ActivatedAbilities[0]
		if a.DiscardCostN != 1 || a.DiscardCostLabel != "a creature card" {
			t.Errorf("discard cost = %d %q, want 1 \"a creature card\"", a.DiscardCostN, a.DiscardCostLabel)
		}
		if len(a.DiscardCostOptions) != 1 || a.DiscardCostOptions[0] != bear.String() {
			t.Errorf("discard_cost_options = %v, want only the creature card %v", a.DiscardCostOptions, bear)
		}
		if a.DiscardSelf {
			t.Error("discard_self set on an ability that does not discard its source")
		}
		return
	}
	t.Fatal("the source is missing from the battlefield view")
}
