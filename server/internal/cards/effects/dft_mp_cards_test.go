package effects

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_mp_cards_test.go — Aetherdrift (DFT), cards from M to P (#4464).
// Each test plays the real catalog entry and asserts what its text says.
// The fixtures are dft_ab_cards_test.go's and dft_st_cards_test.go's.

const (
	dftMPPathcruiserOracle = "60f20f32-b6db-4ee7-a0aa-535647e6fcd5"
	dftMPOverdriveOracle   = "f0e85b74-12fb-4d63-853e-7a3f41f731b4"
	dftMPGuardianOracle    = "dac33359-0317-4b11-9078-fccfd2321e53"
	dftMPManglerOracle     = "c9f8651a-4118-4d80-902e-60473e81d3a9"
	dftMPKetradonOracle    = "5d957045-acb2-4546-8c19-71511fd497ef"
	dftMPMoltTenderOracle  = "9a39ca64-e1a9-42ef-8efd-6aa3edde3d93"
	dftMPNightMarketOracle = "4cdb9f80-d08f-4986-99c8-573166d66082"
	dftMPThopteristOracle  = "01813091-75e8-431b-ba3f-b1ea1ba20a3b"
	dftMPOozePatrolOracle  = "a9e3e8b0-4276-419d-9326-be56af3c41be"
	dftMPParagonOracle     = "1acb289f-fb7f-4d83-aaef-4e7b9462c593"
	dftMPPactdollOracle    = "e02fff6b-9ee4-474f-97fc-d70d4bc8001b"
	dftMPPedalOracle       = "22196948-9f30-4fcd-b43b-c7f81a0f4f36"
	dftMPPlowThroughOracle = "a5492d28-b1bf-408d-a048-8f200158d6b1"
	dftMPPotholeMoleOracle = "8a8bc537-b69f-441d-a10e-d7f17f2f75bb"
)

func TestDftMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftMPPathcruiserOracle: "Marshals' Pathcruiser",
		dftMPOverdriveOracle:   "Maximum Overdrive",
		dftMPGuardianOracle:    "Memory Guardian",
		dftMPManglerOracle:     "Midnight Mangler",
		dftMPKetradonOracle:    "Migrating Ketradon",
		dftMPMoltTenderOracle:  "Molt Tender",
		dftMPNightMarketOracle: "Night Market",
		dftMPThopteristOracle:  "Nimble Thopterist",
		dftMPOozePatrolOracle:  "Ooze Patrol",
		dftMPParagonOracle:     "Pacesetter Paragon",
		dftMPPactdollOracle:    "Pactdoll Terror",
		dftMPPedalOracle:       "Pedal to the Metal",
		dftMPPlowThroughOracle: "Plow Through",
		dftMPPotholeMoleOracle: "Pothole Mole",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// dftMPLibraryTop puts cards on top of p's library, the first argument
// on top, and returns their ids in the same order.
func dftMPLibraryTop(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

// dftMPPassTurn ends the current turn, which also ends every "until end
// of turn" effect.
func dftMPPassTurn(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
}

// --- Marshals' Pathcruiser ----------------------------------------------------------

func TestDftMPMarshalsPathcruiserFetchesABasicLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	dual := pushLibraryCardForTest(me, game.Card{Name: "Breeding Pool", TypeLine: "Land — Forest Island", Owner: me.ID, Controller: me.ID})
	cruiser := dftABCast(t, g, me, game.Card{Name: "Marshals' Pathcruiser", TypeLine: dftABVehicle,
		OracleID: dftMPPathcruiserOracle, Power: 6, Toughness: 5})
	if !g.Battlefield.Contains(cruiser) {
		t.Fatal("the Pathcruiser did not enter")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search when the Pathcruiser entered")
	}
	if !slices.Contains(c.SearchCards, forest) {
		t.Error("the basic land was not offered")
	}
	if slices.Contains(c.SearchCards, dual) {
		t.Error("a nonbasic land was offered")
	}
	answerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land is not in hand")
	}
	if cr, _ := battlefieldCard(g, cruiser); cr.IsCreature() {
		t.Error("the uncrewed Pathcruiser is a creature")
	}
}

func TestDftMPMarshalsPathcruiserExhaustsIntoAPermanentCreatureWithTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cruiser := dftABPush(g, me.ID, "Marshals' Pathcruiser", dftABVehicle, dftMPPathcruiserOracle, 6, 5)
	advanceToMain(t, g)
	payFive := func() {
		for _, c := range []string{"W", "U", "B", "R", "G"} {
			fillPoolColored(me, c, 1)
		}
	}
	payFive()
	if err := g.ActivateCatalogAbility(me.ID, cruiser, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, cruiser)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("the Pathcruiser is not an artifact creature: %q", c.TypeLine)
	}
	if got := countersOn(g, cruiser, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if p, tough := effectivePTOf(t, g, cruiser); p != 6 || tough != 5 {
		t.Errorf("the Pathcruiser's base is %d/%d, want its printed 6/5", p, tough)
	}
	if got := effectivePowerWithCounters(t, g, cruiser); got != 8 {
		t.Errorf("power with its counters %d, want 8", got)
	}
	dftMPPassTurn(t, g)
	if c, _ := battlefieldCard(g, cruiser); !c.IsCreature() {
		t.Error("the exhausted Pathcruiser stopped being a creature when the turn ended")
	}
	payFive()
	if err := g.ActivateCatalogAbility(me.ID, cruiser, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

func TestDftMPMarshalsPathcruiserCrewsFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cruiser := dftABPush(g, me.ID, "Marshals' Pathcruiser", dftABVehicle, dftMPPathcruiserOracle, 6, 5)
	four := b12Creature(g, me.ID, "Four", "Creature — Beast", 4, 4)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cruiser, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{four}}); err == nil {
		t.Fatal("crewed with total power 4; crew 5 wants 5")
	}
	dftABCrew(t, g, me, cruiser, 1, four, one)
}

// --- Maximum Overdrive --------------------------------------------------------------

func TestDftMPMaximumOverdriveAddsACounterAndGrantsDeathtouchAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Maximum Overdrive", "Instant", dftMPOverdriveOracle, dftABCardRef(bear))
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	for _, kw := range []string{"deathtouch", "indestructible"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the Bear did not gain %s", kw)
		}
	}
	dftMPPassTurn(t, g)
	for _, kw := range []string{"deathtouch", "indestructible"} {
		if hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the Bear still has %s after the turn ended", kw)
		}
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter did not stay: %d", got)
	}
}

// --- Memory Guardian ----------------------------------------------------------------

func TestDftMPMemoryGuardianHasAffinityForArtifactsAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	pushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	const typeLine = "Artifact Creature — Robot Artificer"
	if got := selfPricedMV(t, g, me, dftMPGuardianOracle, typeLine, "{4}{U}"); got != 2 {
		t.Errorf("with three artifacts it costs mana value %d, want 2 ({1}{U})", got)
	}
	for i := 0; i < 3; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	if got := selfPricedMV(t, g, me, dftMPGuardianOracle, typeLine, "{4}{U}"); got != 1 {
		t.Errorf("with six artifacts it costs mana value %d, want 1 ({U} is never reduced)", got)
	}
	guardian := dftABPush(g, me.ID, "Memory Guardian", typeLine, dftMPGuardianOracle, 3, 4)
	if !hasEffectiveKeyword(t, g, guardian, "flying") {
		t.Error("Memory Guardian has no flying")
	}
}

// --- Midnight Mangler ---------------------------------------------------------------

func TestDftMPMidnightManglerIsACreatureOnlyOnOtherPlayersTurns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mangler := dftABPush(g, me.ID, "Midnight Mangler", dftABVehicle, dftMPManglerOracle, 3, 3)
	advanceToMain(t, g)
	if c, _ := battlefieldCard(g, mangler); c.IsCreature() {
		t.Fatal("the Mangler is a creature on its controller's own turn")
	}
	dftMPPassTurn(t, g)
	if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		t.Fatal("the turn did not pass")
	}
	c, _ := battlefieldCard(g, mangler)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("the Mangler is not an artifact creature on an opponent's turn: %q", c.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, mangler); p != 3 || tough != 3 {
		t.Errorf("the Mangler is %d/%d, want its printed 3/3", p, tough)
	}
}

func TestDftMPMidnightManglerCrewsTwoOnItsOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mangler := dftABPush(g, me.ID, "Midnight Mangler", dftABVehicle, dftMPManglerOracle, 3, 3)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, mangler, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("crewed with total power 1; crew 2 wants 2")
	}
	dftABCrew(t, g, me, mangler, 0, two)
}

// --- Migrating Ketradon -------------------------------------------------------------

func TestDftMPMigratingKetradonGainsFourLifeAndHasReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	life := me.Life
	ketradon := dftABCast(t, g, me, game.Card{Name: "Migrating Ketradon", TypeLine: "Creature — Dinosaur",
		OracleID: dftMPKetradonOracle, Power: 6, Toughness: 6})
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
	if !hasEffectiveKeyword(t, g, ketradon, "reach") {
		t.Error("the Ketradon has no reach")
	}
}

func TestDftMPMigratingKetradonCycles(t *testing.T) {
	dftSTCycles(t, "Migrating Ketradon", "Creature — Dinosaur", dftMPKetradonOracle, 0)
}

// --- Molt Tender --------------------------------------------------------------------

func TestDftMPMoltTenderMillsThenTapsForAnyColorByExilingAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tender := dftABPush(g, me.ID, "Molt Tender", "Creature — Insect Druid", dftMPMoltTenderOracle, 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	top := dftMPLibraryTop(me, game.Card{Name: "Top Card", TypeLine: "Sorcery"})[0]

	// With an empty graveyard the mana ability's cost cannot be paid.
	if err := g.ActivateManaAbility(me.ID, tender, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Fatal("tapped for mana without exiling a card")
	}
	if err := g.ActivateCatalogAbility(me.ID, tender, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("mill: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Fatal("the top card of the library was not milled")
	}

	untap(g, tender)
	if err := g.ActivateManaAbility(me.ID, tender, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{top}, Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want [U]", got)
	}
	if !g.Exile.Contains(top) {
		t.Error("the graveyard card was not exiled to pay the cost")
	}
	if c, _ := battlefieldCard(g, tender); !c.Tapped {
		t.Error("the Tender did not tap")
	}
}

// --- Night Market -------------------------------------------------------------------

func TestDftMPNightMarketEntersTappedAndTapsForTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	market := b13PlayAs(t, g, 0, "Night Market", "Land", dftMPNightMarketOracle)
	if !b13Tapped(t, g, market) {
		t.Error("Night Market entered untapped")
	}
	prompt := pendingOfKind(g, game.PendingChoiceColor)
	if prompt == nil {
		t.Fatal("no colour prompt")
	}
	if !reflect.DeepEqual(prompt.ColorOptions, game.AllColors) {
		t.Errorf("options = %v, want every colour", prompt.ColorOptions)
	}
	answerColor(t, g, me.ID, "B")
	untap(g, market)
	if err := g.ActivateManaAbility(me.ID, market, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool = %v, want [B]", got)
	}
}

func TestDftMPNightMarketCyclesForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Drawn")
	id := pushCatalogHandCard(me, "Night Market", "Land", dftMPNightMarketOracle)
	hand := me.Hand.Size()
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("cycled for two; cycling costs {3}")
	}
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("Night Market was not discarded to cycle")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d after cycling, want %d", got, hand)
	}
}

// --- Nimble Thopterist and Pactdoll Terror ------------------------------------------

func TestDftMPNimbleThopteristMakesAFlyingThopter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	dftABCast(t, g, me, game.Card{Name: "Nimble Thopterist", TypeLine: "Creature — Vedalken Artificer",
		OracleID: dftMPThopteristOracle, Power: 3, Toughness: 2})
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && c.Controller == me.ID {
			thopter = c.InstanceID
		}
	}
	if thopter == uuid.Nil {
		t.Fatal("no Thopter token")
	}
	c, _ := battlefieldCard(g, thopter)
	if !c.IsArtifact() || !c.IsCreature() || !c.IsToken() {
		t.Errorf("the Thopter is not an artifact creature token: %q", c.TypeLine)
	}
	if !hasEffectiveKeyword(t, g, thopter, "flying") {
		t.Error("the Thopter has no flying")
	}
	if p, tough := effectivePTOf(t, g, thopter); p != 1 || tough != 1 {
		t.Errorf("the Thopter is %d/%d, want 1/1", p, tough)
	}
}

func TestDftMPPactdollTerrorDrainsForItselfAndEachOtherArtifactYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	lives, life := lifeOfOpponents(g), me.Life
	dftABCast(t, g, me, game.Card{Name: "Pactdoll Terror", TypeLine: "Artifact Creature — Toy",
		OracleID: dftMPPactdollOracle, Power: 3, Toughness: 4})
	for i, now := range lifeOfOpponents(g) {
		if now != lives[i]-1 {
			t.Errorf("opponent %d: life %d after the Terror entered, want %d", i, now, lives[i]-1)
		}
	}
	if me.Life != life+1 {
		t.Errorf("life %d after the Terror entered, want %d", me.Life, life+1)
	}

	// The Thopterist is not an artifact; its Thopter token is.
	dftABCast(t, g, me, game.Card{Name: "Nimble Thopterist", TypeLine: "Creature — Vedalken Artificer",
		OracleID: dftMPThopteristOracle, Power: 3, Toughness: 2})
	passPriorityAroundTable(t, g)
	for i, now := range lifeOfOpponents(g) {
		if now != lives[i]-2 {
			t.Errorf("opponent %d: life %d after the Thopter entered, want %d (the creature spell does not count)", i, now, lives[i]-2)
		}
	}
	if me.Life != life+2 {
		t.Errorf("life %d after the Thopter entered, want %d", me.Life, life+2)
	}
}

func TestDftMPPactdollTerrorIgnoresAnOpponentsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dftABPush(g, me.ID, "Pactdoll Terror", "Artifact Creature — Toy", dftMPPactdollOracle, 3, 4)
	dftMPPassTurn(t, g)
	opp := g.Seats[g.Turn.ActiveSeat]
	if opp.ID == me.ID {
		t.Fatal("the turn did not pass")
	}
	advanceToMain(t, g)
	life := me.Life
	dftABCast(t, g, opp, game.Card{Name: "Nimble Thopterist", TypeLine: "Creature — Vedalken Artificer",
		OracleID: dftMPThopteristOracle, Power: 3, Toughness: 2})
	passPriorityAroundTable(t, g)
	if countNamed(g, "Thopter") != 1 {
		t.Fatal("the opponent's Thopter was not made")
	}
	if me.Life != life {
		t.Errorf("an opponent's artifact entering changed my life %d -> %d", life, me.Life)
	}
}

// --- Ooze Patrol --------------------------------------------------------------------

func TestDftMPOozePatrolMillsTwoThenCountsArtifactAndCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	dftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	dftABInGraveyard(me, "Opt", "Instant")
	milled := dftMPLibraryTop(me,
		game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
		game.Card{Name: "Third Card", TypeLine: "Creature — Elf"})
	ooze := dftABCast(t, g, me, game.Card{Name: "Ooze Patrol", TypeLine: "Creature — Ooze",
		OracleID: dftMPOozePatrolOracle, Power: 2, Toughness: 2})
	for _, id := range milled[:2] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s was not milled", id)
		}
	}
	if me.Graveyard.Contains(milled[2]) {
		t.Error("a third card was milled")
	}
	// Bears, Ornithopter (once) and Sol Ring.
	if got := countersOn(g, ooze, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
}

// --- Pacesetter Paragon -------------------------------------------------------------

func TestDftMPPacesetterParagonExhaustsForACounterAndDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	paragon := dftABPush(g, me.ID, "Pacesetter Paragon", "Creature — Human Pilot", dftMPParagonOracle, 2, 3)
	advanceToMain(t, g)
	fillPool(me, 2)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, paragon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, paragon, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, paragon, "double strike") {
		t.Error("the Paragon did not gain double strike")
	}
	fillPool(me, 2)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, paragon, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
	dftMPPassTurn(t, g)
	if hasEffectiveKeyword(t, g, paragon, "double strike") {
		t.Error("double strike outlasted the turn")
	}
	if got := countersOn(g, paragon, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter did not stay: %d", got)
	}
}

// --- Pedal to the Metal -------------------------------------------------------------

func TestDftMPPedalToTheMetalPumpsByXAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Pedal to the Metal", "Instant", dftMPPedalOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Targets: dftABCardRef(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear did not gain first strike")
	}
}

// --- Plow Through -------------------------------------------------------------------

func TestDftMPPlowThroughFightsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Beast", 4, 4)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Giant", 1, 5)
	if _, err := castModalErr(t, g, "Plow Through", "Sorcery", dftMPPlowThroughOracle, []int{0}, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: other, Slot: 1},
	}); err == nil {
		t.Fatal("the second target was a creature I control; it must be an opponent's")
	}
	castModal(t, g, "Plow Through", "Sorcery", dftMPPlowThroughOracle, []int{0}, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, theirs); got != 4 {
		t.Errorf("their creature took %d, want 4", got)
	}
	if got := damageMarkedOn(g, mine); got != 1 {
		t.Errorf("my creature took %d, want 1", got)
	}
}

func TestDftMPPlowThroughDestroysAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := b12Permanent(g, opp.ID, "Their Vehicle", dftABVehicle)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if _, err := castModalErr(t, g, "Plow Through", "Sorcery", dftMPPlowThroughOracle, []int{1}, dftABCardRef(bear)); err == nil {
		t.Fatal("a creature that is not a Vehicle was a legal target")
	}
	castModal(t, g, "Plow Through", "Sorcery", dftMPPlowThroughOracle, []int{1}, dftABCardRef(vehicle))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vehicle) {
		t.Error("the Vehicle was not destroyed")
	}
}

// --- Pothole Mole -------------------------------------------------------------------

func TestDftMPPotholeMoleMillsThreeThenMayReturnALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	milled := dftMPLibraryTop(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Opt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	dftABCast(t, g, me, game.Card{Name: "Pothole Mole", TypeLine: "Creature — Mole",
		OracleID: dftMPPotholeMoleOracle, Power: 2, Toughness: 3})
	for _, id := range milled {
		if !me.Graveyard.Contains(id) {
			t.Fatalf("%s was not milled", id)
		}
	}
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick after the mill")
	}
	if c.ChooseMin != 0 || !slices.Equal(c.ChooseCards, []uuid.UUID{milled[0]}) {
		t.Fatalf("pick offers %v (min %d), want only the land, optionally", c.ChooseCards, c.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, milled[0])
	if !me.Hand.Contains(milled[0]) {
		t.Error("the land is not in hand")
	}
}

func TestDftMPPotholeMoleMayReturnNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	land := dftMPLibraryTop(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Opt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})[0]
	dftABCast(t, g, me, game.Card{Name: "Pothole Mole", TypeLine: "Creature — Mole",
		OracleID: dftMPPotholeMoleOracle, Power: 2, Toughness: 3})
	answerChooseCards(t, g, me.ID)
	if !me.Graveyard.Contains(land) || me.Hand.Contains(land) {
		t.Error("declining still returned the land")
	}
}
