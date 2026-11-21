package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mshB11Equipment puts a plain Equipment onto the battlefield attached to host.
func mshB11Equipment(t *testing.T, g *game.Game, owner, host uuid.UUID) uuid.UUID {
	t.Helper()
	eq := effectstest.B12Permanent(g, owner, "Test Blade", "Artifact — Equipment")
	effectstest.BrAttach(t, g, eq, host)
	return eq
}

// --- Whiplash, Vengeful Engineer ----------------------------------------------

func TestMshB11WhiplashEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Whiplash, Vengeful Engineer", "Legendary Creature — Human Artificer Villain", mshB11WhiplashOracle, 2, 2)
	if !effectstest.TappedForTest(t, g, id) {
		t.Error("Whiplash did not enter tapped")
	}
}

func TestMshB11WhiplashDrainsForEachEquipmentAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.B12Push(g, me.ID, "Whiplash, Vengeful Engineer", "Legendary Creature — Human Artificer Villain", mshB11WhiplashOracle, 2, 2)
	mshB11Equipment(t, g, me.ID, w)
	mshB11Equipment(t, g, me.ID, w)
	myLife, theirLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.BrAttack(t, g, w)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != theirLife-2 {
		t.Errorf("opponent life = %d, want %d", got, theirLife-2)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife+2 {
		t.Errorf("controller life = %d, want %d", got, myLife+2)
	}
}

func TestMshB11WhiplashUnequippedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.B12Push(g, me.ID, "Whiplash, Vengeful Engineer", "Legendary Creature — Human Artificer Villain", mshB11WhiplashOracle, 2, 2)
	myLife, theirLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.BrAttack(t, g, w)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, opp.ID) != theirLife || effectstest.LifeOf(g, me.ID) != myLife {
		t.Errorf("unequipped Whiplash changed life: opp %d->%d, me %d->%d", theirLife, effectstest.LifeOf(g, opp.ID), myLife, effectstest.LifeOf(g, me.ID))
	}
}

// --- White Widow, Free Agent ----------------------------------------------------

func TestMshB11WhiteWidowPutsCountersOnUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "White Widow, Free Agent", "Legendary Creature — Human Hero Villain", mshB11WhiteWidowOracle, 2, 3)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mine) != 1 || effectstest.PlusOnes(g, theirs) != 1 {
		t.Errorf("counters: mine %d, theirs %d, want 1 and 1", effectstest.PlusOnes(g, mine), effectstest.PlusOnes(g, theirs))
	}
	if effectstest.PlusOnes(g, other) != 0 {
		t.Errorf("an unchosen creature got %d counters", effectstest.PlusOnes(g, other))
	}
}

func TestMshB11WhiteWidowReturnsAnArtifactOrEnchantmentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := effectstest.PushCatalogGraveyardCard(me, "Old Relic", "Artifact", "", 0, 0)
	ench := effectstest.PushCatalogGraveyardCard(me, "Old Aura", "Enchantment", "", 0, 0)
	bear := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "White Widow, Free Agent", "Legendary Creature — Human Hero Villain", mshB11WhiteWidowOracle, 2, 3)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, art) || !effectstest.HasID(p.PickTargetCards, ench) || effectstest.HasID(p.PickTargetCards, bear) {
		t.Errorf("offered %v: want the artifact and the enchantment, not the creature", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment card did not return to hand")
	}
	if !effectstest.InGraveyard(me, art) || !effectstest.InGraveyard(me, bear) {
		t.Error("an unchosen card left the graveyard")
	}
}

// --- Winter Soldier, Icy Assassin ----------------------------------------------

func TestMshB11WinterSoldierGetsPlusTwoPerEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	w := effectstest.B12Push(g, me.ID, "Winter Soldier, Icy Assassin", "Legendary Creature — Human Assassin Villain", mshB11WinterOracle, 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, w); p != 2 {
		t.Fatalf("unequipped power = %d, want 2", p)
	}
	mshB11Equipment(t, g, me.ID, w)
	mshB11Equipment(t, g, me.ID, w)
	if p, tough := effectstest.EffectivePTOf(t, g, w); p != 6 || tough != 2 {
		t.Errorf("with two Equipment: %d/%d, want 6/2", p, tough)
	}
}

func TestMshB11WinterSoldierReturnsWithFinalityAndMayAttachAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	blade := effectstest.B12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	id := effectstest.PushCatalogGraveyardCard(me, "Winter Soldier, Icy Assassin", "Legendary Creature — Human Assassin Villain", mshB11WinterOracle, 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}{C}{W}{B}", nil); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, blade)
	effectstest.PassPriorityAroundTable(t, g)
	bf := effectstest.FindBattlefieldByName(g, "Winter Soldier, Icy Assassin")
	if bf == uuid.Nil {
		t.Fatal("Winter Soldier did not return")
	}
	if effectstest.CountersOn(g, bf, game.CounterFinality) != 1 {
		t.Errorf("finality counters = %d, want 1", effectstest.CountersOn(g, bf, game.CounterFinality))
	}
	if !effectstest.FindBattlefieldCardForTest(g, blade).IsAttachedTo(bf) {
		t.Error("the Equipment was not attached to him")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bf); p != 4 {
		t.Errorf("power with the Equipment = %d, want 4", p)
	}
	effectstest.KillCreature(t, g, me.ID, bf)
	if effectstest.InGraveyard(me, bf) {
		t.Error("a finality creature died into the graveyard")
	}
}

func TestMshB11WinterSoldierMayDeclineTheAttach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	blade := effectstest.B12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	id := effectstest.PushCatalogGraveyardCard(me, "Winter Soldier, Icy Assassin", "Legendary Creature — Human Assassin Villain", mshB11WinterOracle, 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}{C}{W}{B}", nil); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	bf := effectstest.FindBattlefieldByName(g, "Winter Soldier, Icy Assassin")
	if bf == uuid.Nil || effectstest.FindBattlefieldCardForTest(g, blade).IsAttachedTo(bf) {
		t.Error("declining must leave him on the battlefield with nothing attached")
	}
}

// --- Thunderbolts Conspiracy ----------------------------------------------------

func TestMshB11ThunderboltsReturnsADeadVillainAsAHeroWithFinality(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Thunderbolts Conspiracy", "Enchantment", mshB11ThunderboltsOracle, 0, 0)
	villain := effectstest.B12Creature(g, me.ID, "Test Villain", "Creature — Human Villain", 2, 2)
	effectstest.KillCreature(t, g, me.ID, villain)
	effectstest.PassPriorityAroundTable(t, g)
	bf := effectstest.FindBattlefieldByName(g, "Test Villain")
	if bf == uuid.Nil {
		t.Fatal("the Villain did not return")
	}
	card := effectstest.RingBFCard(t, g, bf)
	if !card.HasSubtype("Hero") || !card.HasSubtype("Villain") {
		t.Errorf("returned types %q, want Hero in addition to Villain", card.TypeLine)
	}
	if effectstest.CountersOn(g, bf, game.CounterFinality) != 1 {
		t.Errorf("finality counters = %d, want 1", effectstest.CountersOn(g, bf, game.CounterFinality))
	}
	effectstest.KillCreature(t, g, me.ID, bf)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Test Villain") != uuid.Nil || effectstest.InGraveyard(me, bf) {
		t.Error("the second death must exile it, not return it or bin it")
	}
}

func TestMshB11ThunderboltsIgnoresOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thunderbolts Conspiracy", "Enchantment", mshB11ThunderboltsOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Plain Bear") != uuid.Nil || effectstest.FindBattlefieldByName(g, "Their Villain") != uuid.Nil {
		t.Error("only a Villain you control returns")
	}
}

// --- Grim Reaper's Scythe -------------------------------------------------------

func TestMshB11ScytheMakesAZombieWhenACreatureCardLeavesYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Grim Reaper's Scythe", "Legendary Artifact", mshB11ScytheOracle, 0, 0)
	a := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	b := effectstest.PushCatalogGraveyardCard(me, "Dead Elk", "Creature — Elk", "", 2, 2)
	land := effectstest.PushCatalogGraveyardCard(me, "Dead Forest", "Land", "", 0, 0)
	g.WithWriteLock(func() {
		if g.ExileCardsForEffect([]uuid.UUID{a, b}) != 2 {
			t.Error("the two creature cards were not both exiled")
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MshB1Tokens(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("Zombies = %d, want 1 for two cards leaving at once", got)
	}
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{land}) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MshB1Tokens(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("Zombies = %d, want still 1 after a land left", got)
	}
}

func TestMshB11ScytheReturnsACreatureWithFinalityAtSorcerySpeedAndThenAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	scythe := effectstest.B12Push(g, me.ID, "Grim Reaper's Scythe", "Legendary Artifact", mshB11ScytheOracle, 0, 0)
	s1 := effectstest.B12Creature(g, me.ID, "Fodder One", "Creature — Bear", 1, 1)
	s2 := effectstest.B12Creature(g, me.ID, "Fodder Two", "Creature — Bear", 1, 1)
	dead := effectstest.PushCatalogGraveyardCard(me, "Fallen Hero", "Creature — Human Hero", "", 3, 3)
	g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, scythe, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
		SacrificeIDs: []uuid.UUID{s1, s2},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bf := effectstest.FindBattlefieldByName(g, "Fallen Hero")
	if bf == uuid.Nil {
		t.Fatal("the creature card did not return")
	}
	if effectstest.CountersOn(g, bf, game.CounterFinality) != 1 {
		t.Errorf("finality counters = %d, want 1", effectstest.CountersOn(g, bf, game.CounterFinality))
	}
	if effectstest.FindBattlefieldByName(g, "Fodder One") != uuid.Nil || effectstest.FindBattlefieldByName(g, "Fodder Two") != uuid.Nil {
		t.Error("the sacrificed creatures are still on the battlefield")
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("Zombies = %d, want 1 (the card left the graveyard)", got)
	}
}

func TestMshB11ScytheIsSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	scythe := effectstest.B12Push(g, me.ID, "Grim Reaper's Scythe", "Legendary Artifact", mshB11ScytheOracle, 0, 0)
	s1 := effectstest.B12Creature(g, me.ID, "Fodder One", "Creature — Bear", 1, 1)
	s2 := effectstest.B12Creature(g, me.ID, "Fodder Two", "Creature — Bear", 1, 1)
	dead := effectstest.PushCatalogGraveyardCard(me, "Fallen Hero", "Creature — Human Hero", "", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{B}")
	err := g.ActivateCatalogAbility(me.ID, scythe, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
		SacrificeIDs: []uuid.UUID{s1, s2},
	})
	if err == nil {
		t.Error("the Scythe activated outside a main phase")
	}
}
