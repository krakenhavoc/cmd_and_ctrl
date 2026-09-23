package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// other_zone_ability_test.go — the enumerator half of #1221. The
// invariant is #544's, three zones further out than #660 took it:
// offer exactly the activations the engine accepts, from every zone
// an ability can declare, and nothing from the one zone the walk
// deliberately does not visit.

// unearthCard is a graveyard ability with a mana cost and no exile —
// unearth's shape, spelled out here so the legal package's tests need
// no catalog import.
func unearthCard(name, cost string) game.Card {
	return game.Card{
		Name:     name,
		TypeLine: "Creature — Zombie",
		Power:    2, Toughness: 1,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:        "Unearth " + cost,
			Cost:         game.AbilityCost{Mana: cost},
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

// zoneAbilityCard is the same ability declared for an arbitrary zone,
// so one helper covers the exile, command and library cases.
func zoneAbilityCard(name, cost string, zone game.ZoneKind) game.Card {
	return game.Card{
		Name:     name,
		TypeLine: "Creature — Test",
		Power:    1, Toughness: 1,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  string(zone) + " ability " + cost,
			Cost:   game.AbilityCost{Mana: cost},
			Zones:  []game.ZoneKind{zone},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

// A graveyard ability is an enumerated activation when the seat can
// pay for it and no move at all when it cannot — cycling's test one
// zone over, and the move the enumerator offers is one the dispatcher
// accepts.
func TestUnearthFromGraveyardIsEnumeratedOnlyWhenPayable(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	zombie := graveyardCard(active, unearthCard("Dregscape Zombie", "{1}"))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), zombie); len(acts) != 0 {
		t.Fatalf("a seat with no mana: want no unearth move, got %v", labels(acts))
	}

	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, zombie)
	if len(acts) != 1 {
		t.Fatalf("want exactly one unearth move, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// CR 113.6 filters the enumeration in both directions, and the
// graveyard is no different from the hand: the same ability on the
// battlefield is not a move, and a battlefield ability on a card in a
// graveyard is not one either.
func TestGraveyardAbilityZoneFiltersTheEnumeration(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, basic("Forest", "Forest"))

	onBoard := battlefieldCard(g, active, unearthCard("Dregscape Zombie", "{1}"))
	inGraveyard := graveyardCard(active, game.Card{
		Name:     "Sol Ring",
		TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "{2}: do nothing",
			Cost:   game.AbilityCost{Mana: "{2}"},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, onBoard); len(acts) != 0 {
		t.Errorf("a graveyard ability offered on a permanent: %v", labels(acts))
	}
	if acts := activationsOf(moves, inGraveyard); len(acts) != 0 {
		t.Errorf("a battlefield ability offered on a graveyard card: %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// CR 108.4: the enumerator walks a seat's OWN cards. Another seat's
// graveyard card is not this seat's move, and neither is a card
// somebody else owns in the shared exile pile.
func TestAnotherSeatsZoneAbilitiesAreNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	theirs := graveyardCard(other, unearthCard("Dregscape Zombie", "{1}"))
	theirExile := exileCardOwnedBy(g, other, zoneAbilityCard("Exile Test", "{1}", game.ZoneExile))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, theirs); len(acts) != 0 {
		t.Errorf("another seat's graveyard ability offered: %v", labels(acts))
	}
	if acts := activationsOf(moves, theirExile); len(acts) != 0 {
		t.Errorf("another seat's exiled card's ability offered: %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// The walk visits exile and the command zone, and does NOT visit the
// library — CR 401.2 makes it hidden and no printed ability functions
// from one, so an enumeration that read it would be reading cards the
// seat is not entitled to see.
func TestZoneAbilityWalkCoversExileAndCommandButNotTheLibrary(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Island", "Island"))

	inExile := exileCardOwnedBy(g, active, zoneAbilityCard("Exile Test", "{1}", game.ZoneExile))
	inCommand := commandZoneCard(active, zoneAbilityCard("Command Test", "{1}", game.ZoneCommand))
	inLibrary := libraryTopCard(active, zoneAbilityCard("Library Test", "{1}", game.ZoneLibrary))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, inExile); len(acts) != 1 {
		t.Errorf("exile: %d moves, want 1 (%v)", len(acts), labels(acts))
	}
	if acts := activationsOf(moves, inCommand); len(acts) != 1 {
		t.Errorf("command zone: %d moves, want 1 (%v)", len(acts), labels(acts))
	}
	if acts := activationsOf(moves, inLibrary); len(acts) != 0 {
		t.Errorf("library: %v, want none — the walk does not read a hidden zone", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// exileCardOwnedBy drops a face-up card into the shared exile pile,
// known to the whole table the way a face-up exile is.
func exileCardOwnedBy(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	c.KnownBy = make(map[uuid.UUID]bool, len(g.Seats))
	for _, s := range g.Seats {
		c.KnownBy[s.ID] = true
	}
	g.Exile.PushTop(c)
	return c.InstanceID
}

func commandZoneCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Command.PushTop(c)
	return c.InstanceID
}

func libraryTopCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}
