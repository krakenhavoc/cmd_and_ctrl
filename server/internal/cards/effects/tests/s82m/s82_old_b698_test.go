package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b698_test.go — Mercadian Masques, "+1/+1 and other counters" (S82 OLD
// batch 698, #5619): each card driven through a real activation, combat or
// trigger.

const (
	oldB698Lift      = "88c3bb73-c765-4e1c-b4b9-97eaab02182f"
	oldB698Lamprey   = "f5b374d8-bf22-44f7-a18a-18c096a84640"
	oldB698Rushwood  = "d98edd47-c9de-46e9-a184-ba3bc0be42c4"
	oldB698Toy       = "f1a0d112-9e1b-4354-bd7f-0051bff8f904"
	oldB698Hound     = "ace532a3-fa6e-479d-b169-1fb24c3d0f8c"
	oldB698Erithizon = "45645a97-f41d-47fb-b943-293ffc18fc82"
	oldB698LeyLine   = "230d79dd-15dd-431d-af22-bdc6eaf86341"
)

func TestS82OldB698AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB698Lift, oldB698Lamprey, oldB698Rushwood, oldB698Toy, oldB698Hound, oldB698Erithizon, oldB698LeyLine} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB698MercadianLiftWindsAndPutsACreatureOfThatManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lift := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Mercadian Lift", TypeLine: "Artifact", ManaCost: "{2}", OracleID: oldB698Lift})
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, lift, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("wind: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, lift, "winch"); n != 1 {
		t.Fatalf("%d winch counters, want 1", n)
	}
	// Two more counters, then spend two on a mana value 2 creature.
	effectstest.AddCountersForTest(g, lift, "winch", 1)
	effectstest.UntapForTest(g, lift)
	two, three := uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: two, Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: three, Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, lift, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{lift}, CounterCounts: []int{3}}); err == nil {
		t.Fatal("removed more winch counters than the Lift holds")
	}
	if err := g.ActivateCatalogAbility(me.ID, lift, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{lift}, CounterCounts: []int{2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("no put offered")
	}
	for _, id := range p.ChooseCards {
		if id == three {
			t.Error("a creature card with mana value 3 was offered at X = 2")
		}
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{two}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, two) || me.Hand.Contains(two) {
		t.Error("the mana value 2 creature did not enter")
	}
	if n := effectstest.CountersOn(g, lift, "winch"); n != 0 {
		t.Errorf("%d winch counters left, want 0", n)
	}
	if !me.Hand.Contains(three) {
		t.Error("the mana value 3 creature left the hand")
	}
}

func TestS82OldB698KyrenToyBanksAndPaysXPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	toy := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Kyren Toy", TypeLine: "Artifact", ManaCost: "{3}", OracleID: oldB698Toy})
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, toy, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("bank: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, toy, game.CounterCharge); n != 1 {
		t.Fatalf("%d charge counters, want 1", n)
	}
	for _, tc := range []struct {
		name               string
		held, remove, want int
	}{
		{"none removed", 3, 0, 1},
		{"two removed", 3, 2, 3},
		{"none held", 0, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			toy := effectstest.B12Push(g, me.ID, "Kyren Toy", "Artifact", oldB698Toy, 0, 0)
			effectstest.AdvanceToMain(t, g)
			if tc.held > 0 {
				effectstest.AddCountersForTest(g, toy, game.CounterCharge, tc.held)
			}
			effectstest.ActivateManaFor(t, g, me.ID, toy, 0, game.ManaAbilityParams{CounterCounts: []int{tc.remove}})
			if got := effectstest.PoolColours(me); got["C"] != tc.want || me.ManaPool.Total() != tc.want {
				t.Errorf("pool %v, want %d {C} and nothing else", got, tc.want)
			}
			if got := effectstest.CounterCount(g, toy, game.CounterCharge); got != tc.held-tc.remove {
				t.Errorf("charge counters left %d, want %d", got, tc.held-tc.remove)
			}
		})
	}
}

func TestS82OldB698LampreyCountersEachBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lamprey := effectstest.B12Push(g, me.ID, "Quagmire Lamprey", "Creature — Fish", oldB698Lamprey, 1, 1)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 3, 3)
	effectstest.BrAttack(t, g, lamprey)
	for _, blocker := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(blocker, lamprey); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, blocker := range []uuid.UUID{a, b} {
		if n := effectstest.CounterCount(g, blocker, game.CounterMinusOne); n != 1 {
			t.Errorf("a blocker has %d -1/-1 counters, want 1", n)
		}
	}
	if n := effectstest.CounterCount(g, lamprey, game.CounterMinusOne); n != 0 {
		t.Errorf("the Lamprey got %d -1/-1 counters", n)
	}
}

func TestS82OldB698RushwoodElementalMayGrowEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elemental := effectstest.B12Push(g, me.ID, "Rushwood Elemental", "Creature — Elemental", oldB698Rushwood, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, elemental, "trample") {
		t.Error("no trample")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	oldB698AnswerTriggerPrompt(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, elemental, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d +1/+1 counters after accepting, want 1", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	oldB698AnswerTriggerPrompt(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, elemental, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters after declining, want 1", n)
	}
}

func TestS82OldB698BloodHoundTakesCountersForDamageAndLosesThemAtEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hound := effectstest.B12Push(g, me.ID, "Blood Hound", "Creature — Dog", oldB698Hound, 1, 1)
	other := effectstest.B12Creature(g, g.Seats[1].ID, "Elsewhere", "Creature — Bear", 2, 2)
	effectstest.GptB1DamagePlayer(g, g.Seats[1].ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("damage to an opponent asked the Hound's controller")
	}
	effectstest.GptB1DamagePlayer(g, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, hound, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d +1/+1 counters, want 3 (that much damage)", n)
	}
	effectstest.GptB1DamagePlayer(g, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, hound, game.CounterPlusOne); n != 3 {
		t.Errorf("%d +1/+1 counters after declining, want 3", n)
	}
	_ = other
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, hound, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters left at the end step, want 0", n)
	}
}

func TestS82OldB698ErithizonCounterGoesWhereTheDefenderSays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, defender := g.Seats[0], g.Seats[2]
	eri := effectstest.B12Push(g, me.ID, "Erithizon", "Creature — Beast", oldB698Erithizon, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, defender.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, defender.ID, eri)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != defender.ID {
		t.Fatalf("the prompt is for %s, want the defending player", ch.Chooser)
	}
	if err := g.ResolvePickTarget(ch.ID, defender.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
		t.Fatalf("picking the attacker's own creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, mine, game.CounterPlusOne); n != 1 {
		t.Errorf("the chosen creature has %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.CounterCount(g, theirs, game.CounterPlusOne) + effectstest.CounterCount(g, eri, game.CounterPlusOne); n != 0 {
		t.Errorf("another creature got %d counters", n)
	}
}

func TestS82OldB698LeyLineLetsTheUpkeepPlayerPutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Ley Line", oldB698LeyLine, "Enchantment")
	mine := effectstest.B12Creature(g, upkeep.ID, "Upkeep Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, owner.ID, "Owner Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != upkeep.ID {
		t.Fatalf("the prompt is for %s, want the upkeep player", ch.Chooser)
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, upkeep, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("the chosen creature has %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.CounterCount(g, mine, game.CounterPlusOne); n != 0 {
		t.Errorf("the other creature has %d counters", n)
	}
}

// oldB698AnswerTriggerPrompt answers the open "you may" trigger prompt.
func oldB698AnswerTriggerPrompt(t *testing.T, g *game.Game, who *game.Player, yes bool) {
	t.Helper()
	c := effectstest.ConfirmForTriggerPrompt(g, who.ID)
	if c == nil {
		t.Fatalf("no trigger prompt for %s", who.Name)
	}
	if err := g.ResolveTriggerPrompt(c.ID, who.ID, yes); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
}

// An attack on a planeswalker: its controller is the defending player.
func TestS82OldB698ErithizonOnAPlaneswalkerAsksItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, defender := g.Seats[0], g.Seats[1]
	eri := effectstest.B12Push(g, me.ID, "Erithizon", "Creature — Beast", oldB698Erithizon, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	pw := effectstest.PushTestPermanent(g, defender.ID, game.Card{Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4}})
	effectstest.DeclareAttack(t, g, pw, eri)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != defender.ID {
		t.Fatalf("the prompt is for %s, want the planeswalker's controller", ch.Chooser)
	}
	if err := g.ResolvePickTarget(ch.ID, defender.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}
