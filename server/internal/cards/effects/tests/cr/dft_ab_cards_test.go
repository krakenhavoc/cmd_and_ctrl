package testscr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDftABEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.DftABAatchikOracle:      "Aatchik, Emerald Radian",
		effectstest.DftABJockeyOracle:       "Adrenaline Jockey",
		effectstest.DftABAetherjacketOracle: "Aetherjacket",
		effectstest.DftABAirResponseOracle:  "Air Response Unit",
		effectstest.DftABVendettaOracle:     "Ancient Vendetta",
		effectstest.DftABRunnerOracle:       "Apocalypse Runner",
		effectstest.DftABBeastriderOracle:   "Beastrider Vanguard",
		effectstest.DftABBestowOracle:       "Bestow Greatness",
		effectstest.DftABSloopOracle:        "Boosted Sloop",
		effectstest.DftABBounceOffOracle:    "Bounce Off",
		effectstest.DftABGearhulkOracle:     "Brightglass Gearhulk",
		effectstest.DftABRamblerOracle:      "Broadcast Rambler",
		effectstest.DftABBarrageOracle:      "Broadside Barrage",
		effectstest.DftABBrokenWingsOracle:  "Broken Wings",
		effectstest.DftABBroodheartOracle:   "Broodheart Engine",
		effectstest.DftABBurnerRocketOracle: "Burner Rocket",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Aatchik, Emerald Radian ------------------------------------------------------

func TestDftABAatchikMakesAnInsectPerArtifactOrCreatureCardAndGrowsOnInsectDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	effectstest.DftABInGraveyard(me, "Ornithopter", "Artifact Creature — Thopter")
	effectstest.DftABInGraveyard(me, "Sol Ring", "Artifact")
	effectstest.DftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	effectstest.DftABInGraveyard(me, "Opt", "Instant")
	effectstest.DftABInGraveyard(me, "Forest", "Basic Land — Forest")

	aatchik := effectstest.DftABCast(t, g, me, game.Card{Name: "Aatchik, Emerald Radian", TypeLine: "Legendary Creature — Insect Druid",
		OracleID: effectstest.DftABAatchikOracle, Power: 3, Toughness: 3})
	if got := effectstest.CountNamed(g, "Insect"); got != 3 {
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
	effectstest.KillCreature(t, g, me.ID, insect)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aatchik, game.CounterPlusOne); got != 1 {
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
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aatchik, game.CounterPlusOne); got != 1 {
		t.Errorf("a Bear's death changed Aatchik's counters to %d", got)
	}
	// Nor does an opponent's Insect.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Insect", "Creature — Insect", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aatchik, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's Insect dying changed Aatchik's counters to %d", got)
	}
}

// --- Adrenaline Jockey ------------------------------------------------------------

func TestDftABAdrenalineJockeyBurnsOffTurnCastersAndGrowsOnExhaust(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jockey := effectstest.DftABPush(g, me.ID, "Adrenaline Jockey", "Creature — Minotaur Pilot", effectstest.DftABJockeyOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	effectstest.B29CastAs(t, g, opp, "Opt", "Instant", "", "{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("an opponent casting on my turn: life %d, want %d", opp.Life, before-4)
	}
	mine := me.Life
	effectstest.B24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("casting on my own turn: life %d, want %d", me.Life, mine)
	}

	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Prowcatcher Specialist", "Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, jockey, game.CounterPlusOne); got != 1 {
		t.Errorf("the Jockey has %d +1/+1 counters after an exhaust activation, want 1", got)
	}
}

// --- Aetherjacket -----------------------------------------------------------------

func TestDftABAetherjacketSacrificesToDestroyAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jacket := effectstest.DftABPush(g, me.ID, "Aetherjacket", "Artifact Creature — Thopter", effectstest.DftABAetherjacketOracle, 2, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	effectstest.AdvanceToMain(t, g)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, jacket, kw) {
			t.Errorf("Aetherjacket has no %s", kw)
		}
	}

	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, jacket, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DftABCardRef(jacket)}); err == nil {
		t.Fatal("Aetherjacket targeted itself; the text says another artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, jacket, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DftABCardRef(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(jacket) {
		t.Error("Aetherjacket was not sacrificed as part of the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the target artifact was not destroyed")
	}
}

// --- Air Response Unit ------------------------------------------------------------

func TestDftABAirResponseUnitCrewsOneAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	unit := effectstest.DftABPush(g, me.ID, "Air Response Unit", effectstest.DftABVehicle, effectstest.DftABAirResponseOracle, 3, 3)
	pilot := effectstest.B12Creature(g, me.ID, "Pilot", "Creature — Human Pilot", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCrew(t, g, me, unit, 0, pilot)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, unit, kw) {
			t.Errorf("the crewed Air Response Unit has no %s", kw)
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, unit); p != 3 || tough != 3 {
		t.Errorf("the crewed Unit is %d/%d, want 3/3", p, tough)
	}
}

// --- Ancient Vendetta -------------------------------------------------------------

func TestDftABAncientVendettaExilesUpToFourOfTheNamedCard(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	x.Put("lib 2", x.Opp.Library, "Lightning Bolt", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Ancient Vendetta", "Sorcery", effectstest.DftABVendettaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.Me.ID}}); err == nil {
		t.Fatal("Ancient Vendetta targeted its caster; the text says target opponent")
	}
	effectstest.CastCatalogSpell(t, x.G, "Ancient Vendetta", "Sorcery", effectstest.DftABVendettaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.Opp.ID}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.ActNameFor(t, x.G, x.Me.ID, "Lightning Bolt")

	effectstest.WantKeys(t, x.Candidates(t), "gy", "gy 2", "hand", "lib", "lib 2")
	c := effectstest.SearchChoiceFor(x.G, x.Me.ID)
	if c.SearchMax != 4 {
		t.Errorf("the search takes up to %d, want 4", c.SearchMax)
	}
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["gy"], x.Cards["gy 2"], x.Cards["hand"], x.Cards["lib"])
	x.Exiled(t, "gy", "gy 2", "hand", "lib")
	x.NotExiled(t, "lib 2", "gy other", "hand other", "lib other", "gy forest")
	if x.Searches() != 1 {
		t.Errorf("%d searches of their library, want 1", x.Searches())
	}
}

func TestDftABAncientVendettaMayFindNothing(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	effectstest.CastCatalogSpell(t, x.G, "Ancient Vendetta", "Sorcery", effectstest.DftABVendettaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.Opp.ID}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.ActNameFor(t, x.G, x.Me.ID, "Lightning Bolt")
	effectstest.AnswerSearchFailToFind(t, x.G, x.Me.ID)
	x.NotExiled(t, "gy", "gy 2", "hand", "lib")
	if x.Searches() != 1 {
		t.Errorf("%d searches, want 1: the library is shuffled whatever was taken", x.Searches())
	}
}

// --- Apocalypse Runner ------------------------------------------------------------

func TestDftABApocalypseRunnerMakesASmallCreatureUnblockableWithLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.DftABPush(g, me.ID, "Apocalypse Runner", effectstest.DftABVehicle, effectstest.DftABRunnerOracle, 6, 5)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Rat", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, runner, 0, game.ActivateAbilityParams{Targets: effectstest.DftABCardRef(big)}); err == nil {
		t.Fatal("a power-3 creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, runner, 0, game.ActivateAbilityParams{Targets: effectstest.DftABCardRef(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, runner); !c.Tapped {
		t.Error("the Runner did not tap")
	}
	if !effectstest.HasEffectiveKeyword(t, g, small, "lifelink") {
		t.Error("the target did not gain lifelink")
	}
	effectstest.BrAttack(t, g, small)
	effectstest.BrRefusal(t, g.DeclareBlocker(blocker, small), game.BlockReasonCantBeBlocked)
}

func TestDftABApocalypseRunnerCrewsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	runner := effectstest.DftABPush(g, me.ID, "Apocalypse Runner", effectstest.DftABVehicle, effectstest.DftABRunnerOracle, 6, 5)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, runner, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	effectstest.DftABCrew(t, g, me, runner, 1, two, one)
}

// --- Beastrider Vanguard ----------------------------------------------------------

func TestDftABBeastriderVanguardTakesAPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := effectstest.DftABPush(g, me.ID, "Beastrider Vanguard", "Creature — Human Knight", effectstest.DftABBeastriderOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedLibrary(me, "A", "B")
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
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
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{ids[1], ids[0]}) {
		t.Errorf("bottom = %v, want the player's order", got)
	}
}

// --- Bestow Greatness -------------------------------------------------------------

func TestDftABBestowGreatnessPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bestow Greatness", "Instant", effectstest.DftABBestowOracle, effectstest.DftABCardRef(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}

// --- Boosted Sloop ----------------------------------------------------------------

func TestDftABBoostedSloopLootsWheneverYouAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sloop := effectstest.DftABPush(g, me.ID, "Boosted Sloop", effectstest.DftABVehicle, effectstest.DftABSloopOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, sloop, "menace") {
		t.Error("Boosted Sloop has no menace")
	}
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "Drawn")
	pitch := effectstest.HandCard(me, "Pitch", "Sorcery")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard after the attack: the Sloop did not loot")
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d after one loot, want %d (two attackers are one attack)", got, hand)
	}
}

// --- Bounce Off -------------------------------------------------------------------

func TestDftABBounceOffReturnsACreatureOrAnUncrewedVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := effectstest.B12Permanent(g, opp.ID, "Their Vehicle", effectstest.DftABVehicle)
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Bounce Off", "Instant", effectstest.DftABBounceOffOracle, effectstest.DftABCardRef(rock)); err == nil {
		t.Fatal("a noncreature, non-Vehicle artifact was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Bounce Off", "Instant", effectstest.DftABBounceOffOracle, effectstest.DftABCardRef(vehicle))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(vehicle) {
		t.Error("the Vehicle is not in its owner's hand")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bounce Off", "Instant", effectstest.DftABBounceOffOracle, effectstest.DftABCardRef(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the creature is not in its owner's hand")
	}
}

// --- Brightglass Gearhulk ---------------------------------------------------------

func TestDftABBrightglassGearhulkTutorsUpToTwoCheapPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	put := func(name, typeLine, cost string) uuid.UUID {
		return effectstest.PushLibraryCardForTest(me, game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	}
	ring := put("Sol Ring", "Artifact", "{1}")
	elf := put("Llanowar Elves", "Creature — Elf Druid", "{G}")
	aura := put("Rancor", "Enchantment — Aura", "{G}")
	bear := put("Grizzly Bears", "Creature — Bear", "{1}{G}")
	bolt := put("Lightning Bolt", "Instant", "{R}")
	land := put("Forest", "Basic Land — Forest", "")

	hulk := effectstest.DftABCast(t, g, me, game.Card{Name: "Brightglass Gearhulk", TypeLine: "Artifact Creature — Construct",
		OracleID: effectstest.DftABGearhulkOracle, ManaCost: "{G}{G}{W}{W}", Power: 4, Toughness: 4})
	for _, kw := range []string{"first strike", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, hulk, kw) {
			t.Errorf("Brightglass Gearhulk has no %s", kw)
		}
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
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
	effectstest.AnswerSearchByID(t, g, me.ID, ring, aura)
	if !me.Hand.Contains(ring) || !me.Hand.Contains(aura) {
		t.Error("the two found cards are not in hand")
	}
}

// --- Broadcast Rambler ------------------------------------------------------------

func TestDftABBroadcastRamblerMakesAThopterThatCanCrewIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	rambler := effectstest.DftABCast(t, g, me, game.Card{Name: "Broadcast Rambler", TypeLine: effectstest.DftABVehicle,
		OracleID: effectstest.DftABRamblerOracle, Power: 5, Toughness: 4})
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && c.Controller == me.ID {
			thopter = c.InstanceID
		}
	}
	if thopter == uuid.Nil {
		t.Fatal("no Thopter token when the Rambler entered")
	}
	tok, _ := effectstest.BattlefieldCard(g, thopter)
	if !tok.IsArtifact() || !tok.IsCreature() || len(tok.Colors) != 0 || !effectstest.HasEffectiveKeyword(t, g, thopter, "flying") {
		t.Errorf("the token is not a colorless artifact creature with flying: %+v", tok)
	}
	effectstest.DftABCrew(t, g, me, rambler, 0, thopter)
}

// --- Broadside Barrage ------------------------------------------------------------

func TestDftABBroadsideBarrageDealsFiveThenLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Five Toughness", "Creature — Giant", 5, 5)
	effectstest.SeedLibrary(me, "Drawn")
	pitch := effectstest.HandCard(me, "Pitch", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Broadside Barrage", "Instant", effectstest.DftABBarrageOracle, effectstest.DftABCardRef(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("5 damage did not kill a 5/5")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
	land := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	if err := effectstest.CastCatalogSpellErr(t, g, "Broadside Barrage", "Instant", effectstest.DftABBarrageOracle, effectstest.DftABCardRef(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
}

// --- Broken Wings -----------------------------------------------------------------

func TestDftABBrokenWingsDestroysAFlyerOrAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ground := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Broken Wings", "Instant", effectstest.DftABBrokenWingsOracle, effectstest.DftABCardRef(ground)); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Broken Wings", "Instant", effectstest.DftABBrokenWingsOracle, effectstest.DftABCardRef(bird))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flyer was not destroyed")
	}
	aura := effectstest.B12Permanent(g, opp.ID, "Oppression", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Broken Wings", "Instant", effectstest.DftABBrokenWingsOracle, effectstest.DftABCardRef(aura))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment was not destroyed")
	}
}

// --- Broodheart Engine ------------------------------------------------------------

func TestDftABBroodheartEngineSurveilsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.DftABPush(g, me.ID, "Broodheart Engine", "Artifact", effectstest.DftABBroodheartOracle, 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("no surveil 1 at the upkeep: %+v", c)
	}
}

func TestDftABBroodheartEngineReturnsACreatureOrVehicleCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	engine := effectstest.DftABPush(g, me.ID, "Broodheart Engine", "Artifact", effectstest.DftABBroodheartOracle, 0, 0)
	vehicle := effectstest.DftABInGraveyard(me, "Old Vehicle", effectstest.DftABVehicle)
	spell := effectstest.DftABInGraveyard(me, "Old Spell", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{G}{B}{G}")
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DftABCardRef(spell)}); err == nil {
		t.Fatal("a sorcery card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DftABCardRef(vehicle)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(engine) {
		t.Error("the Engine was not sacrificed as part of the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vehicle) {
		t.Error("the Vehicle card did not return to the battlefield")
	}
}

// --- Burner Rocket ----------------------------------------------------------------

func TestDftABBurnerRocketHasFlashAndPumpsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t) // seat 0's draw step: flash is what makes the cast legal
	me := g.Seats[0]
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rocket := game.Card{InstanceID: uuid.New(), Name: "Burner Rocket", TypeLine: effectstest.DftABVehicle, OracleID: effectstest.DftABBurnerRocketOracle,
		Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(rocket)
	if err := g.CastSpell(me.ID, rocket.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Burner Rocket outside a main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if slices.Contains(pick.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered; the text says a creature you control")
	}
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}
