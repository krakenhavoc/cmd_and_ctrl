package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b117_test.go — S82 set GAP, batch 117 (Tales of Middle-earth
// Commander creatures), through the real catalog.

const (
	gapDenethor    = "6669a19e-7944-44c9-ac37-a92016783375"
	gapMerry       = "1b600345-2b89-45bc-98c8-609fdd08a5fd"
	gapTreebeard   = "83860749-1603-46e9-9065-e98001244193"
	gapGandalfWV   = "4563ae4b-6226-4cc9-abde-582d0ab83433"
	gapAragorn     = "be97c400-c690-426d-a39e-4082080a3823"
	gapRadagast    = "50f144f8-29eb-4af7-9a50-eded3b9c83c2"
	gapRiders      = "688a559d-249f-4d03-98e6-581843e7ddc3"
	gapBanquet     = "7f2579ef-0422-4cb1-a0f8-6ad0d942515b"
	gapKnollspine  = "9d834f1b-a4b5-4676-bb6d-2ffc6b3bccd2"
	gapGandalfSF   = "cbad35cd-9026-43ea-8333-94f3b34459e3"
	gapTrapper     = "debbbdb2-2fa8-413f-b394-977212915c65"
	gapKnollspineN = "Knollspine Dragon"
)

func gapB117Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == name {
				n++
			}
		}
	})
	return n
}

func gapB117Counters(g *game.Game, id uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterPlusOne]
			}
		}
	})
	return n
}

func TestGapB117Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gapDenethor: "Denethor, Stone Seer", gapMerry: "Merry, Warden of Isengard",
		gapTreebeard: "Treebeard, Gracious Host", gapGandalfWV: "Gandalf, Westward Voyager",
		gapAragorn: "Aragorn, King of Gondor", gapRadagast: "Radagast, Wizard of Wilds",
		gapRiders: "Riders of Rohan", gapBanquet: "Banquet Guests",
		gapKnollspine: gapKnollspineN, gapGandalfSF: "Gandalf of the Secret Fire",
		gapTrapper: "Mirkwood Trapper",
	} {
		stdRegisteredFull(t, oracle, name)
	}
}

func TestGapB117DenethorCrownsTheTargetPlayerAndDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	id := pushCatalogPermanent(g, me.ID, "Denethor, Stone Seer", "Legendary Creature — Human Noble", gapDenethor, false)
	advanceToMain(t, g)
	// Strict: a pool that can't pay {3}{R} is refused.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID, Slot: 1}}}); err == nil {
		t.Fatal("activated without paying {3}{R}")
	}
	b06AddMana(me, "R", "G", "G", "G")
	before := lifeOf(g, other.ID)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID, Slot: 1}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Denethor was not sacrificed")
	}
	if g.Monarch != opp.ID {
		t.Errorf("monarch = %v, want the targeted player", g.Monarch)
	}
	if got := lifeOf(g, other.ID); got != before-3 {
		t.Errorf("life %d -> %d, want -3", before, got)
	}
}

func TestGapB117DenethorEntersAndScrysTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	if _, err := blbHCastWithMana(t, g, "Denethor, Stone Seer", "Legendary Creature — Human Noble", gapDenethor, "{1}{U}", 1, 3, "U", "U"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a scry-2 prompt, got %+v", c)
	}
}

func TestGapB117TreebeardMakesFoodAndPutsCountersEqualToLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if _, err := blbHCastWithMana(t, g, "Treebeard, Gracious Host", "Legendary Creature — Treefolk", gapTreebeard, "{2}{G}{W}", 0, 5, "G", "W", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := gapB117Count(g, me.ID, "Food"); n != 2 {
		t.Fatalf("%d Food tokens, want 2", n)
	}
	var tree uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treebeard, Gracious Host" {
				tree = c.InstanceID
			}
		}
	})
	if !p1g2HasKeyword(g, tree, "trample") {
		t.Error("Treebeard lacks trample")
	}
	gainLifeNow(t, g, tree, me, 3)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: tree}, true)
	if got := gapB117Counters(g, tree); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
	// A non-Halfling, non-Treefolk is not a legal target.
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	gainLifeNow(t, g, tree, me, 1)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Error("a Bear was accepted as 'target Halfling or Treefolk'")
		}
	}
}

func TestGapB117RidersOfRohanMakesTwoKnightsAndDashes(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if _, err := blbHCastWithMana(t, g, "Riders of Rohan", "Creature — Human Knight", gapRiders, "{3}{R}{W}", 4, 4, "R", "W"); err == nil {
		t.Fatal("cast for {3}{R}{W} with two mana")
	}
	g2 := newCatalogGame(t)
	me2, _, _ := blbB1Seats(g2)
	if _, err := blbHCastWithMana(t, g2, "Riders of Rohan", "Creature — Human Knight", gapRiders, "{3}{R}{W}", 4, 4, "R", "W", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if n := gapB117Count(g2, me2.ID, "Human Knight"); n != 2 {
		t.Fatalf("%d Knight tokens, want 2", n)
	}
	_ = me
	// Dash: {4}{R}{W}, returns to hand at the end step.
	g3 := newCatalogGame(t)
	me3, _, _ := blbB1Seats(g3)
	id := uuid.New()
	me3.Hand.PushTop(game.Card{InstanceID: id, Name: "Riders of Rohan", TypeLine: "Creature — Human Knight", OracleID: gapRiders,
		ManaCost: "{3}{R}{W}", Power: 4, Toughness: 4, Owner: me3.ID, Controller: me3.ID})
	advanceToMain(t, g3)
	b06AddMana(me3, "R", "W", "G", "G", "G", "G")
	if err := g3.CastSpell(me3.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "dash"}); err != nil {
		t.Fatalf("dash cast: %v", err)
	}
	passPriorityAroundTable(t, g3)
	if !p1g2HasKeyword(g3, id, "haste") {
		t.Error("dashed Riders lack haste")
	}
	if n := gapB117Count(g3, me3.ID, "Human Knight"); n != 2 {
		t.Errorf("dashed cast made %d Knights, want 2", n)
	}
	advanceTo(t, g3, game.StepEnd)
	passPriorityAroundTable(t, g3)
	if !me3.Hand.Contains(id) {
		t.Error("the dashed Riders did not return to hand")
	}
}

func TestGapB117BanquetGuestsAffinityCountersAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	// Two Foods make {X}{G}{W} cost {X}{G}{W} - 2 generic when X = 2.
	f1 := blbB1Push(g, me.ID, "Food", "Token Artifact — Food", "", 0, 0)
	f2 := blbB1Push(g, me.ID, "Food", "Token Artifact — Food", "", 0, 0)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Banquet Guests", TypeLine: "Creature — Halfling Citizen", OracleID: gapBanquet,
		ManaCost: "{X}{G}{W}", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	// X = 2 costs {2}{G}{W} before affinity; two Foods leave just {G}{W}.
	b06AddMana(me, "G", "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("cast X=2 with two Foods: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := gapB117Counters(g, id); got != 4 {
		t.Errorf("%d +1/+1 counters, want twice X = 4", got)
	}
	if !p1g2HasKeyword(g, id, "trample") {
		t.Error("lacks trample")
	}
	// {2}, Sacrifice a Food: indestructible.
	b06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{f1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(f1) {
		t.Error("the Food was not sacrificed")
	}
	if !p1g2HasKeyword(g, id, "indestructible") {
		t.Error("Banquet Guests did not gain indestructible")
	}
	_ = f2
	// A pool that can't pay {2} is refused.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{f2}}); err == nil {
		t.Error("activated with no mana")
	}
}

func TestGapB117MerrySoldiersOncePerTurnAndPartnerWith(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	merry, _ := Lookup(gapMerry)
	if len(merry.Triggered) != 2 {
		t.Fatalf("want the partner trigger and the artifact trigger, got %d", len(merry.Triggered))
	}
	blbB1Push(g, me.ID, "Merry, Warden of Isengard", "Legendary Creature — Halfling Advisor", gapMerry, 1, 4)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := gapB117Count(g, me.ID, "Soldier"); n != 1 {
		t.Fatalf("%d Soldiers after the first artifact, want 1", n)
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := gapB117Count(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("%d Soldiers after a second artifact the same turn, want still 1", n)
	}
}

func TestGapB117AragornCrownsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if _, err := blbHCastWithMana(t, g, "Aragorn, King of Gondor", "Legendary Creature — Human Noble", gapAragorn, "{1}{U}{R}{W}", 4, 4, "U", "R", "W", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want the caster", g.Monarch)
	}
}

func TestGapB117AragornAttackStopsOneBlockerAndTheMonarchStopsAll(t *testing.T) {
	for _, monarch := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		id := blbB1Push(g, me.ID, "Aragorn, King of Gondor", "Legendary Creature — Human Noble", gapAragorn, 4, 4)
		for _, kw := range []string{"vigilance", "lifelink"} {
			if !p1g2HasKeyword(g, id, kw) {
				t.Errorf("lacks %s", kw)
			}
		}
		if monarch {
			g.WithWriteLock(func() { g.Monarch = me.ID })
		}
		target := blbB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
		other := blbB1Push(g, opp.ID, "Wall Two", "Creature — Wall", "", 0, 4)
		declareAttack(t, g, opp.ID, id)
		assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: target}, true)
		if auraRestrictions(t, g, target)&game.CantBlock == 0 {
			t.Errorf("monarch=%v: the target can still block", monarch)
		}
		if got := auraRestrictions(t, g, other)&game.CantBlock != 0; got != monarch {
			t.Errorf("monarch=%v: other creature can't block = %v", monarch, got)
		}
	}
}

func TestGapB117KnollspineDiscardsAndDrawsTheDamageDealt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.New(), opp.ID, 4); err != nil {
			t.Fatal(err)
		}
	})
	handBefore := me.Hand.Size()
	if handBefore == 0 {
		t.Fatal("no hand to discard")
	}
	if _, err := blbHCastWithMana(t, g, "Knollspine Dragon", "Creature — Dragon", gapKnollspine, "{5}{R}{R}", 7, 5, "R", "R", "G", "G", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	if got := me.Hand.Size(); got != 4 {
		t.Errorf("hand = %d, want 4 (discarded the hand, drew the 4 damage dealt)", got)
	}
}

func TestGapB117KnollspineDeclinedKeepsTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	before := me.Hand.Size()
	if _, err := blbHCastWithMana(t, g, "Knollspine Dragon", "Creature — Dragon", gapKnollspine, "{5}{R}{R}", 7, 5, "R", "R", "G", "G", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, false)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand = %d, want %d unchanged", got, before)
	}
}

func TestGapB117MirkwoodTrapperShrinksAnAttackerOnYouAndPumpsOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	// The Trapper belongs to a defender: seat 1; seat 0 (active) attacks.
	trapper := blbB1Push(g, opp.ID, "Mirkwood Trapper", "Creature — Elf Scout", gapTrapper, 1, 4)
	atk := blbB1Push(g, me.ID, "Attacker", "Creature — Bear", "", 3, 3)
	declareAttack(t, g, opp.ID, atk)
	assistCardSettle(t, g, opp.ID, &game.TargetRef{Kind: game.TargetCard, ID: atk}, true)
	var p int
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(atk); p = c.PowerForComparison() })
	if p != 1 {
		t.Errorf("attacker power = %d, want 3-2 = 1", p)
	}
	_ = trapper
	_ = other
}

func TestGapB117MirkwoodTrapperLetsTheAttackerPumpWhenNotAttackingYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	// The Trapper is seat 1's; the active player attacks seat 2.
	blbB1Push(g, opp.ID, "Mirkwood Trapper", "Creature — Elf Scout", gapTrapper, 1, 4)
	atk := blbB1Push(g, me.ID, "Attacker", "Creature — Bear", "", 3, 3)
	declareAttack(t, g, other.ID, atk)
	passPriorityAroundTable(t, g)
	// The attacking player (not the Trapper's controller) chooses the creature.
	answerPermanentsPick(t, g, me.ID, atk)
	passPriorityAroundTable(t, g)
	var p int
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(atk); p = c.PowerForComparison() })
	if p != 5 {
		t.Errorf("attacker power = %d, want 3+2 = 5", p)
	}
}

func TestGapB117RadagastGrantsWardAndMakesATokenOnABigSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	spec, _ := Lookup(gapRadagast)
	if len(spec.Triggered) != 3 {
		t.Fatalf("want ward, granted ward and the cast trigger, got %d", len(spec.Triggered))
	}
	blbB1Push(g, me.ID, "Radagast, Wizard of Wilds", "Legendary Creature — Avatar Wizard", gapRadagast, 3, 5)
	// A 5-mana spell: choose the Bird mode.
	big := game.Card{InstanceID: uuid.New(), Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{G}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(big)
	advanceToMain(t, g)
	b06AddMana(me, "G", "G", "G", "G", "G")
	if err := g.CastSpell(me.ID, big.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		passPriorityAroundTable(t, g)
		c = modePickChoiceFor(g, me.ID)
	}
	if c == nil {
		t.Fatal("no mode prompt for the cast trigger")
	}
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)
	if n := gapB117Count(g, me.ID, "Bird"); n != 1 {
		t.Errorf("%d Birds, want 1 (mode 2)", n)
	}
	// A 4-mana spell does not trigger it.
	small := game.Card{InstanceID: uuid.New(), Name: "Small Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{G}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(small)
	b06AddMana(me, "G", "G", "G", "G")
	if err := g.CastSpell(me.ID, small.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := gapB117Count(g, me.ID, "Bird") + gapB117Count(g, me.ID, "Beast"); n != 1 {
		t.Errorf("%d tokens after a 4-mana spell, want still 1", n)
	}
}

func TestGapB117RadagastGrantsWardToBeastsAndBirdsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Radagast, Wizard of Wilds", "Legendary Creature — Avatar Wizard", gapRadagast, 3, 5)
	beast := blbB1Push(g, me.ID, "Beast", "Creature — Beast", "", 3, 3)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	warded := func(victim uuid.UUID) bool {
		castAtWardedCreature(t, g, opp, victim)
		for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
			if stackFullyEmpty(g) {
				return false
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		if hasPayUnlessFor(g, opp.ID) {
			answerPayUnless(t, g, opp.ID, false)
			passPriorityAroundTable(t, g)
			return true
		}
		return false
	}
	if warded(bear) {
		t.Error("a Bear got ward from Radagast")
	}
	passPriorityAroundTable(t, g)
	if !warded(beast) {
		t.Error("a Beast you control did not get ward {1}")
	}
}

func TestGapB117GandalfWestwardVoyagerCopiesOnASharedTypeOtherwiseDraws(t *testing.T) {
	for _, match := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp, other := blbB1Seats(g)
		blbB1Push(g, me.ID, "Gandalf, Westward Voyager", "Legendary Creature — Avatar Wizard", gapGandalfWV, 5, 5)
		// Each opponent's top card: a Sorcery for a match, a Land otherwise.
		typ := "Land"
		if match {
			typ = "Sorcery"
		}
		for _, o := range []*game.Player{opp, other} {
			o.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: typ, Owner: o.ID, Controller: o.ID})
		}
		myHand, oppHand := me.Hand.Size(), opp.Hand.Size()
		life := lifeOf(g, opp.ID)
		big := game.Card{InstanceID: uuid.New(), Name: "Big Burn", TypeLine: "Sorcery", ManaCost: "{4}{R}", OracleID: "", Owner: me.ID, Controller: me.ID}
		me.Hand.PushTop(big)
		advanceToMain(t, g)
		b06AddMana(me, "R", "R", "R", "R", "R")
		if err := g.CastSpell(me.ID, big.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		assistCardSettle(t, g, me.ID, nil, true)
		if match {
			if got := opp.Hand.Size(); got != oppHand+1 {
				t.Errorf("match: opponent hand %d -> %d, want one more", oppHand, got)
			}
			if got := me.Hand.Size(); got != myHand {
				t.Errorf("match: your hand %d -> %d, want unchanged (cast one card, drew none)", myHand, got)
			}
		} else {
			if got := me.Hand.Size(); got != myHand+1 {
				t.Errorf("no match: your hand %d -> %d, want a draw", myHand, got)
			}
			if got := opp.Hand.Size(); got != oppHand {
				t.Errorf("no match: opponent drew (%d -> %d)", oppHand, got)
			}
		}
		_ = life
	}
}

func TestGapB117GandalfOfTheSecretFireSuspendsAnOpponentsTurnSpell(t *testing.T) {
	// Negative first: a spell cast from hand on your own turn goes to the
	// graveyard as usual.
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Gandalf of the Secret Fire", "Legendary Creature — Avatar Wizard", gapGandalfSF, 3, 4)
	own := castCatalogSpell(t, g, "Own Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(own) || g.Exile.Contains(own) {
		t.Fatal("a spell cast on your own turn was exiled")
	}

	// On an opponent's turn the instant is exiled with three time counters.
	g = newCatalogGame(t)
	me = g.Seats[0]
	blbB1Push(g, me.ID, "Gandalf of the Secret Fire", "Legendary Creature — Avatar Wizard", gapGandalfSF, 3, 4)
	advanceToMainOf(t, g, 1)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Instant", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast on the opponent's turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Fatal("the instant went to the graveyard")
	}
	var counters int
	var suspended bool
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				counters, suspended = c.Counters[game.CounterTime], game.CardHasSuspend(c)
			}
		}
	})
	if counters != 3 || !suspended {
		t.Errorf("exiled with %d time counters, suspend=%v; want 3 and suspend", counters, suspended)
	}
}
