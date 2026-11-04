package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b5_cards_test.go — the Jumpstart Selesnya deck request (#4893),
// batch 5. Each test plays the real catalog entry and asserts what its
// text says, and what it refuses.

const (
	jmpB5CradleOracle     = "956250da-532a-4457-8696-73915be56943"
	jmpB5ChargerOracle    = "d3de6b36-c631-4b52-b16b-0749a347d73d"
	jmpB5EmielOracle      = "b11c250c-f191-4c52-ba02-a9176f163447"
	jmpB5FeatOracle       = "2038f505-7b03-4c63-bc44-c5139729dd05"
	jmpB5SentinelsOracle  = "f0e7d147-379e-45b7-bfc4-d926637e060d"
	jmpB5WarlordOracle    = "abc49dc3-03b3-48a4-afe1-5fc435b65d0f"
	jmpB5BeetleOracle     = "693ca077-2b91-4de1-8136-8fe45f8ea5a7"
	jmpB5LenaOracle       = "a8a8407c-d0e3-4311-82ed-1ab1f75159d9"
	jmpB5GiftOracle       = "6052d2d5-6310-42fb-bdec-7b114d74697d"
	jmpB5NaturesWayOracle = "8c04663a-1023-49a5-a0b9-78f24a72fb1e"
)

func TestJmpB5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		jmpB5CradleOracle:     "Cradle of Vitality",
		jmpB5ChargerOracle:    "Daybreak Charger",
		jmpB5EmielOracle:      "Emiel the Blessed",
		jmpB5FeatOracle:       "Feat of Resistance",
		jmpB5SentinelsOracle:  "High Sentinels of Arashin",
		jmpB5WarlordOracle:    "Ironroot Warlord",
		jmpB5BeetleOracle:     "Ironshell Beetle",
		jmpB5LenaOracle:       "Lena, Selfless Champion",
		jmpB5GiftOracle:       "Lifecrafter's Gift",
		jmpB5NaturesWayOracle: "Nature's Way",
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

// --- helpers ---------------------------------------------------------------

// jmpB5Eff is a battlefield permanent's characteristics after the layers.
func jmpB5Eff(t *testing.T, g *game.Game, id uuid.UUID) game.Characteristic {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return effectiveOf(t, g, id)
}

// jmpB5PTWithCounters is a permanent's power and toughness with its
// counters, as it would fight.
func jmpB5PTWithCounters(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.PowerForComparison(), c.CurrentToughness()
}

// jmpB5Plus is a permanent's +1/+1 counters.
func jmpB5Plus(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return b12Counter(t, g, id, game.CounterPlusOne)
}

// jmpB5Card is a target ref to a card.
func jmpB5Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// jmpB5Activate activates a catalog row, paying its cost from the pool.
func jmpB5Activate(g *game.Game, who, source uuid.UUID, idx int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, idx, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

// jmpB5Named counts the permanents a player controls with a name.
func jmpB5Named(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// jmpB5ByName is the first battlefield permanent with a name.
func jmpB5ByName(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s on the battlefield", name)
	return game.Card{}
}

// --- Daybreak Charger ------------------------------------------------------

func TestJmpB5DaybreakChargerGivesTargetCreaturePlusTwoUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Daybreak Charger", "Creature — Unicorn", jmpB5ChargerOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)

	if eff := jmpB5Eff(t, g, theirs); eff.Power != 4 || eff.Toughness != 2 {
		t.Errorf("the target is %d/%d, want 4/2 (any creature, +2/+0)", eff.Power, eff.Toughness)
	}
	endTurn(t, g)
	if eff := jmpB5Eff(t, g, theirs); eff.Power != 2 {
		t.Errorf("the target has power %d after the turn, want 2", eff.Power)
	}
}

func TestJmpB5DaybreakChargerTargetThatLeftGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Daybreak Charger", "Creature — Unicorn", jmpB5ChargerOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
	passPriorityAroundTable(t, g)
	charger := jmpB5ByName(t, g, "Daybreak Charger")
	if eff := jmpB5Eff(t, g, charger.InstanceID); eff.Power != charger.Power {
		t.Errorf("the Charger was pumped when its target left: power %d", eff.Power)
	}
}

// --- Ironshell Beetle ------------------------------------------------------

func TestJmpB5IronshellBeetlePutsACounterOnTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Ironshell Beetle", "Creature — Insect", jmpB5BeetleOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.PickTargetCards {
		offered[id] = true
	}
	if !offered[bear] || !offered[theirs] {
		t.Fatalf("any creature is a legal target; offered %v", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, bear); n != 1 {
		t.Errorf("the Bear has %d counters, want 1", n)
	}
	if n := jmpB5Plus(t, g, theirs); n != 0 {
		t.Errorf("the untargeted creature has %d counters", n)
	}
}

// --- Feat of Resistance ----------------------------------------------------

func TestJmpB5FeatOfResistanceCountersAndGrantsProtectionFromTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Feat of Resistance", "Instant", jmpB5FeatOracle, []game.TargetRef{jmpB5Card(bear)})
	for i := 0; i < 8 && pendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := jmpB5Plus(t, g, bear); n != 1 {
		t.Errorf("the Bear has %d counters before the colour, want 1", n)
	}
	answerColor(t, g, me.ID, "B")
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := battlefieldCard(g, bear); !game.HasKeyword(&c, game.ProtectionFromColor("B")) {
		t.Error("the Bear did not gain protection from black")
	}
	if c, _ := battlefieldCard(g, bear); game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("the Bear has protection from a colour that was not chosen")
	}
	endTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := battlefieldCard(g, bear); game.HasKeyword(&c, game.ProtectionFromColor("B")) {
		t.Error("the protection outlived the turn")
	}
	if n := jmpB5Plus(t, g, bear); n != 1 {
		t.Errorf("the counter is permanent; the Bear has %d", n)
	}
}

func TestJmpB5FeatOfResistanceRefusesACreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Feat of Resistance", "Instant", jmpB5FeatOracle, []game.TargetRef{jmpB5Card(theirs)}); err == nil {
		t.Fatal("an opponent's creature was accepted as the target")
	}
}

// --- Nature's Way ----------------------------------------------------------

func TestJmpB5NaturesWayGrantsVigilanceTrampleAndBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 4)

	castCatalogSpell(t, g, "Nature's Way", "Sorcery", jmpB5NaturesWayOracle, []game.TargetRef{jmpB5Card(mine), jmpB5Card(theirs)})
	passPriorityAroundTable(t, g)

	if c, _ := battlefieldCard(g, theirs); c.DamageMarked != 3 {
		t.Errorf("their creature has %d damage, want 3 (my creature's power)", c.DamageMarked)
	}
	if c, _ := battlefieldCard(g, mine); c.DamageMarked != 0 {
		t.Errorf("my creature took %d damage; the fight is one-sided", c.DamageMarked)
	}
	for _, kw := range []string{"vigilance", "trample"} {
		if !hasKeywordOnBattlefield(t, g, mine, kw) {
			t.Errorf("my creature did not gain %s", kw)
		}
	}
	endTurn(t, g)
	if hasKeywordOnBattlefield(t, g, mine, "trample") {
		t.Error("trample outlived the turn")
	}
}

func TestJmpB5NaturesWayKeywordsStayWhenTheVictimLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 4)

	castCatalogSpell(t, g, "Nature's Way", "Sorcery", jmpB5NaturesWayOracle, []game.TargetRef{jmpB5Card(mine), jmpB5Card(theirs)})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, mine, "vigilance") || !hasKeywordOnBattlefield(t, g, mine, "trample") {
		t.Error("my creature still gains vigilance and trample when the other target is gone")
	}
}

func TestJmpB5NaturesWayRefusesTheWrongSides(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 3, 3)
	mine2 := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 4)

	if err := castCatalogSpellErr(t, g, "Nature's Way", "Sorcery", jmpB5NaturesWayOracle, []game.TargetRef{jmpB5Card(theirs), jmpB5Card(mine)}); err == nil {
		t.Error("the slots were accepted reversed")
	}
	if err := castCatalogSpellErr(t, g, "Nature's Way", "Sorcery", jmpB5NaturesWayOracle, []game.TargetRef{jmpB5Card(mine), jmpB5Card(mine2)}); err == nil {
		t.Error("a creature I control was accepted as the creature I don't control")
	}
}

// --- Lifecrafter's Gift ----------------------------------------------------

func TestJmpB5LifecraftersGiftCountersTheTargetThenEachCounteredCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := b12Creature(g, me.ID, "Target Bear", "Creature — Bear", 2, 2)
	countered := b12Creature(g, me.ID, "Countered Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirCountered := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(countered, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirCountered, game.CounterPlusOne, 1)
	})

	castCatalogSpell(t, g, "Lifecrafter's Gift", "Instant", jmpB5GiftOracle, []game.TargetRef{jmpB5Card(target)})
	passPriorityAroundTable(t, g)

	for _, c := range []struct {
		id   uuid.UUID
		name string
		want int
	}{
		{target, "the target (one, then one more for having one)", 2},
		{countered, "my creature that already had one", 2},
		{plain, "my creature with none", 0},
		{theirCountered, "an opponent's countered creature", 1},
	} {
		if n := jmpB5Plus(t, g, c.id); n != c.want {
			t.Errorf("%s has %d counters, want %d", c.name, n, c.want)
		}
	}
}

func TestJmpB5LifecraftersGiftOnAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	countered := b12Creature(g, me.ID, "Countered Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, game.CounterPlusOne, 1) })

	castCatalogSpell(t, g, "Lifecrafter's Gift", "Instant", jmpB5GiftOracle, []game.TargetRef{jmpB5Card(theirs)})
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, theirs); n != 1 {
		t.Errorf("their creature has %d counters, want 1 (it is not yours)", n)
	}
	if n := jmpB5Plus(t, g, countered); n != 2 {
		t.Errorf("my countered creature has %d counters, want 2", n)
	}
}

func TestJmpB5LifecraftersGiftDoesNothingWhenItsTargetLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := b12Creature(g, me.ID, "Target Bear", "Creature — Bear", 2, 2)
	countered := b12Creature(g, me.ID, "Countered Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, game.CounterPlusOne, 1) })

	castCatalogSpell(t, g, "Lifecrafter's Gift", "Instant", jmpB5GiftOracle, []game.TargetRef{jmpB5Card(target)})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(target) })
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, countered); n != 1 {
		t.Errorf("my countered creature has %d counters; a spell whose only target is gone does nothing", n)
	}
}

// --- High Sentinels of Arashin ---------------------------------------------

func TestJmpB5HighSentinelsGrowsForEachOtherCounteredCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinels := b12Push(g, me.ID, "High Sentinels of Arashin", "Creature — Bird Soldier", jmpB5SentinelsOracle, 3, 4)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if eff := jmpB5Eff(t, g, sentinels); eff.Power != 3 || eff.Toughness != 4 {
		t.Fatalf("Sentinels are %d/%d with no counters around, want 3/4", eff.Power, eff.Toughness)
	}
	if !hasKeywordOnBattlefield(t, g, sentinels, "flying") {
		t.Error("the Sentinels have no flying")
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(a, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(b, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	if eff := jmpB5Eff(t, g, sentinels); eff.Power != 5 || eff.Toughness != 6 {
		t.Errorf("Sentinels are %d/%d with two countered creatures of mine, want 5/6 (per creature, not per counter; an opponent's do not count)", eff.Power, eff.Toughness)
	}
	// Their own counter is a counter, not another creature.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sentinels, game.CounterPlusOne, 1) })
	if p, tough := jmpB5PTWithCounters(t, g, sentinels); p != 6 || tough != 7 {
		t.Errorf("Sentinels are %d/%d with one counter of their own, want 6/7", p, tough)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(b, game.CounterPlusOne, -1) })
	if p, _ := jmpB5PTWithCounters(t, g, sentinels); p != 5 {
		t.Errorf("Sentinels have power %d once Bear B lost its counter, want 5", p)
	}
}

func TestJmpB5HighSentinelsActivationPutsACounterOnTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinels := b12Push(g, me.ID, "High Sentinels of Arashin", "Creature — Bird Soldier", jmpB5SentinelsOracle, 3, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	if err := jmpB5Activate(g, me.ID, sentinels, 0, jmpB5Card(theirs)); err == nil {
		t.Fatal("the activation was accepted with no mana")
	}
	b06AddMana(me, "W", "C", "C", "C")
	if err := jmpB5Activate(g, me.ID, sentinels, 0, jmpB5Card(theirs)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, theirs); n != 1 {
		t.Errorf("the target has %d counters, want 1 (any creature)", n)
	}
	if eff := jmpB5Eff(t, g, sentinels); eff.Power != 3 {
		t.Errorf("an opponent's countered creature grew the Sentinels to power %d", eff.Power)
	}
}

// --- Ironroot Warlord ------------------------------------------------------

func TestJmpB5IronrootWarlordPowerIsCreaturesYouControlAndItMakesSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warlord := b12Push(g, me.ID, "Ironroot Warlord", "Creature — Treefolk Soldier", jmpB5WarlordOracle, 0, 5)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if eff := jmpB5Eff(t, g, warlord); eff.Power != 1 || eff.Toughness != 5 {
		t.Fatalf("the Warlord alone is %d/%d, want 1/5", eff.Power, eff.Toughness)
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if eff := jmpB5Eff(t, g, warlord); eff.Power != 2 {
		t.Errorf("power %d with two creatures, want 2", eff.Power)
	}

	advanceToMain(t, g)
	b06AddMana(me, "G", "W", "C", "C")
	if err := jmpB5Activate(g, me.ID, warlord, 0); err == nil {
		t.Fatal("the activation was accepted for four mana")
	}
	b06AddMana(me, "C")
	if err := jmpB5Activate(g, me.ID, warlord, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := jmpB5Named(g, me.ID, "Soldier"); n != 1 {
		t.Fatalf("%d Soldiers, want 1", n)
	}
	soldier := jmpB5ByName(t, g, "Soldier")
	if eff := jmpB5Eff(t, g, soldier.InstanceID); eff.Power != 1 || eff.Toughness != 1 || !soldier.IsToken() {
		t.Errorf("the Soldier is %d/%d token=%v, want a 1/1 token", eff.Power, eff.Toughness, soldier.IsToken())
	}
	if eff := jmpB5Eff(t, g, warlord); eff.Power != 3 {
		t.Errorf("power %d with three creatures, want 3", eff.Power)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if eff := jmpB5Eff(t, g, warlord); eff.Power != 2 {
		t.Errorf("power %d once the Bear left, want 2", eff.Power)
	}
}

// --- Lena, Selfless Champion -----------------------------------------------

func TestJmpB5LenaMakesASoldierForEachNontokenCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 black Zombie"), 1) })

	castCatalogSpell(t, g, "Lena, Selfless Champion", "Legendary Creature — Human Knight", jmpB5LenaOracle, nil)
	passPriorityAroundTable(t, g)
	if n := jmpB5Named(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("%d Soldiers, want 3 (Lena and two Bears; the token and an opponent's creature do not count)", n)
	}
}

func TestJmpB5LenaSacrificeGivesIndestructibleBelowHerLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lena := b12Push(g, me.ID, "Lena, Selfless Champion", "Legendary Creature — Human Knight", jmpB5LenaOracle, 3, 3)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 4, 4)
	equal := b12Creature(g, me.ID, "Equal", "Creature — Bear", 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Small", "Creature — Bear", 1, 1)
	// Two counters make her 5/5: the last-known power, not the printed 3.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(lena, game.CounterPlusOne, 2) })
	advanceToMain(t, g)

	if err := jmpB5Activate(g, me.ID, lena, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if onBattlefieldByID(g, lena) {
		t.Fatal("Lena is still on the battlefield after paying her sacrifice")
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, small, "indestructible") {
		t.Error("a 4-power creature did not gain indestructible below Lena's 5")
	}
	if hasKeywordOnBattlefield(t, g, equal, "indestructible") {
		t.Error("a creature with power equal to Lena's gained indestructible")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature gained indestructible")
	}
	// Fixed as it resolved: a creature that arrives later gets nothing.
	late := b12Creature(g, me.ID, "Late", "Creature — Bear", 1, 1)
	if hasKeywordOnBattlefield(t, g, late, "indestructible") {
		t.Error("a creature that entered afterwards gained indestructible")
	}
	endTurn(t, g)
	if hasKeywordOnBattlefield(t, g, small, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
}

// --- Cradle of Vitality ----------------------------------------------------

func TestJmpB5CradleOfVitalityPaysForACounterPerLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Cradle of Vitality", "Enchantment", jmpB5CradleOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	b14Gain(t, g, me.ID, 3)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("gaining life does not offer the {1}{W}")
	}
	b06AddMana(me, "W", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, bear); n != 3 {
		t.Errorf("the Bear has %d counters after gaining 3 and paying, want 3", n)
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {1}{W}", poolColors(me))
	}
}

func TestJmpB5CradleOfVitalityDeclineAndUnpayableDoNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Cradle of Vitality", "Enchantment", jmpB5CradleOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	b14Gain(t, g, me.ID, 2)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "W", "C")
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, bear); n != 0 {
		t.Errorf("declining put %d counters", n)
	}
	if len(poolColors(me)) != 2 {
		t.Errorf("declining spent mana: pool = %v", poolColors(me))
	}

	// {1}{W} paid with two colourless is no payment.
	me.ManaPool = game.ManaPool{}
	b06AddMana(me, "C", "C")
	b14Gain(t, g, me.ID, 2)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, true)
	}
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, bear); n != 0 {
		t.Errorf("an unpayable {1}{W} put %d counters", n)
	}

	// An opponent's lifegain is not yours.
	b14Gain(t, g, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's lifegain triggered the Cradle")
	}
}

// --- Emiel the Blessed -----------------------------------------------------

func jmpB5Emiel(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	emiel := b12Push(g, me.ID, "Emiel the Blessed", "Legendary Creature — Unicorn", jmpB5EmielOracle, 4, 4)
	advanceToMain(t, g)
	return emiel
}

// jmpB5EnterCreature casts a vanilla creature for the active player and
// resolves it; Emiel's trigger then waits on its pay prompt.
func jmpB5EnterCreature(t *testing.T, g *game.Game, name, typeLine string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, typeLine, "", nil)
	passPriorityAroundTable(t, g)
	return id
}

func TestJmpB5EmielPaysHybridForACounterOnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emiel := jmpB5Emiel(t, g, me)

	bear := jmpB5EnterCreature(t, g, "Bear", "Creature — Bear")
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("another creature entering does not offer the {G/W}")
	}
	b06AddMana(me, "G")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, bear); n != 1 {
		t.Errorf("the Bear has %d counters, want 1", n)
	}
	if n := jmpB5Plus(t, g, emiel); n != 0 {
		t.Errorf("Emiel has %d counters", n)
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("{G} paid {G/W}; pool = %v", poolColors(me))
	}
}

func TestJmpB5EmielPutsTwoOnAUnicorn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jmpB5Emiel(t, g, me)

	unicorn := jmpB5EnterCreature(t, g, "Pony", "Creature — Unicorn")
	b06AddMana(me, "W")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, unicorn); n != 2 {
		t.Errorf("the Unicorn has %d counters, want 2", n)
	}
}

func TestJmpB5EmielDeclineWrongColourAndOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jmpB5Emiel(t, g, me)

	bear := jmpB5EnterCreature(t, g, "Bear", "Creature — Bear")
	b06AddMana(me, "W")
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, bear); n != 0 {
		t.Errorf("declining put %d counters", n)
	}

	// {U} cannot pay {G/W}.
	me.ManaPool = game.ManaPool{}
	b06AddMana(me, "U")
	bear2 := jmpB5EnterCreature(t, g, "Bear 2", "Creature — Bear")
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, true)
	}
	passPriorityAroundTable(t, g)
	if n := jmpB5Plus(t, g, bear2); n != 0 {
		t.Errorf("{U} paid {G/W}: %d counters", n)
	}

	// An opponent's creature entering is not a creature you control.
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1) })
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's creature entering triggered Emiel")
	}
}

func TestJmpB5EmielCreatureThatLeftBeforeTheAnswerGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jmpB5Emiel(t, g, me)

	bear := jmpB5EnterCreature(t, g, "Bear", "Creature — Bear")
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no {G/W} offer")
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
	b06AddMana(me, "W")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, bear) {
		t.Fatal("the Bear came back")
	}
	if c, ok := g.LookupCardForEffect(bear); ok && c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("the departed Bear carries %d counters", c.Counters[game.CounterPlusOne])
	}
}

func TestJmpB5EmielBlinksAnotherCreatureYouControlUnderItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	emiel := jmpB5Emiel(t, g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	borrowed := b12Creature(g, opp.ID, "Borrowed Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == borrowed {
				g.Battlefield.Cards[i].Controller = me.ID
			}
		}
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })

	b06AddMana(me, "C", "C", "C")
	if err := jmpB5Activate(g, me.ID, emiel, 0, jmpB5Card(emiel)); err == nil {
		t.Error("Emiel accepted itself as \"another\" target")
	}
	if err := jmpB5Activate(g, me.ID, emiel, 0, jmpB5Card(theirs)); err == nil {
		t.Error("Emiel accepted a creature I don't control")
	}
	if err := jmpB5Activate(g, me.ID, emiel, 0, jmpB5Card(bear)); err != nil {
		t.Fatalf("activate on my Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The returned Bear is a new object, which triggers Emiel's second ability.
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
	} else {
		t.Error("the returned Bear did not trigger Emiel")
	}
	back := jmpB5ByName(t, g, "Bear")
	if back.Controller != me.ID || back.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("the Bear came back controller=%v counters=%d, want mine with no counters", back.Controller, back.Counters[game.CounterPlusOne])
	}

	b06AddMana(me, "C", "C", "C")
	if err := jmpB5Activate(g, me.ID, emiel, 0, jmpB5Card(borrowed)); err != nil {
		t.Fatalf("activate on the borrowed Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the borrowed Bear returned under my control and triggered Emiel")
	}
	if home := jmpB5ByName(t, g, "Borrowed Bear"); home.Controller != opp.ID {
		t.Errorf("the borrowed Bear returned under %v, want its owner", home.Controller)
	}
}
