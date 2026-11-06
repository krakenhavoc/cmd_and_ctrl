package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_ab_cards_test.go — Aetherdrift (DFT), the cards from A to B (#4464).
// Each test plays the real catalog entry and asserts what its text says.

const (
	dftABAatchikOracle      = "405c4a7b-dbb3-48a1-9205-7dbe6e2ad363"
	dftABJockeyOracle       = "c3dc6b0e-8790-4bd3-b166-63b2dc77488f"
	dftABAetherjacketOracle = "062c8e40-dea7-4df7-bcaa-f1d696fe61c3"
	dftABAirResponseOracle  = "18a22d9f-81bc-4ed5-bd83-e3f8ec5e38c8"
	dftABVendettaOracle     = "6f053f0d-c9c7-4fcd-b218-a29d1d080104"
	dftABRunnerOracle       = "0ca1f76d-ae77-490a-8b32-d30ad33afc2d"
	dftABBeastriderOracle   = "19a0d191-64c0-4e8a-8cb0-bf680cc681bb"
	dftABBestowOracle       = "58a17110-b2a7-4dbc-bba5-5ed3728ad261"
	dftABSloopOracle        = "cec10abd-2a2d-4d89-9ecc-ea50c2ac1538"
	dftABBounceOffOracle    = "d9c29e6e-5949-4599-bf74-b9539e753f76"
	dftABGearhulkOracle     = "0c7019a7-3cca-481a-9af0-abedb89c15ad"
	dftABRamblerOracle      = "21aaecc7-b9ec-4068-8e47-4dffaa12f102"
	dftABBarrageOracle      = "dbf9feb3-0751-47bd-a8dc-ed1c076cf8c7"
	dftABBrokenWingsOracle  = "5e316864-d55c-496f-8f46-773567896864"
	dftABBroodheartOracle   = "1b759831-abe8-4996-a897-3c6d6d30a1c7"
	dftABBurnerRocketOracle = "0d07eb9f-46f4-466d-acb6-609437c34f13"

	dftABVehicle = "Artifact — Vehicle"
)

func TestDftABEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftABAatchikOracle:      "Aatchik, Emerald Radian",
		dftABJockeyOracle:       "Adrenaline Jockey",
		dftABAetherjacketOracle: "Aetherjacket",
		dftABAirResponseOracle:  "Air Response Unit",
		dftABVendettaOracle:     "Ancient Vendetta",
		dftABRunnerOracle:       "Apocalypse Runner",
		dftABBeastriderOracle:   "Beastrider Vanguard",
		dftABBestowOracle:       "Bestow Greatness",
		dftABSloopOracle:        "Boosted Sloop",
		dftABBounceOffOracle:    "Bounce Off",
		dftABGearhulkOracle:     "Brightglass Gearhulk",
		dftABRamblerOracle:      "Broadcast Rambler",
		dftABBarrageOracle:      "Broadside Barrage",
		dftABBrokenWingsOracle:  "Broken Wings",
		dftABBroodheartOracle:   "Broodheart Engine",
		dftABBurnerRocketOracle: "Burner Rocket",
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

// dftABPush puts a catalog permanent with printed P/T onto the battlefield
// under owner's control, past summoning sickness.
func dftABPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// dftABCast casts a catalog card from the hand with no payment and lets
// it and anything it triggers resolve, stopping at the first prompt.
func dftABCast(t *testing.T, g *game.Game, me *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	return c.InstanceID
}

// dftABCrew crews a Vehicle with the given creatures and resolves it.
func dftABCrew(t *testing.T, g *game.Game, me *game.Player, vehicle uuid.UUID, idx int, crewers ...uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, vehicle, idx, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, vehicle); !ok || !c.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}
}

func dftABCardRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func dftABInGraveyard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Aatchik, Emerald Radian ------------------------------------------------------

func TestDftABAatchikMakesAnInsectPerArtifactOrCreatureCardAndGrowsOnInsectDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	dftABInGraveyard(me, "Ornithopter", "Artifact Creature — Thopter")
	dftABInGraveyard(me, "Sol Ring", "Artifact")
	dftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	dftABInGraveyard(me, "Opt", "Instant")
	dftABInGraveyard(me, "Forest", "Basic Land — Forest")

	aatchik := dftABCast(t, g, me, game.Card{Name: "Aatchik, Emerald Radian", TypeLine: "Legendary Creature — Insect Druid",
		OracleID: dftABAatchikOracle, Power: 3, Toughness: 3})
	if got := countNamed(g, "Insect"); got != 3 {
		t.Fatalf("created %d Insects, want 3 (one per artifact and/or creature card)", got)
	}

	var insect uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Insect" {
			insect = c.InstanceID
			break
		}
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	killCreature(t, g, me.ID, insect)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, aatchik, game.CounterPlusOne); got != 1 {
		t.Errorf("Aatchik has %d +1/+1 counters after an Insect died, want 1", got)
	}
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, want)
		}
	}

	// A creature that is not an Insect does nothing.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, aatchik, game.CounterPlusOne); got != 1 {
		t.Errorf("a Bear's death changed Aatchik's counters to %d", got)
	}
	// Nor does an opponent's Insect.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Insect", "Creature — Insect", 1, 1)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, aatchik, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's Insect dying changed Aatchik's counters to %d", got)
	}
}

// --- Adrenaline Jockey ------------------------------------------------------------

func TestDftABAdrenalineJockeyBurnsOffTurnCastersAndGrowsOnExhaust(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jockey := dftABPush(g, me.ID, "Adrenaline Jockey", "Creature — Minotaur Pilot", dftABJockeyOracle, 3, 3)
	advanceToMain(t, g)
	before := opp.Life
	b29CastAs(t, g, opp, "Opt", "Instant", "", "{U}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("an opponent casting on my turn: life %d, want %d", opp.Life, before-4)
	}
	mine := me.Life
	b24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("casting on my own turn: life %d, want %d", me.Life, mine)
	}

	goblin := pushCatalogPermanent(g, me.ID, "Prowcatcher Specialist", "Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)
	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, jockey, game.CounterPlusOne); got != 1 {
		t.Errorf("the Jockey has %d +1/+1 counters after an exhaust activation, want 1", got)
	}
}

// --- Aetherjacket -----------------------------------------------------------------

func TestDftABAetherjacketSacrificesToDestroyAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jacket := dftABPush(g, me.ID, "Aetherjacket", "Artifact Creature — Thopter", dftABAetherjacketOracle, 2, 1)
	rock := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	advanceToMain(t, g)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, jacket, kw) {
			t.Errorf("Aetherjacket has no %s", kw)
		}
	}

	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, jacket, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(jacket)}); err == nil {
		t.Fatal("Aetherjacket targeted itself; the text says another artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, jacket, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(jacket) {
		t.Error("Aetherjacket was not sacrificed as part of the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the target artifact was not destroyed")
	}
}

// --- Air Response Unit ------------------------------------------------------------

func TestDftABAirResponseUnitCrewsOneAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	unit := dftABPush(g, me.ID, "Air Response Unit", dftABVehicle, dftABAirResponseOracle, 3, 3)
	pilot := b12Creature(g, me.ID, "Pilot", "Creature — Human Pilot", 1, 1)
	advanceToMain(t, g)
	dftABCrew(t, g, me, unit, 0, pilot)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, unit, kw) {
			t.Errorf("the crewed Air Response Unit has no %s", kw)
		}
	}
	if p, tough := effectivePTOf(t, g, unit); p != 3 || tough != 3 {
		t.Errorf("the crewed Unit is %d/%d, want 3/3", p, tough)
	}
}

// --- Ancient Vendetta -------------------------------------------------------------

func TestDftABAncientVendettaExilesUpToFourOfTheNamedCard(t *testing.T) {
	x := extractionSetUp(t, "Lightning Bolt", "Instant")
	x.put("lib 2", x.opp.Library, "Lightning Bolt", "Instant")
	if err := castCatalogSpellErr(t, x.g, "Ancient Vendetta", "Sorcery", dftABVendettaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.me.ID}}); err == nil {
		t.Fatal("Ancient Vendetta targeted its caster; the text says target opponent")
	}
	castCatalogSpell(t, x.g, "Ancient Vendetta", "Sorcery", dftABVendettaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.opp.ID}})
	passPriorityAroundTable(t, x.g)
	actNameFor(t, x.g, x.me.ID, "Lightning Bolt")

	wantKeys(t, x.candidates(t), "gy", "gy 2", "hand", "lib", "lib 2")
	c := searchChoiceFor(x.g, x.me.ID)
	if c.SearchMax != 4 {
		t.Errorf("the search takes up to %d, want 4", c.SearchMax)
	}
	answerSearchByID(t, x.g, x.me.ID, x.cards["gy"], x.cards["gy 2"], x.cards["hand"], x.cards["lib"])
	x.exiled(t, "gy", "gy 2", "hand", "lib")
	x.notExiled(t, "lib 2", "gy other", "hand other", "lib other", "gy forest")
	if x.searches() != 1 {
		t.Errorf("%d searches of their library, want 1", x.searches())
	}
}

func TestDftABAncientVendettaMayFindNothing(t *testing.T) {
	x := extractionSetUp(t, "Lightning Bolt", "Instant")
	castCatalogSpell(t, x.g, "Ancient Vendetta", "Sorcery", dftABVendettaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.opp.ID}})
	passPriorityAroundTable(t, x.g)
	actNameFor(t, x.g, x.me.ID, "Lightning Bolt")
	answerSearchFailToFind(t, x.g, x.me.ID)
	x.notExiled(t, "gy", "gy 2", "hand", "lib")
	if x.searches() != 1 {
		t.Errorf("%d searches, want 1: the library is shuffled whatever was taken", x.searches())
	}
}

// --- Apocalypse Runner ------------------------------------------------------------

func TestDftABApocalypseRunnerMakesASmallCreatureUnblockableWithLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := dftABPush(g, me.ID, "Apocalypse Runner", dftABVehicle, dftABRunnerOracle, 6, 5)
	small := b12Creature(g, me.ID, "Small", "Creature — Rat", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, runner, 0, game.ActivateAbilityParams{Targets: dftABCardRef(big)}); err == nil {
		t.Fatal("a power-3 creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, runner, 0, game.ActivateAbilityParams{Targets: dftABCardRef(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, runner); !c.Tapped {
		t.Error("the Runner did not tap")
	}
	if !hasEffectiveKeyword(t, g, small, "lifelink") {
		t.Error("the target did not gain lifelink")
	}
	brAttack(t, g, small)
	brRefusal(t, g.DeclareBlocker(blocker, small), game.BlockReasonCantBeBlocked)
}

func TestDftABApocalypseRunnerCrewsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	runner := dftABPush(g, me.ID, "Apocalypse Runner", dftABVehicle, dftABRunnerOracle, 6, 5)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, runner, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	dftABCrew(t, g, me, runner, 1, two, one)
}

// --- Beastrider Vanguard ----------------------------------------------------------

func TestDftABBeastriderVanguardTakesAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := dftABPush(g, me.ID, "Beastrider Vanguard", "Creature — Human Knight", dftABBeastriderOracle, 2, 2)
	advanceToMain(t, g)
	ids := seedLibrary(me, "A", "B")
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, "{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 {
		t.Fatalf("no take-one prompt: %+v", pick)
	}
	for _, id := range ids {
		if slices.Contains(pick.ChooseCards, id) {
			t.Errorf("a sorcery card %s was offered; only permanent cards may be taken", id)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("take the Forest: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Fatal("the land card is not in hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryBottomIDs(me, 2); !sameIDs(got, []uuid.UUID{ids[1], ids[0]}) {
		t.Errorf("bottom = %v, want the player's order", got)
	}
}

// --- Bestow Greatness -------------------------------------------------------------

func TestDftABBestowGreatnessPumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Bestow Greatness", "Instant", dftABBestowOracle, dftABCardRef(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}

// --- Boosted Sloop ----------------------------------------------------------------

func TestDftABBoostedSloopLootsWheneverYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sloop := dftABPush(g, me.ID, "Boosted Sloop", dftABVehicle, dftABSloopOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, sloop, "menace") {
		t.Error("Boosted Sloop has no menace")
	}
	a := b12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	seedLibrary(me, "Drawn")
	pitch := handCard(me, "Pitch", "Sorcery")
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard after the attack: the Sloop did not loot")
	}
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d after one loot, want %d (two attackers are one attack)", got, hand)
	}
}

// --- Bounce Off -------------------------------------------------------------------

func TestDftABBounceOffReturnsACreatureOrAnUncrewedVehicle(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := b12Permanent(g, opp.ID, "Their Vehicle", dftABVehicle)
	rock := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	if err := castCatalogSpellErr(t, g, "Bounce Off", "Instant", dftABBounceOffOracle, dftABCardRef(rock)); err == nil {
		t.Fatal("a noncreature, non-Vehicle artifact was a legal target")
	}
	castCatalogSpell(t, g, "Bounce Off", "Instant", dftABBounceOffOracle, dftABCardRef(vehicle))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(vehicle) {
		t.Error("the Vehicle is not in its owner's hand")
	}
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Bounce Off", "Instant", dftABBounceOffOracle, dftABCardRef(bear))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the creature is not in its owner's hand")
	}
}

// --- Brightglass Gearhulk ---------------------------------------------------------

func TestDftABBrightglassGearhulkTutorsUpToTwoCheapPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	put := func(name, typeLine, cost string) uuid.UUID {
		return pushLibraryCardForTest(me, game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	}
	ring := put("Sol Ring", "Artifact", "{1}")
	elf := put("Llanowar Elves", "Creature — Elf Druid", "{G}")
	aura := put("Rancor", "Enchantment — Aura", "{G}")
	bear := put("Grizzly Bears", "Creature — Bear", "{1}{G}")
	bolt := put("Lightning Bolt", "Instant", "{R}")
	land := put("Forest", "Basic Land — Forest", "")

	hulk := dftABCast(t, g, me, game.Card{Name: "Brightglass Gearhulk", TypeLine: "Artifact Creature — Construct",
		OracleID: dftABGearhulkOracle, ManaCost: "{G}{G}{W}{W}", Power: 4, Toughness: 4})
	for _, kw := range []string{"first strike", "trample"} {
		if !hasEffectiveKeyword(t, g, hulk, kw) {
			t.Errorf("Brightglass Gearhulk has no %s", kw)
		}
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search when the Gearhulk entered")
	}
	for _, id := range []uuid.UUID{ring, elf, aura} {
		if !slices.Contains(c.SearchCards, id) {
			t.Errorf("a mana value 1 or less artifact, creature or enchantment %s was not offered", id)
		}
	}
	for _, id := range []uuid.UUID{bear, bolt, land} {
		if slices.Contains(c.SearchCards, id) {
			t.Errorf("%s was offered; it is too expensive or the wrong type", id)
		}
	}
	answerSearchByID(t, g, me.ID, ring, aura)
	if !me.Hand.Contains(ring) || !me.Hand.Contains(aura) {
		t.Error("the two found cards are not in hand")
	}
}

// --- Broadcast Rambler ------------------------------------------------------------

func TestDftABBroadcastRamblerMakesAThopterThatCanCrewIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	rambler := dftABCast(t, g, me, game.Card{Name: "Broadcast Rambler", TypeLine: dftABVehicle,
		OracleID: dftABRamblerOracle, Power: 5, Toughness: 4})
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && c.Controller == me.ID {
			thopter = c.InstanceID
		}
	}
	if thopter == uuid.Nil {
		t.Fatal("no Thopter token when the Rambler entered")
	}
	tok, _ := battlefieldCard(g, thopter)
	if !tok.IsArtifact() || !tok.IsCreature() || len(tok.Colors) != 0 || !hasEffectiveKeyword(t, g, thopter, "flying") {
		t.Errorf("the token is not a colorless artifact creature with flying: %+v", tok)
	}
	dftABCrew(t, g, me, rambler, 0, thopter)
}

// --- Broadside Barrage ------------------------------------------------------------

func TestDftABBroadsideBarrageDealsFiveThenLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Five Toughness", "Creature — Giant", 5, 5)
	seedLibrary(me, "Drawn")
	pitch := handCard(me, "Pitch", "Sorcery")
	castCatalogSpell(t, g, "Broadside Barrage", "Instant", dftABBarrageOracle, dftABCardRef(victim))
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("5 damage did not kill a 5/5")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
	land := b12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	if err := castCatalogSpellErr(t, g, "Broadside Barrage", "Instant", dftABBarrageOracle, dftABCardRef(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
}

// --- Broken Wings -----------------------------------------------------------------

func TestDftABBrokenWingsDestroysAFlyerOrAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ground := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Broken Wings", "Instant", dftABBrokenWingsOracle, dftABCardRef(ground)); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Broken Wings", "Instant", dftABBrokenWingsOracle, dftABCardRef(bird))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flyer was not destroyed")
	}
	aura := b12Permanent(g, opp.ID, "Oppression", "Enchantment")
	castCatalogSpell(t, g, "Broken Wings", "Instant", dftABBrokenWingsOracle, dftABCardRef(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment was not destroyed")
	}
}

// --- Broodheart Engine ------------------------------------------------------------

func TestDftABBroodheartEngineSurveilsAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	dftABPush(g, me.ID, "Broodheart Engine", "Artifact", dftABBroodheartOracle, 0, 0)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("no surveil 1 at the upkeep: %+v", c)
	}
}

func TestDftABBroodheartEngineReturnsACreatureOrVehicleCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	engine := dftABPush(g, me.ID, "Broodheart Engine", "Artifact", dftABBroodheartOracle, 0, 0)
	vehicle := dftABInGraveyard(me, "Old Vehicle", dftABVehicle)
	spell := dftABInGraveyard(me, "Old Spell", "Sorcery")
	advanceToMain(t, g)
	floatMana(t, g, me, "{B}{G}{B}{G}")
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(spell)}); err == nil {
		t.Fatal("a sorcery card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(vehicle)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(engine) {
		t.Error("the Engine was not sacrificed as part of the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vehicle) {
		t.Error("the Vehicle card did not return to the battlefield")
	}
}

// --- Burner Rocket ----------------------------------------------------------------

func TestDftABBurnerRocketHasFlashAndPumpsYourCreature(t *testing.T) {
	g := newCatalogGame(t) // seat 0's draw step: flash is what makes the cast legal
	me := g.Seats[0]
	opp := g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rocket := game.Card{InstanceID: uuid.New(), Name: "Burner Rocket", TypeLine: dftABVehicle, OracleID: dftABBurnerRocketOracle,
		Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(rocket)
	if err := g.CastSpell(me.ID, rocket.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Burner Rocket outside a main phase: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if slices.Contains(pick.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered; the text says a creature you control")
	}
	pickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}
