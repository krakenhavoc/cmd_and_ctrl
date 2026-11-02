package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b4_cards_test.go — Marvel's Spider-Man, batch 4 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB4LizardOracle     = "48d1b217-b068-468c-bc6c-4bef9e44cc6e"
	spmB4MadameWebOracle  = "b1e0841d-b004-40f9-a644-aecb422f6ad5"
	spmB4MaryJaneOracle   = "6d44c87b-8945-4c39-b034-7690cf483992"
	spmB4CarnageOracle    = "7351fca5-8672-4ba3-a015-bd4b6b289efc"
	spmB4MobsterOracle    = "b0672aa3-3207-4ff8-af8f-bdabacbb4d04"
	spmB4EnforcersOracle  = "b32a7b97-0883-4c9f-a23d-ad269efd8cc3"
	spmB4MoltenManOracle  = "c6fcf3c1-0626-4d53-92c0-caa22fa61e9a"
	spmB4MorbiusOracle    = "1a947e4d-8a14-45f1-bb80-760a317b1ed0"
	spmB4MorlunOracle     = "273ad15c-366a-4691-978b-d6028b57005a"
	spmB4PhantasmOracle   = "99419378-8a23-44f3-aebe-e62d6e20e877"
	spmB4PhantasmTypeLine = "Creature — Illusion Villain"
)

func TestSpmB4EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB4LizardOracle:    "Lizard, Connors's Curse",
		spmB4MadameWebOracle: "Madame Web, Clairvoyant",
		spmB4MaryJaneOracle:  "Mary Jane Watson",
		spmB4CarnageOracle:   "Maximum Carnage",
		spmB4MobsterOracle:   "Mechanical Mobster",
		spmB4EnforcersOracle: "Merciless Enforcers",
		spmB4MoltenManOracle: "Molten Man, Inferno Incarnate",
		spmB4MorbiusOracle:   "Morbius the Living Vampire",
		spmB4MorlunOracle:    "Morlun, Devourer of Spiders",
		spmB4PhantasmOracle:  "Mysterio's Phantasm",
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

// spmB4Cast puts a card with real P/T in the active player's hand and casts
// it for free (X given), returning its id and the cast's error.
func spmB4Cast(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, power, toughness, x int) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x})
}

func spmB4Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

// --- Merciless Enforcers ---

func TestSpmB4EnforcersPingEachOpponentAndLifelinkGainsForIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	enf := b12Push(g, me.ID, "Merciless Enforcers", "Creature — Human Mercenary Villain", spmB4EnforcersOracle, 2, 1)
	assertKeywords(t, g, enf, "lifelink")
	life := me.Life
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		lives[s.ID] = s.Life
	}
	floatMana(t, g, me, "{B}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, enf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		if s.Life != lives[s.ID]-1 {
			t.Errorf("seat %d life %d, want %d", s.ID, s.Life, lives[s.ID]-1)
		}
	}
	if me.Life != life+len(g.Seats)-1 {
		t.Errorf("lifelink: life %d -> %d, want +%d (one per damage event)", life, me.Life, len(g.Seats)-1)
	}
}

func TestSpmB4EnforcersRefuseTooLittleMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	enf := b12Push(g, me.ID, "Merciless Enforcers", "Creature — Human Mercenary Villain", spmB4EnforcersOracle, 2, 1)
	floatMana(t, g, me, "{B}{C}")
	if err := g.ActivateCatalogAbility(me.ID, enf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {3}{B} with two mana")
	}
}

// --- Mysterio's Phantasm ---

func TestSpmB4PhantasmHasKeywordsAndMillsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ph := b12Push(g, me.ID, "Mysterio's Phantasm", spmB4PhantasmTypeLine, spmB4PhantasmOracle, 1, 3)
	assertKeywords(t, g, ph, "flying", "vigilance")
	seedLibrary(me, "One", "Two", "Three")
	lib, gy := me.Library.Size(), me.Graveyard.Size()
	declareAttack(t, g, opp.ID, ph)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib-1 || me.Graveyard.Size() != gy+1 {
		t.Errorf("library %d -> %d, graveyard %d -> %d; want one milled", lib, me.Library.Size(), gy, me.Graveyard.Size())
	}
	// A different attacker does not trigger it.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	b12Push(g2, me2.ID, "Mysterio's Phantasm", spmB4PhantasmTypeLine, spmB4PhantasmOracle, 1, 3)
	bear := b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibrary(me2, "One", "Two", "Three")
	lib2 := me2.Library.Size()
	declareAttack(t, g2, opp2.ID, bear)
	passPriorityAroundTable(t, g2)
	if me2.Library.Size() != lib2 {
		t.Error("another creature attacking milled")
	}
}

// --- Mary Jane Watson ---

func TestSpmB4MaryJaneDrawsOnceEachTurnForASpider(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Mary Jane Watson", "Legendary Creature — Human Performer", spmB4MaryJaneOracle, 2, 2)
	seedLibrary(me, "One", "Two", "Three", "Four")
	hand := me.Hand.Size()

	// A Spider an opponent controls does nothing.
	pushPermanent(g, opp.ID, game.Card{Name: "Their Spider", TypeLine: "Creature — Spider", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("an opponent's Spider drew a card")
	}
	// A non-Spider you control does nothing.
	if _, err := spmB4Cast(t, g, "Plain Bear", "Creature — Bear", "", "", 2, 2, 0); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a Bear drew a card: hand %d -> %d", hand, me.Hand.Size())
	}
	// The first Spider draws.
	if _, err := spmB4Cast(t, g, "Spider One", "Creature — Spider Hero", "", "", 2, 2, 0); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("the first Spider: hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	// A second Spider the same turn does not.
	if _, err := spmB4Cast(t, g, "Spider Two", "Creature — Spider Hero", "", "", 2, 2, 0); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("the second Spider drew again: hand %d -> %d", hand, me.Hand.Size())
	}
}

func TestSpmB4MaryJaneCountsATokenSpider(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Mary Jane Watson", "Legendary Creature — Human Performer", spmB4MaryJaneOracle, 2, 2)
	seedLibrary(me, "One", "Two", "Three")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Spider with reach"), 1) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a Spider token: hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

// --- Morbius the Living Vampire ---

func TestSpmB4MorbiusHasKeywordsAndDigsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	// On the battlefield it has its three keywords.
	bf := b12Push(g, me.ID, "Morbius the Living Vampire", "Legendary Creature — Vampire Scientist Villain", spmB4MorbiusOracle, 3, 1)
	assertKeywords(t, g, bf, "flying", "vigilance", "lifelink")

	mor := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: mor, Name: "Morbius the Living Vampire", TypeLine: "Legendary Creature — Vampire Scientist Villain",
		OracleID: spmB4MorbiusOracle, Owner: me.ID, Controller: me.ID,
	})
	ids := seedLibrary(me, "A", "B", "C", "D")
	a, b, c, d := ids[0], ids[1], ids[2], ids[3]
	hand := me.Hand.Size()

	floatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, mor, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {U}{B} with one mana")
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(game.ManaToken{Color: "B"}) })
	if err := g.ActivateCatalogAbility(me.ID, mor, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(mor) {
		t.Fatal("Morbius was not exiled to pay the cost")
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("want a take-exactly-one prompt over three cards, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(b) || me.Hand.Size() != hand+1 {
		t.Fatal("the picked card did not go to hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 2 {
		t.Fatalf("want a bottom ordering of the two others, got %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{c, a}, nil); err != nil {
		t.Fatal(err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != d {
		t.Error("the fourth card should now be on top")
	}
	if got := libraryBottomIDs(me, 2); !sameIDs(got, []uuid.UUID{c, a}) {
		t.Errorf("bottom two %v, want [C A] top-first", got)
	}
}

// --- Morlun, Devourer of Spiders ---

func TestSpmB4MorlunEntersWithXCountersAndHitsAnOpponentForX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	id, err := spmB4Cast(t, g, "Morlun, Devourer of Spiders", "Legendary Creature — Vampire Villain", spmB4MorlunOracle, "{X}{B}{B}", 2, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	// Only an opponent is a legal pick.
	ch := latestPickTarget(g, me.ID)
	if ch == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(ch.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("the caster was a legal target opponent")
	}
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if plusOnes(g, id) != 3 {
		t.Errorf("Morlun has %d counters, want 3", plusOnes(g, id))
	}
	if m := spmB4Live(t, g, id); m.CurrentPower() != 5 || m.CurrentToughness() != 4 {
		t.Errorf("Morlun is %d/%d, want 5/4", m.CurrentPower(), m.CurrentToughness())
	}
	if opp.Life != oppLife-3 {
		t.Errorf("opponent life %d -> %d, want -3", oppLife, opp.Life)
	}
	if me.Life != life+3 {
		t.Errorf("lifelink: life %d -> %d, want +3", life, me.Life)
	}
	assertKeywords(t, g, id, "lifelink")
}

func TestSpmB4MorlunForXZeroDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppLife := opp.Life
	id, err := spmB4Cast(t, g, "Morlun, Devourer of Spiders", "Legendary Creature — Vampire Villain", spmB4MorlunOracle, "{X}{B}{B}", 2, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if ch := latestPickTarget(g, me.ID); ch != nil {
		pickPlayerTarget(t, g, me.ID, opp.ID)
		passPriorityAroundTable(t, g)
	}
	if opp.Life != oppLife {
		t.Errorf("X = 0 dealt damage: %d -> %d", oppLife, opp.Life)
	}
	if plusOnes(g, id) != 0 {
		t.Errorf("X = 0 put %d counters", plusOnes(g, id))
	}
}

func TestSpmB4MorlunIsNotCastableWithoutItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Morlun, Devourer of Spiders", TypeLine: "Legendary Creature — Vampire Villain",
		OracleID: spmB4MorlunOracle, ManaCost: "{X}{B}{B}", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true}); err == nil {
		t.Fatal("cast Morlun for X=2 with one mana")
	}
}

// --- Mechanical Mobster ---

func TestSpmB4MobsterExilesAGraveyardCardAndConnives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushGraveyardCardTyped(opp, "Dead Thing", "Creature — Bear")
	seedLibrary(me, "Draw Me", "Two", "Three")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	spmB4Cast(t, g, "Mechanical Mobster", "Artifact Creature — Human Robot Villain", spmB4MobsterOracle, "{3}", 2, 1, 0)
	passPriorityAroundTable(t, g)
	// One prompt per clause: the graveyard card first, then the creature.
	answerPickTarget(t, g, victim)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) || opp.Graveyard.Contains(victim) {
		t.Error("the graveyard card was not exiled")
	}
	// Connive: one drawn, then a discard prompt; a nonland discard grows the Bear.
	drawn := me.Hand.Size()
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("connive raised no discard prompt")
	}
	nonland := pick.ChooseCards[0]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{nonland}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 {
		t.Errorf("the Bear has %d counters, want 1 (a nonland card was discarded)", plusOnes(g, bear))
	}
	if me.Hand.Size() != drawn-1 {
		t.Errorf("hand %d, want %d after the discard", me.Hand.Size(), drawn-1)
	}
}

func TestSpmB4MobsterNeedsNoGraveyardCardAndCanConniveItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Draw Me", "Two", "Three")
	id, err := spmB4Cast(t, g, "Mechanical Mobster", "Artifact Creature — Human Robot Villain", spmB4MobsterOracle, "{3}", 2, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, []game.TargetRef{{Kind: game.TargetCard, ID: id, Slot: 1}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("connive raised no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pick.ChooseCards[0]}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, id) != 1 {
		t.Errorf("Mobster has %d counters, want 1", plusOnes(g, id))
	}
}

func TestSpmB4MobsterCannotConniveAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spmB4Cast(t, g, "Mechanical Mobster", "Artifact Creature — Human Robot Villain", spmB4MobsterOracle, "{3}", 2, 1, 0)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 1}}); err == nil {
		t.Fatal("an opponent's creature was a legal connive target")
	}
}

// --- Molten Man, Inferno Incarnate ---

func TestSpmB4MoltenManFetchesATappedMountainAndGrowsPerMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	fetched := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: fetched, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	id, err := spmB4Cast(t, g, "Molten Man, Inferno Incarnate", "Legendary Creature — Elemental Villain", spmB4MoltenManOracle, "{2}{R}", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, fetched)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, fetched)
	if !ok {
		t.Fatal("the Mountain was not put onto the battlefield")
	}
	if !card.Tapped {
		t.Error("the Mountain should enter tapped")
	}
	if p, tough := effectivePTOf(t, g, id); p != 2 || tough != 2 {
		t.Errorf("Molten Man is %d/%d, want 2/2 (two Mountains)", p, tough)
	}
	if !me.Library.Contains(forest) {
		t.Error("the Forest should still be in the library")
	}
}

func TestSpmB4MoltenManAsksForALandWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m1 := pushPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	m2 := pushPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	forest := pushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	mm := b12Push(g, me.ID, "Molten Man, Inferno Incarnate", "Legendary Creature — Elemental Villain", spmB4MoltenManOracle, 0, 0)
	if p, _ := effectivePTOf(t, g, mm); p != 2 {
		t.Fatalf("power %d, want 2", p)
	}
	b27Kill(g, mm)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	ch := chooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no sacrifice prompt")
	}
	if len(ch.ChooseCards) != 3 || !hasID(ch.ChooseCards, forest) {
		t.Errorf("sacrifice candidates %v, want all three lands", ch.ChooseCards)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !g.Battlefield.Contains(m1) || !g.Battlefield.Contains(m2) {
		t.Error("exactly the chosen land should have been sacrificed")
	}
}

// --- Lizard, Connors's Curse ---

func TestSpmB4LizardTurnsAnotherCreatureIntoAVanilla4x4GreenLizard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	flier := b12Push(g, opp.ID, "Mysterio's Phantasm", spmB4PhantasmTypeLine, spmB4PhantasmOracle, 1, 3)
	assertKeywords(t, g, flier, "flying", "vigilance")
	id, err := spmB4Cast(t, g, "Lizard, Connors's Curse", "Legendary Creature — Lizard Villain", spmB4LizardOracle, "{2}{G}{G}", 5, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
		t.Fatal("Lizard targeted itself")
	}
	answerPickTarget(t, g, flier)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, flier); p != 4 || tough != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tough)
	}
	c := spmB4Live(t, g, flier).Effective()
	if len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Errorf("colors %v, want green only", c.Colors)
	}
	if !spmB4Live(t, g, flier).HasSubtype("Lizard") || spmB4Live(t, g, flier).HasSubtype("Illusion") {
		t.Errorf("subtypes %v, want only Lizard", c.Subtypes)
	}
	if effectiveAbilitiesContain(t, g, flier, "flying") || effectiveAbilitiesContain(t, g, flier, "vigilance") {
		t.Error("the target kept an ability")
	}
	// Lizard itself is untouched.
	if p, tough := effectivePTOf(t, g, id); p != 5 || tough != 5 {
		t.Errorf("Lizard is %d/%d, want 5/5", p, tough)
	}
	assertKeywords(t, g, id, "trample")
}

func TestSpmB4LizardAllowsNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id, err := spmB4Cast(t, g, "Lizard, Connors's Curse", "Legendary Creature — Lizard Villain", spmB4LizardOracle, "{2}{G}{G}", 5, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if ch := latestPickTarget(g, me.ID); ch != nil {
		if err := g.ResolvePickTargets(ch.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("Bear is %d/%d, want unchanged 2/2", p, tough)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("Lizard is not on the battlefield")
	}
}

// --- Maximum Carnage ---

func TestSpmB4MaximumCarnageForcesEveryCreatureToAttackUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	third := g.Seats[2]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})
	castCatalogSpell(t, g, "Maximum Carnage", "Enchantment — Saga", spmB4CarnageOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasRequirementRecord(g) {
		t.Fatal("chapter I registered no requirement")
	}
	// My creature must attack this turn.
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with my Bear home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with my Bear attacking: %v", err)
	}
	// An opponent's creature must attack on their turn, and not the Saga's controller.
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with their Bear home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Fatal("their Bear attacked the Saga's controller while it could attack another player")
	}
	if err := g.DeclareAttacker(theirs, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with their Bear attacking another player: %v", err)
	}
	// On my next turn the effect is over.
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("on my next turn a creature is still forced to attack: %v", err)
	}
}

func TestSpmB4MaximumCarnageAddsRedRedRedThenBurnsEachOpponentAndLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	saga := castCatalogSpell(t, g, "Maximum Carnage", "Enchantment — Saga", spmB4CarnageOracle, nil)
	passPriorityAroundTable(t, g)
	// II
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Fatalf("chapter II added %d mana, want 3", got)
	}
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		lives[s.ID] = s.Life
	}
	// III
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		if s.Life != lives[s.ID]-5 {
			t.Errorf("seat %d life %d, want %d", s.ID, s.Life, lives[s.ID]-5)
		}
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga should be sacrificed after chapter III")
	}
}

// --- Madame Web, Clairvoyant ---

func TestSpmB4MadameWebCastsSpidersAndNoncreatureSpellsFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	web := pushTopSource(t, g, me.ID, "Madame Web, Clairvoyant", "Legendary Creature — Mutant Advisor", spmB4MadameWebOracle)
	if !topVisibleTo(g, me.ID, me.ID) || topVisibleTo(g, me.ID, opp.ID) {
		t.Error("the look should be private to its controller")
	}
	for _, tc := range []struct {
		name string
		card game.Card
		want bool
	}{
		{"spider creature", game.Card{Name: "Top Spider", TypeLine: "Creature — Spider Hero", ManaCost: "{0}"}, true},
		{"sorcery", game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"}, true},
		{"instant", game.Card{Name: "Top Instant", TypeLine: "Instant", ManaCost: "{0}"}, true},
		{"artifact", game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{0}"}, true},
		{"plain creature", game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"}, false},
		{"basic land", game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			top := seedTopOfLibrary(me, tc.card)
			if got := grantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted(); got != tc.want {
				t.Fatalf("permission granted = %v, want %v", got, tc.want)
			}
			err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true})
			if tc.want && tc.card.ManaCost == "{0}" {
				if err != nil {
					t.Errorf("cast from the top: %v", err)
				}
				passPriorityAroundTable(t, g)
			}
			if !tc.want {
				if err == nil {
					t.Errorf("%s was cast off the top", tc.card.Name)
				}
				if !me.Library.Contains(top) {
					t.Errorf("a refused play moved the card")
				}
			}
		})
	}
	removeFromBattlefield(g, web)
	top := seedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	if grantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
		t.Error("the permission outlived Madame Web")
	}
}

func TestSpmB4MadameWebMayMillOnceWhenYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushTopSource(t, g, me.ID, "Madame Web, Clairvoyant", "Legendary Creature — Mutant Advisor", spmB4MadameWebOracle)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	seedLibrary(me, "One", "Two", "Three", "Four")
	lib := me.Library.Size()
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("two attackers milled %d cards, want exactly 1 (one trigger per attack)", got)
	}
}
