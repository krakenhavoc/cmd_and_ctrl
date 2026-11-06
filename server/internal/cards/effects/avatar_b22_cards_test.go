package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b22_cards_test.go — Avatar: The Last Airbender, batches 22 and
// 23 (#4459): every card cast, triggered or activated through the
// engine.

const (
	avB22FirebenderAscOracle = "96cc63af-69d1-4493-b241-0cb244a9e48a"
	avB22WallsOracle         = "af7d7da5-aff5-48e1-901a-f81257a85f71"
	avB22AzulaOracle         = "2631130e-63e2-43a3-914a-648e76722203"
	avB22IcebergOracle       = "646f9ad8-9d18-4b42-954a-e21a62e9f0e8"
	avB22YueOracle           = "961f1883-2f97-4249-90f9-416279d7a9dd"
	avB22PursuitOracle       = "7e7abb44-47b0-4d80-b8a1-36f0c1032aa3"
	avB22AangOracle          = "96e5d4a1-e59f-4140-823f-a17e15ee5d8d"
	avB22IrohLotusOracle     = "ee8e6ef3-eb96-48bd-8619-8a9112b3d4f0"
	avB22ZukoOracle          = "a129423d-c886-4120-96f7-c2517ea7eb83"
	avB22KataraOracle        = "0972d46e-423b-454e-87c7-a2d40fb6fb6d"
	avB22TophOracle          = "fbfef028-b10c-4ade-9503-d9af5b0c211b"
	avB22HakodaOracle        = "de3a469a-e38b-456b-994d-abec1b010e42"
	avB22VoyagerOracle       = "c5173332-68ad-456e-ad85-dd7475464d94"
	avB22ProtestorsOracle    = "6557813b-4ee7-4881-a37c-10c8ea097360"
	avB22ZhaoOracle          = "94c64d41-d756-42b6-b0e7-0a351d64a68c"
	avB22EarthenAllyOracle   = "913bca28-faf9-42e6-a8c0-469cd9bc0208"
	avB22TeachingsOracle     = "7255705c-b788-4bb1-824a-9262f4a0864e"
	avB22IrohTeaOracle       = "87adda41-113b-4126-8978-25a258b31bdd"
	avB22SokkaOracle         = "bdbec6d5-0f10-4114-b709-4402a44debdf"
)

func TestAvatarB22EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB22FirebenderAscOracle: "Firebender Ascension",
		avB22WallsOracle:         "The Walls of Ba Sing Se",
		avB22AzulaOracle:         "Fire Lord Azula",
		avB22IcebergOracle:       "Aang's Iceberg",
		avB22YueOracle:           "Yue, the Moon Spirit",
		avB22PursuitOracle:       "Obsessive Pursuit",
		avB22AangOracle:          "Aang, at the Crossroads",
		avB22AangOracle + "#1":   "Aang, Destined Savior",
		avB22IrohLotusOracle:     "Iroh, Grand Lotus",
		avB22ZukoOracle:          "Fire Lord Zuko",
		avB22KataraOracle:        "Katara, the Fearless",
		avB22TophOracle:          "Toph, the First Metalbender",
		avB22HakodaOracle:        "Hakoda, Selfless Commander",
		avB22VoyagerOracle:       "South Pole Voyager",
		avB22ProtestorsOracle:    "Wartime Protestors",
		avB22ZhaoOracle:          "Zhao, the Moon Slayer",
		avB22EarthenAllyOracle:   "Earthen Ally",
		avB22TeachingsOracle:     "Elemental Teachings",
		avB22IrohTeaOracle:       "Iroh, Tea Master",
		avB22SokkaOracle:         "Sokka, Bold Boomeranger",
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

// avB22Has reports whether the battlefield permanent has the keyword
// through the layers.
func avB22Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := avB9Live(t, g, id)
	return game.HasKeyword(&c, kw)
}

// avB22AllyEnters casts a vanilla Ally creature from the active seat's
// hand and resolves it.
func avB22AllyEnters(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	return avB1ResolveCreature(t, g, p, name, "Creature — Human Ally", "")
}

// --- The Walls of Ba Sing Se ---------------------------------------------------

func TestAvatarB22WallsGrantOtherPermanentsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walls := b12Push(g, me.ID, "The Walls of Ba Sing Se", "Legendary Artifact Creature — Wall", avB22WallsOracle, 0, 30)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !avB22Has(t, g, bear, "indestructible") || !avB22Has(t, g, land, "indestructible") {
		t.Error("another permanent you control lacks indestructible")
	}
	if avB22Has(t, g, walls, "indestructible") {
		t.Error("the Walls gave themselves indestructible")
	}
	if !avB22Has(t, g, walls, "defender") {
		t.Error("the Walls lack defender")
	}
	if avB22Has(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature has indestructible")
	}
	b27Kill(g, bear)
	b27Kill(g, theirs)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("your creature was destroyed")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature survived destruction")
	}
}

// --- Wartime Protestors ---------------------------------------------------------

func TestAvatarB22WartimeProtestorsCountersAndHastesAnotherAlly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pro := b12Push(g, me.ID, "Wartime Protestors", "Creature — Human Rebel Ally", avB22ProtestorsOracle, 4, 4)
	bear := avB1ResolveCreature(t, g, me, "Plain Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Fatalf("a non-Ally got %d counters", n)
	}
	avB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Human Ally")
	ally := avB22AllyEnters(t, g, me, "My Ally")
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ally, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Ally has %d counters, want 1", n)
	}
	if !avB22Has(t, g, ally, "haste") {
		t.Error("the Ally did not gain haste")
	}
	if n := countersOn(g, pro, game.CounterPlusOne); n != 0 {
		t.Errorf("the Protestors got %d counters", n)
	}
	if !avB22Has(t, g, pro, "haste") {
		t.Error("the Protestors lack haste")
	}
}

// --- South Pole Voyager ---------------------------------------------------------

func TestAvatarB22SouthPoleVoyagerDrawsOnTheSecondResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	life := lifeOf(g, me.ID)
	avB1ResolveCreature(t, g, me, "South Pole Voyager", "Creature — Human Scout Ally", avB22VoyagerOracle)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Fatalf("life %d after the Voyager entered, want %d", got, life+1)
	}
	avB22AllyEnters(t, g, me, "Second Ally")
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Fatalf("life %d after the second Ally, want %d", got, life+2)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d after the second resolution, want %d", got, hand+1)
	}
	avB22AllyEnters(t, g, me, "Third Ally")
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d after the third Ally, want %d", got, life+3)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the third resolution, want %d (only the second draws)", got, hand+1)
	}
}

// --- Sokka, Bold Boomeranger ------------------------------------------------------

func TestAvatarB22SokkaLootsTwoOnEntryAndGrowsOnArtifactsAndLessons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Draw A", "Draw B", "Draw C")
	a := avB1Hand(me, game.Card{Name: "Junk A", TypeLine: "Sorcery"})
	b := avB1Hand(me, game.Card{Name: "Junk B", TypeLine: "Sorcery"})
	sokka := avB1ResolveCreature(t, g, me, "Sokka, Bold Boomeranger", "Legendary Creature — Human Warrior Ally", avB22SokkaOracle)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	discardChoice := discardChoiceFor(g, me.ID)
	if discardChoice == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(discardChoice.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d after discarding two and drawing two, want %d", got, hand)
	}
	if got := me.Graveyard.Size(); got != 2 {
		t.Fatalf("graveyard %d, want 2", got)
	}

	avB1Cast(t, g, me, game.Card{Name: "Trinket", TypeLine: "Artifact"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, sokka, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after an artifact spell, want 1", n)
	}
	avB1Cast(t, g, me, game.Card{Name: "A Lesson", TypeLine: "Sorcery — Lesson"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, sokka, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters after a Lesson spell, want 2", n)
	}
	avB1Cast(t, g, me, game.Card{Name: "Plain Sorcery", TypeLine: "Sorcery"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, sokka, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after a plain sorcery, want still 2", n)
	}
}

// --- Earthen Ally -----------------------------------------------------------------

func TestAvatarB22EarthenAllyCountsColorsAmongAlliesAndEarthbendsFive(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ally := b27Push(g, me.ID, "Earthen Ally", "Creature — Human Soldier Ally", avB22EarthenAllyOracle, "{G}", 0, 2, "G")
	if p := avB9Live(t, g, ally).CurrentPower(); p != 1 {
		t.Fatalf("power %d alone, want 1 (its own green)", p)
	}
	b27Push(g, me.ID, "Gold Ally", "Creature — Human Ally", "", "{W}{U}", 1, 1, "W", "U")
	b27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{R}", 1, 1, "R")
	b27Push(g, me.ID, "Green Ally", "Creature — Elf Ally", "", "{G}", 1, 1, "G")
	if p := avB9Live(t, g, ally).CurrentPower(); p != 3 {
		t.Fatalf("power %d, want 3 (green, white, blue among Allies; the red Bear is no Ally)", p)
	}

	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	activate := func() error {
		fillPool(me, 2)
		for _, c := range []string{"W", "U", "B", "R", "G"} {
			fillPoolColored(me, c, 1)
		}
		return g.ActivateCatalogAbility(me.ID, ally, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
		})
	}
	if err := activate(); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	power, toughness, creatureLand, hasty := earthbentBody(t, g, land)
	if !creatureLand || !hasty || power != 5 || toughness != 5 {
		t.Fatalf("earthbent land %d/%d creature=%v haste=%v, want a 5/5 hasty land creature", power, toughness, creatureLand, hasty)
	}
}

func TestAvatarB22EarthenAllyEarthbendNeedsFiveColors(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ally := b27Push(g, me.ID, "Earthen Ally", "Creature — Human Soldier Ally", avB22EarthenAllyOracle, "{G}", 0, 2, "G")
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	fillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, ally, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Fatal("seven generic mana paid {2}{W}{U}{B}{R}{G}")
	}
}

// --- Toph, the First Metalbender ----------------------------------------------------

func TestAvatarB22TophMakesNontokenArtifactsLandsAndEarthbendsAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Toph, the First Metalbender", "Legendary Creature — Human Warrior Ally", avB22TophOracle, 3, 3)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	var tokenRelic uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, game.Card{Name: "Token Relic", TypeLine: "Token Artifact"}, 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("token: %v %v", ids, err)
		}
		tokenRelic = ids[0]
	})
	passPriorityAroundTable(t, g)
	if c := avB9Live(t, g, relic); !c.IsLand() || !c.IsArtifact() {
		t.Fatalf("your nontoken artifact land=%v artifact=%v, want both", c.IsLand(), c.IsArtifact())
	}
	if avB9Live(t, g, theirs).IsLand() {
		t.Error("an opponent's artifact became a land")
	}
	if avB9Live(t, g, tokenRelic).IsLand() {
		t.Error("an artifact token became a land")
	}

	mkAdvanceTo(t, g, game.StepEnd)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		passPriorityAroundTable(t, g)
		pick = latestPickTarget(g, me.ID)
	}
	if pick == nil {
		t.Fatalf("no earthbend target prompt at the end step: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: relic}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	power, toughness, creatureLand, _ := earthbentBody(t, g, relic)
	if !creatureLand || power != 2 || toughness != 2 {
		t.Fatalf("the artifact land is %d/%d creature land=%v, want an earthbent 2/2", power, toughness, creatureLand)
	}
}

// avB22Tapped reports whether the battlefield permanent is tapped.
func avB22Tapped(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

// --- Zhao, the Moon Slayer --------------------------------------------------------

func TestAvatarB22ZhaoTapsNonbasicsAndMakesThemMountainsWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	zhao := b12Push(g, me.ID, "Zhao, the Moon Slayer", "Legendary Creature — Human Soldier", avB22ZhaoOracle, 2, 2)
	if !avB22Has(t, g, zhao, "menace") {
		t.Error("Zhao lacks menace")
	}
	basic := b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if avB22Tapped(g, basic) {
		t.Error("a basic land entered tapped")
	}
	nonbasic := b12PlayFromHand(t, g, "Some Grove", "Land — Forest", "", game.CastSpellParams{})
	if !avB22Tapped(g, nonbasic) {
		t.Fatal("a nonbasic land entered untapped")
	}
	if c := avB9Live(t, g, nonbasic); c.HasSubtype("Mountain") {
		t.Fatal("a nonbasic land is a Mountain before Zhao has a conqueror counter")
	}

	fillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, zhao, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {7}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, zhao, "conqueror"); n != 1 {
		t.Fatalf("%d conqueror counters, want 1", n)
	}
	c := avB9Live(t, g, nonbasic)
	if !c.HasSubtype("Mountain") || c.HasSubtype("Forest") {
		t.Errorf("the nonbasic land's subtypes are %v, want only Mountain", c.Effective().Subtypes)
	}
	if b := avB9Live(t, g, basic); !b.HasSubtype("Forest") || b.HasSubtype("Mountain") {
		t.Errorf("the basic Forest changed: %v", b.Effective().Subtypes)
	}
}

func TestAvatarB22ZhaoCounterAbilityNeedsSeven(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	zhao := b12Push(g, me.ID, "Zhao, the Moon Slayer", "Legendary Creature — Human Soldier", avB22ZhaoOracle, 2, 2)
	fillPool(me, 6)
	if err := g.ActivateCatalogAbility(me.ID, zhao, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("six mana paid {7}")
	}
}

// --- Hakoda, Selfless Commander -----------------------------------------------------

func TestAvatarB22HakodaCastsAlliesFromTheTopOfTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hakoda := b12Push(g, me.ID, "Hakoda, Selfless Commander", "Legendary Creature — Human Warrior Ally", avB22HakodaOracle, 3, 5)
	if !avB22Has(t, g, hakoda, "vigilance") {
		t.Error("Hakoda lacks vigilance")
	}
	if v := game.CatalogLibraryTopVisible(avB22HakodaOracle); v != game.LibraryTopOwner {
		t.Errorf("LibraryTopVisible = %v, want LibraryTopOwner", v)
	}
	bear := seedLibraryTop(me, "Top Bear", "Creature — Bear", "{G}")
	fillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("a non-Ally creature was cast from the top of the library")
	}
	ally := seedLibraryTop(me, "Top Ally", "Creature — Human Ally", "{G}")
	if err := g.CastSpell(me.ID, ally, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("casting the Ally from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ally) {
		t.Fatal("the Ally cast from the top never resolved")
	}
}

func TestAvatarB22HakodaSacrificesToShieldYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hakoda := b12Push(g, me.ID, "Hakoda, Selfless Commander", "Legendary Creature — Human Warrior Ally", avB22HakodaOracle, 3, 5)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, hakoda, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(hakoda) {
		t.Fatal("Hakoda was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	c := avB9Live(t, g, bear)
	if c.CurrentToughness() != 7 || c.CurrentPower() != 2 || !game.HasKeyword(&c, "indestructible") {
		t.Fatalf("your Bear is %d/%d indestructible=%v, want 2/7 and indestructible",
			c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "indestructible"))
	}
	o := avB9Live(t, g, theirs)
	if o.CurrentToughness() != 2 || game.HasKeyword(&o, "indestructible") {
		t.Error("an opponent's creature was boosted")
	}
}

// --- Aang's Iceberg ------------------------------------------------------------------

func TestAvatarB22IcebergExilesUntilItLeavesThenWaterbendsToScry(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedLibrary(me, "Scry A", "Scry B", "Scry C")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "W", 1)
	fillPool(me, 2)
	ice := castFromHandForTest(t, g, me, "Aang's Iceberg", "Enchantment", "{2}{W}", avB22IcebergOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target prompt for the enters trigger: %+v", g.PendingChoices)
	}
	if hasID(pick.PickTargetCards, ice) {
		t.Error("the Iceberg may target itself; the text says other")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the target was not exiled")
	}

	tappers := pushTapCostSoldiers(g, me.ID, 3)
	if err := avB4Activate(g, me.ID, ice, 0, tappers); err != nil {
		t.Fatalf("waterbend {3} by tapping three creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry after the sacrifice: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry over %d cards, want 2", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ice) {
		t.Error("the Iceberg was not sacrificed")
	}
	returned := false
	for _, bc := range g.Battlefield.Cards {
		if bc.Name == "Their Bear" && bc.Controller == opp.ID {
			returned = true
		}
	}
	if !returned {
		t.Error("the exiled card did not return when the Iceberg left")
	}
}

func TestAvatarB22IcebergWaterbendCannotBePaidShort(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ice := b12Permanent(g, me.ID, "Aang's Iceberg", "Enchantment")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == ice {
			g.Battlefield.Cards[i].OracleID = avB22IcebergOracle
		}
	}
	tappers := pushTapCostSoldiers(g, me.ID, 2)
	if err := avB4Activate(g, me.ID, ice, 0, tappers); err == nil {
		t.Fatal("two taps and no mana paid waterbend {3}")
	}
}

// --- Yue, the Moon Spirit ----------------------------------------------------------------

func TestAvatarB22YueCastsANoncreatureSpellFromHandForFree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	yue := b12Push(g, me.ID, "Yue, the Moon Spirit", "Legendary Creature — Spirit Ally", avB22YueOracle, 3, 3)
	if !avB22Has(t, g, yue, "flying") || !avB22Has(t, g, yue, "vigilance") {
		t.Error("Yue lacks flying or vigilance")
	}
	sorcery := handCardFull(me, "Big Sorcery", "Sorcery", "{6}{U}", "", []string{"U"})
	creature := handCardFull(me, "Big Creature", "Creature — Giant", "{6}{U}", "", []string{"U"})
	tappers := pushTapCostSoldiers(g, me.ID, 5)
	if err := avB4Activate(g, me.ID, yue, 0, tappers); err != nil {
		t.Fatalf("waterbend {5}, {T}: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, sorcery)
	if hasID(p.DiscardOptions, creature) {
		t.Error("a creature spell was offered")
	}
	if err := castThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	castInsideTheResolution(t, g, sorcery)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("the free sorcery did not resolve")
	}
	if !avB22Tapped(g, yue) {
		t.Error("Yue did not tap")
	}
}

// --- Obsessive Pursuit ----------------------------------------------------------------------

func TestAvatarB22ObsessivePursuitDrainsForAClueOnEntryAndUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	life := lifeOf(g, me.ID)
	fillPoolColored(me, "B", 1)
	fillPool(me, 1)
	castFromHandForTest(t, g, me, "Obsessive Pursuit", "Enchantment", "{1}{B}", avB22PursuitOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life-1 {
		t.Fatalf("life %d after entering, want %d", got, life-1)
	}
	if got := len(avB5CluesB7(g, me.ID)); got != 1 {
		t.Fatalf("%d Clues after entering, want 1", got)
	}
}

func TestAvatarB22ObsessivePursuitCountsSacrificesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Obsessive Pursuit", "Enchantment", avB22PursuitOracle, 0, 0)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Fodder", "Artifact"))
	}
	declareAttack(t, g, opp.ID, attacker)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		passPriorityAroundTable(t, g)
		pick = latestPickTarget(g, me.ID)
	}
	if pick == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: attacker}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, attacker, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters, want 3 (three permanents sacrificed)", n)
	}
	if !avB22Has(t, g, attacker, "lifelink") {
		t.Error("X of three did not grant lifelink")
	}
}

func TestAvatarB22ObsessivePursuitWithTwoSacrificesGivesNoLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Obsessive Pursuit", "Enchantment", avB22PursuitOracle, 0, 0)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	for i := 0; i < 2; i++ {
		avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Fodder", "Artifact"))
	}
	declareAttack(t, g, opp.ID, attacker)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		passPriorityAroundTable(t, g)
		pick = latestPickTarget(g, me.ID)
	}
	if pick == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: attacker}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, attacker, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters, want 2", n)
	}
	if avB22Has(t, g, attacker, "lifelink") {
		t.Error("X of two granted lifelink")
	}
}

// --- Fire Lord Zuko -------------------------------------------------------------------------

func TestAvatarB22ZukoFirebendsForHisPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zuko := b27Push(g, me.ID, "Fire Lord Zuko", "Legendary Creature — Human Noble Ally", avB22ZukoOracle, "{R}{W}{B}", 2, 4, "R", "W", "B")
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(zuko, game.CounterPlusOne, 1) })
	declareAttack(t, g, opp.ID, zuko)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepEndCombat)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

func TestAvatarB22ZukoGrowsTheTeamWhenAPermanentEntersFromExile(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	zuko := b12Push(g, me.ID, "Fire Lord Zuko", "Legendary Creature — Human Noble Ally", avB22ZukoOracle, 2, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	flicked := b12Creature(g, me.ID, "Flicked", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(flicked); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	g.WithWriteLock(func() {
		if _, err := g.ReturnFromExileToBattlefieldForEffect(flicked, me.ID, false); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := countersOn(g, zuko, game.CounterPlusOne); n != 1 {
		t.Fatalf("Zuko has %d counters after a permanent entered from exile, want 1", n)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("your Bear has %d counters, want 1", n)
	}
}

// --- Iroh, Grand Lotus ------------------------------------------------------------------------

func TestAvatarB22IrohGrandLotusGivesFlashbackOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Iroh, Grand Lotus", "Legendary Creature — Human Noble Ally", avB22IrohLotusOracle, 5, 5)
	shock := gyCard(me, "Shock", "Instant", "{R}", "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d")
	lesson := gyCard(me, "Old Lesson", "Sorcery — Lesson", "{3}{G}{G}", "")
	creature := gyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")

	fillPool(me, 1)
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("the flashback cost of a non-Lesson instant was paid with generic mana; it is the card's mana cost")
	}
	fillPoolColored(me, "R", 1)
	life := lifeOf(g, opp.ID)
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("flashback Shock for {R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent life %d, want %d", got, life-2)
	}
	if me.Graveyard.Contains(shock) || !g.Exile.Contains(shock) {
		t.Error("the flashed-back Shock was not exiled")
	}

	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, lesson, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback the Lesson for {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(lesson) {
		t.Error("the flashed-back Lesson was not exiled")
	}
	fillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err == nil {
		t.Error("a creature card was given flashback")
	}
}

func TestAvatarB22IrohGrandLotusGivesNothingOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, opp.ID, "Iroh, Grand Lotus", "Legendary Creature — Human Noble Ally", avB22IrohLotusOracle, 5, 5)
	shock := gyCard(opp, "Shock", "Instant", "{R}", "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d")
	fillPoolColored(opp, "R", 1)
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("Iroh's flashback worked during an opponent's turn")
	}
}

func TestAvatarB22IrohGrandLotusFirebendsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iroh := b12Push(g, me.ID, "Iroh, Grand Lotus", "Legendary Creature — Human Noble Ally", avB22IrohLotusOracle, 5, 5)
	declareAttack(t, g, opp.ID, iroh)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 2})
}

// avB22PickOpponentThenPermanent answers Iroh, Tea Master's two clauses,
// the opponent first and then the permanent, each on its own prompt.
func avB22PickOpponentThenPermanent(t *testing.T, g *game.Game, me uuid.UUID, pick *game.PendingChoice, opp, perm uuid.UUID) {
	t.Helper()
	if !hasID(pick.PickTargetPlayers, opp) || hasID(pick.PickTargetPlayers, me) {
		t.Fatalf("first clause offers players %v, want the opponent only", pick.PickTargetPlayers)
	}
	if err := g.ResolvePickTargets(pick.ID, me, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp}}); err != nil {
		t.Fatalf("ResolvePickTargets (opponent): %v", err)
	}
	second := latestPickTarget(g, me)
	if second == nil {
		t.Fatalf("no second target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(second.ID, me, []game.TargetRef{{Kind: game.TargetCard, ID: perm, Slot: 1}}); err != nil {
		t.Fatalf("ResolvePickTargets (permanent): %v", err)
	}
}

// --- Iroh, Tea Master ----------------------------------------------------------------------------

func TestAvatarB22IrohTeaMasterMakesFoodThenGivesAPermanentForAnAlly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fillPoolColored(me, "R", 1)
	fillPoolColored(me, "W", 1)
	fillPool(me, 1)
	castFromHandForTest(t, g, me, "Iroh, Tea Master", "Legendary Creature — Human Citizen Ally", "{1}{R}{W}", avB22IrohTeaOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	food := avB1Named(g, me.ID, "Food")
	if food == uuid.Nil {
		t.Fatal("no Food token")
	}
	gift := b12Permanent(g, me.ID, "Old Relic", "Artifact")

	mkAdvanceTo(t, g, game.StepBeginCombat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		passPriorityAroundTable(t, g)
		pick = latestPickTarget(g, me.ID)
	}
	if pick == nil {
		t.Fatalf("no target prompt at the beginning of combat: %+v", g.PendingChoices)
	}
	avB22PickOpponentThenPermanent(t, g, me.ID, pick, opp.ID, gift)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c := avB9Live(t, g, gift); c.Controller != opp.ID {
		t.Fatalf("the relic is controlled by %s, want the opponent", c.Controller)
	}
	ally := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ally" && IsToken(c) && c.Controller == me.ID {
			ally = c.InstanceID
		}
	}
	if ally == uuid.Nil {
		t.Fatal("no Ally token")
	}
	if c := avB9Live(t, g, ally); c.CurrentPower() != 2 || !c.HasColor("W") {
		t.Errorf("the Ally is %d power, white=%v; want a white 1/1 with one counter (one permanent given)", c.CurrentPower(), c.HasColor("W"))
	}
	if n := countersOn(g, ally, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters on the Ally, want 1", n)
	}
}

func TestAvatarB22IrohTeaMasterDecliningGivesNothing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Iroh, Tea Master", "Legendary Creature — Human Citizen Ally", avB22IrohTeaOracle, 2, 2)
	gift := b12Permanent(g, me.ID, "Old Relic", "Artifact")
	mkAdvanceTo(t, g, game.StepBeginCombat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		passPriorityAroundTable(t, g)
		pick = latestPickTarget(g, me.ID)
	}
	if pick == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	avB22PickOpponentThenPermanent(t, g, me.ID, pick, opp.ID, gift)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if c := avB9Live(t, g, gift); c.Controller != me.ID {
		t.Error("declining gave the relic away")
	}
	if avB1Named(g, me.ID, "Ally") != uuid.Nil {
		t.Error("declining made an Ally")
	}
}

// --- Elemental Teachings ------------------------------------------------------------------------

func TestAvatarB22ElementalTeachingsOpponentBinsTwoTheRestEnterTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lands := seedTypedLibrary(me, "Land", "Grove", "Marsh", "Peak", "Strand", "Grove")
	castCatalogSpell(t, g, "Elemental Teachings", "Instant — Lesson", avB22TeachingsOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{lands[0], lands[4]}); err == nil {
		t.Fatal("two lands with the same name were found")
	}
	found := lands[:4]
	if err := g.ResolveSearchLibrary(c.ID, me.ID, found); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if latestOptionPickFor(g, me.ID) != nil {
		answerChoosePlayer(t, g, me.ID, opp)
	}
	pick := latestRevealPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked to choose: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(pick.ID, opp.ID, []uuid.UUID{found[0], found[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range found[:2] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("chosen land %s is not in the graveyard", id)
		}
	}
	for _, id := range found[2:] {
		if !g.Battlefield.Contains(id) || !avB22Tapped(g, id) {
			t.Errorf("land %s did not enter tapped", id)
		}
	}
	if !me.Library.Contains(lands[4]) {
		t.Error("the unfound land left the library")
	}
}

// --- Katara, the Fearless ------------------------------------------------------------------------

func TestAvatarB22KataraDoublesAnAllysTriggerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Katara, the Fearless", "Legendary Creature — Human Warrior Ally", avB22KataraOracle, 3, 3)
	b12Push(g, me.ID, "Wartime Protestors", "Creature — Human Rebel Ally", avB22ProtestorsOracle, 4, 4)
	ally := avB22AllyEnters(t, g, me, "New Ally")
	for i := 0; i < 4; i++ {
		if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
			break
		}
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ally, game.CounterPlusOne); n != 2 {
		t.Fatalf("the new Ally has %d counters, want 2 (the Protestors' trigger doubled)", n)
	}

	advanceToMain(t, g)
	fillPoolColored(me, "B", 1)
	fillPool(me, 1)
	castFromHandForTest(t, g, me, "Obsessive Pursuit", "Enchantment", "{1}{B}", avB22PursuitOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != 1 {
		t.Errorf("%d Clues from a non-Ally's trigger, want 1 (not doubled)", got)
	}
}

// --- Fire Lord Azula ------------------------------------------------------------------------------

func TestAvatarB22AzulaCopiesASpellCastWhileSheAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	azula := b27Push(g, me.ID, "Fire Lord Azula", "Legendary Creature — Human Noble", avB22AzulaOracle, "{1}{U}{B}{R}", 4, 4, "U", "B", "R")
	declareAttack(t, g, opp.ID, azula)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 2})
	life := lifeOf(g, opp.ID)
	b29CastAs(t, g, me, "Shock", "Instant", "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d", "{R}",
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	for i := 0; i < 6; i++ {
		passPriorityAroundTable(t, g)
		if c := latestConfirmFor(g, me.ID); c != nil {
			answerMayChoice(t, g, me.ID, false)
			continue
		}
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
				t.Fatalf("ResolvePickTargets: %v", err)
			}
			continue
		}
		break
	}
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Fatalf("opponent life %d, want %d (Shock and its copy)", got, life-4)
	}
}

func TestAvatarB22AzulaDoesNotCopyWhenNotAttacking(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b27Push(g, me.ID, "Fire Lord Azula", "Legendary Creature — Human Noble", avB22AzulaOracle, "{1}{U}{B}{R}", 4, 4, "U", "B", "R")
	life := lifeOf(g, opp.ID)
	fillPoolColored(me, "R", 1)
	b29CastAs(t, g, me, "Shock", "Instant", "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d", "{R}",
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Fatalf("opponent life %d, want %d (no copy outside an attack)", got, life-2)
	}
}

// --- Aang, at the Crossroads // Aang, Destined Savior ---------------------------------------------

func avB22AangCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   avB22AangOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Aang, at the Crossroads", TypeLine: "Legendary Creature — Human Avatar Ally",
				ManaCost: "{2}{G}{W}{U}", Colors: []string{"G", "W", "U"}, Power: 3, Toughness: 3},
			{Name: "Aang, Destined Savior", TypeLine: "Legendary Creature — Avatar Ally",
				Colors: []string{"G", "W", "U"}, Power: 4, Toughness: 4},
		},
	}
	c.SetFace(0)
	return c
}

func TestAvatarB22AangPutsACheapCreatureFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	big := seedLibraryTop(me, "Big Beast", "Creature — Beast", "{4}{G}")
	small := seedLibraryTop(me, "Small Beast", "Creature — Beast", "{2}{G}")
	for i := 0; i < 3; i++ {
		seedLibraryTop(me, "Filler", "Sorcery", "{1}")
	}
	aang := avB22AangCard(me.ID)
	me.Hand.PushTop(aang)
	if err := g.CastSpell(me.ID, aang.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !avB22Has(t, g, aang.InstanceID, "flying") {
		t.Error("Aang lacks flying")
	}
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatalf("no choose-cards prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{big}); err == nil {
		t.Fatal("a creature with mana value 5 was put onto the battlefield")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{small}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Fatal("the mana value 3 creature did not enter")
	}
	if !me.Library.Contains(big) {
		t.Error("the rest left the library")
	}
}

func TestAvatarB22AangTransformsAtTheNextUpkeepAfterAnotherCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := avB22AangCard(me.ID)
	pushBattlefieldCardWithTimestamp(g, aang)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	b27Kill(g, other)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if c := battlefieldCardFor(g, aang.InstanceID); c == nil || c.ActiveFace != 0 {
		t.Fatal("Aang transformed before the upkeep")
	}
	mkAdvanceTo(t, g, game.StepEnd)
	mkAdvanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	c := battlefieldCardFor(g, aang.InstanceID)
	if c == nil || c.ActiveFace != 1 || c.Name != "Aang, Destined Savior" {
		t.Fatalf("Aang is %+v, want transformed into Aang, Destined Savior", c)
	}
}

func TestAvatarB22AangDestinedSaviorGivesLandCreaturesVigilanceAndEarthbends(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := avB22AangCard(me.ID)
	aang.SetFace(1)
	pushBattlefieldCardWithTimestamp(g, aang)
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if avB22Has(t, g, land, "vigilance") {
		t.Error("a land that is not a creature has vigilance")
	}
	mkAdvanceTo(t, g, game.StepBeginCombat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		passPriorityAroundTable(t, g)
		pick = latestPickTarget(g, me.ID)
	}
	if pick == nil {
		t.Fatalf("no earthbend prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: land}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	power, _, creatureLand, _ := earthbentBody(t, g, land)
	if !creatureLand || power != 2 {
		t.Fatalf("land power %d creature land %v, want an earthbent 2/2", power, creatureLand)
	}
	if !avB22Has(t, g, land, "vigilance") {
		t.Error("the land creature lacks vigilance")
	}
	if avB22Has(t, g, bear, "vigilance") {
		t.Error("a non-land creature has vigilance")
	}
}

// --- Firebender Ascension ------------------------------------------------------------------------

func TestAvatarB22FirebenderAscensionMakesASoldierThatFirebends(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fillPoolColored(me, "R", 1)
	fillPool(me, 1)
	castFromHandForTest(t, g, me, "Firebender Ascension", "Enchantment", "{1}{R}", avB22FirebenderAscOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	soldier := avB1Named(g, me.ID, "Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Soldier token")
	}
	c := avB9Live(t, g, soldier)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.HasColor("R") || !IsToken(c) {
		t.Errorf("the Soldier is %d/%d red=%v token=%v", c.CurrentPower(), c.CurrentToughness(), c.HasColor("R"), IsToken(c))
	}
}

func TestAvatarB22FirebenderAscensionCountsAttackTriggersAndCopiesAtFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	asc := b12Push(g, me.ID, "Firebender Ascension", "Enchantment", avB22FirebenderAscOracle, 0, 0)
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(asc, "quest", 3) })
	iroh := b12Push(g, me.ID, "Iroh, Grand Lotus", "Legendary Creature — Human Noble Ally", avB22IrohLotusOracle, 5, 5)
	declareAttack(t, g, opp.ID, iroh)
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		if c := latestConfirmFor(g, me.ID); c != nil {
			answerMayChoice(t, g, me.ID, true)
			continue
		}
		break
	}
	if n := countersOn(g, asc, "quest"); n != 4 {
		t.Fatalf("%d quest counters, want 4", n)
	}
	mkWantPool(t, me, map[string]int{"R": 4})
}

func TestAvatarB22FirebenderAscensionBelowFourOnlyCounts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	asc := b12Push(g, me.ID, "Firebender Ascension", "Enchantment", avB22FirebenderAscOracle, 0, 0)
	iroh := b12Push(g, me.ID, "Iroh, Grand Lotus", "Legendary Creature — Human Noble Ally", avB22IrohLotusOracle, 5, 5)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, iroh, plain)
	passPriorityAroundTable(t, g)
	if c := latestConfirmFor(g, me.ID); c != nil {
		t.Fatalf("a copy was offered below four counters: %+v", c)
	}
	if n := countersOn(g, asc, "quest"); n != 1 {
		t.Fatalf("%d quest counters, want 1 (one attack trigger; a creature with none adds nothing)", n)
	}
	mkWantPool(t, me, map[string]int{"R": 2})
}
