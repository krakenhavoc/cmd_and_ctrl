package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b11_cards_test.go — Marvel Super Heroes, half B, batch 11 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB11WhiplashOracle     = "00ed7aca-b3c1-4d48-acd5-b1596daeb4c3"
	mshB11WhiteWidowOracle   = "4f0c055f-f614-44b0-9374-c53541f89185"
	mshB11WinterOracle       = "6f4c94b6-f829-4646-a574-083997b78dfb"
	mshB11ThunderboltsOracle = "d90a5213-3363-4cd1-97d6-4fc057f78b2a"
	mshB11ScytheOracle       = "91b1db8c-4bce-43d1-94ab-14b1435686c7"
)

func TestMshB11EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB11WhiplashOracle:     "Whiplash, Vengeful Engineer",
		mshB11WhiteWidowOracle:   "White Widow, Free Agent",
		mshB11WinterOracle:       "Winter Soldier, Icy Assassin",
		mshB11ThunderboltsOracle: "Thunderbolts Conspiracy",
		mshB11ScytheOracle:       "Grim Reaper's Scythe",
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

// mshB11Equipment puts a plain Equipment onto the battlefield attached to host.
func mshB11Equipment(t *testing.T, g *game.Game, owner, host uuid.UUID) uuid.UUID {
	t.Helper()
	eq := b12Permanent(g, owner, "Test Blade", "Artifact — Equipment")
	brAttach(t, g, eq, host)
	return eq
}

// --- Whiplash, Vengeful Engineer ----------------------------------------------

func TestMshB11WhiplashEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Whiplash, Vengeful Engineer", "Legendary Creature — Human Artificer Villain", mshB11WhiplashOracle, 2, 2)
	if !tappedForTest(t, g, id) {
		t.Error("Whiplash did not enter tapped")
	}
}

func TestMshB11WhiplashDrainsForEachEquipmentAttached(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := b12Push(g, me.ID, "Whiplash, Vengeful Engineer", "Legendary Creature — Human Artificer Villain", mshB11WhiplashOracle, 2, 2)
	mshB11Equipment(t, g, me.ID, w)
	mshB11Equipment(t, g, me.ID, w)
	myLife, theirLife := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	brAttack(t, g, w)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != theirLife-2 {
		t.Errorf("opponent life = %d, want %d", got, theirLife-2)
	}
	if got := lifeOf(g, me.ID); got != myLife+2 {
		t.Errorf("controller life = %d, want %d", got, myLife+2)
	}
}

func TestMshB11WhiplashUnequippedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := b12Push(g, me.ID, "Whiplash, Vengeful Engineer", "Legendary Creature — Human Artificer Villain", mshB11WhiplashOracle, 2, 2)
	myLife, theirLife := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	brAttack(t, g, w)
	passPriorityAroundTable(t, g)
	if lifeOf(g, opp.ID) != theirLife || lifeOf(g, me.ID) != myLife {
		t.Errorf("unequipped Whiplash changed life: opp %d->%d, me %d->%d", theirLife, lifeOf(g, opp.ID), myLife, lifeOf(g, me.ID))
	}
}

// --- White Widow, Free Agent ----------------------------------------------------

func TestMshB11WhiteWidowPutsCountersOnUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "White Widow, Free Agent", "Legendary Creature — Human Hero Villain", mshB11WhiteWidowOracle, 2, 3)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	answerMode(t, g, c, me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, mine) != 1 || plusOnes(g, theirs) != 1 {
		t.Errorf("counters: mine %d, theirs %d, want 1 and 1", plusOnes(g, mine), plusOnes(g, theirs))
	}
	if plusOnes(g, other) != 0 {
		t.Errorf("an unchosen creature got %d counters", plusOnes(g, other))
	}
}

func TestMshB11WhiteWidowReturnsAnArtifactOrEnchantmentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	art := pushCatalogGraveyardCard(me, "Old Relic", "Artifact", "", 0, 0)
	ench := pushCatalogGraveyardCard(me, "Old Aura", "Enchantment", "", 0, 0)
	bear := pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	castHoldCreature(t, g, "White Widow, Free Agent", "Legendary Creature — Human Hero Villain", mshB11WhiteWidowOracle, 2, 3)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	answerMode(t, g, c, me.ID, 1)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, art) || !hasID(p.PickTargetCards, ench) || hasID(p.PickTargetCards, bear) {
		t.Errorf("offered %v: want the artifact and the enchantment, not the creature", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, ench)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment card did not return to hand")
	}
	if !inGraveyard(me, art) || !inGraveyard(me, bear) {
		t.Error("an unchosen card left the graveyard")
	}
}

// --- Winter Soldier, Icy Assassin ----------------------------------------------

func TestMshB11WinterSoldierGetsPlusTwoPerEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := b12Push(g, me.ID, "Winter Soldier, Icy Assassin", "Legendary Creature — Human Assassin Villain", mshB11WinterOracle, 2, 2)
	if p, _ := effectivePTOf(t, g, w); p != 2 {
		t.Fatalf("unequipped power = %d, want 2", p)
	}
	mshB11Equipment(t, g, me.ID, w)
	mshB11Equipment(t, g, me.ID, w)
	if p, tough := effectivePTOf(t, g, w); p != 6 || tough != 2 {
		t.Errorf("with two Equipment: %d/%d, want 6/2", p, tough)
	}
}

func TestMshB11WinterSoldierReturnsWithFinalityAndMayAttachAnEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	blade := b12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	id := pushCatalogGraveyardCard(me, "Winter Soldier, Icy Assassin", "Legendary Creature — Human Assassin Villain", mshB11WinterOracle, 2, 2)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}{C}{C}{W}{B}", nil); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, blade)
	passPriorityAroundTable(t, g)
	bf := findBattlefieldByName(g, "Winter Soldier, Icy Assassin")
	if bf == uuid.Nil {
		t.Fatal("Winter Soldier did not return")
	}
	if countersOn(g, bf, game.CounterFinality) != 1 {
		t.Errorf("finality counters = %d, want 1", countersOn(g, bf, game.CounterFinality))
	}
	if !findBattlefieldCardForTest(g, blade).IsAttachedTo(bf) {
		t.Error("the Equipment was not attached to him")
	}
	if p, _ := effectivePTOf(t, g, bf); p != 4 {
		t.Errorf("power with the Equipment = %d, want 4", p)
	}
	killCreature(t, g, me.ID, bf)
	if inGraveyard(me, bf) {
		t.Error("a finality creature died into the graveyard")
	}
}

func TestMshB11WinterSoldierMayDeclineTheAttach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	blade := b12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	id := pushCatalogGraveyardCard(me, "Winter Soldier, Icy Assassin", "Legendary Creature — Human Assassin Villain", mshB11WinterOracle, 2, 2)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}{C}{C}{W}{B}", nil); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	bf := findBattlefieldByName(g, "Winter Soldier, Icy Assassin")
	if bf == uuid.Nil || findBattlefieldCardForTest(g, blade).IsAttachedTo(bf) {
		t.Error("declining must leave him on the battlefield with nothing attached")
	}
}

// --- Thunderbolts Conspiracy ----------------------------------------------------

func TestMshB11ThunderboltsReturnsADeadVillainAsAHeroWithFinality(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Thunderbolts Conspiracy", "Enchantment", mshB11ThunderboltsOracle, 0, 0)
	villain := b12Creature(g, me.ID, "Test Villain", "Creature — Human Villain", 2, 2)
	killCreature(t, g, me.ID, villain)
	passPriorityAroundTable(t, g)
	bf := findBattlefieldByName(g, "Test Villain")
	if bf == uuid.Nil {
		t.Fatal("the Villain did not return")
	}
	card := ringBFCard(t, g, bf)
	if !card.HasSubtype("Hero") || !card.HasSubtype("Villain") {
		t.Errorf("returned types %q, want Hero in addition to Villain", card.TypeLine)
	}
	if countersOn(g, bf, game.CounterFinality) != 1 {
		t.Errorf("finality counters = %d, want 1", countersOn(g, bf, game.CounterFinality))
	}
	killCreature(t, g, me.ID, bf)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Test Villain") != uuid.Nil || inGraveyard(me, bf) {
		t.Error("the second death must exile it, not return it or bin it")
	}
}

func TestMshB11ThunderboltsIgnoresOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Thunderbolts Conspiracy", "Enchantment", mshB11ThunderboltsOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	killCreature(t, g, me.ID, bear)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Plain Bear") != uuid.Nil || findBattlefieldByName(g, "Their Villain") != uuid.Nil {
		t.Error("only a Villain you control returns")
	}
}

// --- Grim Reaper's Scythe -------------------------------------------------------

func TestMshB11ScytheMakesAZombieWhenACreatureCardLeavesYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Grim Reaper's Scythe", "Legendary Artifact", mshB11ScytheOracle, 0, 0)
	a := pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	b := pushCatalogGraveyardCard(me, "Dead Elk", "Creature — Elk", "", 2, 2)
	land := pushCatalogGraveyardCard(me, "Dead Forest", "Land", "", 0, 0)
	g.WithWriteLock(func() {
		if g.ExileCardsForEffect([]uuid.UUID{a, b}) != 2 {
			t.Error("the two creature cards were not both exiled")
		}
	})
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("Zombies = %d, want 1 for two cards leaving at once", got)
	}
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{land}) })
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("Zombies = %d, want still 1 after a land left", got)
	}
}

func TestMshB11ScytheReturnsACreatureWithFinalityAtSorcerySpeedAndThenAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	scythe := b12Push(g, me.ID, "Grim Reaper's Scythe", "Legendary Artifact", mshB11ScytheOracle, 0, 0)
	s1 := b12Creature(g, me.ID, "Fodder One", "Creature — Bear", 1, 1)
	s2 := b12Creature(g, me.ID, "Fodder Two", "Creature — Bear", 1, 1)
	dead := pushCatalogGraveyardCard(me, "Fallen Hero", "Creature — Human Hero", "", 3, 3)
	g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, scythe, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
		SacrificeIDs: []uuid.UUID{s1, s2},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	bf := findBattlefieldByName(g, "Fallen Hero")
	if bf == uuid.Nil {
		t.Fatal("the creature card did not return")
	}
	if countersOn(g, bf, game.CounterFinality) != 1 {
		t.Errorf("finality counters = %d, want 1", countersOn(g, bf, game.CounterFinality))
	}
	if findBattlefieldByName(g, "Fodder One") != uuid.Nil || findBattlefieldByName(g, "Fodder Two") != uuid.Nil {
		t.Error("the sacrificed creatures are still on the battlefield")
	}
	if got := mshB1Tokens(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("Zombies = %d, want 1 (the card left the graveyard)", got)
	}
}

func TestMshB11ScytheIsSorcerySpeedOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	scythe := b12Push(g, me.ID, "Grim Reaper's Scythe", "Legendary Artifact", mshB11ScytheOracle, 0, 0)
	s1 := b12Creature(g, me.ID, "Fodder One", "Creature — Bear", 1, 1)
	s2 := b12Creature(g, me.ID, "Fodder Two", "Creature — Bear", 1, 1)
	dead := pushCatalogGraveyardCard(me, "Fallen Hero", "Creature — Human Hero", "", 3, 3)
	advanceTo(t, g, game.StepBeginCombat)
	g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{B}")
	err := g.ActivateCatalogAbility(me.ID, scythe, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
		SacrificeIDs: []uuid.UUID{s1, s2},
	})
	if err == nil {
		t.Error("the Scythe activated outside a main phase")
	}
}
