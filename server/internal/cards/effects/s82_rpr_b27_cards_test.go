package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b27_cards_test.go — S82 reprint pass, batch 27 (#5908).

func rprB27Put(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, pw, tf int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: pw, Toughness: tf, Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}

func rprB27Count(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

const rprB27Bishop = "a47d9ea3-95fe-4774-8e72-30e7f56756f4"

func TestRprB27BygoneBishopInvestigatesOnCheapCreatureSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		clues                int
	}{
		{"Cheap", "Creature — Human", "{2}", 1},
		{"Pricey", "Creature — Giant", "{3}{G}", 0},
		{"Spell", "Instant", "{1}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			rprB27Put(g, me, "Bygone Bishop", "Creature — Spirit Cleric", "{2}{W}", rprB27Bishop, 2, 3)
			advanceToMain(t, g)
			castFromHandForTest(t, g, me, tc.name, tc.typeLine, tc.cost, "", game.CastSpellParams{})
			passPriorityAroundTable(t, g)
			if got := rprB27Count(g, "Clue", me.ID); got != tc.clues {
				t.Errorf("%d Clues, want %d", got, tc.clues)
			}
		})
	}
}

const (
	rprB27Ravager    = "8f1016c7-0551-4813-987d-36abf3b2ffeb"
	rprB27Anowon     = "0c201f79-9eca-48c8-88ba-a20fa6918a07"
	rprB27Locket     = "ca13dda2-fbb0-4b6b-9126-3c8df04dee2b"
	rprB27Offering   = "3db8466c-864f-48ba-b617-f2c863dfdcbd"
	rprB27Bloodlord  = "d8560ee9-8cbb-440b-a132-e102a1eb3e89"
	rprB27Steward    = "a5ec7ccd-d1fe-4ca6-b019-d4616c233771"
	rprB27Charger    = "f6dae324-cc79-4093-8cce-667015c1de6d"
	rprB27Guard      = "3e638aa1-31ff-4e57-9e4e-444211031eb5"
	rprB27Soulbinder = "c7a8daff-9913-4970-8765-52ba9d9d01db"
	rprB27Dovin      = "596237da-3d51-4052-a8ca-2218efc1afb3"
)

func TestRprB27CardsAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{rprB27Bishop, rprB27Ravager, rprB27Anowon, rprB27Locket, rprB27Offering,
		rprB27Bloodlord, rprB27Steward, rprB27Charger, rprB27Guard, rprB27Soulbinder, rprB27Dovin} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Anje's Ravager -----------------------------------------------------------

func TestRprB27AnjesRavagerMustAttackAndWheelsOnAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if spec, _ := Lookup(rprB27Ravager); spec.Madness != "{1}{R}" {
		t.Errorf("madness cost %q, want {1}{R}", spec.Madness)
	}
	ravager := pushCatalogPermanent(g, me.ID, "Anje's Ravager", "Creature — Vampire Berserker", rprB27Ravager, false)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("pass with the Ravager home = %v, want ErrAttackRequirement", err)
	}
	handBefore := me.Hand.Size()
	if handBefore == 0 {
		t.Fatal("the test hand is empty")
	}
	gyBefore := me.Graveyard.Size()
	if err := g.DeclareAttacker(ravager, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand = %d, want 3", got)
	}
	if got := me.Graveyard.Size() - gyBefore; got != handBefore {
		t.Errorf("%d cards discarded, want %d", got, handBefore)
	}
}

// --- Anowon, the Ruin Sage ----------------------------------------------------

func TestRprB27AnowonEachPlayerSacrificesANonVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[1], g.Seats[0], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Anowon, the Ruin Sage", "Legendary Creature — Vampire Shaman", rprB27Anowon, false)
	myBear := seedCreature(g, "My Bear", me.ID)
	vamp := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Count", TypeLine: "Creature — Vampire", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID,
	})
	theirBear := seedCreature(g, "Their Bear", b.ID)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myBear) {
		t.Error("Anowon's controller kept their non-Vampire")
	}
	if g.Battlefield.Contains(theirBear) {
		t.Error("an opponent kept their non-Vampire")
	}
	if !g.Battlefield.Contains(vamp) {
		t.Error("a Vampire was sacrificed")
	}
	if findBattlefieldByName(g, "Anowon, the Ruin Sage") == uuid.Nil {
		t.Error("Anowon sacrificed itself")
	}
}

// --- Crimson Honor Guard ------------------------------------------------------

func TestRprB27CrimsonHonorGuardDamagesPlayersWithoutACommander(t *testing.T) {
	for _, tc := range []struct {
		name      string
		commander bool
		loss      int
	}{{"no commander", false, 4}, {"commander", true, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			active, other := g.Seats[0], g.Seats[1]
			pushCatalogPermanent(g, other.ID, "Crimson Honor Guard", "Creature — Vampire Knight", rprB27Guard, false)
			if tc.commander {
				pushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Boss", TypeLine: "Legendary Creature — Avatar",
					Power: 2, Toughness: 2, IsCommander: true, Owner: active.ID, Controller: active.ID,
				})
			}
			lifeActive, lifeOther := active.Life, other.Life
			advanceToEndStep(t, g)
			passPriorityAroundTable(t, g)
			if got := lifeActive - active.Life; got != tc.loss {
				t.Errorf("the active player lost %d, want %d", got, tc.loss)
			}
			if other.Life != lifeOther {
				t.Errorf("the Guard's controller lost %d on someone else's end step", lifeOther-other.Life)
			}
		})
	}
}

func TestRprB27CrimsonHonorGuardHitsItsOwnControllerToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Crimson Honor Guard", "Creature — Vampire Knight", rprB27Guard, false)
	before := me.Life
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := before - me.Life; got != 4 {
		t.Errorf("lost %d, want 4", got)
	}
}

// --- Azorius Locket -----------------------------------------------------------

func TestRprB27AzoriusLocketMakesWhiteOrBlueAndSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	locket := pushCatalogPermanent(g, me.ID, "Azorius Locket", "Artifact", rprB27Locket, false)
	if err := g.ActivateManaAbility(me.ID, locket, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool = %d, want 1", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	// Tapped, so the draw ability (which needs the {T}) is refused.
	if err := g.ActivateCatalogAbility(me.ID, locket, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Locket drew cards")
	}
}

func TestRprB27AzoriusLocketPaysFourHybridAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	locket := pushCatalogPermanent(g, me.ID, "Azorius Locket", "Artifact", rprB27Locket, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "U"}, game.ManaToken{Color: "W"})
	if err := g.ActivateCatalogAbility(me.ID, locket, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("three mana paid for four hybrid symbols")
	}
	if g.Battlefield.Contains(locket) == false {
		t.Fatal("the refused activation still sacrificed the Locket")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, locket, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(locket) {
		t.Error("the Locket was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// --- Benevolent Offering ------------------------------------------------------

func TestRprB27BenevolentOfferingMakesSpiritsThenGainsLifePerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	seedCreature(g, "My Bear", me.ID)
	seedCreature(g, "Their Bear A", second.ID)
	seedCreature(g, "Their Bear B", second.ID)
	castCatalogSpell(t, g, "Benevolent Offering", "Instant", rprB27Offering, nil)
	passPriorityAroundTable(t, g)
	otChooseOpponent(t, g, me.ID, first.ID)
	if got := rprB27Count(g, "Spirit", me.ID); got != 3 {
		t.Errorf("you have %d Spirits, want 3", got)
	}
	if got := rprB27Count(g, "Spirit", first.ID); got != 3 {
		t.Errorf("the first opponent has %d Spirits, want 3", got)
	}
	lifeMe, lifeFirst, lifeSecond := me.Life, first.Life, second.Life
	otChooseOpponent(t, g, me.ID, second.ID)
	if got := me.Life - lifeMe; got != 8 {
		t.Errorf("you gained %d, want 8 (four creatures)", got)
	}
	if got := second.Life - lifeSecond; got != 4 {
		t.Errorf("the second opponent gained %d, want 4 (two creatures)", got)
	}
	if first.Life != lifeFirst {
		t.Error("the first opponent gained life from the second choice")
	}
}

// --- Bloodlord of Vaasgoth ----------------------------------------------------

func rprB27CastVampire(t *testing.T, g *game.Game, me *game.Player, typeLine string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Test Card", TypeLine: typeLine, ManaCost: "{1}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func rprB27DamageOpponent(g *game.Game, from, to uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: from, Target: to, Amount: 2})
	})
}

func TestRprB27BloodlordGivesVampireSpellsBloodthirst(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		damaged  bool
		counters int
	}{
		{"vampire, opponent damaged", "Creature — Vampire", true, 3},
		{"vampire, nobody damaged", "Creature — Vampire", false, 0},
		{"non-vampire, opponent damaged", "Creature — Human", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lord := pushCatalogPermanent(g, me.ID, "Bloodlord of Vaasgoth", "Creature — Vampire Warrior", rprB27Bloodlord, false)
			if tc.damaged {
				rprB27DamageOpponent(g, lord, opp.ID)
			}
			id := rprB27CastVampire(t, g, me, tc.typeLine)
			c, ok := battlefieldCard(g, id)
			if !ok {
				t.Fatal("the creature is not on the battlefield")
			}
			if got := c.Counters[game.CounterPlusOne]; got != tc.counters {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.counters)
			}
		})
	}
}

func TestRprB27BloodlordHasBloodthirstItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := seedCreature(g, "Pinger", me.ID)
	rprB27DamageOpponent(g, src, opp.ID)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Bloodlord of Vaasgoth", TypeLine: "Creature — Vampire Warrior", OracleID: rprB27Bloodlord,
			Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
		})
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Bloodlord is not on the battlefield")
	}
	if got := c.Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("%d counters, want 3", got)
	}
}

// --- Bloodsworn Steward -------------------------------------------------------

func TestRprB27BloodswornStewardPumpsAndHastesOnlyYourCommanders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bloodsworn Steward", TypeLine: "Creature — Vampire Knight", OracleID: rprB27Steward,
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	cmdr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Boss", TypeLine: "Legendary Creature — Avatar", Power: 3, Toughness: 3,
		IsCommander: true, Owner: me.ID, Controller: me.ID,
	})
	bear := seedCreature(g, "Bear", me.ID)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Boss", TypeLine: "Legendary Creature — Avatar", Power: 3, Toughness: 3,
		IsCommander: true, Owner: opp.ID, Controller: opp.ID,
	})
	if p, tough := effectivePower(t, g, cmdr), effectiveToughness(t, g, cmdr); p != 5 || tough != 5 {
		t.Errorf("my commander is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, cmdr, "haste") {
		t.Error("my commander lacks haste")
	}
	if effectivePower(t, g, bear) != 2 || hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a non-commander was pumped or hasted")
	}
	if effectivePower(t, g, theirs) != 3 || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's commander was pumped or hasted")
	}
}

// --- Custodi Soulbinders ------------------------------------------------------

func TestRprB27CustodiSoulbindersEntersWithCountersAndMakesSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedCreature(g, "A", me.ID)
	seedCreature(g, "B", opp.ID)
	seedCreature(g, "C", opp.ID)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Custodi Soulbinders", TypeLine: "Creature — Human Cleric", OracleID: rprB27Soulbinder,
			Owner: me.ID, Controller: me.ID,
		})
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("not on the battlefield")
	}
	if got := c.Counters[game.CounterPlusOne]; got != 3 {
		t.Fatalf("%d counters, want 3 (the other creatures)", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rprB27Count(g, "Spirit", me.ID); got != 1 {
		t.Errorf("%d Spirits, want 1", got)
	}
	if c, _ := battlefieldCard(g, id); c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("%d counters left, want 2", c.Counters[game.CounterPlusOne])
	}
}

// --- Dovin, Grand Arbiter -----------------------------------------------------

func TestRprB27DovinPlusOneCountsEachCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dovin := pushWalkerForTest(g, me.ID, "Dovin, Grand Arbiter", rprB27Dovin, 3)
	a := seedCreature(g, "A", me.ID)
	b := seedCreature(g, "B", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dovin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	loyalty := func() int { c, _ := battlefieldCard(g, dovin); return c.Counters[game.CounterLoyalty] }
	if loyalty() != 4 {
		t.Fatalf("loyalty %d after +1, want 4", loyalty())
	}
	hit := func(src, to uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: src, Target: to, Amount: 2, Combat: true})
		})
		passPriorityAroundTable(t, g)
	}
	hit(a, opp.ID)
	hit(b, opp.ID)
	if loyalty() != 6 {
		t.Errorf("loyalty %d after two creatures connected, want 6", loyalty())
	}
	hit(theirs, me.ID)
	if loyalty() != 6 {
		t.Errorf("an opponent's creature added loyalty (%d)", loyalty())
	}
}

func TestRprB27DovinMinusOneMakesAThopterAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dovin := pushWalkerForTest(g, me.ID, "Dovin, Grand Arbiter", rprB27Dovin, 3)
	advanceToMain(t, g)
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, dovin, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if rprB27Count(g, "Thopter", me.ID) != 1 {
		t.Error("no Thopter")
	}
	if me.Life-before != 1 {
		t.Errorf("gained %d, want 1", me.Life-before)
	}
}

func TestRprB27DovinMinusSevenTakesThreeOfTheTopTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dovin := pushWalkerForTest(g, me.ID, "Dovin, Grand Arbiter", rprB27Dovin, 7)
	advanceToMain(t, g)
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, dovin, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-7: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := chooseCardsPrompt(t, g)
	if len(c.ChooseCards) != 10 {
		t.Fatalf("%d candidates, want the top ten", len(c.ChooseCards))
	}
	picks := append([]uuid.UUID(nil), c.ChooseCards[:3]...)
	if err := g.ResolveChooseCards(c.ID, me.ID, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
	if got := libBefore - me.Library.Size(); got != 3 {
		t.Errorf("library shrank by %d, want 3 (the rest go to the bottom)", got)
	}
}

// --- Boreas Charger -----------------------------------------------------------

func TestRprB27BoreasChargerFetchesPlainsForTheLandDifference(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	charger := pushCatalogPermanent(g, me.ID, "Boreas Charger", "Creature — Pegasus", rprB27Charger, false)
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 4; i++ {
		seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	}
	var plains []uuid.UUID
	for i := 0; i < 4; i++ {
		plains = append(plains, stapleLibraryCard(me, "Plains", "Basic Land — Plains"))
	}
	handBefore := me.Hand.Size()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, charger); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	otChooseOpponent(t, g, me.ID, opp.ID)
	c := searchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 3 {
		t.Fatalf("want a search for three Plains, got %+v", c)
	}
	answerSearchByID(t, g, me.ID, plains[0], plains[1], plains[2])
	passPriorityAroundTable(t, g)
	pick := chooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{plains[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, plains[1])
	if !ok || !got.Tapped {
		t.Errorf("the chosen Plains should be on the battlefield tapped (found %v, %+v)", ok, got)
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
}

func TestRprB27BoreasChargerDoesNothingWhenNoOpponentHasMoreLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	charger := pushCatalogPermanent(g, me.ID, "Boreas Charger", "Creature — Pegasus", rprB27Charger, false)
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	stapleLibraryCard(me, "Plains", "Basic Land — Plains")
	handBefore := me.Hand.Size()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, charger); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil || me.Hand.Size() != handBefore {
		t.Error("searched with no opponent ahead on lands")
	}
}
