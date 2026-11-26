package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b209_test.go — Dominaria counters, batch 209 (#5530): each test plays the
// real catalog entry through the real cast, activation, trigger and turn paths.

const (
	oldB209HallarOracle      = "04559759-1afd-4785-badc-c3416041f016"
	oldB209KazarovOracle     = "3067c653-ac29-4f1d-822f-7e2e5bb38a62"
	oldB209WarcallerOracle   = "b44d3fe8-b1e2-4abb-ab82-7a516a1db1e5"
	oldB209ArchaeoOracle     = "b9d89f7c-a362-48dc-a335-f0e556e83f35"
	oldB209DarigaazOracle    = "c2f42c67-ff88-4248-bfa5-79a18c6473d5"
	oldB209BombardmentOracle = "fc1eb1c9-b22b-41a1-8b13-94a0e758db0f"
)

func TestOldB209CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB209HallarOracle:      "Hallar, the Firefletcher",
		oldB209KazarovOracle:     "Kazarov, Sengir Pureblood",
		oldB209WarcallerOracle:   "Keldon Warcaller",
		oldB209ArchaeoOracle:     "Daring Archaeologist",
		oldB209DarigaazOracle:    "Darigaaz Reincarnated",
		oldB209BombardmentOracle: "Haphazard Bombardment",
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

// --- Hallar, the Firefletcher ---

func TestOldB209HallarGrowsAndBurnsOnlyForYourKickedSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hallar := effectstest.B12Push(g, me.ID, "Hallar, the Firefletcher", "Legendary Creature — Elf Archer", oldB209HallarOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	// An unkicked spell does nothing to it.
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant", effectstest.BurstLightningOracle, target, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, hallar) != 0 {
		t.Error("an unkicked spell triggered Hallar")
	}
	// Two kicked spells: one counter and one damage, then two counters and two damage.
	for want := 1; want <= 2; want++ {
		life := [2]int{g.Seats[2].Life, g.Seats[3].Life}
		before := opp.Life
		if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant", effectstest.BurstLightningOracle, target, []int{0}, nil); err != nil {
			t.Fatalf("kicked cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.PlusOnes(g, hallar); got != want {
			t.Errorf("kicked spell %d: %d counters, want %d", want, got, want)
		}
		// The Bolt dealt 4 to the targeted opponent, and Hallar its counters to each opponent.
		if opp.Life != before-4-want {
			t.Errorf("kicked spell %d: the targeted opponent lost %d, want %d", want, before-opp.Life, 4+want)
		}
		if g.Seats[2].Life != life[0]-want || g.Seats[3].Life != life[1]-want {
			t.Errorf("kicked spell %d: the other opponents did not each take %d", want, want)
		}
	}
}

func TestOldB209HallarIgnoresAnOpponentsKickedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hallar := effectstest.B12Push(g, me.ID, "Hallar, the Firefletcher", "Legendary Creature — Elf Archer", oldB209HallarOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Burst Lightning", TypeLine: "Instant", OracleID: effectstest.BurstLightningOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, hallar) != 0 {
		t.Error("an opponent's kicked spell grew Hallar")
	}
}

// --- Kazarov, Sengir Pureblood ---

func TestOldB209KazarovGrowsWhenAnOpponentsCreatureIsDamagedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kazarov := effectstest.B12Push(g, me.ID, "Kazarov, Sengir Pureblood", "Legendary Creature — Vampire", oldB209KazarovOracle, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	mine := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(pinger, mine, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, kazarov) != 0 {
		t.Error("damage to my own creature grew Kazarov")
	}
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error { return g.DealDamageToCreatureForEffect(pinger, theirs, 2) })
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, kazarov); got != 1 {
		t.Errorf("Kazarov has %d counters after an opponent's creature was dealt damage, want 1", got)
	}
}

func TestOldB209KazarovDealsTwoToATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kazarov := effectstest.PushCatalogPermanent(g, me.ID, "Kazarov, Sengir Pureblood", "Legendary Creature — Vampire", oldB209KazarovOracle, false)
	prey := effectstest.B12Creature(g, opp.ID, "Prey", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, kazarov, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: prey}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(prey) {
		t.Error("the creature survived 2 damage")
	}
	// The damage was dealt to an opponent's creature, so Kazarov grew too.
	if effectstest.PlusOnes(g, kazarov) < 1 {
		t.Error("Kazarov did not grow from its own damage to an opponent's creature")
	}
	effectstest.B06AddMana(me, "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, kazarov, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a player was a legal target for the creature-only ability")
	}
}

// --- Keldon Warcaller ---

func TestOldB209KeldonWarcallerAddsALoreCounterToYourSaga(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warcaller := effectstest.B12Push(g, me.ID, "Keldon Warcaller", "Creature — Human Warrior", oldB209WarcallerOracle, 2, 2)
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "History of Benalia", TypeLine: "Enchantment — Saga", OracleID: effectstest.HistoryOfBenaliaOracle,
		Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "History of Benalia", TypeLine: "Enchantment — Saga", OracleID: effectstest.HistoryOfBenaliaOracle,
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLore: 1},
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(warcaller, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		c := effectstest.LatestPickTarget(g, me.ID)
		if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Error("an opponent's Saga was a legal target")
		}
		effectstest.AnswerPickTarget(t, g, saga)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Errorf("my Saga has %d lore counters, want 2", got)
	}
	if got := effectstest.LoreCountersOn(g, theirs); got != 1 {
		t.Errorf("their Saga has %d lore counters, want 1", got)
	}
	// My precombat main added the first lore counter (chapter I made a Knight);
	// the Warcaller's counter is the second (chapter II made another).
	if n := effectstest.BattlefieldNamed(g, me.ID, "Knight"); n != 2 {
		t.Errorf("%d Knight tokens, want 2: the Warcaller's lore counter did not trigger chapter II", n)
	}
}

// --- Daring Archaeologist ---

func TestOldB209DaringArchaeologistReturnsAnArtifactAndGrowsOnHistoricSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := effectstest.GyCard(me, "Old Relic", "Artifact", "", "")
	bear := effectstest.GyCard(me, "Old Bear", "Creature — Bear", "", "")
	arch := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Daring Archaeologist", TypeLine: "Creature — Human Artificer", OracleID: oldB209ArchaeoOracle, Power: 3, Toughness: 3})
	effectstest.AnswerPickTarget(t, g, art)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(art) || me.Hand.Contains(bear) {
		t.Error("the artifact card was not returned, or a creature card was")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, arch) != 0 {
		t.Error("a nonhistoric spell grew the Archaeologist")
	}
	effectstest.CastCatalogSpell(t, g, "Old Relic", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Legend", "Legendary Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Old Saga", "Enchantment — Saga", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, arch); got != 3 {
		t.Errorf("%d counters after an artifact, a legendary and a Saga, want 3", got)
	}
}

func TestOldB209DaringArchaeologistMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := effectstest.GyCard(me, "Old Relic", "Artifact", "", "")
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Daring Archaeologist", TypeLine: "Creature — Human Artificer", OracleID: oldB209ArchaeoOracle, Power: 3, Toughness: 3})
	effectstest.AnswerPickTarget(t, g, art)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(art) {
		t.Error("the artifact was returned after declining")
	}
}

// --- Darigaaz Reincarnated ---

func TestOldB209DarigaazIsExiledWithThreeEggCountersAndReturnsOnTheThirdUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	darigaaz := effectstest.B12Push(g, me.ID, "Darigaaz Reincarnated", "Legendary Creature — Dragon", oldB209DarigaazOracle, 7, 7)
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Test", 9, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(pinger, darigaaz, 7) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(darigaaz) {
		t.Fatal("Darigaaz stayed on the battlefield after lethal damage")
	}
	if me.Graveyard.Contains(darigaaz) {
		t.Fatal("Darigaaz was put into the graveyard, not exiled")
	}
	c, ok := effectstest.CardInZone(g.Exile, darigaaz)
	if !ok || c.Counters["egg"] != 3 {
		t.Fatalf("Darigaaz is in exile = %v with %d egg counters, want 3", ok, c.Counters["egg"])
	}
	for want := 2; want >= 0; want-- {
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if want > 0 {
			c, ok := effectstest.CardInZone(g.Exile, darigaaz)
			if !ok || c.Counters["egg"] != want {
				t.Fatalf("after an upkeep Darigaaz has %d egg counters (in exile %v), want %d", c.Counters["egg"], ok, want)
			}
		}
	}
	back := effectstest.BattlefieldNamed(g, me.ID, "Darigaaz Reincarnated")
	if back != 1 {
		t.Errorf("%d Darigaaz on the battlefield after the third upkeep, want 1", back)
	}
}

func TestOldB209DarigaazIsNotCountedOnAnOpponentsUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	darigaaz := effectstest.B12Push(g, me.ID, "Darigaaz Reincarnated", "Legendary Creature — Dragon", oldB209DarigaazOracle, 7, 7)
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Test", 9, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(pinger, darigaaz, 7) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.CardInZone(g.Exile, darigaaz); !ok || c.Counters["egg"] != 3 {
		t.Error("an opponent's upkeep removed an egg counter")
	}
}

// --- Haphazard Bombardment ---

func TestOldB209HaphazardBombardmentAimsAtFourAndDestroysOneAtRandomEachEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var theirs []uuid.UUID
	for range 5 {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1))
	}
	aura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 1, 1)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Haphazard Bombardment", TypeLine: "Enchantment", OracleID: oldB209BombardmentOracle})
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.SlChoose(t, g, me.ID, theirs[0], theirs[1], theirs[2], theirs[3])
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range theirs {
		c, _ := effectstest.BattlefieldCard(g, id)
		want := 0
		if i < 4 {
			want = 1
		}
		if c.Counters["aim"] != want {
			t.Errorf("creature %d has %d aim counters, want %d", i, c.Counters["aim"], want)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, aura); c.Counters["aim"] != 0 {
		t.Error("an enchantment got an aim counter")
	}
	if c, _ := effectstest.BattlefieldCard(g, mine); c.Counters["aim"] != 0 {
		t.Error("my own permanent got an aim counter")
	}
	// Each of my end steps destroys one aimed permanent while two or more remain.
	left := func() int {
		n := 0
		for _, id := range theirs[:4] {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	for want := 3; want >= 1; want-- {
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if got := left(); got != want {
			t.Fatalf("%d aimed creatures left after an end step, want %d", got, want)
		}
		effectstest.AdvanceToMainOf(t, g, 0)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := left(); got != 1 {
		t.Errorf("%d aimed creatures left, want 1: a lone aimed permanent was destroyed", got)
	}
	if !g.Battlefield.Contains(theirs[4]) {
		t.Error("an unaimed creature was destroyed")
	}
}

func TestOldB209HaphazardBombardmentTakesEveryEligiblePermanentWhenThereAreFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, g.Seats[2].ID, "Bear B", "Creature — Bear", 1, 1)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Haphazard Bombardment", TypeLine: "Enchantment", OracleID: oldB209BombardmentOracle})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Counters["aim"] != 1 {
			t.Error("an eligible permanent did not get an aim counter")
		}
	}
}
