package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_mp2_cards_test.go — Foundations, M–P, second batch (#4465): the
// cards from that stretch of the alphabet that the first batch
// (fdn_mp_cards_test.go) left and that needed no new seam. Each test
// plays the real catalog entry through the engine and asserts what its
// text says.

const (
	fdnMP2MicromancerOracle    = "d2605782-09f1-4d52-b3e7-5ec1442bd9b5"
	fdnMP2MindsparkerOracle    = "54866c25-b332-46c8-a2c5-1a71b702ddc0"
	fdnMP2MischievousMystic    = "542831c6-6279-4977-9b29-0fb9c0ebf04d"
	fdnMP2MoldAdderOracle      = "2091eb7e-2357-4ef8-b15e-1e6f82176340"
	fdnMP2MysticalTeachings    = "81d72511-88b4-45ec-a5a9-e722caf55577"
	fdnMP2NeedletoothPack      = "ac68fe30-a4a5-45f6-8782-fed57718c42e"
	fdnMP2NessianHornbeetle    = "ab097f46-a7bf-4cf1-a96c-36b681fe79da"
	fdnMP2NivMizzetVisionary   = "d9f1dc61-831e-4244-a2da-aa9da6fd5b50"
	fdnMP2NullpriestOracle     = "c1111228-ef88-4b64-91e0-66ee32a1f5e4"
	fdnMP2PiratesCutlassOracle = "ac77591c-39d7-4fd1-a6e9-805f57c6614a"
	fdnMP2PrayerOfBinding      = "4505ff9f-97f1-45d1-aea9-5590fefb4a4d"
	fdnMP2PredatorOozeOracle   = "a51ec70d-1356-4403-a858-b99445bac54a"
)

func TestFdnMP2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnMP2MicromancerOracle:    "Micromancer",
		fdnMP2MindsparkerOracle:    "Mindsparker",
		fdnMP2MischievousMystic:    "Mischievous Mystic",
		fdnMP2MoldAdderOracle:      "Mold Adder",
		fdnMP2MysticalTeachings:    "Mystical Teachings",
		fdnMP2NeedletoothPack:      "Needletooth Pack",
		fdnMP2NessianHornbeetle:    "Nessian Hornbeetle",
		fdnMP2NivMizzetVisionary:   "Niv-Mizzet, Visionary",
		fdnMP2NullpriestOracle:     "Nullpriest of Oblivion",
		fdnMP2PiratesCutlassOracle: "Pirate's Cutlass",
		fdnMP2PrayerOfBinding:      "Prayer of Binding",
		fdnMP2PredatorOozeOracle:   "Predator Ooze",
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

// fdnMP2OpponentInstant has `opp` cast a vanilla instant of `colors` at
// instant speed on the active seat's turn, and returns it.
func fdnMP2OpponentInstant(t *testing.T, g *game.Game, opp *game.Player, name string, colors ...string) uuid.UUID {
	t.Helper()
	return b09OpponentCastsCard(t, g, opp, game.Card{Name: name, TypeLine: "Instant", Colors: colors})
}

// --- Micromancer --------------------------------------------------------------

func TestFdnMP2MicromancerFindsOnlyAnInstantOrSorceryWithManaValueOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "One Instant", TypeLine: "Instant", ManaCost: "{U}"},
		game.Card{Name: "One Sorcery", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Two Instant", TypeLine: "Instant", ManaCost: "{1}{U}"},
		game.Card{Name: "Zero Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"},
		game.Card{Name: "One Creature", TypeLine: "Creature — Bear", ManaCost: "{G}"},
	)
	castAndResolveCreature(t, g, "Micromancer", "Creature — Human Wizard", fdnMP2MicromancerOracle)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, offered := range []string{"One Instant", "One Sorcery"} {
		if searchOptionNamed(g, c, offered) == uuid.Nil {
			t.Errorf("%s should be offered", offered)
		}
	}
	for _, refused := range []string{"Two Instant", "Zero Sorcery", "One Creature"} {
		if searchOptionNamed(g, c, refused) != uuid.Nil {
			t.Errorf("%s should not be offered", refused)
		}
	}
	pick := searchOptionNamed(g, c, "One Sorcery")
	answerSearchByID(t, g, me.ID, pick)
	if !me.Hand.Contains(pick) {
		t.Error("the chosen sorcery did not reach the hand")
	}
}

func TestFdnMP2MicromancerMayFindNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, game.Card{Name: "One Instant", TypeLine: "Instant", ManaCost: "{U}"})
	castAndResolveCreature(t, g, "Micromancer", "Creature — Human Wizard", fdnMP2MicromancerOracle)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt")
	}
	before := me.Hand.Size()
	answerSearchFailToFind(t, g, me.ID)
	if me.Hand.Size() != before {
		t.Error("declining the search put a card in hand")
	}
}

// --- Mindsparker --------------------------------------------------------------

func TestFdnMP2MindsparkerBurnsAnOpponentForAWhiteOrBlueInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sparker := b12Push(g, me.ID, "Mindsparker", "Creature — Elemental", fdnMP2MindsparkerOracle, 3, 2)
	assertKeywords(t, g, sparker, "first strike")

	life := lifeOf(g, opp.ID)
	fdnMP2OpponentInstant(t, g, opp, "Blue Instant", "U")
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Fatalf("after a blue instant the caster is at %d, want %d", got, life-2)
	}
	fdnMP2OpponentInstant(t, g, opp, "White Instant", "W")
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Fatalf("after a white instant the caster is at %d, want %d", got, life-4)
	}
	fdnMP2OpponentInstant(t, g, opp, "Red Instant", "R")
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("a red instant cost the caster life: %d, want %d", got, life-4)
	}
	b09OpponentCastsCard(t, g, opp, game.Card{Name: "Blue Flash Bear", TypeLine: "Creature — Bear",
		Colors: []string{"U"}, Power: 2, Toughness: 2, Keywords: []string{"flash"}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("a blue creature spell cost the caster life: %d, want %d", got, life-4)
	}
}

func TestFdnMP2MindsparkerIgnoresYourOwnSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Mindsparker", "Creature — Elemental", fdnMP2MindsparkerOracle, 3, 2)
	life := lifeOf(g, me.ID)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "My Blue Sorcery", TypeLine: "Sorcery",
		Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life {
		t.Errorf("your own blue sorcery cost you life: %d, want %d", got, life)
	}
}

// --- Mischievous Mystic -------------------------------------------------------

func TestFdnMP2MischievousMysticMakesAFaerieOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	// The Mystic's controller is not the active player, so nothing has
	// been drawn by them this turn yet.
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mystic := b12Push(g, owner.ID, "Mischievous Mystic", "Creature — Human Wizard", fdnMP2MischievousMystic, 2, 1)
	assertKeywords(t, g, mystic, "flying")
	seedLibrary(owner, "A", "B", "C", "D")

	draw := func() {
		t.Helper()
		g.WithWriteLock(func() {
			if err := g.DrawNForEffect(owner.ID, 1); err != nil {
				t.Fatal(err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	draw()
	if n := tmtB1Count(g, owner.ID, "Faerie"); n != 0 {
		t.Fatalf("%d Faeries after the first draw, want 0", n)
	}
	draw()
	if n := tmtB1Count(g, owner.ID, "Faerie"); n != 1 {
		t.Fatalf("%d Faeries after the second draw, want 1", n)
	}
	draw()
	if n := tmtB1Count(g, owner.ID, "Faerie"); n != 1 {
		t.Errorf("%d Faeries after the third draw, want 1", n)
	}
	var faerie uuid.UUID
	var blue bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner.ID && c.Name == "Faerie" {
				faerie, blue = c.InstanceID, c.HasColor("U")
			}
		}
	})
	assertKeywords(t, g, faerie, "flying")
	if p, tough := effectivePower(t, g, faerie), effectiveToughness(t, g, faerie); p != 1 || tough != 1 || !blue {
		t.Errorf("the token is %d/%d (blue %v), want a 1/1 blue flier", p, tough, blue)
	}
}

// --- Mold Adder ---------------------------------------------------------------

func TestFdnMP2MoldAdderMayGrowWhenAnOpponentCastsBlueOrBlack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adder := b12Push(g, me.ID, "Mold Adder", "Creature — Fungus Snake", fdnMP2MoldAdderOracle, 1, 1)

	fdnMP2OpponentInstant(t, g, opp, "Blue Instant", "U")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, adder, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after a blue spell (accepted), want 1", got)
	}

	fdnMP2OpponentInstant(t, g, opp, "Black Instant", "B")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, adder, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after a black spell (declined), want 1", got)
	}

	fdnMP2OpponentInstant(t, g, opp, "Red Instant", "R")
	if assistCardTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a red spell asked the Mold Adder's question")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, adder, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a red spell, want 1", got)
	}
}

// --- Mystical Teachings -------------------------------------------------------

func TestFdnMP2MysticalTeachingsFindsAnInstantOrFlashCardAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := pushLibraryCardForTest(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant"})
	snapper := pushLibraryCardForTest(me, game.Card{Name: "Ambush Snapper", TypeLine: "Creature — Turtle", Keywords: []string{"flash"}})
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})

	teachings := castCatalogSpell(t, g, "Mystical Teachings", "Instant", fdnMP2MysticalTeachings, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if !hasID(c.SearchCards, bolt) || !hasID(c.SearchCards, snapper) || hasID(c.SearchCards, bear) {
		t.Errorf("offered %v; want the instant and the flash creature, not the bear", c.SearchCards)
	}
	answerSearchByID(t, g, me.ID, snapper)
	if !me.Hand.Contains(snapper) {
		t.Fatal("the chosen card did not reach the hand")
	}
	if !me.Graveyard.Contains(teachings) {
		t.Fatal("the resolved Teachings is not in the graveyard")
	}

	if err := g.CastSpell(me.ID, teachings, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c == nil || !hasID(c.SearchCards, bolt) {
		t.Fatal("the flashback cast did not search for the instant")
	}
	answerSearchByID(t, g, me.ID, bolt)
	if !me.Hand.Contains(bolt) {
		t.Error("the flashback search did not put the instant in hand")
	}
	if !g.Exile.Contains(teachings) {
		t.Error("the flashed-back Teachings is not in exile")
	}
}

// --- Needletooth Pack ---------------------------------------------------------

func TestFdnMP2NeedletoothPackPutsTwoCountersAfterACreatureDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Needletooth Pack", "Creature — Dinosaur", fdnMP2NeedletoothPack, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	killCreature(t, g, opp.ID, theirs)

	advanceToEndStep(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 2 {
		t.Errorf("the bear has %d counters, want 2", got)
	}
}

func TestFdnMP2NeedletoothPackDoesNothingWhenNoCreatureDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pack := b12Push(g, me.ID, "Needletooth Pack", "Creature — Dinosaur", fdnMP2NeedletoothPack, 4, 5)
	advanceToEndStep(t, g)
	if latestPickTarget(g, me.ID) != nil || !stackFullyEmpty(g) {
		t.Fatal("the end step trigger went on the stack with no creature having died")
	}
	if got := plusOneCounters(g, pack); got != 0 {
		t.Errorf("the Pack has %d counters, want 0", got)
	}
}

// --- Nessian Hornbeetle -------------------------------------------------------

func TestFdnMP2NessianHornbeetleGrowsOnlyBesideAnotherBigCreature(t *testing.T) {
	for _, tc := range []struct {
		name      string
		ownPower  int
		allyPower int // 0: no ally
		oppBig    bool
		want      int
	}{
		{"another power-4 creature", 2, 4, false, 1},
		{"no other creature", 2, 0, false, 0},
		{"a smaller ally", 2, 3, false, 0},
		{"its own power does not count", 4, 0, false, 0},
		{"an opponent's big creature does not count", 2, 0, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			beetle := b12Push(g, me.ID, "Nessian Hornbeetle", "Creature — Insect", fdnMP2NessianHornbeetle, tc.ownPower, 2)
			if tc.allyPower > 0 {
				pushVanillaCreature(g, me.ID, "Ally", tc.allyPower, 4)
			}
			if tc.oppBig {
				pushVanillaCreature(g, opp.ID, "Their Giant", 5, 5)
			}
			advanceTo(t, g, game.StepBeginCombat)
			passPriorityAroundTable(t, g)
			if got := countersOn(g, beetle, game.CounterPlusOne); got != tc.want {
				t.Errorf("counters %d, want %d", got, tc.want)
			}
		})
	}
}

// --- Niv-Mizzet, Visionary ----------------------------------------------------

func TestFdnMP2NivMizzetDrawsForNoncombatDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	niv := pushCatalogPermanent(g, me.ID, "Niv-Mizzet, Visionary", "Legendary Creature — Dragon Wizard", fdnMP2NivMizzetVisionary, false)
	assertKeywords(t, g, niv, "flying")
	if spec, _ := Lookup(fdnMP2NivMizzetVisionary); !spec.NoMaxHandSize {
		t.Error("Niv-Mizzet does not declare no maximum hand size")
	}
	pinger := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)
	theirs := b12Creature(g, other.ID, "Their Pinger", "Creature — Wizard", 1, 1)
	seedLibrary(me, "A", "B", "C", "D", "E")

	before := fdnMPHandSize(g, me)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, opp.ID, 3) })
	passPriorityAroundTable(t, g)
	if got := fdnMPHandSize(g, me); got != before+3 {
		t.Fatalf("hand %d -> %d after 3 damage to an opponent, want +3", before, got)
	}

	before = fdnMPHandSize(g, me)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if got := fdnMPHandSize(g, me); got != before {
		t.Errorf("an opponent's source drew you %d cards", got-before)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := fdnMPHandSize(g, me); got != before {
		t.Errorf("damage to yourself drew you %d cards", got-before)
	}
}

// --- Nullpriest of Oblivion ---------------------------------------------------

func TestFdnMP2NullpriestReanimatesOnlyWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := seedGraveyardCreature(me, "Dead Giant", "{3}{B}")

	priest, err := castWithOptionalCosts(t, g, "Nullpriest of Oblivion", "Creature — Vampire Cleric", fdnMP2NullpriestOracle, nil, nil, nil)
	if err != nil {
		t.Fatalf("unkicked CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Fatal("an unkicked Nullpriest triggered")
	}
	if !me.Graveyard.Contains(dead) {
		t.Fatal("an unkicked Nullpriest returned a creature")
	}
	assertKeywords(t, g, priest, "lifelink", "menace")

	if _, err := castWithOptionalCosts(t, g, "Nullpriest of Oblivion", "Creature — Vampire Cleric", fdnMP2NullpriestOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("kicked CastSpell: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, dead)
	if !ok || c.Controller != me.ID {
		t.Error("the kicked Nullpriest did not return the creature card to the battlefield under your control")
	}
}

// --- Pirate's Cutlass ---------------------------------------------------------

func TestFdnMP2PiratesCutlassAttachesToAPirateOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pirate := b12Creature(g, me.ID, "Test Pirate", "Creature — Human Pirate", 2, 2)
	bear := b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	cutlass := castCatalogSpell(t, g, "Pirate's Cutlass", "Artifact — Equipment", fdnMP2PiratesCutlassOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: pirate}, true)

	if c, _ := battlefieldCard(g, cutlass); !c.IsAttachedTo(pirate) {
		t.Fatal("the Cutlass is not attached to the Pirate")
	}
	if p, tt := ptOf(t, g, pirate); p != 4 || tt != 3 {
		t.Errorf("the Pirate is %d/%d, want 4/3", p, tt)
	}
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("the bear is %d/%d, want an unchanged 2/2", p, tt)
	}
}

func TestFdnMP2PiratesCutlassStaysUnattachedWithoutAPirateAndEquipsForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	cutlass := castCatalogSpell(t, g, "Pirate's Cutlass", "Artifact — Equipment", fdnMP2PiratesCutlassOracle, nil)
	assistCardSettle(t, g, me.ID, nil, true)
	if c, _ := battlefieldCard(g, cutlass); c.IsAttachedTo(bear) {
		t.Fatal("the Cutlass attached to a creature that is not a Pirate")
	}
	equipTo(t, g, me.ID, cutlass, bear)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 3 {
		t.Errorf("after equipping the bear is %d/%d, want 4/3", p, tt)
	}
}

// --- Prayer of Binding --------------------------------------------------------

func TestFdnMP2PrayerOfBindingExilesANonlandPermanentUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := pushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	land := seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life

	prayer := castHoldCreature(t, g, "Prayer of Binding", "Enchantment", fdnMP2PrayerOfBinding, 0, 0)
	assertKeywords(t, g, prayer, "flash")
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{land, mine} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || tmtB1Count(g, opp.ID, "Their Rock") != 0 {
		t.Fatal("the artifact was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(prayer) })
	g.RunStateChecksForTest()
	if tmtB1Count(g, opp.ID, "Their Rock") != 1 {
		t.Error("the artifact did not return when the Prayer left")
	}
}

func TestFdnMP2PrayerOfBindingWithNoTargetStillGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := pushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	life := me.Life
	castHoldCreature(t, g, "Prayer of Binding", "Enchantment", fdnMP2PrayerOfBinding, 0, 0)
	b04WaitForPick(t, g, me.ID)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("choosing no target still exiled a permanent")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Predator Ooze ------------------------------------------------------------

func TestFdnMP2PredatorOozeGrowsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := b12Push(g, me.ID, "Predator Ooze", "Creature — Ooze", fdnMP2PredatorOozeOracle, 1, 1)
	assertKeywords(t, g, ooze, "indestructible")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ooze, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ooze, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after attacking, want 1", got)
	}
}

func TestFdnMP2PredatorOozeGrowsWhenACreatureItDamagedDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := b12Push(g, me.ID, "Predator Ooze", "Creature — Ooze", fdnMP2PredatorOozeOracle, 1, 1)
	hit := b12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 3)
	other := b12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 3)
	pinger := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(ooze, hit, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(pinger, other, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ooze, game.CounterPlusOne); n != 0 {
		t.Fatalf("counters = %d after a creature another source damaged died, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hit) })
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ooze, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d after the creature the Ooze damaged died, want 1", n)
	}
}
