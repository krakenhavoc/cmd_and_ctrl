package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kamigawa: Neon Dynasty batch 11 (S82, #5493): the cards without a file of
// their own for tests.

const (
	kamiOfCelebrationOracle  = "ecfabd06-c059-4f3d-9e5f-587364033cac"
	akkiBattleSquadOracle    = "5bd81d2f-46a1-4536-92e7-28c7e1892735"
	koseiOracle              = "41f7e725-bdb8-4c3a-adc7-a8184ec7c394"
	concordWithTheKamiOracle = "bec04768-c573-4886-9ca1-2511abc05a14"
	kaimaOracle              = "3d8f9b58-d011-4ed2-99ae-c626243fab26"
	ascendantAcolyteOracle   = "8b9918c3-1135-4c63-991f-f0a38340a33b"
	ironApprenticeOracle     = "cd0c8db0-67ee-483b-8298-455cce35fef7"
	dramatistsPuppetOracle   = "e3473ed5-f2b2-4a41-aed9-ee37caa9d7be"
)

func TestS82NeoB11CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		kamiOfCelebrationOracle:  "Kami of Celebration",
		akkiBattleSquadOracle:    "Akki Battle Squad",
		koseiOracle:              "Kosei, Penitent Warlord",
		concordWithTheKamiOracle: "Concord with the Kami",
		kaimaOracle:              "Kaima, the Fractured Calm",
		ascendantAcolyteOracle:   "Ascendant Acolyte",
		ironApprenticeOracle:     "Iron Apprentice",
		dramatistsPuppetOracle:   "Dramatist's Puppet",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// attachForTest attaches an already-placed permanent to a host.
func attachForTest(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
}

// counterBear is a 2/2 Bear with the given counters, placed as a permanent.
func counterBear(g *game.Game, owner uuid.UUID, counters map[string]int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Counters: counters,
	})
}

// setCounters sets a permanent's counters and refreshes the layers.
func s82SetCounters(g *game.Game, id uuid.UUID, counters map[string]int) {
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, id).Counters = counters
		g.BumpLayerVersionForTest()
		g.RecomputeLayersIfStaleLocked()
	})
}

// --- Kami of Celebration ---

func TestKamiOfCelebrationExilesOncePerModifiedAttackerAndGrowsOnACastFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kami := b12Push(g, me.ID, "Kami of Celebration", "Creature — Spirit", kamiOfCelebrationOracle, 3, 3)
	modified := modifiedBear(g, me.ID)
	plain := bigBear(g, me.ID, 2, false)
	topID := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: topID, Name: "Top Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})

	declareAttack(t, g, opp.ID, modified, plain)
	if n := triggersOnStackFrom(g, kami); n != 1 {
		t.Fatalf("triggers = %d, want 1 (the plain attacker does not trigger it)", n)
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(topID); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is not in exile: %v", z)
	}
	advanceToMain(t, g)
	for g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, topID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, plain)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, plain, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the chosen creature = %d, want 1", n)
	}
}

func TestKamiOfCelebrationIgnoresACastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Kami of Celebration", "Creature — Spirit", kamiOfCelebrationOracle, 3, 3)
	bear := bigBear(g, me.ID, 2, false)
	tdCastCreature(t, g, "Hand Bear", 1, 1)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}

// --- Akki Battle Squad ---

func TestAkkiBattleSquadUntapsModifiedCreaturesAndAddsOneCombatOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Akki Battle Squad", "Creature — Goblin Samurai", akkiBattleSquadOracle, 6, 6)
	modified := modifiedBear(g, me.ID)
	plain := bigBear(g, me.ID, 2, false)

	declareAttack(t, g, opp.ID, modified, plain)
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(modified); c.Tapped {
		t.Error("the modified attacker was not untapped")
	}
	if c, _ := g.LookupCardForEffect(plain); !c.Tapped {
		t.Error("the plain attacker was untapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %v, want an additional combat", next.Step)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(modified, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("the second attack added a third combat; it triggers only once each turn")
	}
}

func TestAkkiBattleSquadIgnoresAPlainAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squad := b12Push(g, me.ID, "Akki Battle Squad", "Creature — Goblin Samurai", akkiBattleSquadOracle, 6, 6)
	plain := bigBear(g, me.ID, 2, false)
	declareAttack(t, g, opp.ID, plain)
	if n := triggersOnStackFrom(g, squad); n != 0 {
		t.Errorf("triggers = %d, want 0", n)
	}
}

// --- Kaima, the Fractured Calm ---

func TestKaimaGoadsEnchantedOpponentCreaturesAndGrowsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kaima := b12Push(g, me.ID, "Kaima, the Fractured Calm", "Legendary Creature — Spirit", kaimaOracle, 3, 3)
	mine := bigBear(g, opp.ID, 2, false)
	theirsAlso := bigBear(g, third.ID, 2, false)
	unenchanted := bigBear(g, opp.ID, 2, false)
	oursEnchanted := bigBear(g, me.ID, 2, false)
	myAura := b12Push(g, me.ID, "My Aura", "Enchantment — Aura", "", 0, 0)
	myAura2 := b12Push(g, me.ID, "My Aura Two", "Enchantment — Aura", "", 0, 0)
	myAura3 := b12Push(g, me.ID, "My Aura Three", "Enchantment — Aura", "", 0, 0)
	theirAura := b12Push(g, opp.ID, "Their Aura", "Enchantment — Aura", "", 0, 0)
	attachForTest(t, g, myAura, mine)
	attachForTest(t, g, myAura2, theirsAlso)
	attachForTest(t, g, myAura3, oursEnchanted)
	attachForTest(t, g, theirAura, unenchanted)

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirsAlso} {
		if c, _ := g.LookupCardForEffect(id); !c.IsGoadedBy(me.ID) {
			t.Errorf("%s was not goaded", c.Name)
		}
	}
	for name, id := range map[string]uuid.UUID{"unenchanted-by-me": unenchanted, "my own": oursEnchanted} {
		if c, _ := g.LookupCardForEffect(id); c.IsGoaded() {
			t.Errorf("%s creature was goaded", name)
		}
	}
	if n := countersOn(g, kaima, game.CounterPlusOne); n != 2 {
		t.Errorf("Kaima counters = %d, want 2", n)
	}
}

// --- Kosei, Penitent Warlord ---

func TestKoseiGainsItsTriggerOnlyWhileEnchantedEquippedAndCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	kosei := b12Push(g, me.ID, "Kosei, Penitent Warlord", "Legendary Creature — Ogre Samurai", koseiOracle, 0, 5)
	aura := b12Push(g, me.ID, "Some Aura", "Enchantment — Aura", "", 0, 0)
	gear := b12Push(g, me.ID, "Some Equipment", "Artifact — Equipment", "", 0, 0)
	attachForTest(t, g, aura, kosei)
	attachForTest(t, g, gear, kosei)
	s82SetCounters(g, kosei, map[string]int{"+1/+1": 3})
	// With a +1/+1 counter Kosei is 3/5 (the counter is layer 7d).
	handBefore := me.Hand.Size()
	oppLife, thirdLife := opp.Life, third.Life
	declareAttack(t, g, opp.ID, kosei)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("cards drawn = %d, want 3", got)
	}
	if opp.Life != oppLife-3 {
		t.Errorf("the damaged opponent's life = %d, want %d", opp.Life, oppLife-3)
	}
	if third.Life != thirdLife-3 {
		t.Errorf("another opponent's life = %d, want %d", third.Life, thirdLife-3)
	}
}

func TestKoseiNeedsAllThreeConditions(t *testing.T) {
	for name, build := range map[string]func(t *testing.T, g *game.Game, me *game.Player, kosei uuid.UUID){
		"no counter": func(t *testing.T, g *game.Game, me *game.Player, kosei uuid.UUID) {
			attachForTest(t, g, b12Push(g, me.ID, "Aura", "Enchantment — Aura", "", 0, 0), kosei)
			attachForTest(t, g, b12Push(g, me.ID, "Gear", "Artifact — Equipment", "", 0, 0), kosei)
		},
		"no equipment": func(t *testing.T, g *game.Game, me *game.Player, kosei uuid.UUID) {
			attachForTest(t, g, b12Push(g, me.ID, "Aura", "Enchantment — Aura", "", 0, 0), kosei)
			s82SetCounters(g, kosei, map[string]int{"+1/+1": 3})
		},
		"no aura": func(t *testing.T, g *game.Game, me *game.Player, kosei uuid.UUID) {
			attachForTest(t, g, b12Push(g, me.ID, "Gear", "Artifact — Equipment", "", 0, 0), kosei)
			s82SetCounters(g, kosei, map[string]int{"+1/+1": 3})
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			kosei := b12Push(g, me.ID, "Kosei, Penitent Warlord", "Legendary Creature — Ogre Samurai", koseiOracle, 0, 5)
			build(t, g, me, kosei)
			handBefore := me.Hand.Size()
			declareAttack(t, g, opp.ID, kosei)
			passPriorityAroundTable(t, g)
			advanceTo(t, g, game.StepCombatDamage)
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - handBefore; got != 0 {
				t.Errorf("cards drawn = %d, want 0", got)
			}
		})
	}
}

// --- Concord with the Kami ---

func concordEndStep(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	mode := latestChoiceOfKind(g, game.PendingChoiceModePick)
	if mode == nil {
		t.Fatal("no mode prompt at the end step")
	}
	return mode
}

func TestConcordWithTheKamiRunsEveryChosenBulletThatHasItsCondition(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Concord with the Kami", "Enchantment", concordWithTheKamiOracle, 0, 0)
	host := modifiedBear(g, me.ID)
	aura := b12Push(g, me.ID, "Some Aura", "Enchantment — Aura", "", 0, 0)
	gear := b12Push(g, me.ID, "Some Equipment", "Artifact — Equipment", "", 0, 0)
	attachForTest(t, g, aura, host)
	attachForTest(t, g, gear, host)
	handBefore := me.Hand.Size()

	mode := concordEndStep(t, g)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0, 1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	pickCard(t, g, me.ID, host)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, host, game.CounterPlusOne); n != 2 {
		t.Errorf("counters = %d, want 2", n)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("cards drawn = %d, want 1", got)
	}
	if n := countTokensControlled(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("Spirits = %d, want 1", n)
	}
}

func TestConcordWithTheKamiDoesNothingWithoutAnEnchantedOrEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Concord with the Kami", "Enchantment", concordWithTheKamiOracle, 0, 0)
	bigBear(g, me.ID, 2, false)
	handBefore := me.Hand.Size()
	mode := concordEndStep(t, g)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore || countTokensControlled(g, me.ID, "Spirit") != 0 {
		t.Error("a bullet with no enchanted or equipped creature did something")
	}
}

// --- Ascendant Acolyte ---

func TestAscendantAcolyteEntersWithACounterPerPlusOnePlusOneCounterAmongOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	counterBear(g, me.ID, map[string]int{"+1/+1": 2})
	counterBear(g, me.ID, map[string]int{"+1/+1": 1, "stun": 5})
	theirs := bigBear(g, g.Seats[1].ID, 2, true) // an opponent's counter does not count
	_ = theirs
	acolyte := rfdCast(t, g, "Ascendant Acolyte", "Creature — Human Monk", ascendantAcolyteOracle, 1, 1)
	if n := countersOn(g, acolyte, game.CounterPlusOne); n != 3 {
		t.Errorf("counters = %d, want 3 (2 + 1)", n)
	}
}

func TestAscendantAcolyteDoublesItsCountersEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	acolyte := b12Push(g, me.ID, "Ascendant Acolyte", "Creature — Human Monk", ascendantAcolyteOracle, 1, 1)
	s82SetCounters(g, acolyte, map[string]int{"+1/+1": 3, "stun": 1})
	for i := 0; i < 400 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, acolyte, game.CounterPlusOne); n != 3 {
		t.Fatalf("an opponent's upkeep doubled them: %d", n)
	}
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, acolyte, game.CounterPlusOne); n != 6 {
		t.Errorf("+1/+1 counters = %d, want 6", n)
	}
	if n := countersOn(g, acolyte, "stun"); n != 1 {
		t.Errorf("stun counters = %d, want 1 (only +1/+1 counters are doubled)", n)
	}
}

// --- Iron Apprentice ---

func TestIronApprenticeEntersWithACounterAndPassesItsCountersOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	friend := bigBear(g, me.ID, 2, false)
	apprentice := rfdCast(t, g, "Iron Apprentice", "Artifact Creature — Construct", ironApprenticeOracle, 0, 0)
	if n := countersOn(g, apprentice, game.CounterPlusOne); n != 1 {
		t.Fatalf("enters with %d counters, want 1", n)
	}
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(apprentice, "stun", 2) })
	killCreature(t, g, me.ID, apprentice)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, friend)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, friend, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the other creature = %d, want 1", n)
	}
	if n := countersOn(g, friend, "stun"); n != 2 {
		t.Errorf("stun counters on the other creature = %d, want 2", n)
	}
}

func TestIronApprenticeWithoutCountersDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	friend := bigBear(g, me.ID, 2, false)
	apprentice := b12Push(g, me.ID, "Iron Apprentice", "Artifact Creature — Construct", ironApprenticeOracle, 0, 1)
	killCreature(t, g, me.ID, apprentice)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		t.Fatal("a target prompt for a creature that had no counters")
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, friend, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}

// --- Dramatist's Puppet ---

func TestDramatistsPuppetAsksOncePerKindOfCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := bigBear(g, g.Seats[1].ID, 2, false)
	s82SetCounters(g, target, map[string]int{"+1/+1": 2, "stun": 1, "charge": 3})
	rfdCast(t, g, "Dramatist's Puppet", "Artifact Creature — Construct", dramatistsPuppetOracle, 2, 4)
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	// Kinds are asked in sorted order: "+1/+1", "charge", "stun".
	answerOptionPick(t, g, me.ID, 0) // +1/+1: put another
	answerOptionPick(t, g, me.ID, 1) // charge: remove one
	answerOptionPick(t, g, me.ID, 1) // stun: remove its only counter
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("asked more than once per kind")
	}
	if n := countersOn(g, target, "+1/+1"); n != 3 {
		t.Errorf("+1/+1 = %d, want 3", n)
	}
	if n := countersOn(g, target, "charge"); n != 2 {
		t.Errorf("charge = %d, want 2", n)
	}
	if n := countersOn(g, target, "stun"); n != 0 {
		t.Errorf("stun = %d, want 0", n)
	}
}

func TestDramatistsPuppetOnAPermanentWithNoCountersDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := bigBear(g, g.Seats[1].ID, 2, false)
	rfdCast(t, g, "Dramatist's Puppet", "Artifact Creature — Construct", dramatistsPuppetOracle, 2, 4)
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("asked about a kind of counter the permanent doesn't have")
	}
}
