package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b08_cards_test.go — Commander deck cards wave 2, batch 8 (enrage,
// damage triggers and damage modifiers; #5368). Each test plays the real
// catalog entry through the engine.

const (
	cmdr2B08ApexOID     = "fc467fe5-85ab-4c95-8d21-49096778b107"
	cmdr2B08RangingOID  = "2a86b70b-34ca-45df-8aa1-c3fb2da72dc3"
	cmdr2B08AegisaurOID = "7063abe1-87dd-4526-9867-f7ce255b2cf0"
	cmdr2B08TempleOID   = "1f499d39-b963-45c0-9604-d3cab98c9e05"
	cmdr2B08CrasherOID  = "178a79e2-64fd-4777-95fe-98b7834a642a"
	cmdr2B08ReckonerOID = "5f079a75-899e-467d-b443-3b87ce2fb548"
	cmdr2B08AmbusherOID = "ad3add32-cb87-4cd4-8c36-a737a897a8c8"
	cmdr2B08BodyOID     = "7fc6a5c1-f0da-42f6-93d0-6d94467e2ac0"
	cmdr2B08WaytaOID    = "e507df33-d150-4b21-8ddb-90799f412dd8"
	cmdr2B08GiselaOID   = "66f9f325-5e8e-4ebf-b5b3-c6410d80f2c5"
)

func TestCmdr2B08EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2B08ApexOID:     "Apex Altisaur",
		cmdr2B08RangingOID:  "Ranging Raptors",
		cmdr2B08AegisaurOID: "Bellowing Aegisaur",
		cmdr2B08TempleOID:   "Temple Altisaur",
		cmdr2B08CrasherOID:  "Quartzwood Crasher",
		cmdr2B08ReckonerOID: "Boros Reckoner",
		cmdr2B08AmbusherOID: "Illusory Ambusher",
		cmdr2B08BodyOID:     "Body of Knowledge",
		cmdr2B08WaytaOID:    "Wayta, Trainer Prodigy",
		cmdr2B08GiselaOID:   "Gisela, Blade of Goldnight",
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

// --- Apex Altisaur -------------------------------------------------------

func TestCmdr2B08ApexAltisaurFightsOnEntryAndWhenDamaged(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	small := b16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	big := b16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3, "R")
	apex, err := trekB37CastFromHand(t, g, game.Card{Name: "Apex Altisaur", TypeLine: "Creature — Dinosaur", OracleID: cmdr2B08ApexOID, Power: 10, Toughness: 10}, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, mine) {
			t.Error("my own creature was offered")
		}
		pickCard(t, g, me.ID, small)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, small); ok {
		t.Error("the fought 1/1 survived")
	}
	if got := cardByID(g, apex).DamageMarked; got != 1 {
		t.Errorf("the Altisaur took %d from the fight, want 1", got)
	}
	// Enrage: it was dealt damage by the fight, which fires its second trigger.
	g.RunStateChecksForTest()
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, big)
	}
	passPriorityAroundTable(t, g)
	if got := cardByID(g, big).DamageMarked; got != 10 {
		t.Errorf("the Ogre took %d from the enrage fight, want 10", got)
	}
}

// --- Ranging Raptors -----------------------------------------------------

func TestCmdr2B08RangingRaptorsFetchesATappedBasicWhenDamaged(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raptors := b27Push(g, me.ID, "Ranging Raptors", "Creature — Dinosaur", cmdr2B08RangingOID, "{2}{G}", 2, 3, "G")
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Wastes Nonbasic", "Land"),
	)
	cmdr2B08Damage(g, raptors, raptors, 1)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search was offered")
	}
	if searchOptionNamed(g, c, "Wastes Nonbasic") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	ids := battlefieldIDsNamed(g, "Forest")
	if len(ids) != 1 {
		t.Fatalf("%d Forests on the battlefield, want 1", len(ids))
	}
	if c, _ := battlefieldCard(g, ids[0]); !c.Tapped {
		t.Error("the land did not enter tapped")
	}
}

// --- Bellowing Aegisaur --------------------------------------------------

func TestCmdr2B08AegisaurPutsACounterOnEachOtherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aegis := b27Push(g, me.ID, "Bellowing Aegisaur", "Creature — Dinosaur", cmdr2B08AegisaurOID, "{5}{W}", 3, 5, "W")
	a := b16Creature(g, me.ID, "A", "Creature — Bear", 2, 2, "G")
	b := b16Creature(g, me.ID, "B", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2, "G")
	cmdr2B08Damage(g, theirs, aegis, 2)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := counterCount(g, id, "+1/+1"); n != 1 {
			t.Errorf("%v has %d counters, want 1", id, n)
		}
	}
	if counterCount(g, aegis, "+1/+1") != 0 || counterCount(g, theirs, "+1/+1") != 0 {
		t.Error("the Aegisaur or an opponent's creature got a counter")
	}
}

// --- Temple Altisaur -----------------------------------------------------

func TestCmdr2B08TempleAltisaurShieldsOtherDinosaursOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := b27Push(g, me.ID, "Temple Altisaur", "Creature — Dinosaur", cmdr2B08TempleOID, "{4}{W}", 3, 9, "W")
	dino := b16Creature(g, me.ID, "My Raptor", "Creature — Dinosaur", 2, 9, "R")
	bear := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 9, "G")
	theirs := b16Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 9, "R")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{dino, 1}, {temple, 5}, {bear, 5}, {theirs, 5}} {
		cmdr2B08Damage(g, src, c.id, 5)
		if got := cardByID(g, c.id).DamageMarked; got != c.want {
			t.Errorf("%v took %d, want %d", c.id, got, c.want)
		}
	}
	// Damage of 1 is left alone.
	cmdr2B08Damage(g, src, dino, 1)
	if got := cardByID(g, dino).DamageMarked; got != 2 {
		t.Errorf("after a 1-damage hit the Dinosaur has %d marked, want 2", got)
	}
}

// --- Quartzwood Crasher --------------------------------------------------

func TestCmdr2B08CrasherMakesATokenAsBigAsTheTrampleDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	crasher := b27Push(g, me.ID, "Quartzwood Crasher", "Creature — Dinosaur Beast", cmdr2B08CrasherOID, "{2}{R}{R}{G}", 6, 6, "R", "G")
	plain := b16Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2, "G")
	before := opp.Life
	attackWith(t, g, opp.ID, crasher, plain)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 8 {
		t.Fatalf("the opponent lost %d, want 8", got)
	}
	toks := trekB37Tokens(g, me.ID, "Dinosaur Beast")
	if len(toks) != 1 {
		t.Fatalf("%d Dinosaur Beast tokens, want 1", len(toks))
	}
	id := toks[0].InstanceID
	if effectivePower(t, g, id) != 6 || effectiveToughness(t, g, id) != 6 {
		t.Errorf("the token is %d/%d, want 6/6 (only the trampler's damage counts)", effectivePower(t, g, id), effectiveToughness(t, g, id))
	}
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("the token lacks trample")
	}
}

// --- Boros Reckoner ------------------------------------------------------

func TestCmdr2B08ReckonerReflectsTheDamageToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reck := b27Push(g, me.ID, "Boros Reckoner", "Creature — Minotaur Wizard", cmdr2B08ReckonerOID, "{R/W}{R/W}{R/W}", 3, 3, "R", "W")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	before := opp.Life
	cmdr2B08Damage(g, src, reck, 2)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2 (that much)", got)
	}
}

func TestCmdr2B08ReckonerGainsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reck := b27Push(g, me.ID, "Boros Reckoner", "Creature — Minotaur Wizard", cmdr2B08ReckonerOID, "{R/W}{R/W}{R/W}", 3, 3, "R", "W")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, reck, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	cmdr2B07Fund(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, reck, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with {W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, reck, "first strike") {
		t.Error("no first strike")
	}
}

// --- Illusory Ambusher ---------------------------------------------------

func TestCmdr2B08AmbusherDrawsThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amb := b27Push(g, me.ID, "Illusory Ambusher", "Creature — Cat Illusion", cmdr2B08AmbusherOID, "{4}{U}", 4, 9, "U")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	before := me.Hand.Size()
	cmdr2B08Damage(g, src, amb, 3)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

// --- Body of Knowledge ---------------------------------------------------

func TestCmdr2B08BodyOfKnowledgeIsAsBigAsYourHandAndDrawsThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	body := b27Push(g, me.ID, "Body of Knowledge", "Creature — Avatar", cmdr2B08BodyOID, "{3}{U}{U}", 0, 0, "U")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	n := me.Hand.Size()
	if p, tough := effectivePower(t, g, body), effectiveToughness(t, g, body); p != n || tough != n {
		t.Fatalf("%d/%d with %d cards in hand", p, tough, n)
	}
	cmdr2B08Damage(g, src, body, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != n+1 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), n+1)
	}
	if p, tough := effectivePower(t, g, body), effectiveToughness(t, g, body); p != n+1 || tough != n+1 {
		t.Errorf("%d/%d after drawing, want %d/%d", p, tough, n+1, n+1)
	}
	spec, _ := Lookup(cmdr2B08BodyOID)
	if !spec.NoMaxHandSize {
		t.Error("no maximum hand size is not declared")
	}
}

// --- Wayta, Trainer Prodigy ----------------------------------------------

func TestCmdr2B08WaytaFightsAndDiscountsTwoOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	wayta := b27Push(g, me.ID, "Wayta, Trainer Prodigy", "Legendary Creature — Human Warrior", cmdr2B08WaytaOID, "{R}{G}{W}", 1, 5, "R", "G", "W")
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3, "G")
	mine2 := b16Creature(g, me.ID, "My Other Bear", "Creature — Bear", 1, 3, "G")
	theirs := b16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	advanceToMain(t, g)
	two := []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: mine2}}
	vs := []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}
	// Against an opponent's creature the full {2}{G} is owed.
	cmdr2B07Fund(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, wayta, 0, game.ActivateAbilityParams{Strict: true, Targets: vs}); err == nil {
		t.Fatal("paid {G} for a fight with an opponent's creature")
	}
	cmdr2B07Fund(me, "C", 2)
	if err := g.ActivateCatalogAbility(me.ID, wayta, 0, game.ActivateAbilityParams{Strict: true, Targets: vs}); err != nil {
		t.Fatalf("full-price fight: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the opponent's 1/1 survived the fight")
	}
	// Two of your creatures: {G} is enough. Untap Wayta for the second go.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(wayta) })
	cmdr2B07Fund(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, wayta, 0, game.ActivateAbilityParams{Strict: true, Targets: two}); err != nil {
		t.Fatalf("discounted fight: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := cardByID(g, mine).DamageMarked; got != 2 {
		t.Errorf("my Bear has %d damage, want 1 from the Elf fight and 1 from the other Bear", got)
	}
}

func TestCmdr2B08WaytaDoublesTheEnrageTriggersOfYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Wayta, Trainer Prodigy", "Legendary Creature — Human Warrior", cmdr2B08WaytaOID, "{R}{G}{W}", 1, 5, "R", "G", "W")
	aegis := b27Push(g, me.ID, "Bellowing Aegisaur", "Creature — Dinosaur", cmdr2B08AegisaurOID, "{5}{W}", 3, 9, "W")
	friend := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	cmdr2B08Damage(g, src, aegis, 1)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, friend, "+1/+1"); n != 2 {
		t.Errorf("the Bear has %d counters, want 2 (the trigger happens twice)", n)
	}
}

func TestCmdr2B08WaytaDoesNotDoubleDamageToAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Wayta, Trainer Prodigy", "Legendary Creature — Human Warrior", cmdr2B08WaytaOID, "{R}{G}{W}", 1, 5, "R", "G", "W")
	theirAegis := b27Push(g, opp.ID, "Bellowing Aegisaur", "Creature — Dinosaur", cmdr2B08AegisaurOID, "{5}{W}", 3, 9, "W")
	theirFriend := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	src := b16Creature(g, me.ID, "Source", "Creature — Ogre", 5, 5, "R")
	cmdr2B08Damage(g, src, theirAegis, 1)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, theirFriend, "+1/+1"); n != 1 {
		t.Errorf("their Bear has %d counters, want 1 (not a permanent or creature of Wayta's controller)", n)
	}
}

// --- Gisela, Blade of Goldnight ------------------------------------------

func TestCmdr2B08GiselaDoublesDamageToOpponentsAndHalvesDamageToYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Gisela, Blade of Goldnight", "Legendary Creature — Angel", cmdr2B08GiselaOID, "{4}{R}{W}{W}", 5, 5, "R", "W")
	mine := b16Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 20, "W")
	theirs := b16Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 20, "W")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")

	cmdr2B08Damage(g, src, theirs, 3)
	if got := cardByID(g, theirs).DamageMarked; got != 6 {
		t.Errorf("an opponent's permanent took %d from 3, want 6", got)
	}
	cmdr2B08Damage(g, src, mine, 5)
	if got := cardByID(g, mine).DamageMarked; got != 2 {
		t.Errorf("my permanent took %d from 5, want 2 (half rounded up is prevented)", got)
	}
	cmdr2B08Damage(g, src, mine, 1)
	if got := cardByID(g, mine).DamageMarked; got != 2 {
		t.Errorf("a 1-damage hit left %d marked, want 2 (all of it was prevented)", got)
	}

	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 4) })
	if got := before - opp.Life; got != 8 {
		t.Errorf("an opponent lost %d from 4, want 8", got)
	}
	mine0 := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 7) })
	if got := mine0 - me.Life; got != 3 {
		t.Errorf("I lost %d from 7, want 3", got)
	}
}

// cmdr2B08Damage deals noncombat damage and lets the batch settle, so the
// "whenever this is dealt damage" triggers go on the stack.
func cmdr2B08Damage(g *game.Game, source, target uuid.UUID, n int) {
	b27Damage(g, source, target, n)
	g.RunStateChecksForTest()
}
