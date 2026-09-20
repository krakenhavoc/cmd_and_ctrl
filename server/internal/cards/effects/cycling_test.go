package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cycling_test.go — the catalog half of #660. The engine half (the
// zone dimension, the cost components, the commander case) is in
// game/hand_ability_test.go; what is pinned here is that the six
// cards that used to declare "cycling isn't implemented" now do it,
// that typecycling searches rather than draws, and that Fauna Shaman
// pays a real discard for its tutor.

const (
	raffinesTowerOracle      = "6e9ef5ef-6aed-4d3e-a59b-9e3dc8740b1b"
	ketriaTriomeOracle       = "6bae00e8-06cf-4ac4-a1cc-757e454109fe"
	scroungingSkyrayOracle   = "3a46d85b-ce1a-4842-a342-92a5bddb1053"
	magmakinArtilleristOracl = "900b9409-9c16-414d-8674-2ea42c2415a1"
	faunaShamanOracle        = "35b8fa77-4e85-418b-b335-cd1af127075c"
)

// pushCatalogHandCard seeds a catalog card into a player's hand.
func pushCatalogHandCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// cycleFromHand seeds `oracle` into the active seat's hand, floats
// `mana` (a produced-mana string, so colourless pips are spelled
// {C}{C}), and activates the card's cycling ability.
func cycleFromHand(t *testing.T, g *game.Game, name, typeLine, oracle, mana string) (uuid.UUID, *game.Player) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogHandCard(me, name, typeLine, oracle)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: cycle: %v", name, err)
	}
	return id, me
}

// Every cycling card in the catalog declares exactly one hand ability
// that discards itself, and none of them declares a cost component
// that needs a permanent.
func TestCyclingCardsDeclareAHandAbility(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana string
	}{
		{"Raffine's Tower", raffinesTowerOracle, "{3}"},
		{"Ketria Triome", ketriaTriomeOracle, "{3}"},
		{"Indatha Triome", indathaTriomeOracle, "{3}"},
		{"Marauding Mako", maraudingMakoOracle, "{2}"},
		{"Scrounging Skyray", scroungingSkyrayOracle, "{2}"},
		{"Magmakin Artillerist", magmakinArtilleristOracl, "{1}{R}"},
		{"Sylvan Reclamation", sylvanReclamationOracle, "{2}"},
	} {
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abilities) != 1 {
			t.Errorf("%s: %d activated abilities, want the one cycling ability", tc.name, len(abilities))
			continue
		}
		ab := abilities[0]
		if !ab.Cycling {
			t.Errorf("%s: not marked as a cycling ability, so EventCycle never fires", tc.name)
		}
		if !ab.Cost.DiscardSelf {
			t.Errorf("%s: cost %+v does not discard the card", tc.name, ab.Cost)
		}
		if ab.Cost.Mana != tc.mana {
			t.Errorf("%s: cycling cost %q, want %q", tc.name, ab.Cost.Mana, tc.mana)
		}
		if !game.AbilityFunctionsFromZone(ab, game.ZoneHand) {
			t.Errorf("%s: the cycling ability does not function from hand", tc.name)
		}
		if game.AbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
			t.Errorf("%s: the cycling ability is offered on the battlefield too", tc.name)
		}
		if why := game.AbilityNeedsPermanentSource(ab.Cost); why != "" {
			t.Errorf("%s: cycling declares %s, which needs a permanent", tc.name, why)
		}
	}
}

// Ketria Triome: cycling {3} bins the land and draws a card. The land
// half still works, which is what the caveat used to be protecting.
func TestKetriaTriomeCyclesForACard(t *testing.T) {
	g := newCatalogGame(t)
	id, me := cycleFromHand(t, g, "Ketria Triome", "Land — Forest Island Mountain", ketriaTriomeOracle, "{C}{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Triome is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}

// Raffine's Tower: the same, one colour wheel over — the assertion the
// Esper tri-land's own caveat used to stand in for.
func TestRaffinesTowerCyclesForACard(t *testing.T) {
	g := newCatalogGame(t)
	id, me := cycleFromHand(t, g, "Raffine's Tower", "Land — Plains Island Swamp", raffinesTowerOracle, "{C}{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Tower is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}

// Marauding Mako and Scrounging Skyray watch discards, and a cycling's
// cost IS a discard — so cycling one feeds the other, which is the
// whole reason the cost goes through the one discard helper.
func TestCyclingFeedsADiscardWatcher(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mako := pushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", maraudingMakoOracle, false)

	id, _ := cycleFromHand(t, g, "Scrounging Skyray", "Creature — Fish Pirate", scroungingSkyrayOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Skyray is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, mako, game.CounterPlusOne); got != 1 {
		t.Errorf("Mako has %d +1/+1 counters, want 1 from the cycling's discard", got)
	}
}

// Magmakin Artillerist: cycling it deals 1 to each opponent through
// the DISCARD trigger. Its own "when you cycle this card" trigger is
// the declared caveat (ADR 0062 Decision 7), so the total is 1 rather
// than 2 — asserted so the day the cycle trigger lands, this test says
// so.
func TestMagmakinArtilleristCyclesForOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Magmakin Artillerist", "Creature — Elemental Pirate", magmakinArtilleristOracl, false)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := opp.Life

	cycleFromHand(t, g, "Magmakin Artillerist", "Creature — Elemental Pirate", magmakinArtilleristOracl, "{C}{R}")
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 1 {
		t.Errorf("opponent lost %d life, want 1 (the discard trigger; the cycle trigger is a declared caveat)", got)
	}
}

// Sylvan Reclamation: basic landcycling searches instead of drawing
// (CR 702.29e), and finds only BASIC land cards — a Triome is a land
// with basic land TYPES, not a basic land card.
func TestSylvanReclamationBasicLandcyclesForABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Ketria Triome", "Land — Forest Island Mountain"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	forest, island := ids[0], ids[1]

	id, _ := cycleFromHand(t, g, "Sylvan Reclamation", "Instant", sylvanReclamationOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the landcycled instant is not in the graveyard")
	}
	handBefore := len(me.Hand.Cards)
	libraryBefore := len(me.Library.Cards)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after landcycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two basics %v %v", len(c.SearchCards), forest, island)
	}
	for _, got := range c.SearchCards {
		if got != forest && got != island {
			t.Errorf("search offers %v, which is not a basic land card", got)
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want the fetched basic", handBefore, got)
	}
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not reach the hand")
	}
	// "then shuffle" — CR 702.29e. One card left the library.
	if got := len(me.Library.Cards); got != libraryBefore-1 {
		t.Errorf("library %d -> %d, want one card taken", libraryBefore, got)
	}
}

// Fauna Shaman: the general discard component. The creature card is
// discarded at ANNOUNCE — so it is in the graveyard while the ability
// is still on the stack — and the search follows at resolution.
func TestFaunaShamanDiscardsACreatureToTutor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	shaman := pushCatalogPermanent(g, me.ID, "Fauna Shaman", "Creature — Elf Shaman", faunaShamanOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	bear := pushCatalogHandCard(me, "Bear", "Creature — Bear", "")
	bolt := pushCatalogHandCard(me, "Bolt", "Instant", "")
	ids := seedSearchLibrary(me,
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Elk", TypeLine: "Creature — Elk"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	wolf, elk := ids[0], ids[1]
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	// The instant is not a legal payment for "a creature card".
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{bolt},
	}); err == nil {
		t.Fatal("an instant paid a \"discard a creature card\" cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the discard cost was not paid at announce")
	}
	if !me.Hand.Contains(bolt) {
		t.Error("the discard took a card the activator did not name")
	}

	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two creature cards %v %v", len(c.SearchCards), wolf, elk)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{wolf}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(wolf) {
		t.Error("the tutored creature did not reach the hand")
	}
}

// WheneverYouCycle is the battlefield-watcher constructor: it fires
// for the cycling player and nobody else.
func TestWheneverYouCycleWatchesTheCyclingPlayer(t *testing.T) {
	trig := WheneverYouCycle("probe", func(*game.Game, *game.StackItem) error { return nil })
	if len(trig.Watches) != 1 || trig.Watches[0] != game.EventCycle {
		t.Fatalf("watches %v, want EventCycle", trig.Watches)
	}
	me, them := uuid.New(), uuid.New()
	source := &game.Card{Controller: me}
	if !trig.AppliesTo(game.Event{Kind: game.EventCycle, Actor: me}, source, game.Characteristic{}, nil) {
		t.Error("did not fire for its own controller's cycling")
	}
	if trig.AppliesTo(game.Event{Kind: game.EventCycle, Actor: them}, source, game.Characteristic{}, nil) {
		t.Error("fired for an opponent's cycling")
	}
}
