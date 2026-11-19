package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func playerRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }

func TestS82BnsB02HypothesizzleDrawsTwoThenDiscardingANonlandDealsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	bear := handCardForTest(me, "Bear", "Creature — Bear", "")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Hypothesizzle", "Instant", s82BnsHypothesizzleOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand is %d, want %d after drawing two (the spell left it)", me.Hand.Size(), hand+2)
	}
	answerChooseCards(t, g, me.ID, bear)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("four damage did not kill the 4/4")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestS82BnsB02HypothesizzleRefusesALandAndDeclineMakesNoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	castCatalogSpell(t, g, "Hypothesizzle", "Instant", s82BnsHypothesizzleOracle, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("a land was accepted as the nonland discard")
	}
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a reflexive trigger went on the stack without a discard")
	}
	if !g.Battlefield.Contains(victim) || !me.Hand.Contains(land) {
		t.Error("something happened without a discard")
	}
}

func TestS82BnsB02BackForMoreReturnsThenFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	theirs := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 1, 3)
	if err := castCatalogSpellErrStrict(t, g, "Back for More", "Instant", s82BnsBackForMoreOracle, cardRefs(theirs)); err == nil {
		t.Fatal("targeted a creature on the battlefield")
	}
	castCatalogSpell(t, g, "Back for More", "Instant", s82BnsBackForMoreOracle, cardRefs(dead))
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature did not return")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("3 damage did not kill the 1/3 Wall")
	}
	if got := bnsDamageOn(g, dead); got != 1 {
		t.Errorf("the returned creature has %d damage, want 1", got)
	}
}

func bnsDamageOn(g *game.Game, id uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

func TestS82BnsB02RideDownDestroysABlockerAndGivesTheBlockedTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Unblocked", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	if err := castCatalogSpellErrStrict(t, g, "Ride Down", "Instant", s82BnsRideDownOracle, cardRefs(blocker)); err == nil {
		t.Fatal("Ride Down targeted a creature that is not blocking")
	}
	attackIntoBlocks(t, g, opp.ID, attacker, other)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	g2CastFrom(t, g, me, "Ride Down", "Instant", s82BnsRideDownOracle, game.CastSpellParams{Targets: cardRefs(blocker)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocker survived")
	}
	if !hasEffectiveKeyword(t, g, attacker, "trample") {
		t.Error("the creature it blocked did not gain trample")
	}
	if hasEffectiveKeyword(t, g, other, "trample") {
		t.Error("an unblocked creature gained trample")
	}
}

func TestS82BnsB02GrindstoneRepeatsWhileTheMilledPairSharesAColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := func(n string) game.Card {
		return game.Card{Name: n, TypeLine: "Creature — Bear", ManaCost: "{R}", Colors: []string{"R"}}
	}
	green := game.Card{Name: "Green", TypeLine: "Creature — Bear", ManaCost: "{G}", Colors: []string{"G"}}
	// Top-first: R1 R2 (share red, repeat), then R3 and Green (no shared
	// color, stop), leaving Deep untouched.
	ids := seedSearchLibrary(opp, red("R1"), red("R2"), red("R3"), green, game.Card{Name: "Deep", TypeLine: "Land"})
	stone := b12Push(g, me.ID, "Grindstone", "Artifact", s82BnsGrindstoneOracle, 0, 0)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerRef(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range ids[:4] {
		if !opp.Graveyard.Contains(id) {
			t.Errorf("a card that should be milled is still in the library")
		}
	}
	if !opp.Library.Contains(ids[4]) {
		t.Error("the repeat did not stop after a pair that shares no color")
	}
}

func TestS82BnsB02GrindstoneStopsAfterOnePassWhenNothingIsShared(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := seedSearchLibrary(opp,
		game.Card{Name: "A", TypeLine: "Creature — Bear", Colors: []string{"R"}},
		game.Card{Name: "B", TypeLine: "Creature — Bear", Colors: []string{"G"}},
		game.Card{Name: "Deep", TypeLine: "Creature — Bear", Colors: []string{"R"}})
	stone := b12Push(g, me.ID, "Grindstone", "Artifact", s82BnsGrindstoneOracle, 0, 0)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerRef(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(ids[0]) || !opp.Graveyard.Contains(ids[1]) || !opp.Library.Contains(ids[2]) {
		t.Error("exactly two cards should have been milled")
	}
}

func TestS82BnsB02ContagionEngineCountersTheChosenPlayersCreaturesAndProliferatesTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCounterCreature(g, opp.ID, "Their A", "", 0)
	b := pushCounterCreature(g, opp.ID, "Their B", "", 0)
	mine := pushCounterCreature(g, me.ID, "My Bear", "", 0)
	engine := castCatalogSpell(t, g, "Contagion Engine", "Artifact", s82BnsContagionEngineOracle, nil)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, playerRef(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	settleAnsweringProliferate(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := counterCount(g, id, game.CounterMinusOne); got != 1 {
			t.Errorf("their creature has %d -1/-1 counters, want 1", got)
		}
	}
	if got := counterCount(g, mine, game.CounterMinusOne); got != 0 {
		t.Errorf("my creature has %d counters, want 0", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	settleAnsweringProliferate(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := counterCount(g, id, game.CounterMinusOne); got != 3 {
			t.Errorf("after proliferating twice their creature has %d counters, want 3", got)
		}
	}
}

func TestS82BnsB02DetentionSphereExilesEverySameNamedPermanentAndReturnsThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf1 := b12Permanent(g, opp.ID, "Wolf Totem", "Artifact")
	wolf2 := b12Permanent(g, me.ID, "Wolf Totem", "Artifact")
	other := b12Permanent(g, opp.ID, "Other Totem", "Artifact")
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if detentionSphereNotNamedSphere(g, me.ID, game.Card{Name: "Detention Sphere"}) {
		t.Error("a permanent named Detention Sphere is a legal target")
	}
	sphere := castCatalogSpell(t, g, "Detention Sphere", "Enchantment", s82BnsDetentionSphereOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Fatal("a land was a legal target")
	}
	pickCard(t, g, me.ID, wolf1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wolf1) || g.Battlefield.Contains(wolf2) {
		t.Fatal("both same-named permanents should be exiled")
	}
	if !g.Battlefield.Contains(other) || !g.Battlefield.Contains(land) {
		t.Error("a differently named permanent was exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(sphere) })
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	owners := map[uuid.UUID]int{}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Wolf Totem" {
				owners[c.Controller]++
			}
		}
	})
	if owners[opp.ID] != 1 || owners[me.ID] != 1 {
		t.Errorf("after the Sphere left, Wolf Totems by controller: %v, want one each under their owners", owners)
	}
}

func TestS82BnsB02OutlawsMerrimentMakesOneOfThreeTokensEachUpkeep(t *testing.T) {
	seen := map[string]bool{}
	for key := byte(0); key < 40 && len(seen) < 3; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Outlaws' Merriment", "Enchantment", s82BnsMerrimentOracle, 0, 0)
		advanceToUpkeepOf(t, g, 0)
		life := opp.Life
		passPriorityAroundTable(t, g)
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, playerRef(opp.ID)); err != nil {
				t.Fatalf("Rogue ping target: %v", err)
			}
			passPriorityAroundTable(t, g)
		}
		var made []game.Card
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID && c.IsToken() {
					made = append(made, c)
				}
			}
		})
		if len(made) != 1 {
			t.Fatalf("key %d: %d tokens, want exactly 1", key, len(made))
		}
		c := made[0]
		seen[c.Name] = true
		if !c.HasColor("R") || !c.HasColor("W") || len(c.EffectiveColors()) != 2 {
			t.Errorf("%s is not red and white", c.Name)
		}
		want := map[string][2]int{"Human Warrior": {3, 1}, "Human Cleric": {2, 1}, "Human Rogue": {1, 2}}[c.Name]
		if c.Power != want[0] || c.Toughness != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.Power, c.Toughness, want[0], want[1])
		}
		if !hasEffectiveKeyword(t, g, c.InstanceID, "haste") {
			t.Errorf("%s lacks haste", c.Name)
		}
		switch c.Name {
		case "Human Warrior":
			if !hasEffectiveKeyword(t, g, c.InstanceID, "trample") {
				t.Error("the Warrior lacks trample")
			}
		case "Human Cleric":
			if !hasEffectiveKeyword(t, g, c.InstanceID, "lifelink") {
				t.Error("the Cleric lacks lifelink")
			}
		case "Human Rogue":
			if opp.Life != life-1 {
				t.Errorf("the Rogue's enters ping left the opponent at %d, want %d", opp.Life, life-1)
			}
		}
	}
	if len(seen) != 3 {
		t.Errorf("reached %v; a uniform pick reaches all three", seen)
	}
}
