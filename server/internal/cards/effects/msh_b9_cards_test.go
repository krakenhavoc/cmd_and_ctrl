package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b9_cards_test.go — Marvel Super Heroes, half B, ninth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB9TakeUpTheShieldOracle = "e8d0eafe-540a-4b2e-a989-e98ff3c31105"
	mshB9ThanosOracle          = "090d815b-b2c8-48cd-a0b3-dc662d5352ac"
	mshB9GalactusSagaOracle    = "e05f23ba-f2b2-4c40-83fa-8defce3cb978"
	mshB9MastersOfEvilOracle   = "fcb1caea-e1ac-48f0-a67a-d4aadad490d9"
	mshB9ScarletWitchOracle    = "d8ae8c38-501e-44ea-9faf-cf9d11ba6616"
	mshB9ThingOracle           = "d566d5f1-5abd-405b-8553-d267ecbe72ab"
	mshB9WaspOracle            = "5ad8963b-8091-4787-9147-20a1fe7bf0fb"
	mshB9ThorOracle            = "2842217a-173e-4879-8b25-fefe2fc1123f"
	mshB9TigraOracle           = "39d993be-30ad-4756-a494-fd8808d474c9"
)

func TestMshB9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB9TakeUpTheShieldOracle: "Take Up the Shield",
		mshB9ThanosOracle:          "Thanos, the Mad Titan",
		mshB9GalactusSagaOracle:    "The Coming of Galactus",
		mshB9MastersOfEvilOracle:   "The Masters of Evil",
		mshB9ScarletWitchOracle:    "The Scarlet Witch",
		mshB9ThingOracle:           "The Thing, Ben Grimm",
		mshB9WaspOracle:            "The Wondrous Wasp",
		mshB9ThorOracle:            "Thor, God of Thunder",
		mshB9TigraOracle:           "Tigra, Feline Fury",
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

// mshB9Damage fakes one creature's damage event, which is all the Thing's
// trigger watches.
func mshB9Damage(g *game.Game, source, victim uuid.UUID, amount int, combat bool) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:   game.EventDealDamage,
			Source: source,
			Target: victim,
			Amount: amount,
			Combat: combat,
		})
	})
}

// --- Take Up the Shield ------------------------------------------------------

func TestMshB9TakeUpTheShieldCountersAndGrantsUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Take Up the Shield", "Instant", mshB9TakeUpTheShieldOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)

	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the target lacks %s", kw)
		}
	}

	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	for _, kw := range []string{"lifelink", "indestructible"} {
		if hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("%s did not end with the turn", kw)
		}
	}
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the counter did not stay: %d", got)
	}
}

func TestMshB9TakeUpTheShieldCanTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Take Up the Shield", "Instant", mshB9TakeUpTheShieldOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)

	if got := plusOnes(g, theirs); got != 1 || !hasKeywordOnBattlefield(t, g, theirs, "indestructible") {
		t.Errorf("an opponent's creature: counters %d, want 1 and indestructible", got)
	}
}

// --- Tigra, Feline Fury ------------------------------------------------------

func TestMshB9TigraGrowsOncePerLifeGainEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tigra := b12Push(g, me.ID, "Tigra, Feline Fury", "Legendary Creature — Cat Human Hero", mshB9TigraOracle, 2, 1)
	for _, kw := range []string{"flash", "trample"} {
		if !hasKeywordOnBattlefield(t, g, tigra, kw) {
			t.Errorf("Tigra lacks %s", kw)
		}
	}

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, tigra); got != 1 {
		t.Errorf("counters after gaining 3 = %d, want 1 (one event, not one per point)", got)
	}

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 2) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, tigra); got != 1 {
		t.Errorf("counters = %d, want 1: an opponent's gain and my loss are not mine to count", got)
	}

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, tigra); got != 2 {
		t.Errorf("counters after a second gain = %d, want 2", got)
	}
}

// --- The Wondrous Wasp -------------------------------------------------------

func TestMshB9WaspTapsTheTargetAndStripsItsAbilitiesWhileSheRemains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	tigra := b12Push(g, me.ID, "Tigra, Feline Fury", "Legendary Creature — Cat Human Hero", mshB9TigraOracle, 2, 1)

	wasp := castPowerUpCreature(t, g, me, "The Wondrous Wasp", "Legendary Creature — Human Hero", "{1}{U}", "{C}{U}", mshB9WaspOracle, 2)
	for _, kw := range []string{"flash", "flying"} {
		if !hasKeywordOnBattlefield(t, g, wasp, kw) {
			t.Errorf("the Wasp lacks %s", kw)
		}
	}
	mshB2PickIfAsked(t, g, me.ID, tigra)
	passPriorityAroundTable(t, g)

	if !ringBFCard(t, g, tigra).Tapped {
		t.Error("the target was not tapped")
	}
	if hasKeywordOnBattlefield(t, g, tigra, "trample") {
		t.Error("the target kept trample while the Wasp remains")
	}

	killCreature(t, g, me.ID, wasp)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !hasKeywordOnBattlefield(t, g, tigra, "trample") {
		t.Error("the target did not get its abilities back when the Wasp left")
	}
	if !ringBFCard(t, g, tigra).Tapped {
		t.Error("the tap should not be undone when the Wasp leaves")
	}
}

// --- The Masters of Evil -----------------------------------------------------

func TestMshB9MastersOfEvilGivesOtherVillainsYouControlPlusTwoPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	masters := b12Push(g, me.ID, "The Masters of Evil", "Legendary Creature — Human Villain", mshB9MastersOfEvilOracle, 5, 6)
	villain := b12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)

	if p, tough := effectivePTOf(t, g, villain); p != 4 || tough != 3 {
		t.Errorf("my Villain = %d/%d, want 4/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a non-Villain = %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Villain = %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, masters); p != 5 || tough != 6 {
		t.Errorf("the Masters themselves = %d/%d, want 5/6 (other Villains only)", p, tough)
	}
}

func TestMshB9MastersOfEvilDiscardsToFetchAPlan(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	masters := pushCatalogHandCard(me, "The Masters of Evil", "Legendary Creature — Human Villain", mshB9MastersOfEvilOracle)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Test Plan", TypeLine: "Enchantment — Plan"},
		game.Card{Name: "Test Bear", TypeLine: "Creature — Bear"})

	if err := hobB4Activate(t, g, me.ID, masters, 0, "{C}{B}", nil); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(masters) {
		t.Error("the discard was not a cost paid at announce")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(ids[0]) {
		t.Error("the Plan card did not reach the hand")
	}
	if me.Hand.Contains(ids[1]) {
		t.Error("the non-Plan card reached the hand")
	}
}

// --- The Scarlet Witch -------------------------------------------------------

func TestMshB9ScarletWitchDiscountsBigInstantsAndSorceriesByHerPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "The Scarlet Witch", "Legendary Creature — Mutant Warlock Hero", mshB9ScarletWitchOracle, 2, 3)

	for _, tc := range []struct {
		name, typeLine, cost string
		who                  *game.Player
		want                 int
	}{
		{"Three Instant", "Instant", "{2}{R}", me, 3},
		{"Four Instant", "Instant", "{3}{R}", me, 2},
		{"Five Sorcery", "Sorcery", "{4}{U}", me, 3},
		{"Four Creature", "Creature — Bear", "{3}{G}", me, 4},
		{"Coloured Four", "Instant", "{R}{R}{R}{R}", me, 4},
		{"Their Four Instant", "Instant", "{3}{R}", opp, 4},
	} {
		if got := priceInHand(t, g, tc.who, tc.name, tc.typeLine, tc.cost); got != tc.want {
			t.Errorf("%s %s costs %d, want %d", tc.name, tc.cost, got, tc.want)
		}
	}
}

func TestMshB9ScarletWitchDiscountFollowsHerCurrentPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "The Scarlet Witch", "Legendary Creature — Mutant Warlock Hero", mshB9ScarletWitchOracle, 4, 3)

	if got := priceInHand(t, g, me, "Big Instant", "Instant", "{5}{R}"); got != 2 {
		t.Errorf("a six-mana instant costs %d, want 2 with power 4", got)
	}
}

// --- The Thing, Ben Grimm ----------------------------------------------------

func TestMshB9ThingGrowsWhenAHeroYouControlDamagesAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	thing := b12Push(g, me.ID, "The Thing, Ben Grimm", "Legendary Creature — Human Hero", mshB9ThingOracle, 7, 7)
	if !hasKeywordOnBattlefield(t, g, thing, "trample") {
		t.Error("the Thing lacks trample")
	}
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human Hero", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirHero := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)

	mshB9Damage(g, bear, opp.ID, 2, true)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 0 {
		t.Fatalf("a Bear's damage grew the Thing: %d counters", got)
	}
	mshB9Damage(g, theirHero, me.ID, 2, true)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 0 {
		t.Fatalf("an opponent's Hero grew the Thing: %d counters", got)
	}
	mshB9Damage(g, hero, opp.ID, 0, true)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 0 {
		t.Fatalf("zero damage grew the Thing: %d counters", got)
	}

	mshB9Damage(g, hero, opp.ID, 2, true)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 2 {
		t.Errorf("counters after a Hero connected = %d, want 2", got)
	}
}

func TestMshB9ThingCountsNoncombatDamageAndOwnDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	thing := b12Push(g, me.ID, "The Thing, Ben Grimm", "Legendary Creature — Human Hero", mshB9ThingOracle, 7, 7)
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human Hero", 2, 2)

	mshB9Damage(g, hero, opp.ID, 3, false)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 2 {
		t.Errorf("counters after noncombat Hero damage = %d, want 2", got)
	}
	mshB9Damage(g, thing, opp.ID, 7, true)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 4 {
		t.Errorf("counters after the Thing himself connected = %d, want 4", got)
	}
}

func TestMshB9ThingTriggersOncePerPlayerPerBatch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	advanceToMain(t, g)
	thing := b12Push(g, me.ID, "The Thing, Ben Grimm", "Legendary Creature — Human Hero", mshB9ThingOracle, 7, 7)
	h1 := b12Creature(g, me.ID, "Hero One", "Creature — Human Hero", 2, 2)
	h2 := b12Creature(g, me.ID, "Hero Two", "Creature — Human Hero", 2, 2)

	// Two Heroes hit one player in one damage step: one trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: h1, Target: opp.ID, Amount: 2, Combat: true})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: h2, Target: opp.ID, Amount: 2, Combat: true})
	})
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 2 {
		t.Fatalf("two Heroes on one player: %d counters, want 2", got)
	}

	// Two Heroes hit two players in one step: two triggers.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: h1, Target: opp.ID, Amount: 2, Combat: true})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: h2, Target: third.ID, Amount: 2, Combat: true})
	})
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, thing); got != 6 {
		t.Errorf("two Heroes on two players: %d counters, want 6", got)
	}
}

// --- The Coming of Galactus --------------------------------------------------

func TestMshB9GalactusSagaPlaysThroughAllFourChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)

	saga := castCatalogSpell(t, g, "The Coming of Galactus", "Enchantment — Saga", mshB9GalactusSagaOracle, nil)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, theirBear)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirBear) {
		t.Error("chapter I did not destroy the target nonland permanent")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("chapter I destroyed a land")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	for chapter := 2; chapter <= 3; chapter++ {
		advanceToPrecombatMainOf(t, g, 0)
		passPriorityAroundTable(t, g)
		for i, p := range g.Seats {
			want := lives[p.ID]
			if i != 0 {
				want -= 2 * (chapter - 1)
			}
			if p.Life != want {
				t.Errorf("chapter %d: seat %d life = %d, want %d", chapter, i, p.Life, want)
			}
		}
	}
	if countBattlefieldByName(g, "Galactus") != 0 {
		t.Fatal("Galactus arrived before chapter IV")
	}

	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga was not sacrificed after chapter IV")
	}
	if got := countBattlefieldByName(g, "Galactus"); got != 1 {
		t.Fatalf("%d Galactus tokens, want 1", got)
	}
	galactus := findBattlefieldByName(g, "Galactus")
	if p, tough := effectivePTOf(t, g, galactus); p != 16 || tough != 16 {
		t.Errorf("Galactus = %d/%d, want 16/16", p, tough)
	}
	for _, kw := range []string{"flying", "trample"} {
		if !hasKeywordOnBattlefield(t, g, galactus, kw) {
			t.Errorf("Galactus lacks %s", kw)
		}
	}
	card := ringBFCard(t, g, galactus)
	if !card.IsToken() || card.Controller != me.ID || !strings.Contains(card.TypeLine, "Legendary") ||
		!card.HasSubtype("Elder") || !card.HasSubtype("Alien") || len(card.Colors) != 1 || card.Colors[0] != "B" {
		t.Errorf("Galactus = %+v, want my legendary black Elder Alien token", card)
	}

	// Next turn he attacks and destroys a land. The Saga is gone, so the
	// ability is the token's own.
	advanceToPrecombatMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, galactus)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("Galactus's attack did not destroy the land")
	}
}

// --- Thanos, the Mad Titan ---------------------------------------------------

// mshB9ThanosTable sets up Thanos beside creatures of mana values 0 to 5 and
// a noncreature artifact, and returns the ids by name.
func mshB9ThanosTable(t *testing.T) (g *game.Game, thanos uuid.UUID, ids map[string]uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	ids = map[string]uuid.UUID{}
	thanos = pushPowerUpPermanent(g, me.ID, "Thanos, the Mad Titan", "Legendary Creature — Eternal Villain", "{R}{W}{B}", mshB9ThanosOracle)
	ids["one"] = pushPowerUpPermanent(g, me.ID, "One", "Creature — Bear", "{1}", "")
	ids["two"] = pushPowerUpPermanent(g, me.ID, "Two", "Creature — Bear", "{1}{G}", "")
	ids["three"] = pushPowerUpPermanent(g, me.ID, "Three", "Creature — Bear", "{2}{G}", "")
	ids["four"] = pushPowerUpPermanent(g, opp.ID, "Four", "Creature — Bear", "{3}{G}", "")
	ids["five"] = pushPowerUpPermanent(g, opp.ID, "Five", "Creature — Bear", "{4}{G}", "")
	ids["zero"] = b12Creature(g, opp.ID, "Zero", "Creature — Bear", 2, 2)
	ids["rock"] = pushPowerUpPermanent(g, me.ID, "Odd Rock", "Artifact", "{1}", "")
	return g, thanos, ids
}

func mshB9ThanosPowerUp(t *testing.T, g *game.Game, thanos uuid.UUID, optionIndex int) {
	t.Helper()
	me := g.Seats[0]
	floatMana(t, g, me, "{C}{W}{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, thanos, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if c == nil {
		t.Fatal("Thanos did not ask odd or even")
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, optionIndex); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}

func TestMshB9ThanosOddDestroysOddCreaturesAndSparesTheRest(t *testing.T) {
	g, thanos, ids := mshB9ThanosTable(t)
	if !hasKeywordOnBattlefield(t, g, thanos, "deathtouch") || !hasKeywordOnBattlefield(t, g, thanos, "lifelink") {
		t.Error("Thanos lacks deathtouch or lifelink")
	}
	mshB9ThanosPowerUp(t, g, thanos, 0)

	if got := plusOnes(g, thanos); got != 2 {
		t.Errorf("Thanos has %d counters, want 2", got)
	}
	for name, wantAlive := range map[string]bool{
		"one": false, "two": true, "three": false, "four": true, "five": false, "zero": true, "rock": true,
	} {
		if g.Battlefield.Contains(ids[name]) != wantAlive {
			t.Errorf("odd: %s alive = %v, want %v", name, !wantAlive, wantAlive)
		}
	}
	if !g.Battlefield.Contains(thanos) {
		t.Error("Thanos destroyed himself (he is a 3-mana creature)")
	}
}

func TestMshB9ThanosEvenDestroysEvenCreaturesIncludingZero(t *testing.T) {
	g, thanos, ids := mshB9ThanosTable(t)
	mshB9ThanosPowerUp(t, g, thanos, 1)

	for name, wantAlive := range map[string]bool{
		"one": true, "two": false, "three": true, "four": false, "five": true, "zero": false, "rock": true,
	} {
		if g.Battlefield.Contains(ids[name]) != wantAlive {
			t.Errorf("even: %s alive = %v, want %v", name, !wantAlive, wantAlive)
		}
	}
	if !g.Battlefield.Contains(thanos) {
		t.Error("Thanos did not survive his own sweep")
	}
}

func TestMshB9ThanosPowerUpIsOnceOnly(t *testing.T) {
	g, thanos, _ := mshB9ThanosTable(t)
	mshB9ThanosPowerUp(t, g, thanos, 0)

	floatMana(t, g, g.Seats[0], "{C}{W}{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, thanos, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second power-up activation was allowed")
	}
}

// --- Thor, God of Thunder ----------------------------------------------------

func TestMshB9ThorExilesAGraveyardCardWithAPlayPermissionThroughYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	bolt := pushCatalogGraveyardCard(me, "Test Bolt", "Instant", "", 0, 0)
	sword := pushCatalogGraveyardCard(me, "Test Sword", "Artifact — Equipment", "", 0, 0)
	bear := pushCatalogGraveyardCard(me, "Test Bear", "Creature — Bear", "", 2, 2)

	thor := castPowerUpCreature(t, g, me, "Thor, God of Thunder", "Legendary Creature — God Warrior Hero", "{3}{R}{R}", "{C}{C}{C}{R}{R}", mshB9ThorOracle, 5)
	if !hasKeywordOnBattlefield(t, g, thor, "flying") {
		t.Error("Thor lacks flying")
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, id := range p.PickTargetCards {
			if id == bear {
				t.Error("a creature card was offered as Thor's target")
			}
		}
	}
	mshB2PickIfAsked(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Fatal("the chosen card was not exiled")
	}
	if grantedPermissionOn(g, me.ID, bolt, game.ZoneExile) == nil {
		t.Error("the exiled card carries no permission to play it")
	}
	if !me.Graveyard.Contains(sword) || !me.Graveyard.Contains(bear) {
		t.Error("an unchosen card left the graveyard")
	}

	// Still playable on my next turn, gone the turn after.
	advanceToPrecombatMainOf(t, g, 0)
	if grantedPermissionOn(g, me.ID, bolt, game.ZoneExile) == nil {
		t.Error("the permission ended before my next turn")
	}
	advanceToPrecombatMainOf(t, g, 0)
	if grantedPermissionOn(g, me.ID, bolt, game.ZoneExile) != nil {
		t.Error("the permission outlasted my next turn")
	}
}

func TestMshB9ThorEntersWithNoLegalGraveyardCardAndDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	bear := pushCatalogGraveyardCard(me, "Test Bear", "Creature — Bear", "", 2, 2)

	thor := castPowerUpCreature(t, g, me, "Thor, God of Thunder", "Legendary Creature — God Warrior Hero", "{3}{R}{R}", "{C}{C}{C}{R}{R}", mshB9ThorOracle, 5)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(thor) || !me.Graveyard.Contains(bear) || g.Exile.Contains(bear) {
		t.Error("Thor exiled a card that is not an Equipment, instant or sorcery")
	}
}

func TestMshB9ThorDealsDamageEqualToANoncreatureSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Thor, God of Thunder", "Legendary Creature — God Warrior Hero", mshB9ThorOracle, 5, 5)

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Test Bolt", TypeLine: "Instant", ManaCost: "{2}{R}",
		Owner: me.ID, Controller: me.ID,
	})
	floatMana(t, g, me, "{C}{C}{R}")
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != life-3 {
		t.Errorf("opponent life = %d, want %d (3 damage for a three-mana spell)", opp.Life, life-3)
	}
}

func TestMshB9ThorIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	b12Push(g, me.ID, "Thor, God of Thunder", "Legendary Creature — God Warrior Hero", mshB9ThorOracle, 5, 5)

	life := opp.Life
	castPowerUpCreature(t, g, me, "Test Bear", "Creature — Bear", "{1}{G}", "{C}{G}", "", 2)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("a creature spell triggered Thor")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life = %d, want %d", opp.Life, life)
	}
}
