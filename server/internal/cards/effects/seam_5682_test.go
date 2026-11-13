package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seam_5682_test.go — #5682: dice rolled to pay an activation cost
// (game.AbilityCost.RollDie, Clay Golem) and dice rolled as a permanent
// enters (game.EntryCountersFromCast.RollXSides, Neverwinter Hydra). Both
// go through the one dice door (game/dice_roll.go, game/dice_cost.go).

const (
	clayGolemOracle        = "8773c2ea-ef41-4403-bb52-07f0d70d9bbe"
	neverwinterHydraOracle = "e71e7a68-d96e-4ebb-8d4f-9c45ff8b8b66"
)

func TestSeam5682CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		clayGolemOracle:        "Clay Golem",
		neverwinterHydraOracle: "Neverwinter Hydra",
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

// s5682Rolls lists the dice announced after the first `from` events, and
// fails on one that is not a d<sides> from `source` by `roller`.
func s5682Rolls(t *testing.T, g *game.Game, from int, roller, source uuid.UUID, sides int) []int {
	t.Helper()
	var out []int
	g.WithWriteLock(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind != game.EventRollDie {
				continue
			}
			if ev.Actor != roller || ev.Source != source || ev.Sides != sides || ev.Amount < 1 || ev.Amount > sides {
				t.Errorf("roll %+v, want a d%d by the roller from the source", ev, sides)
			}
			out = append(out, ev.Amount)
		}
	})
	return out
}

// s5682Shield is Squid Fire Knight's "the next time <roller> would roll
// one or more dice this turn, instead they roll that many dice plus one,
// then you choose one of those rolls to ignore", with `chooser` as you.
func s5682Shield(t *testing.T, g *game.Game, chooser, roller uuid.UUID) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() { ok = g.RollExtraDieNextTimeForEffect(uuid.New(), chooser, roller, "Squid Fire Knight") })
	if !ok {
		t.Fatal("the shield was not registered")
	}
}

func s5682Golem(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	golem := b12Push(g, me.ID, "Clay Golem", "Artifact Creature — Golem", clayGolemOracle, 4, 4)
	return g, me, opp, golem
}

func s5682GolemItem(t *testing.T, g *game.Game, golem uuid.UUID) *game.StackItem {
	t.Helper()
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.SourceCardID == golem {
			return item
		}
	}
	t.Fatal("the Golem's ability is not on the stack")
	return nil
}

// "{6}, Roll a d8: Monstrosity X, where X is the result." The d8 is
// rolled as the ability is activated; the Golem gets that many counters
// as it resolves; "Berserk — When this creature becomes monstrous,
// destroy target permanent."
func TestClayGolemRollsAD8AsACostAndBerserkDestroysAPermanent(t *testing.T) {
	g, me, opp, golem := s5682Golem(t)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	from := len(g.Events)
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	rolls := s5682Rolls(t, g, from, me.ID, golem, 8)
	if len(rolls) != 1 {
		t.Fatalf("rolled %v as the ability was activated, want one d8", rolls)
	}
	x := rolls[0]
	if got := s5682GolemItem(t, g, golem).Paid.Rolled; !sameInts(got, rolls) {
		t.Fatalf("the ability's payment record rolled %v, want %v", got, rolls)
	}
	if counterCount(g, golem, "+1/+1") != 0 || monstrousNow(g, golem) {
		t.Fatal("monstrous before the ability resolved")
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, golem, "+1/+1"); n != x || !monstrousNow(g, golem) {
		t.Fatalf("%d counters, monstrous %v; want %d counters and monstrous", n, monstrousNow(g, golem), x)
	}
	if got := becameMonstrousEvents(g, golem); !sameInts(got, []int{x}) {
		t.Errorf("became monstrous with X %v, want [%d]", got, x)
	}
	if len(s5682Rolls(t, g, from, me.ID, golem, 8)) != 1 {
		t.Error("the resolution rolled again")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Berserk asked for no target")
	}
	if !hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, golem) {
		t.Errorf("target permanent offers %v, want any permanent", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target permanent survived")
	}
	if !g.Battlefield.Contains(golem) {
		t.Error("the Golem left")
	}
}

// A second activation in response pays and rolls again, and its
// resolution does nothing (CR 701.37a): X is the first to resolve's roll.
func TestClayGolemActivatedInResponseRollsAgainForNothing(t *testing.T) {
	g, me, _, golem := s5682Golem(t)
	from := len(g.Events)
	for range 2 {
		if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
	}
	rolls := s5682Rolls(t, g, from, me.ID, golem, 8)
	if len(rolls) != 2 {
		t.Fatalf("rolled %v, want one d8 per activation", rolls)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, golem)
	}
	// The second activation is the first to resolve.
	if n := counterCount(g, golem, "+1/+1"); g.Battlefield.Contains(golem) && n != rolls[1] {
		t.Errorf("%d counters, want the roll of the ability that resolved first, %d", n, rolls[1])
	}
	if got := becameMonstrousEvents(g, golem); !sameInts(got, []int{rolls[1]}) {
		t.Errorf("became monstrous with X %v, want [%d]", got, rolls[1])
	}
}

// The cost's roll is a replaceable roll: Squid Fire Knight's shield adds
// a die and asks its controller which to ignore, mid-payment, with the
// ability already on the stack. Nobody can act until the answer is in,
// the open question is a restore point, and X is the roll that was kept.
func TestClayGolemCostRollIsReplaceableAndWaitsForTheIgnoredRoll(t *testing.T) {
	g, me, opp, golem := s5682Golem(t)
	s5682Shield(t, g, opp.ID, me.ID)
	from := len(g.Events)
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := s5682Rolls(t, g, from, me.ID, golem, 8); len(got) != 0 {
		t.Fatalf("a roll was announced before one was ignored: %v", got)
	}
	if got := s5682GolemItem(t, g, golem).Paid.Rolled; len(got) != 0 {
		t.Fatalf("the result was recorded before one was ignored: %v", got)
	}
	c := sfkPrompt(g, "dice/ignore-roll")
	if c == nil || c.Chooser != opp.ID {
		t.Fatalf("no roll-to-ignore question for the shield's controller: %+v", c)
	}
	dice := sfkRolls(t, c)
	if len(dice) != 2 {
		t.Fatalf("%d dice rolled, want 2", len(dice))
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("priority was passed with the cost's roll still open")
	}

	restored := restoreRoundTrip(t, g, true)
	rc := sfkPrompt(restored, "dice/ignore-roll")
	if rc == nil || !sameInts(sfkRolls(t, rc), dice) {
		t.Fatalf("the restored question: %+v, want dice %v", rc, dice)
	}
	sfkAnswer(t, restored, rc, opp.ID, 0)
	if got := s5682Rolls(t, restored, from, me.ID, golem, 8); !sameInts(got, dice[1:]) {
		t.Fatalf("announced %v, want the kept roll %v", got, dice[1:])
	}
	if got := s5682GolemItem(t, restored, golem).Paid.Rolled; !sameInts(got, dice[1:]) {
		t.Fatalf("the payment record rolled %v, want %v", got, dice[1:])
	}
	passPriorityAroundTable(t, restored)
	if n := counterCount(restored, golem, "+1/+1"); n != dice[1] || !monstrousNow(restored, golem) {
		t.Errorf("%d counters, monstrous %v; want the kept roll, %d", n, monstrousNow(restored, golem), dice[1])
	}
}

// A copy of the ability copies the result with the rest of its costs
// (CR 707.10), and the payment record survives an undo's clone.
func TestClayGolemCostRollSurvivesACloneOfTheGame(t *testing.T) {
	g, me, _, golem := s5682Golem(t)
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	want := append([]int(nil), s5682GolemItem(t, g, golem).Paid.Rolled...)
	restored := restoreRoundTrip(t, g, true)
	if got := s5682GolemItem(t, restored, golem).Paid.Rolled; len(want) != 1 || !sameInts(got, want) {
		t.Fatalf("restored payment record rolled %v, want %v", got, want)
	}
	passPriorityAroundTable(t, restored)
	if n := counterCount(restored, golem, "+1/+1"); n != want[0] {
		t.Errorf("%d counters after the restore, want %d", n, want[0])
	}
}

// --- Neverwinter Hydra -------------------------------------------------

func s5682CastHydra(t *testing.T, g *game.Game, x int) uuid.UUID {
	t.Helper()
	return castXSpell(t, g, "Neverwinter Hydra", "Creature — Hydra", neverwinterHydraOracle, "{X}{X}{G}{G}", x, nil)
}

// "As this creature enters, roll X d6. It enters with a number of +1/+1
// counters on it equal to the total of those results." Nothing is rolled
// while the spell waits on the stack; X dice are rolled as it resolves,
// and the Hydra enters with their total. Trample.
func TestNeverwinterHydraRollsXD6AndEntersWithTheirTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	from := len(g.Events)
	hydra := s5682CastHydra(t, g, 3)
	if got := s5682Rolls(t, g, from, me.ID, hydra, 6); len(got) != 0 {
		t.Fatalf("rolled %v while the spell was on the stack", got)
	}
	passPriorityAroundTable(t, g)
	rolls := s5682Rolls(t, g, from, me.ID, hydra, 6)
	if len(rolls) != 3 {
		t.Fatalf("rolled %v, want three d6", rolls)
	}
	total := rolls[0] + rolls[1] + rolls[2]
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("the Hydra is not on the battlefield")
	}
	if n := counterCount(g, hydra, "+1/+1"); n != total || cardByID(g, hydra).CurrentPower() != total {
		t.Errorf("%d counters, power %d; want the total of %v", n, cardByID(g, hydra).CurrentPower(), rolls)
	}
	if !hasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("no trample")
	}
	// One batch: "whenever you roll one or more dice" sees one roll.
	batches := map[uint64]bool{}
	g.WithWriteLock(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == game.EventRollDie {
				batches[ev.BatchSeq] = true
			}
		}
	})
	if len(batches) != 1 {
		t.Errorf("the dice were announced in %d batches, want one", len(batches))
	}
}

// Cast for X=0 it rolls nothing and is a 0/0 (CR 704.5f).
func TestNeverwinterHydraCastForZeroRollsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	from := len(g.Events)
	hydra := s5682CastHydra(t, g, 0)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == hydra {
				g.Stack.Cards[i].PrintedPTKnown = true
			}
		}
	})
	passPriorityAroundTable(t, g)
	if got := s5682Rolls(t, g, from, me.ID, hydra, 6); len(got) != 0 {
		t.Errorf("rolled %v for X=0", got)
	}
	if g.Battlefield.Contains(hydra) || !me.Graveyard.Contains(hydra) {
		t.Error("a 0/0 Hydra should be in its owner's graveyard")
	}
}

// Put onto the battlefield without being cast, X is 0 (CR 107.3m): no
// dice, no counters.
func TestNeverwinterHydraPutOntoTheBattlefieldRollsNothing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hydra := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: hydra, Name: "Neverwinter Hydra", TypeLine: "Creature — Hydra", OracleID: neverwinterHydraOracle,
		ManaCost: "{X}{X}{G}{G}", PrintedPTKnown: true, Owner: me.ID, Controller: me.ID,
	})
	from := len(g.Events)
	var err error
	g.WithWriteLock(func() {
		_, err = g.ReturnToBattlefieldForEffect(hydra, me.ID, false)
	})
	if err != nil {
		t.Fatalf("put onto the battlefield: %v", err)
	}
	g.RunStateChecksForTest()
	if got := s5682Rolls(t, g, from, me.ID, hydra, 6); len(got) != 0 {
		t.Errorf("rolled %v for a Hydra that was not cast", got)
	}
	if g.Battlefield.Contains(hydra) {
		t.Error("a Hydra that was not cast has no counters and should have died")
	}
}

// The entry's roll is a replaceable roll. With Squid Fire Knight's shield
// on the caster, X=2 rolls three dice and the shield's controller ignores
// one before the Hydra enters (CR 614.12a); the open question is a
// restore point, and the Hydra enters with the two that were kept.
func TestNeverwinterHydraEntryRollWaitsForTheIgnoredRollAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	from := len(g.Events)
	hydra := s5682CastHydra(t, g, 2)
	s5682Shield(t, g, opp.ID, me.ID)
	passPriorityAroundTable(t, g)
	c := sfkPrompt(g, "dice/ignore-roll")
	if c == nil || c.Chooser != opp.ID {
		t.Fatalf("no roll-to-ignore question for the shield's controller: %+v", c)
	}
	dice := sfkRolls(t, c)
	if len(dice) != 3 {
		t.Fatalf("%d dice rolled, want 3", len(dice))
	}
	if g.Battlefield.Contains(hydra) {
		t.Fatal("the Hydra entered before its dice were settled")
	}
	if got := s5682Rolls(t, g, from, me.ID, hydra, 6); len(got) != 0 {
		t.Fatalf("a roll was announced before one was ignored: %v", got)
	}

	restored := restoreRoundTrip(t, g, true)
	rc := sfkPrompt(restored, "dice/ignore-roll")
	if rc == nil || !sameInts(sfkRolls(t, rc), dice) {
		t.Fatalf("the restored question: %+v, want dice %v", rc, dice)
	}
	sfkAnswer(t, restored, rc, opp.ID, 1)
	restored.SettleResolution()
	kept := []int{dice[0], dice[2]}
	if got := s5682Rolls(t, restored, from, me.ID, hydra, 6); !sameInts(got, kept) {
		t.Fatalf("announced %v, want the kept rolls %v", got, kept)
	}
	if !restored.Battlefield.Contains(hydra) {
		t.Fatal("the Hydra did not enter once its dice were settled")
	}
	if n := counterCount(restored, hydra, "+1/+1"); n != kept[0]+kept[1] {
		t.Errorf("%d counters, want the total of the kept rolls %v", n, kept)
	}
}

// A copy of the spell copies X (CR 707.10) and rolls its own dice as it
// resolves: the token enters with their total.
func TestNeverwinterHydraSpellCopyRollsItsOwnDice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	from := len(g.Events)
	hydra := s5682CastHydra(t, g, 2)
	var err error
	g.WithWriteLock(func() { err = g.CopySpellForEffect(hydra, me.ID, false, nil) })
	if err != nil {
		t.Fatalf("copy the spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Neverwinter Hydra" && c.InstanceID != hydra {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("the copy did not become a token")
	}
	bySource := map[uuid.UUID][]int{}
	g.WithWriteLock(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == game.EventRollDie && ev.Sides == 6 && ev.Actor == me.ID {
				bySource[ev.Source] = append(bySource[ev.Source], ev.Amount)
			}
		}
	})
	if len(bySource) != 2 || len(bySource[hydra]) != 2 {
		t.Fatalf("rolls by source %v, want two d6 for the spell and two for its copy", bySource)
	}
	for source, rolls := range bySource {
		id := token
		if source == hydra {
			id = hydra
		}
		if len(rolls) != 2 || counterCount(g, id, "+1/+1") != rolls[0]+rolls[1] {
			t.Errorf("rolled %v and has %d counters (original: %v)", rolls, counterCount(g, id, "+1/+1"), source == hydra)
		}
	}
}

// Ward {4}: an opponent's spell that targets the Hydra is countered
// unless they pay.
func TestNeverwinterHydraWardCountersUnlessFourIsPaid(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Neverwinter Hydra", TypeLine: "Creature — Hydra", OracleID: neverwinterHydraOracle,
		Counters: map[string]int{"+1/+1": 4}, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, hydra)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) || !opp.Graveyard.Contains(blade) {
		t.Error("declining the ward must counter the spell")
	}
}
