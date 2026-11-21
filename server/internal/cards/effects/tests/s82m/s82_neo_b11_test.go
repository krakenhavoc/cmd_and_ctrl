package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// --- Kami of Celebration ---

func TestKamiOfCelebrationExilesOncePerModifiedAttackerAndGrowsOnACastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kami := effectstest.B12Push(g, me.ID, "Kami of Celebration", "Creature — Spirit", kamiOfCelebrationOracle, 3, 3)
	modified := effectstest.ModifiedBear(g, me.ID)
	plain := effectstest.BigBear(g, me.ID, 2, false)
	topID := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: topID, Name: "Top Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})

	effectstest.DeclareAttack(t, g, opp.ID, modified, plain)
	if n := effectstest.TriggersOnStackFrom(g, kami); n != 1 {
		t.Fatalf("triggers = %d, want 1 (the plain attacker does not trigger it)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(topID); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is not in exile: %v", z)
	}
	effectstest.AdvanceToMain(t, g)
	for g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, topID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, plain)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, plain, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the chosen creature = %d, want 1", n)
	}
}

func TestKamiOfCelebrationIgnoresACastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Kami of Celebration", "Creature — Spirit", kamiOfCelebrationOracle, 3, 3)
	bear := effectstest.BigBear(g, me.ID, 2, false)
	effectstest.TdCastCreature(t, g, "Hand Bear", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}

// --- Akki Battle Squad ---

func TestAkkiBattleSquadUntapsModifiedCreaturesAndAddsOneCombatOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Akki Battle Squad", "Creature — Goblin Samurai", akkiBattleSquadOracle, 6, 6)
	modified := effectstest.ModifiedBear(g, me.ID)
	plain := effectstest.BigBear(g, me.ID, 2, false)

	effectstest.DeclareAttack(t, g, opp.ID, modified, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(modified); c.Tapped {
		t.Error("the modified attacker was not untapped")
	}
	if c, _ := g.LookupCardForEffect(plain); !c.Tapped {
		t.Error("the plain attacker was untapped")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %v, want an additional combat", next.Step)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(modified, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("the second attack added a third combat; it triggers only once each turn")
	}
}

func TestAkkiBattleSquadIgnoresAPlainAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squad := effectstest.B12Push(g, me.ID, "Akki Battle Squad", "Creature — Goblin Samurai", akkiBattleSquadOracle, 6, 6)
	plain := effectstest.BigBear(g, me.ID, 2, false)
	effectstest.DeclareAttack(t, g, opp.ID, plain)
	if n := effectstest.TriggersOnStackFrom(g, squad); n != 0 {
		t.Errorf("triggers = %d, want 0", n)
	}
}

// --- Kaima, the Fractured Calm ---

func TestKaimaGoadsEnchantedOpponentCreaturesAndGrowsPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	kaima := effectstest.B12Push(g, me.ID, "Kaima, the Fractured Calm", "Legendary Creature — Spirit", kaimaOracle, 3, 3)
	mine := effectstest.BigBear(g, opp.ID, 2, false)
	theirsAlso := effectstest.BigBear(g, third.ID, 2, false)
	unenchanted := effectstest.BigBear(g, opp.ID, 2, false)
	oursEnchanted := effectstest.BigBear(g, me.ID, 2, false)
	myAura := effectstest.B12Push(g, me.ID, "My Aura", "Enchantment — Aura", "", 0, 0)
	myAura2 := effectstest.B12Push(g, me.ID, "My Aura Two", "Enchantment — Aura", "", 0, 0)
	myAura3 := effectstest.B12Push(g, me.ID, "My Aura Three", "Enchantment — Aura", "", 0, 0)
	theirAura := effectstest.B12Push(g, opp.ID, "Their Aura", "Enchantment — Aura", "", 0, 0)
	effectstest.AttachForTest(t, g, myAura, mine)
	effectstest.AttachForTest(t, g, myAura2, theirsAlso)
	effectstest.AttachForTest(t, g, myAura3, oursEnchanted)
	effectstest.AttachForTest(t, g, theirAura, unenchanted)

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
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
	if n := effectstest.CountersOn(g, kaima, game.CounterPlusOne); n != 2 {
		t.Errorf("Kaima counters = %d, want 2", n)
	}
}

// --- Kosei, Penitent Warlord ---

func TestKoseiGainsItsTriggerOnlyWhileEnchantedEquippedAndCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	kosei := effectstest.B12Push(g, me.ID, "Kosei, Penitent Warlord", "Legendary Creature — Ogre Samurai", koseiOracle, 0, 5)
	aura := effectstest.B12Push(g, me.ID, "Some Aura", "Enchantment — Aura", "", 0, 0)
	gear := effectstest.B12Push(g, me.ID, "Some Equipment", "Artifact — Equipment", "", 0, 0)
	effectstest.AttachForTest(t, g, aura, kosei)
	effectstest.AttachForTest(t, g, gear, kosei)
	effectstest.S82SetCounters(g, kosei, map[string]int{"+1/+1": 3})
	// With a +1/+1 counter Kosei is 3/5 (the counter is layer 7d).
	handBefore := me.Hand.Size()
	oppLife, thirdLife := opp.Life, third.Life
	effectstest.DeclareAttack(t, g, opp.ID, kosei)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
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
			effectstest.AttachForTest(t, g, effectstest.B12Push(g, me.ID, "Aura", "Enchantment — Aura", "", 0, 0), kosei)
			effectstest.AttachForTest(t, g, effectstest.B12Push(g, me.ID, "Gear", "Artifact — Equipment", "", 0, 0), kosei)
		},
		"no equipment": func(t *testing.T, g *game.Game, me *game.Player, kosei uuid.UUID) {
			effectstest.AttachForTest(t, g, effectstest.B12Push(g, me.ID, "Aura", "Enchantment — Aura", "", 0, 0), kosei)
			effectstest.S82SetCounters(g, kosei, map[string]int{"+1/+1": 3})
		},
		"no aura": func(t *testing.T, g *game.Game, me *game.Player, kosei uuid.UUID) {
			effectstest.AttachForTest(t, g, effectstest.B12Push(g, me.ID, "Gear", "Artifact — Equipment", "", 0, 0), kosei)
			effectstest.S82SetCounters(g, kosei, map[string]int{"+1/+1": 3})
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			kosei := effectstest.B12Push(g, me.ID, "Kosei, Penitent Warlord", "Legendary Creature — Ogre Samurai", koseiOracle, 0, 5)
			build(t, g, me, kosei)
			handBefore := me.Hand.Size()
			effectstest.DeclareAttack(t, g, opp.ID, kosei)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceTo(t, g, game.StepCombatDamage)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - handBefore; got != 0 {
				t.Errorf("cards drawn = %d, want 0", got)
			}
		})
	}
}

// --- Concord with the Kami ---

func concordEndStep(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	mode := effectstest.LatestChoiceOfKind(g, game.PendingChoiceModePick)
	if mode == nil {
		t.Fatal("no mode prompt at the end step")
	}
	return mode
}

func TestConcordWithTheKamiRunsEveryChosenBulletThatHasItsCondition(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Concord with the Kami", "Enchantment", concordWithTheKamiOracle, 0, 0)
	host := effectstest.ModifiedBear(g, me.ID)
	aura := effectstest.B12Push(g, me.ID, "Some Aura", "Enchantment — Aura", "", 0, 0)
	gear := effectstest.B12Push(g, me.ID, "Some Equipment", "Artifact — Equipment", "", 0, 0)
	effectstest.AttachForTest(t, g, aura, host)
	effectstest.AttachForTest(t, g, gear, host)
	handBefore := me.Hand.Size()

	mode := concordEndStep(t, g)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0, 1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, host, game.CounterPlusOne); n != 2 {
		t.Errorf("counters = %d, want 2", n)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("cards drawn = %d, want 1", got)
	}
	if n := effectstest.CountTokensControlled(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("Spirits = %d, want 1", n)
	}
}

func TestConcordWithTheKamiDoesNothingWithoutAnEnchantedOrEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Concord with the Kami", "Enchantment", concordWithTheKamiOracle, 0, 0)
	effectstest.BigBear(g, me.ID, 2, false)
	handBefore := me.Hand.Size()
	mode := concordEndStep(t, g)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore || effectstest.CountTokensControlled(g, me.ID, "Spirit") != 0 {
		t.Error("a bullet with no enchanted or equipped creature did something")
	}
}

// --- Ascendant Acolyte ---

func TestAscendantAcolyteEntersWithACounterPerPlusOnePlusOneCounterAmongOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CounterBear(g, me.ID, map[string]int{"+1/+1": 2})
	effectstest.CounterBear(g, me.ID, map[string]int{"+1/+1": 1, "stun": 5})
	theirs := effectstest.BigBear(g, g.Seats[1].ID, 2, true) // an opponent's counter does not count
	_ = theirs
	acolyte := effectstest.RfdCast(t, g, "Ascendant Acolyte", "Creature — Human Monk", ascendantAcolyteOracle, 1, 1)
	if n := effectstest.CountersOn(g, acolyte, game.CounterPlusOne); n != 3 {
		t.Errorf("counters = %d, want 3 (2 + 1)", n)
	}
}

func TestAscendantAcolyteDoublesItsCountersEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	acolyte := effectstest.B12Push(g, me.ID, "Ascendant Acolyte", "Creature — Human Monk", ascendantAcolyteOracle, 1, 1)
	effectstest.S82SetCounters(g, acolyte, map[string]int{"+1/+1": 3, "stun": 1})
	for i := 0; i < 400 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, acolyte, game.CounterPlusOne); n != 3 {
		t.Fatalf("an opponent's upkeep doubled them: %d", n)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, acolyte, game.CounterPlusOne); n != 6 {
		t.Errorf("+1/+1 counters = %d, want 6", n)
	}
	if n := effectstest.CountersOn(g, acolyte, "stun"); n != 1 {
		t.Errorf("stun counters = %d, want 1 (only +1/+1 counters are doubled)", n)
	}
}

// --- Iron Apprentice ---

func TestIronApprenticeEntersWithACounterAndPassesItsCountersOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	friend := effectstest.BigBear(g, me.ID, 2, false)
	apprentice := effectstest.RfdCast(t, g, "Iron Apprentice", "Artifact Creature — Construct", ironApprenticeOracle, 0, 0)
	if n := effectstest.CountersOn(g, apprentice, game.CounterPlusOne); n != 1 {
		t.Fatalf("enters with %d counters, want 1", n)
	}
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(apprentice, "stun", 2) })
	effectstest.KillCreature(t, g, me.ID, apprentice)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, friend)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, friend, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the other creature = %d, want 1", n)
	}
	if n := effectstest.CountersOn(g, friend, "stun"); n != 2 {
		t.Errorf("stun counters on the other creature = %d, want 2", n)
	}
}

func TestIronApprenticeWithoutCountersDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	friend := effectstest.BigBear(g, me.ID, 2, false)
	apprentice := effectstest.B12Push(g, me.ID, "Iron Apprentice", "Artifact Creature — Construct", ironApprenticeOracle, 0, 1)
	effectstest.KillCreature(t, g, me.ID, apprentice)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		t.Fatal("a target prompt for a creature that had no counters")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, friend, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}

// --- Dramatist's Puppet ---

func TestDramatistsPuppetAsksOncePerKindOfCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.BigBear(g, g.Seats[1].ID, 2, false)
	effectstest.S82SetCounters(g, target, map[string]int{"+1/+1": 2, "stun": 1, "charge": 3})
	effectstest.RfdCast(t, g, "Dramatist's Puppet", "Artifact Creature — Construct", dramatistsPuppetOracle, 2, 4)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	// Kinds are asked in sorted order: "+1/+1", "charge", "stun".
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // +1/+1: put another
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // charge: remove one
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // stun: remove its only counter
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Fatal("asked more than once per kind")
	}
	if n := effectstest.CountersOn(g, target, "+1/+1"); n != 3 {
		t.Errorf("+1/+1 = %d, want 3", n)
	}
	if n := effectstest.CountersOn(g, target, "charge"); n != 2 {
		t.Errorf("charge = %d, want 2", n)
	}
	if n := effectstest.CountersOn(g, target, "stun"); n != 0 {
		t.Errorf("stun = %d, want 0", n)
	}
}

func TestDramatistsPuppetOnAPermanentWithNoCountersDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.BigBear(g, g.Seats[1].ID, 2, false)
	effectstest.RfdCast(t, g, "Dramatist's Puppet", "Artifact Creature — Construct", dramatistsPuppetOracle, 2, 4)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("asked about a kind of counter the permanent doesn't have")
	}
}
