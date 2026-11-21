package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestUndyingPersistPool2Registered(t *testing.T) {
	for _, oracle := range []string{effectstest.P2UndyingEvil, effectstest.P2CauldronHaze, effectstest.P2CauldronOfSouls, effectstest.P2AntlerSkulkin, effectstest.P2Rhys,
		effectstest.P2DuskLegion, effectstest.P2Endling, effectstest.P2Mikaeus, effectstest.P2Rattleblaze, effectstest.P2Wingrattle, effectstest.P2Gargoyle, effectstest.P2Heartmender, effectstest.P2Hancock,
		effectstest.P2TwilightShep, effectstest.P2Demonlord} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// Undying Evil: the granted undying returns the creature that dies.
func TestUndyingEvilReturnsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Undying Evil", "Instant", effectstest.P2UndyingEvil, []game.TargetRef{effectstest.P2Card(bear)})
	effectstest.SettleProwess(t, g)
	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, bear) || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("the creature given undying did not return with a +1/+1 counter")
	}
}

// Cauldron Haze: any number of targets, each returns under its owner.
func TestCauldronHazeGivesEachTargetPersist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Cauldron Haze", "Instant", effectstest.P2CauldronHaze, []game.TargetRef{effectstest.P2Card(mine), effectstest.P2Card(theirs)})
	effectstest.SettleProwess(t, g)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mine, theirs}) })
	effectstest.SettleProwess(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effects.OnBattlefield(g, id) || effectstest.CountersOn(g, id, game.CounterMinusOne) != 1 {
			t.Errorf("%s did not persist", id)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("the opponent's creature returned under the wrong control")
	}
}

// Cauldron of Souls: the tap ability, zero targets allowed.
func TestCauldronOfSoulsGivesPersist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cauldron := effectstest.P2Push(g, me.ID, "Cauldron of Souls", "Artifact", effectstest.P2CauldronOfSouls, 0, 0)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.P2Activate(t, g, me, cauldron, 0, "", effectstest.P2Card(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.SettleProwess(t, g)
	if !effectstest.P2Keyword(g, bear, game.KeywordPersist) {
		t.Fatal("the target has no persist")
	}
	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Error("the creature did not persist")
	}
}

// Antler Skulkin: only a white creature.
func TestAntlerSkulkinTargetsOnlyWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skulkin := effectstest.P2Push(g, me.ID, "Antler Skulkin", "Artifact Creature — Scarecrow", effectstest.P2AntlerSkulkin, 3, 3)
	white := effectstest.P2Push(g, me.ID, "Knight", "Creature — Human Knight", "", 2, 2, "W")
	green := effectstest.P2Push(g, me.ID, "Elf", "Creature — Elf", "", 1, 1, "G")
	if err := effectstest.P2Activate(t, g, me, skulkin, 0, "CC", effectstest.P2Card(green)); err == nil {
		t.Error("a green creature was accepted")
	}
	if err := effectstest.P2Activate(t, g, me, skulkin, 0, "CC", effectstest.P2Card(white)); err != nil {
		t.Fatalf("activate on a white creature: %v", err)
	}
	effectstest.SettleProwess(t, g)
	if !effectstest.P2Keyword(g, white, game.KeywordPersist) {
		t.Error("the white creature has no persist")
	}
}

// Rhys, the Evermore: the enter trigger gives another creature persist;
// the activation removes as many of each kind of counter as you choose.
func TestRhysTheEvermore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rhys := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Rhys, the Evermore", TypeLine: "Legendary Creature — Elf Warrior", OracleID: effectstest.P2Rhys, Power: 2, Toughness: 2})
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		for _, id := range p.PickTargetCards {
			if id == rhys {
				t.Error("Rhys offered itself as \"another\" creature")
			}
		}
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.SettleProwess(t, g)
	if !effectstest.P2Keyword(g, bear, game.KeywordPersist) {
		t.Fatal("the bear has no persist")
	}
	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
	if effectstest.CountersOn(g, bear, game.CounterMinusOne) != 1 {
		t.Fatal("setup: the bear did not persist")
	}
	effectstest.AddCounters(t, g, bear, "charge", 2)

	nextUpkeepOf(t, g, 0) // Rhys loses summoning sickness
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.P2Activate(t, g, me, rhys, 0, "W", effectstest.P2Card(bear)); err != nil {
		t.Fatalf("activate Rhys: %v", err)
	}
	for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	// Kinds are asked in name order: "-1/-1" before "charge".
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // remove the -1/-1 counter
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // keep one charge counter
	effectstest.SettleProwess(t, g)
	if effectstest.CountersOn(g, bear, game.CounterMinusOne) != 0 || effectstest.CountersOn(g, bear, "charge") != 1 {
		t.Errorf("counters after Rhys: -1/-1 %d, charge %d; want 0 and 1",
			effectstest.CountersOn(g, bear, game.CounterMinusOne), effectstest.CountersOn(g, bear, "charge"))
	}
}

// Dusk Legion Sergeant: nontoken Vampires you control gain persist.
func TestDuskLegionSergeantGivesNontokenVampiresPersist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sgt := effectstest.P2Push(g, me.ID, "Dusk Legion Sergeant", "Creature — Vampire Soldier", effectstest.P2DuskLegion, 2, 2)
	vamp := effectstest.P2Push(g, me.ID, "Vampire", "Creature — Vampire", "", 2, 2)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Vampire Token", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1}, 1)
	})
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire Token" {
			token = c.InstanceID
		}
	}
	if err := effectstest.P2Activate(t, g, me, sgt, 0, "BB"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.SettleProwess(t, g)
	if !effectstest.P2Keyword(g, vamp, game.KeywordPersist) {
		t.Error("the nontoken Vampire has no persist")
	}
	if effectstest.P2Keyword(g, bear, game.KeywordPersist) || effectstest.P2Keyword(g, token, game.KeywordPersist) {
		t.Error("a non-Vampire or a token got persist")
	}
}

// Endling: three keyword grants and the +1/-1 or -1/+1 choice.
func TestEndling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	endling := effectstest.P2Push(g, me.ID, "Endling", "Creature — Zombie Shapeshifter", effectstest.P2Endling, 3, 3)
	for i, kw := range []string{"menace", "deathtouch", game.KeywordUndying} {
		if err := effectstest.P2Activate(t, g, me, endling, i, "B"); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.SettleProwess(t, g)
		if !effectstest.P2Keyword(g, endling, kw) {
			t.Errorf("no %s", kw)
		}
	}
	if err := effectstest.P2Activate(t, g, me, endling, 3, "C"); err != nil {
		t.Fatalf("activate the last ability: %v", err)
	}
	ask := effectstest.PassUntilConfirmFor(t, g, me.ID)
	if ask == nil {
		t.Fatal("no +1/-1 or -1/+1 question")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.SettleProwess(t, g)
	if p, tough := effectstest.P2PT(g, endling); p != 2 || tough != 4 {
		t.Errorf("after -1/+1: %d/%d, want 2/4", p, tough)
	}
	effectstest.B25Destroy(g, endling)
	effectstest.SettleProwess(t, g)
	if effectstest.CountersOn(g, endling, game.CounterPlusOne) != 1 {
		t.Error("the undying it gained did not return it")
	}
}

// Mikaeus: other non-Humans get +1/+1 and undying; a Human that damages
// you is destroyed; a non-Human dying alongside Mikaeus still returns.
func TestMikaeusTheUnhallowed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mik := effectstest.P2Push(g, me.ID, "Mikaeus, the Unhallowed", "Legendary Creature — Zombie Cleric", effectstest.P2Mikaeus, 5, 5)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	human := effectstest.P2Push(g, me.ID, "Soldier", "Creature — Human Soldier", "", 2, 2)
	foe := effectstest.P2Push(g, opp.ID, "Raider", "Creature — Human Warrior", "", 2, 2)
	if p, _ := effectstest.P2PT(g, bear); p != 3 || !effectstest.P2Keyword(g, bear, game.KeywordUndying) {
		t.Errorf("the bear is %d power, undying %v; want 3 and true", p, effectstest.P2Keyword(g, bear, game.KeywordUndying))
	}
	if p, _ := effectstest.P2PT(g, human); p != 2 || effectstest.P2Keyword(g, human, game.KeywordUndying) {
		t.Error("a Human got the bonus")
	}
	if effectstest.P2Keyword(g, mik, game.KeywordUndying) {
		t.Error("Mikaeus gave itself undying")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(foe, me.ID, 1) })
	effectstest.SettleProwess(t, g)
	if effects.OnBattlefield(g, foe) {
		t.Error("the Human that damaged me survived")
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mik, bear}) })
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, bear) || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("a non-Human that died with Mikaeus did not return")
	}
}

// The Scarecrows have their keywords while you control a creature of
// the colour.
func TestScarecrowsKeywordsByColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rattle := effectstest.P2Push(g, me.ID, "Rattleblaze Scarecrow", "Artifact Creature — Scarecrow", effectstest.P2Rattleblaze, 5, 3)
	wing := effectstest.P2Push(g, me.ID, "Wingrattle Scarecrow", "Artifact Creature — Scarecrow", effectstest.P2Wingrattle, 2, 2)
	if effectstest.P2Keyword(g, rattle, game.KeywordPersist) || effectstest.P2Keyword(g, wing, "flying") {
		t.Fatal("keywords with no coloured creature")
	}
	black := effectstest.P2Push(g, me.ID, "Black Knight", "Creature — Human Knight", "", 2, 2, "B")
	effectstest.P2Push(g, me.ID, "Merfolk", "Creature — Merfolk", "", 1, 1, "U")
	if !effectstest.P2Keyword(g, rattle, game.KeywordPersist) || !effectstest.P2Keyword(g, wing, game.KeywordPersist) || !effectstest.P2Keyword(g, wing, "flying") {
		t.Error("missing persist or flying with black and blue creatures")
	}
	if effectstest.P2Keyword(g, rattle, "haste") {
		t.Error("haste with no red creature")
	}
	_ = black
	effectstest.B25Destroy(g, rattle)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, rattle) {
		t.Error("Rattleblaze did not persist while I controlled a black creature")
	}
}

// Obstinate Gargoyle flies once persist has modified it.
func TestObstinateGargoyleFliesWhenModified(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	garg := effectstest.P2Push(g, me.ID, "Obstinate Gargoyle", "Artifact Creature — Gargoyle", effectstest.P2Gargoyle, 2, 2)
	if effectstest.P2Keyword(g, garg, "flying") {
		t.Fatal("flying while unmodified")
	}
	effectstest.B25Destroy(g, garg)
	effectstest.SettleProwess(t, g)
	if !effectstest.P2Keyword(g, garg, "flying") {
		t.Error("no flying with a -1/-1 counter")
	}
}

// Heartmender removes a -1/-1 counter from each of your creatures.
func TestHeartmenderRemovesMinusCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hm := effectstest.P2Push(g, me.ID, "Heartmender", "Creature — Elemental", effectstest.P2Heartmender, 2, 2)
	mine := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs := effectstest.P2Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
	effectstest.AddCounters(t, g, mine, game.CounterMinusOne, 2)
	effectstest.AddCounters(t, g, theirs, game.CounterMinusOne, 1)
	effectstest.AddCounters(t, g, hm, game.CounterMinusOne, 1)
	nextUpkeepOf(t, g, 0)
	effectstest.SettleProwess(t, g)
	if effectstest.CountersOn(g, mine, game.CounterMinusOne) != 1 || effectstest.CountersOn(g, hm, game.CounterMinusOne) != 0 {
		t.Error("my creatures did not lose one -1/-1 counter each")
	}
	if effectstest.CountersOn(g, theirs, game.CounterMinusOne) != 1 {
		t.Error("an opponent's creature lost a counter")
	}
}

// Hancock pumps other Zombies and Mutants by its counters.
func TestHancockScalesWithItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hancock := effectstest.P2Push(g, me.ID, "Hancock, Ghoulish Mayor", "Legendary Creature — Zombie Mutant Advisor", effectstest.P2Hancock, 2, 1)
	zombie := effectstest.P2Push(g, me.ID, "Zombie", "Creature — Zombie", "", 2, 2)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B25Destroy(g, hancock)
	effectstest.SettleProwess(t, g)
	effectstest.AddCounters(t, g, hancock, "charge", 1)
	if p, _ := effectstest.P2PT(g, zombie); p != 4 {
		t.Errorf("the Zombie has %d power, want 4 (two counters on Hancock)", p)
	}
	if p, _ := effectstest.P2PT(g, bear); p != 2 {
		t.Error("a non-Zombie, non-Mutant was pumped")
	}
}

// Twilight Shepherd returns what died this turn when persist brings it
// back.
func TestTwilightShepherdReturnsTheDeadToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shep := effectstest.P2Push(g, me.ID, "Twilight Shepherd", "Creature — Angel", effectstest.P2TwilightShep, 5, 5)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Corpse", TypeLine: "Creature — Zombie", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{shep, bear}) })
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, shep) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lone := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Demonlord of Ashmouth", TypeLine: "Creature — Demon", OracleID: effectstest.P2Demonlord, Power: 5, Toughness: 4})
	effectstest.SettleProwess(t, g)
	if effects.OnBattlefield(g, lone) || !g.Exile.Contains(lone) {
		t.Fatal("with nothing to sacrifice it should be exiled")
	}
	fodder := effectstest.P2Push(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)
	kept := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Demonlord of Ashmouth", TypeLine: "Creature — Demon", OracleID: effectstest.P2Demonlord, Power: 5, Toughness: 4})
	for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, kept) || effects.OnBattlefield(g, fodder) {
		t.Error("sacrificing another creature should keep the Demonlord")
	}
}
