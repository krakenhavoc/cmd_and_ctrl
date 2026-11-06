package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_b1_cards_test.go — Wilds of Eldraine, batch 1 (#4471). Each test
// plays the real catalog entry and asserts what its text says.

const (
	woeB1ArchiveDragonOracle = "cddbd819-4895-4712-b44b-d6b51f3d8646"
	woeB1ReaperOracle        = "6f1e5571-ddda-4cea-84ef-36a571d8fd51"
	woeB1TaleOracle          = "682a0770-8fe0-4142-aa0c-b60af3628a33"
	woeB1NaughtyOracle       = "95664c29-9801-4dfc-9dbd-465abf6bb4ee"
	woeB1BeanstalkOracle     = beanstalkWurmOracleID
	woeB1BruiserOracle       = bellowingBruiserOracleID
	woeB1GatekeeperOracle    = belunasGatekeeperOracleID
	woeB1BloodlineOracle     = "a3ae6694-6314-42e2-bc12-d0e32d49b347"
	woeB1GrappleOracle       = "9588a7fd-bbe5-4a81-9a7e-d8e8f6c0537d"
	woeB1CommuneOracle       = "d4ed4260-5f38-4b55-ba6e-9f54f04d2360"
)

func TestWoeB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeB1ArchiveDragonOracle:     "Archive Dragon",
		woeB1ReaperOracle:            "Ashiok's Reaper",
		woeB1TaleOracle:              "A Tale for the Ages",
		woeB1NaughtyOracle:           "Barrow Naughty",
		woeB1BeanstalkOracle:         "Beanstalk Wurm",
		woeB1BeanstalkOracle + "#1":  "Plant Beans",
		woeB1BruiserOracle:           "Bellowing Bruiser",
		woeB1BruiserOracle + "#1":    "Beat a Path",
		woeB1GatekeeperOracle:        "Beluna's Gatekeeper",
		woeB1GatekeeperOracle + "#1": "Entry Denied",
		woeB1BloodlineOracle:         "Bestial Bloodline",
		woeB1GrappleOracle:           "Candy Grapple",
		woeB1CommuneOracle:           "Commune with Nature",
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

// woeB1Seats is the active player and the next one.
func woeB1Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// woeB1Destroy destroys a permanent and settles whatever that triggers.
func woeB1Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
}

// woeB1Attach attaches an Aura already on the battlefield to a creature.
func woeB1Attach(t *testing.T, g *game.Game, aura, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
}

// woeB1Adventure puts an Adventure card into p's hand, on its creature face.
func woeB1Adventure(p *game.Player, oracle string, creature, adventure game.Face) game.Card {
	card := adventureCard(p.ID, adventureCase{oracle: oracle, faces: [2]game.Face{creature, adventure}})
	p.Hand.PushTop(card)
	return card
}

// woeB1CastAdventure casts the Adventure half of a card in hand and
// settles it.
func woeB1CastAdventure(t *testing.T, g *game.Game, p *game.Player, card game.Card, targets []game.TargetRef) {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(p.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: targets}); err != nil {
		t.Fatalf("cast the Adventure: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card is not in exile after it resolved")
	}
}

// --- Archive Dragon -----------------------------------------------------------

func TestWoeB1ArchiveDragonFliesAndScriesTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	dragon := castCatalogSpell(t, g, "Archive Dragon", "Creature — Dragon Wizard", woeB1ArchiveDragonOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Fatal("Archive Dragon did not resolve")
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	answerScryKeepAll(t, g, me.ID)
	if !hasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("Archive Dragon has no flying")
	}
}

func TestWoeB1ArchiveDragonHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushCatalogPermanent(g, me.ID, "Archive Dragon", "Creature — Dragon Wizard", woeB1ArchiveDragonOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, dragon)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting Archive Dragon")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Ashiok's Reaper ----------------------------------------------------------

func TestWoeB1AshioksReaperDrawsWhenYourEnchantmentGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Ashiok's Reaper", "Creature — Nightmare", woeB1ReaperOracle, false)
	mine := b12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirs := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	hand := me.Hand.Size()

	woeB1Destroy(t, g, theirs)
	woeB1Destroy(t, g, rock)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d -> %d after an opponent's enchantment and my artifact died, want no draw", hand, got)
	}
	woeB1Destroy(t, g, mine)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d after my enchantment died, want one draw", hand, got)
	}

	// An Aura that goes with its creature is my enchantment going to the graveyard too.
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := b12Permanent(g, me.ID, "My Aura", auraTypeLine)
	woeB1Attach(t, g, aura, host)
	woeB1Destroy(t, g, host)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d after the Aura fell off its dead creature, want a second draw", hand, got)
	}
}

// --- A Tale for the Ages --------------------------------------------------------

func TestWoeB1ATaleForTheAgesPumpsOnlyYourEnchantedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "A Tale for the Ages", "Enchantment", woeB1TaleOracle, false)
	enchanted := b12Creature(g, me.ID, "Enchanted Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	hexed := b12Creature(g, me.ID, "Hexed Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	woeB1Attach(t, g, b12Permanent(g, me.ID, "My Aura", auraTypeLine), enchanted)
	woeB1Attach(t, g, b12Permanent(g, opp.ID, "Their Aura on mine", auraTypeLine), hexed)
	woeB1Attach(t, g, b12Permanent(g, me.ID, "My Aura on theirs", auraTypeLine), theirs)

	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{
		"my creature with my Aura":            {enchanted, 4},
		"my creature with an opponent's Aura": {hexed, 4},
		"my creature with no Aura":            {plain, 2},
		"an opponent's enchanted creature":    {theirs, 2},
	} {
		if p, tough := effectivePTOf(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, tc.want, tc.want)
		}
	}
}

// --- Barrow Naughty -------------------------------------------------------------

func TestWoeB1BarrowNaughtyHasLifelinkOnlyBesideAnotherFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	naughty := pushCatalogPermanent(g, me.ID, "Barrow Naughty", "Creature — Faerie", woeB1NaughtyOracle, false)
	if !hasEffectiveKeyword(t, g, naughty, "flying") {
		t.Error("Barrow Naughty has no flying")
	}
	if hasEffectiveKeyword(t, g, naughty, "lifelink") {
		t.Fatal("Barrow Naughty has lifelink with no other Faerie; it is not its own other Faerie")
	}
	friend := b12Creature(g, me.ID, "Friend", "Creature — Faerie Rogue", 1, 1)
	if !hasEffectiveKeyword(t, g, naughty, "lifelink") {
		t.Fatal("Barrow Naughty has no lifelink beside another Faerie")
	}
	woeB1Destroy(t, g, friend)
	if hasEffectiveKeyword(t, g, naughty, "lifelink") {
		t.Error("the lifelink outlived the other Faerie")
	}
}

func TestWoeB1BarrowNaughtyPumpsItselfForTwoAndBlack(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	naughty := pushCatalogPermanent(g, me.ID, "Barrow Naughty", "Creature — Faerie", woeB1NaughtyOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	before := effectivePower(t, g, naughty)
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, naughty, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump activated without {B}")
	}
	floatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, naughty, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{B}: +1/+0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, naughty); got != before+1 {
		t.Errorf("power %d after the pump, want %d", got, before+1)
	}
}

// --- Beanstalk Wurm // Plant Beans ----------------------------------------------

func woeB1Beanstalk(p *game.Player) game.Card {
	return woeB1Adventure(p, woeB1BeanstalkOracle,
		game.Face{Name: "Beanstalk Wurm", TypeLine: "Creature — Plant Wurm", ManaCost: "{4}{G}", Colors: []string{"G"}, Power: 5, Toughness: 4},
		game.Face{Name: "Plant Beans", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{G}", Colors: []string{"G"}})
}

func TestWoeB1PlantBeansGrantsALandPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	card := woeB1Beanstalk(me)
	woeB1CastAdventure(t, g, me, card, nil)
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 1 {
		t.Errorf("extra land plays = %d, want 1", got)
	}
}

func TestWoeB1BeanstalkWurmHasReach(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	wurm := pushCatalogPermanent(g, me.ID, "Beanstalk Wurm", "Creature — Plant Wurm", woeB1BeanstalkOracle, false)
	if !hasEffectiveKeyword(t, g, wurm, "reach") {
		t.Error("Beanstalk Wurm has no reach")
	}
}

// --- Bellowing Bruiser // Beat a Path -------------------------------------------

func woeB1Bruiser(p *game.Player) game.Card {
	return woeB1Adventure(p, woeB1BruiserOracle,
		game.Face{Name: "Bellowing Bruiser", TypeLine: "Creature — Ogre", ManaCost: "{4}{R}", Colors: []string{"R"}, Power: 4, Toughness: 4},
		game.Face{Name: "Beat a Path", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{R}", Colors: []string{"R"}})
}

func TestWoeB1BeatAPathStopsUpToTwoCreaturesBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	a := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 4)
	free := b12Creature(g, opp.ID, "Wall C", "Creature — Wall", 0, 4)
	card := woeB1Bruiser(me)
	woeB1CastAdventure(t, g, me, card, cardRefs(a, b))
	for name, id := range map[string]uuid.UUID{"the first target": a, "the second target": b} {
		if auraRestrictions(t, g, id)&game.CantBlock == 0 {
			t.Errorf("%s can still block", name)
		}
	}
	if auraRestrictions(t, g, free)&game.CantBlock != 0 {
		t.Error("an untargeted creature can't block")
	}
}

func TestWoeB1BeatAPathWithNoTargetStillResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	card := woeB1Bruiser(me)
	woeB1CastAdventure(t, g, me, card, nil)
}

func TestWoeB1BellowingBruiserHasHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bruiser := pushCatalogPermanent(g, me.ID, "Bellowing Bruiser", "Creature — Ogre", woeB1BruiserOracle, true)
	if !hasEffectiveKeyword(t, g, bruiser, "haste") {
		t.Error("Bellowing Bruiser has no haste")
	}
}

// --- Beluna's Gatekeeper // Entry Denied ----------------------------------------

func woeB1Gatekeeper(p *game.Player) game.Card {
	return woeB1Adventure(p, woeB1GatekeeperOracle,
		game.Face{Name: "Beluna's Gatekeeper", TypeLine: "Creature — Giant Soldier", ManaCost: "{5}{U}", Colors: []string{"U"}, Power: 6, Toughness: 5},
		game.Face{Name: "Entry Denied", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}})
}

func woeB1CreatureOfCost(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func TestWoeB1EntryDeniedBouncesASmallCreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	small := woeB1CreatureOfCost(g, opp.ID, "Small", "{1}{G}{G}")
	card := woeB1Gatekeeper(me)
	woeB1CastAdventure(t, g, me, card, cardRefs(small))
	if !opp.Hand.Contains(small) {
		t.Error("the mana value 3 creature was not returned to its owner's hand")
	}
}

func TestWoeB1EntryDeniedRefusesABigCreatureOrYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	big := woeB1CreatureOfCost(g, opp.ID, "Big", "{2}{G}{G}")
	mine := woeB1CreatureOfCost(g, me.ID, "Mine", "{G}")
	card := woeB1Gatekeeper(me)
	advanceTo(t, g, game.StepPrecombatMain)
	for name, id := range map[string]uuid.UUID{"a mana value 4 creature": big, "a creature you control": mine} {
		if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(id)}); err == nil {
			t.Fatalf("%s was a legal target for Entry Denied", name)
		}
	}
}

// --- Bestial Bloodline ------------------------------------------------------------

func TestWoeB1BestialBloodlinePumpsItsHostAndReturnsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	host := auraBear(g, me.ID)
	aura := auraCast(t, g, "Bestial Bloodline", woeB1BloodlineOracle, host)
	if p, tough := effectivePTOf(t, g, host); p != 4 || tough != 4 {
		t.Fatalf("host is %d/%d, want 4/4", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the return ability activated while the Aura was on the battlefield")
	}
	woeB1Destroy(t, g, host)
	if !me.Graveyard.Contains(aura) {
		t.Fatal("the Aura is not in the graveyard after its creature died")
	}
	floatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the return activated without {G}")
	}
	floatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4}{G}: return this card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("Bestial Bloodline did not return to hand")
	}
}

// --- Candy Grapple ----------------------------------------------------------------

func TestWoeB1CandyGrappleShrinksByThreeUnbargained(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := woeB1Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	castCatalogSpell(t, g, "Candy Grapple", "Instant", woeB1GrappleOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, foe); p != 2 || tough != 2 {
		t.Errorf("the 5/5 is %d/%d, want 2/2", p, tough)
	}
}

func TestWoeB1CandyGrappleShrinksByFiveInsteadWhenBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: me.ID, Controller: me.ID,
	})
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 7, 7)
	p1g2Cast(t, g, "Candy Grapple", "Instant", woeB1GrappleOracle, "", game.CastSpellParams{
		Targets: cardRefs(foe), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, foe); p != 2 || tough != 2 {
		t.Errorf("the 7/7 is %d/%d, want 2/2 (-5/-5 instead of -3/-3, not both)", p, tough)
	}
}

// --- Commune with Nature ----------------------------------------------------------

func TestWoeB1CommuneWithNatureTakesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	var top [5]uuid.UUID
	cards := []struct{ name, typeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Bear", "Creature — Bear"},
		{"Shock", "Instant"},
		{"Golem", "Artifact Creature — Golem"},
		{"Rancor", "Enchantment — Aura"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].name, TypeLine: cards[i].typeLine, Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Commune with Nature", "Sorcery", woeB1CommuneOracle, nil)
	passPriorityAroundTable(t, g)

	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[1]] || !offered[top[3]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the two creature cards, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[3]) {
		t.Fatal("the creature card is not in hand")
	}
	if putInLibraryChoiceFor(g, me.ID) == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
}

func TestWoeB1CommuneWithNatureMayTakeNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D", "E")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Commune with Nature", "Sorcery", woeB1CommuneOracle, nil)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		answerChooseCards(t, g, me.ID)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d after taking nothing", hand, me.Hand.Size())
	}
}
