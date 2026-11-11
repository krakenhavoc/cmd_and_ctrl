package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2B17ChatterstormOracle = "7e6c7b57-bd6a-4ac8-bcab-c3a879f479c2"
	cmdr2B17SquirrelMobOracle  = "451486df-6160-4ff7-b47c-dd3760417a31"
	cmdr2B17AntQueenOracle     = "50de88d4-0beb-45e5-9664-d4f0e280e652"
	cmdr2B17MarrowGnawerOracle = "2b934761-5720-477b-94b7-d1d91dc581a6"
)

func TestCmdr2B17ChatterstormStormCopiesEachMakeASquirrel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 0)
	stormFiller(t, g)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Chatterstorm", "Sorcery", cmdr2B17ChatterstormOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Squirrel", me.ID); got != 3 {
		t.Errorf("%d Squirrels, want 3 (the spell and two storm copies)", got)
	}
}

func TestCmdr2B17SquirrelMobCountsEveryOtherSquirrel(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mob := b41CatalogCreature(g, me.ID, "Squirrel Mob", "Creature — Squirrel", cmdr2B17SquirrelMobOracle, 2, 2)
	if got := effectivePower(t, g, mob); got != 2 {
		t.Fatalf("alone the Mob is %d power, want 2", got)
	}
	b12Creature(g, me.ID, "Mine", "Creature — Squirrel", 1, 1)
	b12Creature(g, opp.ID, "Theirs", "Creature — Squirrel", 1, 1)
	b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tg := effectivePower(t, g, mob), effectiveToughness(t, g, mob); p != 4 || tg != 4 {
		t.Errorf("Mob is %d/%d, want 4/4 (two other Squirrels, either side, a Bear doesn't count)", p, tg)
	}
	mob2 := b41CatalogCreature(g, me.ID, "Squirrel Mob", "Creature — Squirrel", cmdr2B17SquirrelMobOracle, 2, 2)
	if got := effectivePower(t, g, mob2); got != 5 {
		t.Errorf("second Mob power %d, want 5 (counts the first Mob)", got)
	}
}

func TestCmdr2B17AntQueenMakesInsects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	q := pushCatalogPermanent(g, me.ID, "Ant Queen", "Creature — Insect", cmdr2B17AntQueenOracle, false)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activating with only {G} in the pool should be refused")
	}
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Insect", me.ID); got != 1 {
		t.Errorf("%d Insects, want 1", got)
	}
}

func TestCmdr2B17MarrowGnawerGivesEveryRatFearAndCountsAfterTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gn := pushCatalogPermanent(g, me.ID, "Marrow-Gnawer", "Legendary Creature — Rat Rogue", cmdr2B17MarrowGnawerOracle, false)
	mine := b12Creature(g, me.ID, "Rat A", "Creature — Rat", 1, 1)
	theirs := b12Creature(g, opp.ID, "Rat B", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{gn, mine, theirs} {
		if !hasEffectiveKeyword(t, g, id, "fear") {
			t.Errorf("a Rat lacks fear")
		}
	}
	if hasEffectiveKeyword(t, g, bear, "fear") {
		t.Error("a Bear has fear")
	}
	// A non-Rat is not a legal sacrifice.
	if err := g.ActivateCatalogAbility(me.ID, gn, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("sacrificing a Bear was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, gn, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{mine}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Rats left under my control: Marrow-Gnawer only (the other Rat is
	// theirs), so X = 1.
	if got := countOnBattlefield(g, "Rat", me.ID); got != 1 {
		t.Errorf("%d Rat tokens, want 1", got)
	}
}

func TestCmdr2B17SwarmyardMassacreCountsAfterTheTokensAndSparesTheTribes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 5, 5)
	spider := b12Creature(g, opp.ID, "Spider", "Creature — Spider", 3, 3)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Swarmyard Massacre", "Sorcery", "50e7ae38-2a66-4e32-bfb3-fefbae0caa61", nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Squirrel", me.ID); got != 2 {
		t.Fatalf("%d Squirrels, want 2", got)
	}
	// Three of my creatures are Insect/Rat/Spider/Squirrel (Rat + 2 Squirrels).
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2 (5 - 3)", got)
	}
	if got := effectivePower(t, g, ogre); got != 1 {
		t.Errorf("Ogre power %d, want 1 (4 - 3)", got)
	}
	if got := effectivePower(t, g, spider); got != 3 {
		t.Errorf("an opponent's Spider is spared: power %d, want 3 (and doesn't count for me)", got)
	}
	if got := effectivePower(t, g, rat); got != 1 {
		t.Errorf("my Rat is spared: power %d", got)
	}
}

const cmdr2B17RootcastOracle = "ffe13100-99dd-41e3-82ba-eaead6fd0113"

func TestCmdr2B17RootcastApprenticeshipEveryModeAndTheSacrificeIsTheOpponentsPick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Squirrel", TypeLine: "Token Creature — Squirrel",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	keep := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	treasure := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: opp.ID, Controller: opp.ID,
	})
	castModal(t, g, "Rootcast Apprenticeship", "Sorcery", cmdr2B17RootcastOracle,
		[]int{0, 1, 3},
		[]game.TargetRef{
			modeRef(game.TargetCard, bear, 0, 0),
			modeRef(game.TargetCard, tok, 1, 0),
			modeRef(game.TargetPlayer, opp.ID, 2, 0),
		})
	passPriorityAroundTable(t, g)
	if p := latestChooseCardsFor(g, opp.ID); p != nil {
		if err := g.ResolveChooseCards(p.ID, opp.ID, []uuid.UUID{keep}); err != nil {
			t.Fatalf("sacrifice pick: %v", err)
		}
	} else {
		t.Fatal("the opponent was never asked to sacrifice")
	}
	passPriorityAroundTable(t, g)
	if got := effectivePowerWithCounters(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4 (two +1/+1 counters)", got)
	}
	if got := countOnBattlefield(g, "Squirrel", me.ID); got != 2 {
		t.Errorf("%d Squirrels, want 2 (the token and its copy)", got)
	}
	if g.Battlefield.Contains(keep) {
		t.Error("the nontoken artifact was not sacrificed")
	}
	if !g.Battlefield.Contains(treasure) {
		t.Error("a TOKEN artifact was sacrificed; only nontoken artifacts qualify")
	}
}

func TestCmdr2B17RootcastApprenticeshipSquirrelsForTheTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castModal(t, g, "Rootcast Apprenticeship", "Sorcery", cmdr2B17RootcastOracle,
		[]int{2, 2, 2},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, opp.ID, 0, 0),
			modeRef(game.TargetPlayer, opp.ID, 1, 0),
			modeRef(game.TargetPlayer, me.ID, 2, 0),
		})
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Squirrel", opp.ID); got != 2 {
		t.Errorf("opponent has %d Squirrels, want 2", got)
	}
	if got := countOnBattlefield(g, "Squirrel", me.ID); got != 1 {
		t.Errorf("I have %d Squirrels, want 1", got)
	}
}

const cmdr2B17HermitOracle = "80e9a42e-f4f4-4bfc-91a2-ad124737dca2"

// cmdr2B17NextUpkeepOfSeat0 walks to the next time seat 0's upkeep
// begins and settles what that put on the stack.
func cmdr2B17NextUpkeepOfSeat0(t *testing.T, g *game.Game) {
	t.Helper()
	// Leave the current turn first.
	for i := 0; i < 400 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
}

func TestCmdr2B17DeepForestHermitVanishesAfterThreeUpkeeps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hermit := castCatalogSpell(t, g, "Deep Forest Hermit", "Creature — Elf Druid", cmdr2B17HermitOracle, nil)
	passPriorityAroundTable(t, g)
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, bc := range g.Battlefield.Cards {
			if bc.InstanceID == hermit {
				c, ok = bc, true
			}
		}
	})
	if !ok {
		t.Fatal("the Hermit did not enter")
	}
	if c.Counters[game.CounterTime] != 3 {
		t.Fatalf("%d time counters, want 3", c.Counters[game.CounterTime])
	}
	if got := countOnBattlefield(g, "Squirrel", me.ID); got != 4 {
		t.Fatalf("%d Squirrels, want 4", got)
	}
	sq := firstNamedOnBattlefield(g, "Squirrel", me.ID)
	if got := effectivePower(t, g, sq); got != 2 {
		t.Errorf("Squirrel power %d, want 2 (Hermit's +1/+1)", got)
	}
	for want := 2; want >= 0; want-- {
		cmdr2B17NextUpkeepOfSeat0(t, g)
		if want > 0 {
			bc, here := battlefieldCard(g, hermit)
			if !here {
				t.Fatalf("the Hermit left with %d counters still to go", want)
			}
			if bc.Counters[game.CounterTime] != want {
				t.Fatalf("after an upkeep: %d time counters, want %d", bc.Counters[game.CounterTime], want)
			}
		}
	}
	if _, here := battlefieldCard(g, hermit); here {
		t.Error("the Hermit survived the removal of its last time counter")
	}
	if got := countOnBattlefield(g, "Squirrel", me.ID); got != 4 {
		t.Errorf("the Squirrels stay: %d, want 4", got)
	}
}

func firstNamedOnBattlefield(g *game.Game, name string, controller uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

const (
	cmdr2B17OddAcornOracle = "1d9d8028-d946-4245-ae0c-7508023f619a"
	cmdr2B17BadgerOracle   = "f61c0fc0-7241-455e-bbf5-f1f3862d46e3"
)

func TestCmdr2B17OddAcornGangGrantsTheSorceryPumpToEverySquirrelOfMine(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Odd Acorn Gang", "Legendary Creature — Squirrel Warrior", cmdr2B17OddAcornOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Squirrel", "Token Creature — Squirrel", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Squirrel", "Token Creature — Squirrel", "", false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	advanceToMainOf(t, g, 0)

	activate := func(source, target uuid.UUID) error {
		var idx []int
		var origins game.AbilityOrigins
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(source)
			shapes, o := game.ActivatedAbilitiesWithOrigins(c)
			origins = o
			for i := range shapes {
				idx = append(idx, i)
			}
		})
		if len(idx) == 0 {
			return errNoGrantedAbility
		}
		return g.ActivateCatalogAbility(me.ID, source, idx[0], game.ActivateAbilityParams{
			Ref: origins.Ref(idx[0]), Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	}
	if err := activate(bear, mine); err == nil {
		t.Error("a Bear was given the Squirrel pump")
	}
	if err := activate(mine, theirs); err != nil {
		t.Fatalf("my Squirrel could not pump the opponent's Squirrel: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 3 || tg != 3 {
		t.Errorf("target Squirrel is %d/%d, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("the target did not gain trample")
	}
	if err := activate(mine, bear); err == nil {
		t.Error("a non-Squirrel was a legal target")
	}
}

var errNoGrantedAbility = errors.New("no granted ability")

func TestCmdr2B17OddAcornGangDrawsOncePerPlayerConnectedWith(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Odd Acorn Gang", "Legendary Creature — Squirrel Warrior", cmdr2B17OddAcornOracle, false)
	a := pushCatalogPermanent(g, me.ID, "Squirrel", "Token Creature — Squirrel", "", false)
	b := pushCatalogPermanent(g, me.ID, "Squirrel", "Token Creature — Squirrel", "", false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Hand.Size()
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards for two Squirrels connecting, want 1 (one or more)", got)
	}
	_ = bear
}

func TestCmdr2B17OddAcornGangIgnoresNonSquirrels(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Odd Acorn Gang", "Legendary Creature — Squirrel Warrior", cmdr2B17OddAcornOracle, true)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Hand.Size()
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("a Bear connecting drew %d cards, want 0", got)
	}
}

func TestCmdr2B17BrightcapBadgerMakesFungusAndSaprolingsTapForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	badger := pushCatalogPermanent(g, me.ID, "Brightcap Badger", "Creature — Badger Druid", cmdr2B17BadgerOracle, false)
	sap := pushCatalogPermanent(g, me.ID, "Saproling", "Token Creature — Saproling", "", false)
	fungus := pushCatalogPermanent(g, me.ID, "Mold", "Creature — Fungus", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Saproling", "Token Creature — Saproling", "", false)
	manaOf := func(id uuid.UUID) (int, game.AbilityOrigins) {
		var n int
		var o game.AbilityOrigins
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(id)
			shapes, origins := game.ManaAbilitiesWithOrigins(c)
			n, o = len(shapes), origins
		})
		return n, o
	}
	if n, _ := manaOf(badger); n != 0 {
		t.Errorf("the Badger itself has %d mana abilities, want 0", n)
	}
	if n, _ := manaOf(theirs); n != 0 {
		t.Errorf("an opponent's Saproling has %d mana abilities, want 0", n)
	}
	for _, id := range []uuid.UUID{sap, fungus} {
		n, o := manaOf(id)
		if n != 1 {
			t.Fatalf("a Fungus/Saproling has %d mana abilities, want 1", n)
		}
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Ref: o.Ref(0)}); err != nil {
			t.Fatalf("tap for mana: %v", err)
		}
	}
	if n := len(me.ManaPool); n != 2 {
		t.Errorf("pool has %d mana, want {G}{G}", n)
	}
}

func TestCmdr2B17BrightcapBadgerMakesASaprolingAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Brightcap Badger", "Creature — Badger Druid", cmdr2B17BadgerOracle, false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Saproling", me.ID); got != 1 {
		t.Errorf("%d Saprolings at my end step, want 1", got)
	}
}

func TestCmdr2B17FungusFrolicMakesTwoSaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Fungus Frolic", "Instant — Adventure", cmdr2B17BadgerOracle+"#1", nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Saproling", me.ID); got != 2 {
		t.Errorf("%d Saprolings, want 2", got)
	}
}
