package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b30_cards_test.go — Secrets of Strixhaven, batch 30 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB30MuseSeekerOracle   = "31f3d37f-e613-427f-b34c-cc87611ea078"
	socB30EncouragementOracl = "62421084-96e6-40dd-8d34-7c8c1091d4a4"
	socB30NitaOracle         = "f589e5ee-399c-4613-b7a2-9ab2866cc830"
	socB30NewtOracle         = "1232c461-643c-40a3-901e-09b665f817da"
	socB30EducatorOracle     = "2b7362a9-3518-4071-ab1b-40689a50e12a"
	socB30RestorationOracle  = "00e75a6b-699e-4824-86e5-66d45245cd9b"
	socB30OrysaOracle        = "cf56a5b4-48fa-4d53-a4b9-14afe079c1d3"
	socB30HistorianOracle    = "e1ebb60e-fbf3-4f8f-a76f-d637ec59ffb0"
	socB30PageOracle         = "3bbad38b-7d88-4ff2-9fd3-5081536e7c56"
)

func TestSocB30EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB30MuseSeekerOracle:   "Muse Seeker",
		socB30EncouragementOracl: "Muse's Encouragement",
		socB30NitaOracle:         "Nita, Forum Conciliator",
		socB30NewtOracle:         "Noxious Newt",
		socB30EducatorOracle:     "Old-Growth Educator",
		socB30RestorationOracle:  "Oracle's Restoration",
		socB30OrysaOracle:        "Orysa, Tide Choreographer",
		socB30HistorianOracle:    "Owlin Historian",
		socB30PageOracle:         "Page, Loose Leaf",
	}
	for oracle, name := range want {
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

// --- Muse Seeker ------------------------------------------------------------

func socB30Seeker(g *game.Game, owner uuid.UUID) uuid.UUID {
	return socB17Perm(g, owner, "Muse Seeker", "Creature — Elf Wizard", socB30MuseSeekerOracle, "{1}{U}", 1, 2)
}

// socB30CastOpus casts a plain spell for `cost`, paid with exactly `pool`,
// and resolves the stack up to the discard prompt the Opus trigger may open.
func socB30CastOpus(t *testing.T, g *game.Game, typeLine, cost, pool string) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := socB17Hand(me, "Test Spell", typeLine, cost, socB17PlainSorceryOracle)
	floatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", cost, err)
	}
	for i := 0; i < 32; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) || discardChoiceFor(g, me.ID) != nil {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
		if g.Stack.Size() == 0 {
			return
		}
	}
}

func TestSocB30MuseSeekerDrawsThenDiscardsOnALowCostSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	socB30Seeker(g, me.ID)
	advanceToMain(t, g)
	before := me.Hand.Size()
	socB30CastOpus(t, g, "Instant", "{3}{U}", "UUUU")
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand is %d after the draw, want %d (the discard is still owed)", got, before+1)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("owes %d discards after a four-mana spell, want 1", discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand is %d after the discard, want %d", got, before)
	}
}

func TestSocB30MuseSeekerSkipsTheDiscardOnFiveOrMoreMana(t *testing.T) {
	for _, tc := range []struct{ name, cost, pool string }{
		{"five mana", "{4}{U}", "UUUUU"},
		{"six mana", "{5}{U}", "UUUUUU"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			socB30Seeker(g, me.ID)
			advanceToMain(t, g)
			before := me.Hand.Size()
			socB30CastOpus(t, g, "Sorcery", tc.cost, tc.pool)
			if got := me.Hand.Size(); got != before+1 {
				t.Errorf("hand is %d, want %d (one card drawn, nothing discarded)", got, before+1)
			}
			if discardOwed(g, me.ID) != 0 {
				t.Error("a discard is owed after a spell that cost five or more")
			}
		})
	}
}

func TestSocB30MuseSeekerIgnoresACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	socB30Seeker(g, me.ID)
	advanceToMain(t, g)
	before := me.Hand.Size()
	socB17CastSpellFor(t, g, "Creature — Bear", "{1}{U}", "UU")
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand is %d after a creature spell, want %d", got, before)
	}
}

// --- Muse's Encouragement ---------------------------------------------------

func TestSocB30EncouragementMakesAFlyingElementalThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Second", "Third")
	castCatalogSpell(t, g, "Muse's Encouragement", "Instant", socB30EncouragementOracl, nil)
	passPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := g.Battlefield.Cards[i]; c.Controller == me.ID && c.IsToken() {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no token was created")
	}
	if token.Power != 3 || token.Toughness != 3 || !hasEffectiveKeyword(t, g, token.InstanceID, "flying") {
		t.Errorf("token is %d/%d (flying %v), want a 3/3 flyer", token.Power, token.Toughness,
			hasEffectiveKeyword(t, g, token.InstanceID, "flying"))
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
	first, second := c.ScryCards[0], c.ScryCards[1]
	grave := me.Graveyard.Size()
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{first}, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(first) || me.Graveyard.Size() != grave+2 {
		t.Errorf("graveyard is %d, want the surveilled card and the spell (%d)", me.Graveyard.Size(), grave+2)
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != second {
		t.Error("the kept card is not back on top of the library")
	}
}

// --- Nita, Forum Conciliator -------------------------------------------------

func socB30Nita(g *game.Game, owner uuid.UUID) uuid.UUID {
	return socB17Perm(g, owner, "Nita, Forum Conciliator", "Legendary Creature — Human Advisor", socB30NitaOracle, "{1}{W}{B}", 2, 3)
}

// socB30CastNotMine casts an instant for `cost` that the opponent owns, paid
// with `pool`, and settles the table.
func socB30CastNotMine(t *testing.T, g *game.Game, me, owner *game.Player, cost, pool string) {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Borrowed Spell", TypeLine: "Instant", ManaCost: cost,
		Owner: owner.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	floatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast a spell I don't own: %v", err)
	}
	for answerAnyTriggerOrderPrompt(t, g, me.ID) {
	}
	passPriorityAroundTable(t, g)
}

func TestSocB30NitaPutsACounterOnEachCreatureWhenYouCastASpellYouDontOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	nita := socB30Nita(g, me.ID)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	socB30CastNotMine(t, g, me, opp, "{R}", "R")
	if counterCount(g, nita, game.CounterPlusOne) != 1 || counterCount(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("counters: Nita %d, Bear %d, want 1 each",
			counterCount(g, nita, game.CounterPlusOne), counterCount(g, bear, game.CounterPlusOne))
	}
	if counterCount(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestSocB30NitaIgnoresASpellYouOwn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nita := socB30Nita(g, me.ID)
	socB17CastSpellFor(t, g, "Instant", "{R}", "R")
	if got := counterCount(g, nita, game.CounterPlusOne); got != 0 {
		t.Errorf("Nita has %d counters after the controller cast their own spell, want 0", got)
	}
}

func TestSocB30NitaExilesAnOpponentsSpellAndLetsYouCastItWithAnyMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	nita := socB30Nita(g, me.ID)
	fodder := b12Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	spell := socB24Graveyard(opp, "Stolen Bolt", "Instant", "{R}")
	advanceToMain(t, g)
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, nita, 0, game.ActivateAbilityParams{
		Strict:       true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
		SacrificeIDs: []uuid.UUID{fodder},
	})
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the spell is in %v, want exile", z)
	}
	if _, still := battlefieldCard(g, fodder); still {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	// Mana of any type pays the {R}: a white mana will do.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "W")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast the exiled spell with white mana: %v", err)
	}
	for answerAnyTriggerOrderPrompt(t, g, me.ID) {
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("after resolving, the spell is in %v: it goes to exile, not the graveyard", z)
	}
	if opp.Graveyard.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the spell reached a graveyard")
	}
	// It was a spell Nita's controller does not own: one counter on Nita.
	if got := counterCount(g, nita, game.CounterPlusOne); got != 1 {
		t.Errorf("Nita has %d counters after casting the stolen spell, want 1", got)
	}
}

func TestSocB30NitaCantTargetYourOwnGraveyardOrACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	nita := socB30Nita(g, me.ID)
	fodder := b12Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	mine := socB24Graveyard(me, "My Bolt", "Instant", "{R}")
	creature := socB24Graveyard(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	advanceToMain(t, g)
	for _, target := range []uuid.UUID{mine, creature} {
		floatForTest(g, me, "CC")
		err := g.ActivateCatalogAbility(me.ID, nita, 0, game.ActivateAbilityParams{
			Strict:       true,
			Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: target}},
			SacrificeIDs: []uuid.UUID{fodder},
		})
		if err == nil {
			t.Errorf("activated against an illegal target %s", target)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}
}

func TestSocB30NitaActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	nita := socB30Nita(g, me.ID)
	fodder := b12Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	spell := socB24Graveyard(opp, "Stolen Bolt", "Instant", "{R}")
	advanceToMainOf(t, g, 1)
	floatForTest(g, me, "CC")
	err := g.ActivateCatalogAbility(me.ID, nita, 0, game.ActivateAbilityParams{
		Strict:       true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
		SacrificeIDs: []uuid.UUID{fodder},
	})
	if err == nil {
		t.Error("activated on an opponent's turn")
	}
}

// --- Noxious Newt -----------------------------------------------------------

func TestSocB30NewtHasDeathtouchAndTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	newt := socB17Perm(g, me.ID, "Noxious Newt", "Creature — Salamander", socB30NewtOracle, "{1}{G}", 1, 2)
	if !hasEffectiveKeyword(t, g, newt, "deathtouch") {
		t.Error("the Newt has no deathtouch")
	}
	untapForTest(g, newt)
	if err := g.ActivateManaAbility(me.ID, newt, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {G}: %v", err)
	}
	if card, _ := battlefieldCard(g, newt); !card.Tapped {
		t.Error("the {T} cost did not tap the Newt")
	}
	if err := g.ActivateManaAbility(me.ID, newt, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a tapped Newt made mana a second time")
	}
}

// --- Old-Growth Educator ----------------------------------------------------

func TestSocB30EducatorGrowsOnlyIfYouGainedLifeThisTurn(t *testing.T) {
	for _, gained := range []bool{false, true} {
		name := "no life gained"
		if gained {
			name = "life gained"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			if gained {
				gainLifeNow(t, g, uuid.Nil, me, 1)
			}
			id := castAndResolveCreature(t, g, "Old-Growth Educator", "Creature — Treefolk Druid", socB30EducatorOracle)
			passPriorityAroundTable(t, g)
			want := 0
			if gained {
				want = 2
			}
			if got := counterCount(g, id, game.CounterPlusOne); got != want {
				t.Errorf("Educator has %d +1/+1 counters, want %d", got, want)
			}
			for _, kw := range []string{"reach", "vigilance"} {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("no %s", kw)
				}
			}
		})
	}
}

func TestSocB30EducatorChecksLifeGainAsTheTriggerResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castAndResolveCreature(t, g, "Old-Growth Educator", "Creature — Treefolk Druid", socB30EducatorOracle)
	// The trigger is on the stack; gain life in response.
	gainLifeNow(t, g, uuid.Nil, me, 1)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("Educator has %d counters after a life gain in response, want 2", got)
	}
}

// --- Oracle's Restoration ---------------------------------------------------

func TestSocB30RestorationPumpsDrawsAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	id := socB17Hand(me, "Oracle's Restoration", "Sorcery", "{G}", socB30RestorationOracle)
	floatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := s58p6View(t, g, bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Bear is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand is %d, want %d (the spell left, one card drawn)", me.Hand.Size(), hand+1)
	}
	if me.Life != life+1 {
		t.Errorf("life is %d, want %d", me.Life, life+1)
	}
	advanceToMainOf(t, g, 1)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 2 {
		t.Errorf("Bear is still %d/%d on the next turn: the pump ends with the turn", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB30RestorationCantTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	id := socB17Hand(me, "Oracle's Restoration", "Sorcery", "{G}", socB30RestorationOracle)
	floatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("cast targeting an opponent's creature")
	}
}

// --- Orysa, Tide Choreographer ---------------------------------------------

func TestSocB30OrysaCostsThreeLessAtTotalToughnessTen(t *testing.T) {
	for _, tc := range []struct {
		name      string
		toughness []int
		pool      string
		castable  bool
	}{
		{"nine toughness pays full price", []int{4, 5}, "UUUUU", true},
		{"nine toughness is short at a discounted price", []int{4, 5}, "UU", false},
		{"ten toughness", []int{4, 6}, "UU", true},
		{"ten toughness short", []int{4, 6}, "U", false},
		{"no creatures short", nil, "UU", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			for _, tough := range tc.toughness {
				socB17Perm(g, me.ID, "Wall", "Creature — Wall", "", "{1}", 0, tough)
			}
			advanceToMain(t, g)
			id := socB17Hand(me, "Orysa, Tide Choreographer", "Legendary Creature — Merfolk Bard", "{4}{U}", socB30OrysaOracle)
			floatForTest(g, me, tc.pool)
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if tc.castable && err != nil {
				t.Errorf("could not cast with %s: %v", tc.pool, err)
			}
			if !tc.castable && err == nil {
				t.Errorf("cast with only %s", tc.pool)
			}
		})
	}
}

func TestSocB30OrysaDrawsTwoWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	before := me.Hand.Size()
	id := socB17Hand(me, "Orysa, Tide Choreographer", "Legendary Creature — Merfolk Bard", "{4}{U}", socB30OrysaOracle)
	floatForTest(g, me, "UUUUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand is %d, want %d (Orysa left the hand, two cards drawn)", got, before+2)
	}
}

// --- Owlin Historian --------------------------------------------------------

func socB30Historian(g *game.Game, owner uuid.UUID) uuid.UUID {
	return socB17Perm(g, owner, "Owlin Historian", "Creature — Bird Cleric", socB30HistorianOracle, "{2}{W}", 2, 3)
}

func TestSocB30HistorianSurveilsOneWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	castAndResolveCreature(t, g, "Owlin Historian", "Creature — Bird Cleric", socB30HistorianOracle)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %v, want one card", c)
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
}

func TestSocB30HistorianGrowsWhenACardLeavesYourGraveyardOncePerEvent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	owl := socB30Historian(g, me.ID)
	first := socB24Graveyard(me, "First", "Sorcery", "{1}")
	second := socB24Graveyard(me, "Second", "Sorcery", "{1}")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, first); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, owl); c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Fatalf("Historian is %d/%d after one card left, want 3/4", c.CurrentPower(), c.CurrentToughness())
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, second); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, owl); c.CurrentPower() != 4 || c.CurrentToughness() != 5 {
		t.Errorf("Historian is %d/%d after a second card left, want 4/5", c.CurrentPower(), c.CurrentToughness())
	}
	advanceToMainOf(t, g, 1)
	if c := s58p6View(t, g, owl); c.CurrentPower() != 2 {
		t.Errorf("Historian is %d/%d on the next turn: the bonus ends with the turn", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB30HistorianIgnoresAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	owl := socB30Historian(g, me.ID)
	theirs := socB24Graveyard(opp, "Theirs", "Sorcery", "{1}")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID},
		game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, theirs); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, owl); c.CurrentPower() != 2 {
		t.Errorf("Historian is %d/%d after an opponent's card left their graveyard, want 2/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Page, Loose Leaf -------------------------------------------------------

func socB30Page(g *game.Game, owner uuid.UUID) uuid.UUID {
	return socB17Perm(g, owner, "Page, Loose Leaf", "Legendary Artifact Creature — Construct", socB30PageOracle, "{2}", 0, 2)
}

func TestSocB30PageTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	page := socB30Page(g, me.ID)
	untapForTest(g, page)
	if err := g.ActivateManaAbility(me.ID, page, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if card, _ := battlefieldCard(g, page); !card.Tapped {
		t.Error("the {T} cost did not tap Page")
	}
}

func TestSocB30PageGrandeurFindsTheFirstInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	page := socB30Page(g, me.ID)
	extra := socB17Hand(me, "Page, Loose Leaf", "Legendary Artifact Creature — Construct", "{2}", socB30PageOracle)
	// Top to bottom: a creature, a land, an instant, then the old library.
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deep Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	instant := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: instant, Name: "Found Instant", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	libBefore := me.Library.Size()
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, page, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{extra}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(instant) {
		t.Fatal("the first instant was not put into the hand")
	}
	if !me.Graveyard.Contains(extra) {
		t.Error("the discarded Page is not in the graveyard")
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library is %d, want %d: only the found card left it", me.Library.Size(), libBefore-1)
	}
	// The revealed creature and land are now the bottom two cards, in some order.
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if !bottom[bear] || !bottom[land] {
		t.Error("the revealed cards are not on the bottom of the library")
	}
}

func TestSocB30PageGrandeurWithNoInstantOrSorceryBottomsTheWholeLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	page := socB30Page(g, me.ID)
	extra := socB17Hand(me, "Page, Loose Leaf", "Legendary Artifact Creature — Construct", "{2}", socB30PageOracle)
	me.Library.Cards = me.Library.Cards[:0]
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, page, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{extra}})
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 3 {
		t.Errorf("library is %d, want 3: nothing is put into the hand", me.Library.Size())
	}
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand is %d, want %d (only the discard)", me.Hand.Size(), hand-1)
	}
}

func TestSocB30PageGrandeurNeedsAnotherCopyInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	page := socB30Page(g, me.ID)
	other := socB17Hand(me, "Something Else", "Instant", "{1}", "")
	advanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, page, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{other}})
	if err == nil {
		t.Error("Grandeur was paid by discarding a card with another name")
	}
}
