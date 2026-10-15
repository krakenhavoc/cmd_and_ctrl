package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tidus_b_test.go — the Tidus, Yuna's Guardian deck request (#3246),
// slice B: Tidus and the counter-moving guardians.

const (
	tdAuronOracle            = "5e166b42-252e-4f34-9610-dcddc04a62ff"
	tdForgottenAncientOracle = "2ef75cf3-30dc-4852-a6d1-1d6391fad022"
	tdGattaOracle            = "c01d6fde-1bd2-4412-932a-9bc97d4e5aef"
	tdJyscalOracle           = "52791629-0f22-4023-bb04-c829d1600568"
	tdRikkuOracle            = "81ba75af-eddc-4344-befb-419ebfb855a3"
	tdSphereGridOracle       = "f4770345-f15e-4e31-b6ef-4c2608df9644"
	tdTidusOracle            = "e8fbcafd-3617-4cd7-b55d-457fb7b60162"
	tdTromellOracle          = "f088817a-570d-4480-bc59-f2ad6675a71b"
	tdWakkaOracle            = "fbd37284-35ff-4409-b6d5-f813c06c4931"
)

func TestTidusBSliceIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		tdAuronOracle:            "Auron, Venerated Guardian",
		tdForgottenAncientOracle: "Forgotten Ancient",
		tdGattaOracle:            "Gatta and Luzzu",
		tdJyscalOracle:           "Lord Jyscal Guado",
		tdRikkuOracle:            "Rikku, Resourceful Guardian",
		tdSphereGridOracle:       "Sphere Grid",
		tdTidusOracle:            "Tidus, Yuna's Guardian",
		tdTromellOracle:          "Tromell, Seymour's Butler",
		tdWakkaOracle:            "Wakka, Devoted Guardian",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func tdAddCounter(g *game.Game, by, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(by, id, kind, n) })
}

func TestTidusBAuronCountersThenExilesASmallerCreatureUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auron := b12Push(g, me.ID, "Auron, Venerated Guardian", "Legendary Creature — Human Spirit Samurai", tdAuronOracle, 2, 5)
	small := tdCreature(g, opp.ID, "Small", 2, 2)
	big := tdCreature(g, opp.ID, "Big", 3, 3)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(auron, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Error("a creature with power equal to Auron's was a legal target")
	}
	if got := countersOn(g, auron, game.CounterPlusOne); got != 1 {
		t.Fatalf("Auron has %d +1/+1 counters before the reflexive trigger, want 1", got)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(small) {
		t.Fatal("the smaller creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(auron) })
	g.RunStateChecksForTest()
	if findBattlefieldByName(g, "Small") == uuid.Nil {
		t.Error("the exiled creature did not return when Auron left")
	}
}

func TestTidusBForgottenAncientGrowsOnCastsAndMovesCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ancient := b12Push(g, me.ID, "Forgotten Ancient", "Creature — Elemental", tdForgottenAncientOracle, 0, 3)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	tdCreature(g, opp.ID, "Their Bear", 2, 2)

	tdCastCreature(t, g, "Spell", 1, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ancient, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Ancient has %d counters after a cast, want 1", got)
	}
	tdAddCounter(g, me.ID, ancient, game.CounterPlusOne, 2)

	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if c == nil {
			continue
		}
		switch c.Kind {
		case game.PendingChoiceOwnPermanents:
			if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
		case game.PendingChoiceTheirPermanents:
			if err := g.ResolveTheirPermanents(c.ID, me.ID, nil); err != nil {
				t.Fatalf("ResolveTheirPermanents: %v", err)
			}
		}
	}
	answerNumber(t, g, me.ID, 2)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d counters, want 2 moved onto it", got)
	}
	if got := countersOn(g, ancient, game.CounterPlusOne); got != 1 {
		t.Errorf("the Ancient has %d counters left, want 1", got)
	}
}

func TestTidusBGattaAndLuzzuTurnDamageIntoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Gatta and Luzzu", "Legendary Creature — Human Soldier", tdGattaOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the shielded creature died")
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the creature has %d +1/+1 counters, want 3", got)
	}
}

func TestTidusBLordJyscalInvestigatesAfterYouCounterACreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	b12Push(g, me.ID, "Lord Jyscal Guado", "Legendary Creature — Spirit Cleric", tdJyscalOracle, 2, 1)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	clues := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Clue" {
				n++
			}
		}
		return n
	}
	tdAddCounter(g, me.ID, bear, game.CounterPlusOne, 1)
	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if clues() != 1 {
		t.Fatalf("%d Clues after a counter this turn, want 1", clues())
	}
	// The next player's end step, with no counter that turn: nothing.
	advanceToEndStepOf(t, g, (seat+1)%len(g.Seats))
	passPriorityAroundTable(t, g)
	if clues() != 1 {
		t.Errorf("%d Clues after a turn without counters, want still 1", clues())
	}
}

func TestTidusBRikkuStealsACounterAndMakesTheCreatureUnblockableByOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rikku := b12Push(g, me.ID, "Rikku, Resourceful Guardian", "Legendary Creature — Human Artificer", tdRikkuOracle, 2, 3)
	mine := tdCreature(g, me.ID, "Mine", 2, 2)
	theirs := tdCreature(g, opp.ID, "Theirs", 2, 2)
	blocker := tdCreature(g, opp.ID, "Blocker", 2, 2)
	tdAddCounter(g, opp.ID, theirs, game.CounterPlusOne, 1)

	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, rikku, 0, game.ActivateAbilityParams{
		Targets: append(tdCard(theirs), tdCard(mine)...),
	}); err != nil {
		t.Fatalf("activate Steal: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, mine, game.CounterPlusOne) != 1 || countersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Fatalf("after Steal: mine %d, theirs %d; want 1 and 0",
			countersOn(g, mine, game.CounterPlusOne), countersOn(g, theirs, game.CounterPlusOne))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, mine); err == nil {
		t.Error("an opponent's creature blocked the creature Rikku's trigger covered")
	}
}

func TestTidusBSphereGridCountersConnectingCreaturesAndGrantsReachTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Sphere Grid", "Enchantment", tdSphereGridOracle, false)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Bear has %d counters after connecting, want 1", got)
	}
	for _, kw := range []string{"reach", "trample"} {
		if !containsString(effectiveAbilities(t, g, bear), kw) {
			t.Errorf("a creature with a +1/+1 counter has no %s", kw)
		}
	}
}

func TestTidusBTidusMovesACounterAndCheersOnceATurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tidus, Yuna's Guardian", "Legendary Creature — Human Warrior", tdTidusOracle, 3, 3)
	from := tdCreature(g, me.ID, "From", 2, 2)
	to := tdCreature(g, me.ID, "To", 2, 2)
	tdAddCounter(g, me.ID, from, game.CounterPlusOne, 1)

	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, from)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, to)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if countersOn(g, from, game.CounterPlusOne) != 0 || countersOn(g, to, game.CounterPlusOne) != 1 {
		t.Fatalf("after the move: from %d, to %d; want 0 and 1",
			countersOn(g, from, game.CounterPlusOne), countersOn(g, to, game.CounterPlusOne))
	}

	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, to)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, to)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Cheer drew %d, want 1", me.Hand.Size()-hand)
	}
	if got := countersOn(g, to, game.CounterPlusOne); got != 2 {
		t.Errorf("after the proliferate the creature has %d counters, want 2", got)
	}
}

func TestTidusBTromellAddsACounterAndProliferatesPerEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tromell := b12Push(g, me.ID, "Tromell, Seymour's Butler", "Legendary Creature — Elf Advisor", tdTromellOracle, 2, 3)
	a := tdCastCreature(t, g, "A", 1, 1)
	passPriorityAroundTable(t, g)
	b := tdCastCreature(t, g, "B", 1, 1)
	passPriorityAroundTable(t, g)
	if countersOn(g, a, game.CounterPlusOne) != 1 || countersOn(g, b, game.CounterPlusOne) != 1 {
		t.Fatal("a nontoken creature entered without the extra counter")
	}
	pushToken(g, me.ID, game.Card{Name: "Token", TypeLine: "Token Creature — Test", Power: 1, Toughness: 1})

	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, tromell, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Tromell: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, a)
	answerProliferate(t, g, a)
	if proliferatePrompt(g) != nil {
		t.Error("a third proliferate was asked for two entries")
	}
	if got := countersOn(g, a, game.CounterPlusOne); got != 3 {
		t.Errorf("A has %d counters after two proliferates, want 3", got)
	}
}

func TestTidusBWakkaDestroysAnArtifactAndCaptainsTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	wakka := b12Push(g, me.ID, "Wakka, Devoted Guardian", "Legendary Creature — Human Warrior", tdWakkaOracle, 4, 4)
	friend := tdCreature(g, me.ID, "Friend", 2, 2)
	relic := b12Push(g, opp.ID, "Their Relic", "Artifact", "", 0, 0)

	attackWith(t, g, opp.ID, wakka)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
	if got := countersOn(g, wakka, game.CounterPlusOne); got != 1 {
		t.Fatalf("Wakka has %d counters, want 1", got)
	}
	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, friend, game.CounterPlusOne); got != 1 {
		t.Errorf("the other creature has %d counters at the end step, want 1", got)
	}
	if got := countersOn(g, wakka, game.CounterPlusOne); got != 1 {
		t.Errorf("Wakka put a counter on himself (%d)", got)
	}
}
