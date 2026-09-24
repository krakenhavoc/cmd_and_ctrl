package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// station_ability_test.go — #759: the station ABILITY (CR 702.184a),
// Station(), on #758's tap-another cost, and the four proof cards that
// carry it. The threshold half is station_test.go.

const (
	galvanizingSawshipOracle  = "dfe8f77a-cc26-438b-92ae-2ca7a91f813b"
	uthrosResearchCraftOracle = "e1c9783a-1d1b-40d7-872e-0ca11b229ce6"
	adagiaOracle              = "70d35dbd-1d91-4a2a-a643-6870d168f4f5"
)

// stationBoard seats a station permanent and one creature of the given
// power under the active seat, at its precombat main phase, and
// returns the seat and both IDs.
func stationBoard(t *testing.T, name, typeLine, oracle string, power int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	crew := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Station Hand",
		TypeLine:   "Creature — Human Pilot",
		Power:      power,
		Toughness:  3,
		Owner:      me.ID,
		Controller: me.ID,
	})
	return g, me, ship, crew
}

// stationIndex is the position of the station ability on a card, found
// by its label so a card that declares others first still works.
func stationIndex(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c := layeredCard(t, g, id)
	for i, a := range game.ActivatedAbilitiesForCard(c) {
		if a.Label == StationLabel {
			return i
		}
	}
	t.Fatalf("%s has no station ability", c.Name)
	return -1
}

func station(t *testing.T, g *game.Game, me *game.Player, ship, crew uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, ship, stationIndex(t, g, ship), game.ActivateAbilityParams{TapIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("station: %v", err)
	}
}

// The headline: tap a 4-power creature, get four charge counters; do
// it again with another and the Seriema crosses 7+ and becomes the
// flying 5/5 it prints.
func TestSeriemaStationsToACreature(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", theSeriemaOracle, 4)

	station(t, g, me, ship, crew)
	if !layeredCard(t, g, crew).Tapped {
		t.Error("the station cost did not tap the creature")
	}
	if layeredCard(t, g, ship).Tapped {
		t.Error("station taps ANOTHER creature — the Spacecraft itself stays untapped")
	}
	if got := counterOn(g, ship, game.CounterCharge); got != 0 {
		t.Errorf("charge counters at announce = %d, want 0 — the effect waits for the stack", got)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, game.CounterCharge); got != 4 {
		t.Fatalf("charge counters = %d, want 4 (the tapped creature's power)", got)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("four counters is below 7+")
	}

	second := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Second Hand", TypeLine: "Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	station(t, g, me, ship, second)
	passPriorityAroundTable(t, g)
	at7 := layeredCard(t, g, ship)
	if got := at7.Counters[game.CounterCharge]; got != 7 {
		t.Fatalf("charge counters = %d, want 7", got)
	}
	if !at7.IsCreature() || !game.HasKeyword(&at7, "flying") || at7.CurrentPower() != 5 {
		t.Errorf("at 7+ The Seriema is a 5/5 flying artifact creature; got creature=%v flying=%v power=%d",
			at7.IsCreature(), game.HasKeyword(&at7, "flying"), at7.CurrentPower())
	}
}

// CR 702.184a's "another", and "activate only as a sorcery".
func TestStationRefusesItselfAndInstantSpeed(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Galvanizing Sawship", "Artifact — Spacecraft", galvanizingSawshipOracle, 2)
	idx := stationIndex(t, g, ship)

	if err := g.ActivateCatalogAbility(me.ID, ship, idx, game.ActivateAbilityParams{TapIDs: []uuid.UUID{ship}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("stationing with the Spacecraft itself: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ship, idx, game.ActivateAbilityParams{}); err == nil {
		t.Error("station with no creature named was accepted")
	}

	// Leave the main phase: sorcery timing is shut (CR 307.1).
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ship, idx, game.ActivateAbilityParams{TapIDs: []uuid.UUID{crew}}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("station outside a main phase: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if layeredCard(t, g, crew).Tapped {
		t.Error("a refused station tapped its creature")
	}
}

// #1352, the test #759 wanted first: station is "activate only as a
// sorcery", and a station ability already on the stack means the
// stack is not empty (CR 307.1, CR 405.1). The first activation is an
// ABILITY — no card in the stack zone — which is the half the
// sorcery-speed gate used to miss.
func TestStationRefusesASecondActivationOverTheFirst(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Galvanizing Sawship", "Artifact — Spacecraft", galvanizingSawshipOracle, 2)
	second := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Second Hand", TypeLine: "Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	idx := stationIndex(t, g, ship)
	station(t, g, me, ship, crew)

	err := g.ActivateCatalogAbility(me.ID, ship, idx, game.ActivateAbilityParams{TapIDs: []uuid.UUID{second}})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("station over a station ability on the stack: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if layeredCard(t, g, second).Tapped {
		t.Error("a refused station tapped its creature")
	}

	passPriorityAroundTable(t, g)
	station(t, g, me, ship, second)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, game.CounterCharge); got != 5 {
		t.Errorf("charge counters = %d, want 5 — both stations resolved once the stack was empty", got)
	}
}

// CR 608.2h: the power is read as the ability RESOLVES. A pump in
// response puts more counters on.
func TestStationReadsPowerAsItResolves(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", theSeriemaOracle, 2)
	station(t, g, me, ship, crew)
	if err := g.AddCounter(crew, "+1/+1", 3); err != nil {
		t.Fatalf("pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, game.CounterCharge); got != 5 {
		t.Errorf("charge counters = %d, want 5 — the 2-power creature was pumped to 5 in response", got)
	}
}

// The other half of CR 608.2h: the creature left in response, so its
// power as it last existed on the battlefield is used.
func TestStationUsesLastKnownPowerWhenTheCreatureLeaves(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", theSeriemaOracle, 3)
	station(t, g, me, ship, crew)
	if err := g.SacrificePermanent(me.ID, crew); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, game.CounterCharge); got != 3 {
		t.Errorf("charge counters = %d, want 3 — the sacrificed creature's last-known power", got)
	}
}

// Release notes: "If the tapped creature has negative power, no charge
// counters are put onto or removed from the permanent with station."
func TestStationWithNoPowerChangesNothing(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", theSeriemaOracle, 1)
	if err := g.AddCounter(ship, game.CounterCharge, 4); err != nil {
		t.Fatalf("seed counters: %v", err)
	}
	station(t, g, me, ship, crew)
	if err := g.AddCounter(crew, "-1/-1", 2); err == nil {
		// A 1/3 with two -1/-1 counters is a -1/1: alive, with
		// negative power.
		passPriorityAroundTable(t, g)
	} else {
		t.Fatalf("shrink: %v", err)
	}
	if got := counterOn(g, ship, game.CounterCharge); got != 4 {
		t.Errorf("charge counters = %d, want 4 — negative power neither adds nor removes", got)
	}
}

// A Spacecraft destroyed in response has no "this permanent" to put
// counters on; they must not land on the card in the graveyard.
func TestStationSourceGoneGetsNothing(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Galvanizing Sawship", "Artifact — Spacecraft", galvanizingSawshipOracle, 4)
	station(t, g, me, ship, crew)
	if err := g.SacrificePermanent(me.ID, ship); err != nil {
		t.Fatalf("sacrifice the Spacecraft: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == ship && c.Counters[game.CounterCharge] != 0 {
			t.Errorf("the Spacecraft in the graveyard has %d charge counters", c.Counters[game.CounterCharge])
		}
	}
}

// Galvanizing Sawship: 3+ is a 6/5 with flying AND haste.
func TestGalvanizingSawshipAtThree(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Galvanizing Sawship", "Artifact — Spacecraft", galvanizingSawshipOracle, 3)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 6 || c.CurrentToughness() != 5 {
		t.Errorf("at 3+ the Sawship is a 6/5 creature; got creature=%v %d/%d", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "haste") {
		t.Error("at 3+ the Sawship has flying and haste")
	}
}

// Uthros Research Craft: the 3+ cast trigger does not exist below
// three counters, and above it draws AND adds a charge counter.
func TestUthrosResearchCraftTriggerIsGatedAtThree(t *testing.T) {
	g, me, craft, crew := stationBoard(t, "Uthros Research Craft", "Artifact — Spacecraft", uthrosResearchCraftOracle, 2)
	station(t, g, me, craft, crew)
	passPriorityAroundTable(t, g)

	// castCatalogSpell puts the artifact into hand and casts it, so a
	// cast with no draw leaves the hand where it was.
	hand := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Test Trinket", "Artifact", "uthros-test-trinket", nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("hand = %d after casting at 2 counters, want %d — the 3+ trigger is not live", got, hand)
	}
	if got := counterOn(g, craft, game.CounterCharge); got != 2 {
		t.Errorf("charge counters = %d, want 2", got)
	}

	if err := g.AddCounter(craft, game.CounterCharge, 1); err != nil {
		t.Fatalf("to three: %v", err)
	}
	hand = len(me.Hand.Cards)
	castCatalogSpell(t, g, "Test Trinket", "Artifact", "uthros-test-trinket", nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d after casting at 3 counters, want %d — the trigger draws", got, hand+1)
	}
	if got := counterOn(g, craft, game.CounterCharge); got != 4 {
		t.Errorf("charge counters = %d, want 4 — the trigger puts one on", got)
	}
}

// Uthros Research Craft at 12+: a 0/8 flier that gets +1/+0 for each
// artifact you control, itself included.
func TestUthrosResearchCraftAtTwelve(t *testing.T) {
	g, me, craft, _ := stationBoard(t, "Uthros Research Craft", "Artifact — Spacecraft", uthrosResearchCraftOracle, 1)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.AddCounter(craft, game.CounterCharge, 12); err != nil {
		t.Fatalf("to twelve: %v", err)
	}
	c := layeredCard(t, g, craft)
	if !c.IsCreature() || !game.HasKeyword(&c, "flying") {
		t.Fatal("at 12+ the Craft is a flying creature")
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 8 {
		t.Errorf("P/T = %d/%d, want 2/8 (base 0/8, +1/+0 for each of two artifacts)", c.CurrentPower(), c.CurrentToughness())
	}
}

// Adagia: a Planet is a LAND with station. It enters tapped, taps for
// {W}, stations from another creature, and its 12+ ability does not
// exist below twelve.
func TestAdagiaIsALandWithStation(t *testing.T) {
	g, me, land, crew := stationBoard(t, "Adagia, Windswept Bastion", "Land — Planet", adagiaOracle, 4)
	station(t, g, me, land, crew)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, land, game.CounterCharge); got != 4 {
		t.Fatalf("charge counters = %d, want 4", got)
	}
	c := layeredCard(t, g, land)
	if c.IsCreature() {
		t.Error("a Planet has no P/T box and never becomes a creature")
	}
	if n := len(game.ActivatedAbilitiesForCard(c)); n != 1 {
		t.Errorf("below 12+ Adagia has %d activated abilities, want 1 (station only)", n)
	}
	if err := g.AddCounter(land, game.CounterCharge, 8); err != nil {
		t.Fatalf("to twelve: %v", err)
	}
	if n := len(game.ActivatedAbilitiesForCard(layeredCard(t, g, land))); n != 2 {
		t.Errorf("at 12+ Adagia has %d activated abilities, want 2", n)
	}
}

// Adagia's 12+ copy is LEGENDARY — "except it's legendary".
func TestAdagiaCopiesAnArtifactAsLegendary(t *testing.T) {
	g, me, land, _ := stationBoard(t, "Adagia, Windswept Bastion", "Land — Planet", adagiaOracle, 1)
	relic := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain Relic", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.AddCounter(land, game.CounterCharge, 12); err != nil {
		t.Fatalf("to twelve: %v", err)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, land, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err != nil {
		t.Fatalf("activate the 12+ ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plain Relic" && c.InstanceID != relic {
			copies++
			if !c.IsLegendary() {
				t.Errorf("the copy's type line %q is not legendary", c.TypeLine)
			}
			if !c.IsToken() {
				t.Error("the copy is not a token")
			}
		}
	}
	if copies != 1 {
		t.Errorf("found %d copies, want 1", copies)
	}
}

func TestLegendaryTypeLine(t *testing.T) {
	for in, want := range map[string]string{
		"Token Artifact":           "Legendary Token Artifact",
		"Legendary Token Artifact": "Legendary Token Artifact",
		"":                         "Legendary",
	} {
		if got := legendaryTypeLine(in); got != want {
			t.Errorf("legendaryTypeLine(%q) = %q, want %q", in, got, want)
		}
	}
}

// Plus used to drop TapOthers, so a composed "{T}, tap another
// untapped creature you control" became the bare {T} ability — the
// #259 direction.
func TestPlusCarriesTheTapAnotherComponent(t *testing.T) {
	c := Plus(ManaCost("{1}"), TapCost(), TapAnotherUntapped("another untapped creature you control", Creature()))
	if c.TapOthers.Empty() || !c.Tap || c.Mana != "{1}" {
		t.Errorf("Plus lost a component: %+v", c)
	}
}
