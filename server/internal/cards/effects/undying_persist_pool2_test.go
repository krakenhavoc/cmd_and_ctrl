package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undying_persist_pool2_test.go — the second pool PR after #2075 (ADR
// 0113 §4, owner decision 2): the cards that GIVE undying or persist,
// and the rest of the creatures whose text needed a little more. One
// test per card.

const (
	p2UndyingEvil     = "3ceaeb1b-8b25-453d-ac09-e466cfdd9c77"
	p2CauldronHaze    = "56309bbe-86aa-4003-a531-0ef1e318ba15"
	p2CauldronOfSouls = "5811d905-cae5-4f92-a78e-abb43f8864a7"
	p2AntlerSkulkin   = "30572dd1-c982-4872-9f35-a91ed8b238f7"
	p2Rhys            = "9f013b5d-fff0-4c50-8621-158b6dc34834"
	p2DuskLegion      = "f8b12c43-d83a-4e5b-a525-5ee1ae850817"
	p2Endling         = "4da3e510-e495-4077-997f-8efc04f01892"
	p2Mikaeus         = "5d27c63e-d1ef-48af-b51d-01ebc6daeac9"
	p2Rattleblaze     = "313130ab-13db-4d0a-b3f3-b60e060b303c"
	p2Wingrattle      = "034264da-a404-4037-b215-b095bba77490"
	p2Gargoyle        = "f3bd9bfe-847f-4ed1-88aa-6aa6b3fa7199"
	p2Heartmender     = "ab4cf0ba-617f-4ef6-a831-eb3b41b4ab34"
	p2Hancock         = "adf9aed9-dd63-48da-b799-ea94839fc22a"
	p2TwilightShep    = "eff736bc-d311-4bef-9c7e-de4ebe32e46f"
	p2Demonlord       = "049777d0-910d-46d2-9ca1-10277b2fb845"
)

func TestUndyingPersistPool2Registered(t *testing.T) {
	for _, oracle := range []string{p2UndyingEvil, p2CauldronHaze, p2CauldronOfSouls, p2AntlerSkulkin, p2Rhys,
		p2DuskLegion, p2Endling, p2Mikaeus, p2Rattleblaze, p2Wingrattle, p2Gargoyle, p2Heartmender, p2Hancock,
		p2TwilightShep, p2Demonlord} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

func p2Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

func p2Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, index int, mana string, targets ...game.TargetRef) error {
	t.Helper()
	floatForTest(g, p, mana)
	return g.ActivateCatalogAbility(p.ID, id, index, game.ActivateAbilityParams{Targets: targets})
}

func p2Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func p2Keyword(g *game.Game, id uuid.UUID, kw string) bool {
	has := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}

func p2PT(g *game.Game, id uuid.UUID) (int, int) {
	p, tough := 0, 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(id); ok {
			p, tough = c.PowerForComparison(), c.CurrentToughness()
		}
	})
	return p, tough
}

// Undying Evil: the granted undying returns the creature that dies.
func TestUndyingEvilReturnsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Undying Evil", "Instant", p2UndyingEvil, []game.TargetRef{p2Card(bear)})
	settleProwess(t, g)
	b25Destroy(g, bear)
	settleProwess(t, g)
	if !onBattlefield(g, bear) || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("the creature given undying did not return with a +1/+1 counter")
	}
}

// Cauldron Haze: any number of targets, each returns under its owner.
func TestCauldronHazeGivesEachTargetPersist(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Cauldron Haze", "Instant", p2CauldronHaze, []game.TargetRef{p2Card(mine), p2Card(theirs)})
	settleProwess(t, g)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mine, theirs}) })
	settleProwess(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !onBattlefield(g, id) || countersOn(g, id, game.CounterMinusOne) != 1 {
			t.Errorf("%s did not persist", id)
		}
	}
	if c, _ := battlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("the opponent's creature returned under the wrong control")
	}
}

// Cauldron of Souls: the tap ability, zero targets allowed.
func TestCauldronOfSoulsGivesPersist(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cauldron := p2Push(g, me.ID, "Cauldron of Souls", "Artifact", p2CauldronOfSouls, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := p2Activate(t, g, me, cauldron, 0, "", p2Card(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	settleProwess(t, g)
	if !p2Keyword(g, bear, game.KeywordPersist) {
		t.Fatal("the target has no persist")
	}
	b25Destroy(g, bear)
	settleProwess(t, g)
	if !onBattlefield(g, bear) {
		t.Error("the creature did not persist")
	}
}

// Antler Skulkin: only a white creature.
func TestAntlerSkulkinTargetsOnlyWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skulkin := p2Push(g, me.ID, "Antler Skulkin", "Artifact Creature — Scarecrow", p2AntlerSkulkin, 3, 3)
	white := p2Push(g, me.ID, "Knight", "Creature — Human Knight", "", 2, 2, "W")
	green := p2Push(g, me.ID, "Elf", "Creature — Elf", "", 1, 1, "G")
	if err := p2Activate(t, g, me, skulkin, 0, "CC", p2Card(green)); err == nil {
		t.Error("a green creature was accepted")
	}
	if err := p2Activate(t, g, me, skulkin, 0, "CC", p2Card(white)); err != nil {
		t.Fatalf("activate on a white creature: %v", err)
	}
	settleProwess(t, g)
	if !p2Keyword(g, white, game.KeywordPersist) {
		t.Error("the white creature has no persist")
	}
}

// Rhys, the Evermore: the enter trigger gives another creature persist;
// the activation removes as many of each kind of counter as you choose.
func TestRhysTheEvermore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rhys := enterCard(t, g, me.ID, game.Card{Name: "Rhys, the Evermore", TypeLine: "Legendary Creature — Elf Warrior", OracleID: p2Rhys, Power: 2, Toughness: 2})
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, id := range p.PickTargetCards {
			if id == rhys {
				t.Error("Rhys offered itself as \"another\" creature")
			}
		}
	}
	pickCard(t, g, me.ID, bear)
	settleProwess(t, g)
	if !p2Keyword(g, bear, game.KeywordPersist) {
		t.Fatal("the bear has no persist")
	}
	b25Destroy(g, bear)
	settleProwess(t, g)
	if countersOn(g, bear, game.CounterMinusOne) != 1 {
		t.Fatal("setup: the bear did not persist")
	}
	addCounters(t, g, bear, "charge", 2)

	nextUpkeepOf(t, g, 0) // Rhys loses summoning sickness
	advanceToMain(t, g)
	if err := p2Activate(t, g, me, rhys, 0, "W", p2Card(bear)); err != nil {
		t.Fatalf("activate Rhys: %v", err)
	}
	for i := 0; i < 8 && latestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	// Kinds are asked in name order: "-1/-1" before "charge".
	answerOptionPick(t, g, me.ID, 1) // remove the -1/-1 counter
	answerOptionPick(t, g, me.ID, 1) // keep one charge counter
	settleProwess(t, g)
	if countersOn(g, bear, game.CounterMinusOne) != 0 || countersOn(g, bear, "charge") != 1 {
		t.Errorf("counters after Rhys: -1/-1 %d, charge %d; want 0 and 1",
			countersOn(g, bear, game.CounterMinusOne), countersOn(g, bear, "charge"))
	}
}

// Dusk Legion Sergeant: nontoken Vampires you control gain persist.
func TestDuskLegionSergeantGivesNontokenVampiresPersist(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sgt := p2Push(g, me.ID, "Dusk Legion Sergeant", "Creature — Vampire Soldier", p2DuskLegion, 2, 2)
	vamp := p2Push(g, me.ID, "Vampire", "Creature — Vampire", "", 2, 2)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Vampire Token", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1}, 1)
	})
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire Token" {
			token = c.InstanceID
		}
	}
	if err := p2Activate(t, g, me, sgt, 0, "BB"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	settleProwess(t, g)
	if !p2Keyword(g, vamp, game.KeywordPersist) {
		t.Error("the nontoken Vampire has no persist")
	}
	if p2Keyword(g, bear, game.KeywordPersist) || p2Keyword(g, token, game.KeywordPersist) {
		t.Error("a non-Vampire or a token got persist")
	}
}

// Endling: three keyword grants and the +1/-1 or -1/+1 choice.
func TestEndling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	endling := p2Push(g, me.ID, "Endling", "Creature — Zombie Shapeshifter", p2Endling, 3, 3)
	for i, kw := range []string{"menace", "deathtouch", game.KeywordUndying} {
		if err := p2Activate(t, g, me, endling, i, "B"); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		settleProwess(t, g)
		if !p2Keyword(g, endling, kw) {
			t.Errorf("no %s", kw)
		}
	}
	if err := p2Activate(t, g, me, endling, 3, "C"); err != nil {
		t.Fatalf("activate the last ability: %v", err)
	}
	ask := passUntilConfirmFor(t, g, me.ID)
	if ask == nil {
		t.Fatal("no +1/-1 or -1/+1 question")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	settleProwess(t, g)
	if p, tough := p2PT(g, endling); p != 2 || tough != 4 {
		t.Errorf("after -1/+1: %d/%d, want 2/4", p, tough)
	}
	b25Destroy(g, endling)
	settleProwess(t, g)
	if countersOn(g, endling, game.CounterPlusOne) != 1 {
		t.Error("the undying it gained did not return it")
	}
}

// Mikaeus: other non-Humans get +1/+1 and undying; a Human that damages
// you is destroyed; a non-Human dying alongside Mikaeus still returns.
func TestMikaeusTheUnhallowed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mik := p2Push(g, me.ID, "Mikaeus, the Unhallowed", "Legendary Creature — Zombie Cleric", p2Mikaeus, 5, 5)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	human := p2Push(g, me.ID, "Soldier", "Creature — Human Soldier", "", 2, 2)
	foe := p2Push(g, opp.ID, "Raider", "Creature — Human Warrior", "", 2, 2)
	if p, _ := p2PT(g, bear); p != 3 || !p2Keyword(g, bear, game.KeywordUndying) {
		t.Errorf("the bear is %d power, undying %v; want 3 and true", p, p2Keyword(g, bear, game.KeywordUndying))
	}
	if p, _ := p2PT(g, human); p != 2 || p2Keyword(g, human, game.KeywordUndying) {
		t.Error("a Human got the bonus")
	}
	if p2Keyword(g, mik, game.KeywordUndying) {
		t.Error("Mikaeus gave itself undying")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(foe, me.ID, 1) })
	settleProwess(t, g)
	if onBattlefield(g, foe) {
		t.Error("the Human that damaged me survived")
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mik, bear}) })
	settleProwess(t, g)
	if !onBattlefield(g, bear) || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("a non-Human that died with Mikaeus did not return")
	}
}

// The Scarecrows have their keywords while you control a creature of
// the colour.
func TestScarecrowsKeywordsByColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rattle := p2Push(g, me.ID, "Rattleblaze Scarecrow", "Artifact Creature — Scarecrow", p2Rattleblaze, 5, 3)
	wing := p2Push(g, me.ID, "Wingrattle Scarecrow", "Artifact Creature — Scarecrow", p2Wingrattle, 2, 2)
	if p2Keyword(g, rattle, game.KeywordPersist) || p2Keyword(g, wing, "flying") {
		t.Fatal("keywords with no coloured creature")
	}
	black := p2Push(g, me.ID, "Black Knight", "Creature — Human Knight", "", 2, 2, "B")
	p2Push(g, me.ID, "Merfolk", "Creature — Merfolk", "", 1, 1, "U")
	if !p2Keyword(g, rattle, game.KeywordPersist) || !p2Keyword(g, wing, game.KeywordPersist) || !p2Keyword(g, wing, "flying") {
		t.Error("missing persist or flying with black and blue creatures")
	}
	if p2Keyword(g, rattle, "haste") {
		t.Error("haste with no red creature")
	}
	_ = black
	b25Destroy(g, rattle)
	settleProwess(t, g)
	if !onBattlefield(g, rattle) {
		t.Error("Rattleblaze did not persist while I controlled a black creature")
	}
}

// Obstinate Gargoyle flies once persist has modified it.
func TestObstinateGargoyleFliesWhenModified(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	garg := p2Push(g, me.ID, "Obstinate Gargoyle", "Artifact Creature — Gargoyle", p2Gargoyle, 2, 2)
	if p2Keyword(g, garg, "flying") {
		t.Fatal("flying while unmodified")
	}
	b25Destroy(g, garg)
	settleProwess(t, g)
	if !p2Keyword(g, garg, "flying") {
		t.Error("no flying with a -1/-1 counter")
	}
}

// Heartmender removes a -1/-1 counter from each of your creatures.
func TestHeartmenderRemovesMinusCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hm := p2Push(g, me.ID, "Heartmender", "Creature — Elemental", p2Heartmender, 2, 2)
	mine := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs := p2Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
	addCounters(t, g, mine, game.CounterMinusOne, 2)
	addCounters(t, g, theirs, game.CounterMinusOne, 1)
	addCounters(t, g, hm, game.CounterMinusOne, 1)
	nextUpkeepOf(t, g, 0)
	settleProwess(t, g)
	if countersOn(g, mine, game.CounterMinusOne) != 1 || countersOn(g, hm, game.CounterMinusOne) != 0 {
		t.Error("my creatures did not lose one -1/-1 counter each")
	}
	if countersOn(g, theirs, game.CounterMinusOne) != 1 {
		t.Error("an opponent's creature lost a counter")
	}
}

// Hancock pumps other Zombies and Mutants by its counters.
func TestHancockScalesWithItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hancock := p2Push(g, me.ID, "Hancock, Ghoulish Mayor", "Legendary Creature — Zombie Mutant Advisor", p2Hancock, 2, 1)
	zombie := p2Push(g, me.ID, "Zombie", "Creature — Zombie", "", 2, 2)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	b25Destroy(g, hancock)
	settleProwess(t, g)
	addCounters(t, g, hancock, "charge", 1)
	if p, _ := p2PT(g, zombie); p != 4 {
		t.Errorf("the Zombie has %d power, want 4 (two counters on Hancock)", p)
	}
	if p, _ := p2PT(g, bear); p != 2 {
		t.Error("a non-Zombie, non-Mutant was pumped")
	}
}

// Twilight Shepherd returns what died this turn when persist brings it
// back.
func TestTwilightShepherdReturnsTheDeadToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shep := p2Push(g, me.ID, "Twilight Shepherd", "Creature — Angel", p2TwilightShep, 5, 5)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Corpse", TypeLine: "Creature — Zombie", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{shep, bear}) })
	settleProwess(t, g)
	if !onBattlefield(g, shep) {
		t.Fatal("the Shepherd did not persist")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature that died this turn is not in my hand")
	}
	if me.Hand.Contains(old) {
		t.Error("a card that was already in the graveyard came back")
	}
}

// Demonlord of Ashmouth: exiled with nothing to sacrifice; with another
// creature you choose.
func TestDemonlordOfAshmouth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lone := enterCard(t, g, me.ID, game.Card{Name: "Demonlord of Ashmouth", TypeLine: "Creature — Demon", OracleID: p2Demonlord, Power: 5, Toughness: 4})
	settleProwess(t, g)
	if onBattlefield(g, lone) || !g.Exile.Contains(lone) {
		t.Fatal("with nothing to sacrifice it should be exiled")
	}
	fodder := p2Push(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)
	kept := enterCard(t, g, me.ID, game.Card{Name: "Demonlord of Ashmouth", TypeLine: "Creature — Demon", OracleID: p2Demonlord, Power: 5, Toughness: 4})
	for i := 0; i < 8 && latestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerOptionPick(t, g, me.ID, 1)
	answerChooseCards(t, g, me.ID, fodder)
	settleProwess(t, g)
	if !onBattlefield(g, kept) || onBattlefield(g, fodder) {
		t.Error("sacrificing another creature should keep the Demonlord")
	}
}
