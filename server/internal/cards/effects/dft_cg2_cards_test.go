package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_cg2_cards_test.go — Aetherdrift (DFT), cards from C to G, second
// batch (#4464). Each test plays the real catalog entry and asserts what
// its text says. The fixtures are the other dft_*_cards_test.go files'.

const (
	dftCG2CarrionOracle     = "98a61659-9882-417c-9c16-16ebb26f9037"
	dftCG2ChitinOracle      = "acac15c8-877d-49d4-9b42-90181c05a684"
	dftCG2SkycycleOracle    = "f7467fce-61e4-401a-8a3d-b0a2bc347088"
	dftCG2CollisionOracle   = "49ca0ca8-0733-4aa9-bd0c-1c661725b81b"
	dftCG2DetentionOracle   = "0ea2c5ad-2982-4041-8c9d-a1d31ced11aa"
	dftCG2EarthrumblerOracl = "ba5982a9-e8c1-4e7c-9a73-c4caad0f7336"
	dftCG2FloodOracle       = "e16e2199-d821-49fd-94a0-30ecebd8bd4c"
	dftCG2BlockbusterOracle = "3b601721-90ca-45d3-9a3a-19c719799bfa"
	dftCG2ThrillrollerOracl = "86d67952-19b8-4bb8-b7d7-b2751e86591c"
	dftCG2GreasewrenchOracl = "952300de-766e-4ceb-bd36-e0b721bebc17"
	dftCG2OptimizerOracle   = "045488b3-0206-4dcd-91ab-59ec542f06c0"
	dftCG2SynergistOracle   = "c0bc3704-9bf9-4039-b3f8-205c1e5eb6e9"
)

func TestDftCG2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftCG2CarrionOracle:     "Carrion Cruiser",
		dftCG2ChitinOracle:      "Chitin Gravestalker",
		dftCG2SkycycleOracle:    "Cloudspire Skycycle",
		dftCG2CollisionOracle:   "Collision Course",
		dftCG2DetentionOracle:   "Detention Chariot",
		dftCG2EarthrumblerOracl: "Earthrumbler",
		dftCG2FloodOracle:       "Flood the Engine",
		dftCG2BlockbusterOracle: "Gastal Blockbuster",
		dftCG2ThrillrollerOracl: "Gastal Thrillroller",
		dftCG2GreasewrenchOracl: "Greasewrench Goblin",
		dftCG2OptimizerOracle:   "Guidelight Optimizer",
		dftCG2SynergistOracle:   "Guidelight Synergist",
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

// dftCG2Tapped reports whether a battlefield permanent is tapped.
func dftCG2Tapped(g *game.Game, id uuid.UUID) bool {
	c, ok := battlefieldCard(g, id)
	return ok && c.Tapped
}

// --- Carrion Cruiser --------------------------------------------------------------

func TestDftCG2CarrionCruiserMillsTwoThenReturnsTheChosenCreatureOrVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	vehicle := dftABInGraveyard(me, "Old Vehicle", dftABVehicle)
	dftABInGraveyard(me, "Opt", "Instant")
	milled := dftMPLibraryTop(me,
		game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"})
	cruiser := dftABCast(t, g, me, game.Card{Name: "Carrion Cruiser", TypeLine: dftABVehicle,
		OracleID: dftCG2CarrionOracle, Power: 3, Toughness: 2})
	if !g.Battlefield.Contains(cruiser) {
		t.Fatal("the Cruiser did not enter")
	}
	for _, id := range milled {
		if !me.Graveyard.Contains(id) {
			t.Errorf("milled card %s is not in the graveyard", id)
		}
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no return pick with two creature or Vehicle cards in the graveyard")
	}
	if len(c.ChooseCards) != 2 || c.ChooseMin != 1 {
		t.Fatalf("the pick offers %d cards with floor %d, want the Bear and the Vehicle with floor 1", len(c.ChooseCards), c.ChooseMin)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{vehicle}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(vehicle) {
		t.Error("the Vehicle card did not return to the hand")
	}
	if !me.Graveyard.Contains(milled[0]) {
		t.Error("the Bear left the graveyard too")
	}
	if c, _ := battlefieldCard(g, cruiser); c.IsCreature() {
		t.Error("the uncrewed Cruiser is a creature")
	}
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	dftABCrew(t, g, me, cruiser, 0, one)
}

func TestDftCG2CarrionCruiserReturnsALoneCandidateWithoutAsking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	milled := dftMPLibraryTop(me,
		game.Card{Name: "Milled Opt", TypeLine: "Instant"},
		game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	dftABCast(t, g, me, game.Card{Name: "Carrion Cruiser", TypeLine: dftABVehicle,
		OracleID: dftCG2CarrionOracle, Power: 3, Toughness: 2})
	if c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); c != nil {
		t.Fatalf("asked a pick with one candidate: %+v", c.ChooseCards)
	}
	if !me.Hand.Contains(milled[1]) {
		t.Error("the milled Bear did not return to the hand")
	}
	if !me.Graveyard.Contains(milled[0]) {
		t.Error("the instant left the graveyard")
	}
}

// --- Chitin Gravestalker -----------------------------------------------------------

func TestDftCG2ChitinGravestalkerCostsLessPerArtifactOrCreatureCardAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	const typeLine = "Creature — Insect Warrior"
	if got := selfPricedMV(t, g, me, dftCG2ChitinOracle, typeLine, "{5}{B}"); got != 6 {
		t.Errorf("with an empty graveyard it costs mana value %d, want 6", got)
	}
	dftABInGraveyard(me, "Ornithopter", "Artifact Creature — Thopter")
	dftABInGraveyard(me, "Sol Ring", "Artifact")
	dftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	dftABInGraveyard(me, "Opt", "Instant")
	dftABInGraveyard(me, "Forest", "Basic Land — Forest")
	if got := selfPricedMV(t, g, me, dftCG2ChitinOracle, typeLine, "{5}{B}"); got != 3 {
		t.Errorf("with three artifact and/or creature cards it costs mana value %d, want 3 ({2}{B})", got)
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 4; i++ {
		dftABInGraveyard(opp, "Their Bear", "Creature — Bear")
	}
	if got := selfPricedMV(t, g, me, dftCG2ChitinOracle, typeLine, "{5}{B}"); got != 3 {
		t.Errorf("an opponent's graveyard changed the cost to mana value %d, want 3", got)
	}
	for i := 0; i < 3; i++ {
		dftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	}
	if got := selfPricedMV(t, g, me, dftCG2ChitinOracle, typeLine, "{5}{B}"); got != 1 {
		t.Errorf("with six cards it costs mana value %d, want 1 ({B} is never reduced)", got)
	}
	dftSTCycles(t, "Chitin Gravestalker", typeLine, dftCG2ChitinOracle, 0)
}

// --- Cloudspire Skycycle -----------------------------------------------------------

func TestDftCG2CloudspireSkycycleDistributesTwoCountersAmongOtherCreaturesAndVehicles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := dftABPush(g, me.ID, "My Vehicle", dftABVehicle, "", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	skycycle := dftABCast(t, g, me, game.Card{Name: "Cloudspire Skycycle", TypeLine: dftABVehicle,
		OracleID: dftCG2SkycycleOracle, Power: 2, Toughness: 3})
	if !hasEffectiveKeyword(t, g, skycycle, "flying") {
		t.Error("Cloudspire Skycycle has no flying")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no targets")
	}
	for name, split := range map[string]map[uuid.UUID]int{
		"the Skycycle itself":    {skycycle: 1, bear: 1},
		"an opponent's creature": {theirs: 1, bear: 1},
	} {
		var ids []uuid.UUID
		for id := range split {
			ids = append(ids, id)
		}
		if err := g.ResolvePickTargetsDivided(p.ID, me.ID, cardRefs(ids...), split); err == nil {
			t.Fatalf("the split took %s", name)
		}
	}
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{vehicle: 1, bear: 1}, vehicle, bear)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{vehicle, bear} {
		if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%d +1/+1 counters on %s, want 1", got, id)
		}
	}
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	dftABCrew(t, g, me, skycycle, 0, one)
}

func TestDftCG2CloudspireSkycyclePutsBothCountersOnOneTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	dftABCast(t, g, me, game.Card{Name: "Cloudspire Skycycle", TypeLine: dftABVehicle,
		OracleID: dftCG2SkycycleOracle, Power: 2, Toughness: 3})
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{bear: 2}, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters on the Bear, want both", got)
	}
}

// --- Collision Course --------------------------------------------------------------

func TestDftCG2CollisionCourseDealsDamageEqualToYourCreaturesAndVehicles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	dftABPush(g, me.ID, "My Vehicle", dftABVehicle, "", 3, 3)
	pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	b12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 7)
	dftCGCast(t, g, me, game.Card{Name: "Collision Course", TypeLine: "Sorcery", OracleID: dftCG2CollisionOracle},
		[]int{0}, dftABCardRef(wall))
	if got := damageOn(t, g, wall); got != 3 {
		t.Errorf("%d damage on the Wall, want 3 (two creatures and a Vehicle)", got)
	}
}

func TestDftCG2CollisionCourseDestroysAnArtifactOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := dftCGCastErr(g, me, game.Card{Name: "Collision Course", TypeLine: "Sorcery", OracleID: dftCG2CollisionOracle},
		[]int{1}, dftABCardRef(bear)); err == nil {
		t.Fatal("the destroy bullet took a creature that is not an artifact")
	}
	dftCGCast(t, g, me, game.Card{Name: "Collision Course", TypeLine: "Sorcery", OracleID: dftCG2CollisionOracle},
		[]int{1}, dftABCardRef(rock))
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

// --- Detention Chariot -------------------------------------------------------------

func TestDftCG2DetentionChariotExilesUntilItLeavesAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := pushPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	advanceToMain(t, g)
	chariot := dftABCast(t, g, me, game.Card{Name: "Detention Chariot", TypeLine: dftABVehicle,
		OracleID: dftCG2DetentionOracle, Power: 6, Toughness: 6})
	dftCGPickRefused(t, g, me.ID, mine, land)
	pickTriggerTarget(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) || g.Battlefield.Contains(rock) {
		t.Fatal("their artifact was not exiled")
	}
	three := b12Creature(g, me.ID, "Three", "Creature — Giant", 3, 3)
	dftABCrew(t, g, me, chariot, 0, three)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(chariot) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Signet", opp.ID); n != 1 {
		t.Errorf("after the Chariot left, %d of the exiled artifact under its owner's control, want 1", n)
	}
	dftSTCyclesFor(t, "Detention Chariot", dftABVehicle, dftCG2DetentionOracle, 1, "W")
}

// dftSTCyclesFor is dftSTCycles for a cycling cost of one coloured mana:
// colourless mana is refused, and the colour pays it.
func dftSTCyclesFor(t *testing.T, name, typeLine, oracle string, idx int, color string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Drawn")
	id := pushCatalogHandCard(me, name, typeLine, oracle)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatalf("%s cycled for colourless mana", name)
	}
	fillPoolColored(me, color, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("%s: cycle: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) || me.Hand.Size() != hand {
		t.Errorf("%s: cycling did not discard it for a card: graveyard %v, hand %d want %d",
			name, me.Graveyard.Contains(id), me.Hand.Size(), hand)
	}
}

// --- Earthrumbler ------------------------------------------------------------------

func TestDftCG2EarthrumblerAnimatesByExilingAnArtifactOrCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rumbler := dftABPush(g, me.ID, "Earthrumbler", dftABVehicle, dftCG2EarthrumblerOracl, 7, 6)
	for _, kw := range []string{"vigilance", "trample"} {
		if !hasEffectiveKeyword(t, g, rumbler, kw) {
			t.Errorf("Earthrumbler has no %s", kw)
		}
	}
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	opt := dftABInGraveyard(me, "Opt", "Instant")
	rock := dftABInGraveyard(me, "Sol Ring", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, rumbler, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{opt}}); err == nil {
		t.Fatal("an instant card paid the exile cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, rumbler, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(rock) {
		t.Error("the artifact card was not exiled as the cost")
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, rumbler); !ok || !c.IsCreature() {
		t.Fatal("Earthrumbler is not a creature after its ability resolved")
	}
	if p, tough := effectivePTOf(t, g, rumbler); p != 7 || tough != 6 {
		t.Errorf("Earthrumbler is %d/%d, want 7/6", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, rumbler, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{opt}}); err == nil {
		t.Error("activated with no artifact or creature card left in the graveyard")
	}
	dftMPPassTurn(t, g)
	if c, _ := battlefieldCard(g, rumbler); c.IsCreature() {
		t.Error("the animation outlasted the turn")
	}
}

func TestDftCG2EarthrumblerCrewsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rumbler := dftABPush(g, me.ID, "Earthrumbler", dftABVehicle, dftCG2EarthrumblerOracl, 7, 6)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, rumbler, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	dftABCrew(t, g, me, rumbler, 1, two, one)
}

// --- Flood the Engine --------------------------------------------------------------

func TestDftCG2FloodTheEngineTapsAndSilencesAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// A real catalog Vehicle with trample and crew 2.
	beetle := dftABPush(g, opp.ID, "Debris Beetle", dftABVehicle, dftCGBeetleOracle, 6, 6)
	land := pushPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Flood the Engine", "Enchantment — Aura", dftCG2FloodOracle, dftABCardRef(land)); err == nil {
		t.Fatal("a land was a legal target for enchant creature or Vehicle")
	}
	castCatalogSpell(t, g, "Flood the Engine", "Enchantment — Aura", dftCG2FloodOracle, dftABCardRef(beetle))
	passPriorityAroundTable(t, g)
	if !dftCG2Tapped(g, beetle) {
		t.Error("the enchanted Vehicle is not tapped")
	}
	if hasEffectiveKeyword(t, g, beetle, "trample") {
		t.Error("the enchanted Vehicle kept trample")
	}
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(opp.ID, beetle, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{a, b}}); err == nil {
		t.Error("the enchanted Vehicle could still be crewed")
	}
}

func TestDftCG2FloodTheEngineKeepsACreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Flood the Engine", "Enchantment — Aura", dftCG2FloodOracle, dftABCardRef(bear))
	passPriorityAroundTable(t, g)
	if !dftCG2Tapped(g, bear) {
		t.Fatal("the enchanted creature is not tapped")
	}
	dftMPPassTurn(t, g)
	if g.Seats[g.Turn.ActiveSeat].ID != opp.ID {
		t.Fatalf("the next turn is not the creature's controller's: seat %d", g.Turn.ActiveSeat)
	}
	advanceToMain(t, g)
	if !dftCG2Tapped(g, bear) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Gastal Blockbuster ------------------------------------------------------------

func TestDftCG2GastalBlockbusterSacrificesAVehicleToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := dftABPush(g, me.ID, "My Vehicle", dftABVehicle, "", 3, 3)
	myRock := pushPermanent(g, me.ID, game.Card{Name: "My Signet", TypeLine: "Artifact"})
	rock := pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	advanceToMain(t, g)
	blockbuster := dftABCast(t, g, me, game.Card{Name: "Gastal Blockbuster", TypeLine: "Creature — Human Berserker",
		OracleID: dftCG2BlockbusterOracle, Power: 3, Toughness: 2})
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[vehicle] || !offered[blockbuster] || offered[myRock] {
		t.Errorf("offered %v: want the Vehicle and the Blockbuster, not the plain artifact", c.ChooseCards)
	}
	answerOwnPermanents(t, g, me.ID, vehicle)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(vehicle) {
		t.Fatal("the Vehicle was not sacrificed before the reflexive trigger")
	}
	dftCGPickRefused(t, g, me.ID, myRock)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("their artifact was not destroyed")
	}
	if !g.Battlefield.Contains(myRock) || !g.Battlefield.Contains(blockbuster) {
		t.Error("something else left the battlefield")
	}
}

func TestDftCG2GastalBlockbusterMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	advanceToMain(t, g)
	dftABCast(t, g, me, game.Card{Name: "Gastal Blockbuster", TypeLine: "Creature — Human Berserker",
		OracleID: dftCG2BlockbusterOracle, Power: 3, Toughness: 2})
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("declining changed the battlefield")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no sacrifice, no reflexive trigger")
	}
}

// --- Gastal Thrillroller -----------------------------------------------------------

func TestDftCG2GastalThrillrollerEntersAsAHastyCreatureForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	roller := dftABCast(t, g, me, game.Card{Name: "Gastal Thrillroller", TypeLine: dftABVehicle,
		OracleID: dftCG2ThrillrollerOracl, Power: 4, Toughness: 2})
	if c, ok := battlefieldCard(g, roller); !ok || !c.IsCreature() {
		t.Fatal("the Thrillroller is not a creature the turn it entered")
	}
	for _, kw := range []string{"trample", "haste"} {
		if !hasEffectiveKeyword(t, g, roller, kw) {
			t.Errorf("the Thrillroller has no %s", kw)
		}
	}
	dftMPPassTurn(t, g)
	if c, _ := battlefieldCard(g, roller); c.IsCreature() {
		t.Error("the Thrillroller is still a creature after the turn ended")
	}
}

func TestDftCG2GastalThrillrollerReturnsFromTheGraveyardWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	roller := pushCatalogGraveyardCard(me, "Gastal Thrillroller", dftABVehicle, dftCG2ThrillrollerOracl, 4, 2)
	fodder := handCardForTest(me, "Fodder", "Instant", "")
	pay := func() {
		fillPool(me, 2)
		fillPoolColored(me, "R", 1)
	}
	pay()
	if err := g.ActivateCatalogAbility(me.ID, roller, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	if err := g.ActivateCatalogAbility(me.ID, roller, 1, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the card was not discarded as the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(roller) {
		t.Fatal("the Thrillroller did not return")
	}
	if got := countersOn(g, roller, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	if c, _ := battlefieldCard(g, roller); !c.IsCreature() {
		t.Error("the returned Thrillroller's enters trigger did not animate it")
	}
}

// --- Greasewrench Goblin -----------------------------------------------------------

func TestDftCG2GreasewrenchGoblinRummagesUpToTwoThenGrowsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := dftABPush(g, me.ID, "Greasewrench Goblin", "Creature — Goblin Artificer", dftCG2GreasewrenchOracl, 2, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a := handCardForTest(me, "A", "Instant", "")
	b := handCardForTest(me, "B", "Instant", "")
	handCardForTest(me, "C", "Instant", "")
	seedLibrary(me, "Drawn 1", "Drawn 2", "Drawn 3")
	fillPool(me, 2)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if got := countersOn(g, goblin, game.CounterPlusOne); got != 0 {
		t.Errorf("the counter came before the discard was chosen: %d", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the chosen cards were not discarded")
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand is %d, want 3 (two discarded, two drawn)", got)
	}
	if got := countersOn(g, goblin, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	fillPool(me, 2)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Guidelight Optimizer ----------------------------------------------------------

func TestDftCG2GuidelightOptimizerManaPaysArtifactSpellsAndAbilitiesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opt := pushCatalogPermanent(g, me.ID, "Guidelight Optimizer", "Artifact Creature — Robot", dftCG2OptimizerOracle, false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, opt, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Optimizer: %v", err)
	}
	cost, _ := game.ParseCost("{U}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("refused an artifact spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("refused an ability of a nonartifact source")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a nonartifact spell")
	}
}

// --- Guidelight Synergist ----------------------------------------------------------

func TestDftCG2GuidelightSynergistGetsPlusOnePerArtifactYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	synergist := dftABPush(g, me.ID, "Guidelight Synergist", "Artifact Creature — Robot Artificer", dftCG2SynergistOracle, 0, 4)
	if !hasEffectiveKeyword(t, g, synergist, "flying") {
		t.Error("Guidelight Synergist has no flying")
	}
	if p, tough := effectivePTOf(t, g, synergist); p != 1 || tough != 4 {
		t.Errorf("alone the Synergist is %d/%d, want 1/4 (it counts itself)", p, tough)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	dftABPush(g, me.ID, "My Vehicle", dftABVehicle, "", 3, 3)
	pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectivePTOf(t, g, synergist); p != 3 || tough != 4 {
		t.Errorf("with two more artifacts the Synergist is %d/%d, want 3/4", p, tough)
	}
}
