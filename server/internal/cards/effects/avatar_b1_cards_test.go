package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b1_cards_test.go — Avatar: The Last Airbender, batch 1 (#4459).
// Each test plays the real catalog entry and asserts what its text says.

const (
	avB1AangsJourneyOracle   = "ddf461b9-a205-4dc7-a9c3-c047b1ff709f"
	avB1AscensionOracle      = "c4d2fdf9-637d-4e97-ae33-45f2d27bf8cd"
	avB1AppaOracle           = "c6e28cf0-81c7-4f52-a29b-988cc6fd31ee"
	avB1EnthusiastsOracle    = "7b0d9d79-2e86-4079-8db3-500483c8807e"
	avB1HealerOracle         = "6ef464bd-2ed0-4f99-af4a-ebf276f55384"
	avB1FarmAnimalsOracle    = "e451b5aa-5832-42b3-8d72-f67e126a529f"
	avB1ConfrontationOracle  = "9703140f-36ee-4427-a390-c1dbdb9e4ee9"
	avB1JailerOracle         = "1b799c30-4a43-403d-b822-0555bd513913"
	avB1ProtectorsOracle     = "3bfb3913-808d-444b-97c3-a90af3bc585b"
	avB1EnterAvatarOracle    = "4e552be1-0702-4e18-822f-98611f8a8e70"
	avB1AscensionQuestCounts = "quest"
)

func TestAvatarB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB1AangsJourneyOracle:  "Aang's Journey",
		avB1AscensionOracle:     "Airbender Ascension",
		avB1AppaOracle:          "Appa, Loyal Sky Bison",
		avB1EnthusiastsOracle:   "Avatar Enthusiasts",
		avB1HealerOracle:        "Compassionate Healer",
		avB1FarmAnimalsOracle:   "Curious Farm Animals",
		avB1ConfrontationOracle: "Destined Confrontation",
		avB1JailerOracle:        "Earth Kingdom Jailer",
		avB1ProtectorsOracle:    "Earth Kingdom Protectors",
		avB1EnterAvatarOracle:   "Enter the Avatar State",
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

// avB1Hand puts a card in p's hand and returns its id.
func avB1Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// avB1Cast casts a card from p's hand from a main phase, failing the test on
// a refusal.
func avB1Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := avB1Hand(p, c)
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("casting %s: %v", c.Name, err)
	}
	return id
}

// avB1Plains seeds n Plains for owner.
func avB1Plains(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		pushLandFor(g, owner, "Plains", "Basic Land — Plains")
	}
}

// avB1Card reads a card through the layers.
func avB1Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

// avB1Named returns the battlefield permanent controlled by owner with the
// given name, or uuid.Nil.
func avB1Named(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// avB1ResolveCreature casts a creature from p's hand and resolves it, its
// entry triggers left for the caller.
func avB1ResolveCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := avB1Cast(t, g, p, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s never reached the battlefield", name)
	}
	return id
}

// avB1OpponentCreatureEnters has a creature token enter under owner's control
// outside their turn, so its entry is an event the source can see.
func avB1OpponentCreatureEnters(t *testing.T, g *game.Game, owner uuid.UUID, typeLine string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, game.Card{Name: "Their Token", TypeLine: "Token " + typeLine, Power: 1, Toughness: 1}, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
}

// --- Aang's Journey -------------------------------------------------------------

func TestAvatarB1AangsJourneyFindsABasicLandAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Plains Card", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Some Shrine", TypeLine: "Legendary Enchantment — Shrine"},
		game.Card{Name: "Nonbasic", TypeLine: "Land — Plains Island"},
	)
	avB1Plains(g, me.ID, 2)
	life := me.Life
	avB1Cast(t, g, me, game.Card{Name: "Aang's Journey", TypeLine: "Sorcery — Lesson", ManaCost: "{2}", OracleID: avB1AangsJourneyOracle},
		game.CastSpellParams{Strict: true, AutoTap: true})
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	beginSearchIfAsked(t, g, c)
	for _, bad := range [][]uuid.UUID{{ids[1]}, {ids[2]}, {ids[0], ids[1]}} {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, bad); err == nil {
			t.Fatalf("unkicked, the search accepted %v", bad)
		}
	}
	answerSearchByID(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land card should be in hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestAvatarB1AangsJourneyKickedFindsABasicAndAShrine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Plains A", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Plains B", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Shrine A", TypeLine: "Legendary Enchantment — Shrine"},
		game.Card{Name: "Shrine B", TypeLine: "Legendary Enchantment — Shrine"},
	)
	life := me.Life
	// Kicker {2} needs four mana; three cannot pay it.
	avB1Plains(g, me.ID, 3)
	advanceToMain(t, g)
	short := avB1Hand(me, game.Card{Name: "Aang's Journey", TypeLine: "Sorcery — Lesson", ManaCost: "{2}", OracleID: avB1AangsJourneyOracle})
	if err := g.CastSpell(me.ID, short, game.CastSpellParams{Strict: true, AutoTap: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a kicked cast was paid with three mana")
	}
	avB1Plains(g, me.ID, 1)
	if err := g.CastSpell(me.ID, short, game.CastSpellParams{Strict: true, AutoTap: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	beginSearchIfAsked(t, g, c)
	for _, bad := range [][]uuid.UUID{{ids[0], ids[1]}, {ids[2], ids[3]}, {ids[0], ids[2], ids[3]}} {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, bad); err == nil {
			t.Fatalf("kicked, the search accepted %v", bad)
		}
	}
	answerSearchByID(t, g, me.ID, ids[1], ids[3])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[3]) {
		t.Error("the basic land card and the Shrine card should be in hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Airbender Ascension ---------------------------------------------------------

func TestAvatarB1AirbenderAscensionAirbendsCountsAndFlickers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	asc := avB1Cast(t, g, me, game.Card{Name: "Airbender Ascension", TypeLine: "Enchantment", OracleID: avB1AscensionOracle}, game.CastSpellParams{})
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	assertAirbent(t, g, victim, opp.ID)

	// Three counters, then a creature entering under my control makes four;
	// one entering under an opponent's does not.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, 2) })
	avB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Elf")
	if n := avB1Card(t, g, asc).Counters[avB1AscensionQuestCounts]; n != 2 {
		t.Fatalf("an opponent's creature added a counter: %d", n)
	}
	avB1ResolveCreature(t, g, me, "My Elf", "Creature — Elf", "")
	passPriorityAroundTable(t, g)
	if n := avB1Card(t, g, asc).Counters[avB1AscensionQuestCounts]; n != 3 {
		t.Fatalf("quest counters %d, want 3", n)
	}

	// Three counters at my end step: nothing triggers.
	advanceTo(t, g, game.StepEnd)
	if p := latestPickTarget(g, me.ID); p != nil || g.Stack.Size() > 0 || len(g.StackMeta) > 0 {
		t.Fatal("the end step triggered with three quest counters")
	}
	if !g.Battlefield.Contains(mine) {
		t.Fatal("my creature moved without the fourth counter")
	}
	_ = mine
}

func TestAvatarB1AirbenderAscensionFlickersAtFourCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	asc := b12Push(g, me.ID, "Airbender Ascension", "Enchantment", avB1AscensionOracle, 0, 0)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, 4)
		_ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1)
	})
	advanceTo(t, g, game.StepEnd)
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Fatal("the creature should have left as the old object")
	}
	back := avB1Named(g, me.ID, "My Bear")
	if back == uuid.Nil {
		t.Fatal("the creature never returned to the battlefield")
	}
	if c := avB1Card(t, g, back); c.Owner != me.ID || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("returned as %+v; want a new object under its owner, counters gone", c.Counters)
	}
}

// The intervening "if" is checked again on resolution: a counter removed in
// response stops the flicker.
func TestAvatarB1AirbenderAscensionRechecksOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	asc := b12Push(g, me.ID, "Airbender Ascension", "Enchantment", avB1AscensionOracle, 0, 0)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, 4) })
	advanceTo(t, g, game.StepEnd)
	pickCard(t, g, me.ID, mine)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, -1) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Fatal("the flicker ran though the enchantment had three counters on resolution")
	}
}

// --- Appa, Loyal Sky Bison --------------------------------------------------------

func TestAvatarB1AppaEntersGrantsFlyingOrAirbends(t *testing.T) {
	for _, mode := range []int{0, 1} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		appa := avB1ResolveCreature(t, g, me, "Appa, Loyal Sky Bison", "Legendary Creature — Bison Ally", avB1AppaOracle)
		if !effectiveAbilitiesContain(t, g, appa, "flying") {
			t.Fatal("Appa has printed flying")
		}
		c := modePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("mode %d: no mode prompt on entry", mode)
		}
		answerMode(t, g, c, me.ID, mode)
		if mode == 1 {
			// "Another": Appa is no target for his own airbend.
			if p := latestPickTarget(g, me.ID); p != nil {
				if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: appa}}); err == nil {
					t.Fatal("Appa airbent himself")
				}
			}
		}
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		switch mode {
		case 0:
			if !effectiveAbilitiesContain(t, g, bear, "flying") {
				t.Error("the Bear should have flying until end of turn")
			}
		case 1:
			assertAirbent(t, g, bear, me.ID)
		}
	}
}

func TestAvatarB1AppaTriggersOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	appa := b12Push(g, me.ID, "Appa, Loyal Sky Bison", "Legendary Creature — Bison Ally", avB1AppaOracle, 4, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(appa, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt on attack")
	}
	answerMode(t, g, c, me.ID, 0)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the Bear should have flying until end of turn")
	}
}

// --- Avatar Enthusiasts -----------------------------------------------------------

func TestAvatarB1EnthusiastsGrowOnAnotherAllyOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fan := b12Push(g, me.ID, "Avatar Enthusiasts", "Creature — Human Peasant Ally", avB1EnthusiastsOracle, 2, 2)
	avB1ResolveCreature(t, g, me, "Plain Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	avB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Human Ally")
	if n := avB1Card(t, g, fan).Counters[game.CounterPlusOne]; n != 0 {
		t.Fatalf("a non-Ally or an opponent's Ally added a counter: %d", n)
	}
	avB1ResolveCreature(t, g, me, "My Ally", "Creature — Human Ally", "")
	passPriorityAroundTable(t, g)
	if n := avB1Card(t, g, fan).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("+1/+1 counters %d, want 1", n)
	}
}

// --- Compassionate Healer --------------------------------------------------------

func TestAvatarB1CompassionateHealerGainsAndScriesWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	healer := b12Push(g, me.ID, "Compassionate Healer", "Creature — Human Cleric Ally", avB1HealerOracle, 2, 2)
	life := me.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(healer, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if sc := scryChoiceFor(g, me.ID); sc == nil {
		t.Fatal("attacking tapped the Healer: a scry 1 is owed")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

// --- Curious Farm Animals --------------------------------------------------------

func TestAvatarB1CuriousFarmAnimalsDestroyAndGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	animals := b12Push(g, me.ID, "Curious Farm Animals", "Creature — Boar Elk Bird Ox", avB1FarmAnimalsOracle, 1, 1)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	advanceToMain(t, g)
	avB1Plains(g, me.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{
		Targets: cardRefs(rock), Strict: true, AutoTap: true,
	}); err == nil {
		t.Fatal("one mana paid {2}")
	}
	avB1Plains(g, me.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), Strict: true, AutoTap: true,
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{
		Targets: cardRefs(rock), Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(animals) || g.Battlefield.Contains(rock) {
		t.Error("the Animals are sacrificed and the artifact destroyed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (the sacrifice is a death)", me.Life, life+3)
	}
}

func TestAvatarB1CuriousFarmAnimalsActivatesWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	animals := b12Push(g, me.ID, "Curious Farm Animals", "Creature — Boar Elk Bird Ox", avB1FarmAnimalsOracle, 1, 1)
	life := me.Life
	advanceToMain(t, g)
	avB1Plains(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("up to one: zero targets is legal: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Destined Confrontation ------------------------------------------------------

func TestAvatarB1DestinedConfrontationKeepsTotalPowerFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := b12Creature(g, me.ID, "My Big", "Creature — Wurm", 5, 5)
	myTwo := b12Creature(g, me.ID, "My Two", "Creature — Bear", 2, 2)
	myOther := b12Creature(g, me.ID, "My Other Two", "Creature — Bear", 2, 2)
	theirFour := b12Creature(g, opp.ID, "Their Four", "Creature — Beast", 4, 4)
	theirOne := b12Creature(g, opp.ID, "Their One", "Creature — Elf", 1, 1)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Destined Confrontation", "Sorcery", avB1ConfrontationOracle, nil)
	passPriorityAroundTable(t, g)
	mine := chooseCardsChoiceFor(g, me.ID)
	if mine == nil || chooseCardsChoiceFor(g, opp.ID) == nil {
		t.Fatal("every player with creatures is asked")
	}
	if err := g.ResolveChooseCards(mine.ID, me.ID, []uuid.UUID{myBig}); err == nil {
		t.Fatal("a set with total power 5 was accepted")
	}
	answerChooseCards(t, g, me.ID, myTwo, myOther)
	answerChooseCards(t, g, opp.ID, theirFour)
	passPriorityAroundTable(t, g)
	for _, keep := range []uuid.UUID{myTwo, myOther, theirFour, rock} {
		if !g.Battlefield.Contains(keep) {
			t.Errorf("%s should be kept", keep)
		}
	}
	if g.Battlefield.Contains(myBig) || g.Battlefield.Contains(theirOne) {
		t.Error("everything not chosen is sacrificed")
	}
}

// --- Earth Kingdom Jailer --------------------------------------------------------

func TestAvatarB1EarthKingdomJailerExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushTypedCard(g, opp.ID, "Their Three", "Creature — Beast", "{2}{G}")
	small := pushTypedCard(g, opp.ID, "Their Two", "Artifact", "{2}")
	mineBig := pushTypedCard(g, me.ID, "My Three", "Enchantment", "{2}{W}")
	land := pushTypedCard(g, opp.ID, "Their Land", "Land", "")

	jailer := avB1ResolveCreature(t, g, me, "Earth Kingdom Jailer", "Creature — Human Soldier Ally", avB1JailerOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	for _, bad := range []uuid.UUID{small, mineBig, land} {
		if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(bad)); err == nil {
			t.Fatalf("%s is not a legal target", bad)
		}
	}
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Exile.Contains(big) {
		t.Fatal("the target should be in exile")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(jailer) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if avB1Named(g, opp.ID, "Their Three") == uuid.Nil {
		t.Fatal("the exiled card should return when the Jailer leaves")
	}
}

// --- Earth Kingdom Protectors ----------------------------------------------------

func TestAvatarB1EarthKingdomProtectorsShieldAnotherAlly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := b12Push(g, me.ID, "Earth Kingdom Protectors", "Creature — Human Soldier Ally", avB1ProtectorsOracle, 1, 1)
	ally := b12Creature(g, me.ID, "My Ally", "Creature — Human Ally", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ally", "Creature — Human Ally", 2, 2)
	if !effectiveAbilitiesContain(t, g, guard, "vigilance") {
		t.Error("printed vigilance is missing")
	}
	advanceToMain(t, g)
	for _, bad := range []uuid.UUID{guard, bear, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Targets: cardRefs(bad), Strict: true}); err == nil {
			t.Fatalf("%s is not another Ally you control", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Targets: cardRefs(ally), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(guard) {
		t.Error("the Protectors are sacrificed as the cost")
	}
	if !effectiveAbilitiesContain(t, g, ally, "indestructible") {
		t.Error("the Ally should be indestructible until end of turn")
	}
}

// --- Enter the Avatar State ------------------------------------------------------

func TestAvatarB1EnterTheAvatarState(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	avB1Plains(g, me.ID, 1)
	advanceToMain(t, g)
	spell := avB1Hand(me, game.Card{Name: "Enter the Avatar State", TypeLine: "Instant — Lesson", ManaCost: "{W}", OracleID: avB1EnterAvatarOracle})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardRefs(theirs), Strict: true, AutoTap: true}); err == nil {
		t.Fatal("an opponent's creature is not a creature you control")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardRefs(mine), Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !avB1Card(t, g, mine).HasSubtype("Avatar") {
		t.Error("the creature should be an Avatar until end of turn")
	}
	if !avB1Card(t, g, mine).HasSubtype("Bear") {
		t.Error("the Avatar type is in addition to its other types")
	}
	for _, kw := range []string{"flying", "first strike", "lifelink", "hexproof"} {
		if !effectiveAbilitiesContain(t, g, mine, kw) {
			t.Errorf("missing %s", kw)
		}
	}
}
