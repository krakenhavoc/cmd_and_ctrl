package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTidusBSliceIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.TdAuronOracle:            "Auron, Venerated Guardian",
		effectstest.TdForgottenAncientOracle: "Forgotten Ancient",
		effectstest.TdGattaOracle:            "Gatta and Luzzu",
		effectstest.TdJyscalOracle:           "Lord Jyscal Guado",
		effectstest.TdRikkuOracle:            "Rikku, Resourceful Guardian",
		effectstest.TdSphereGridOracle:       "Sphere Grid",
		effectstest.TdTidusOracle:            "Tidus, Yuna's Guardian",
		effectstest.TdTromellOracle:          "Tromell, Seymour's Butler",
		effectstest.TdWakkaOracle:            "Wakka, Devoted Guardian",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestTidusBAuronCountersThenExilesASmallerCreatureUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auron := effectstest.B12Push(g, me.ID, "Auron, Venerated Guardian", "Legendary Creature — Human Spirit Samurai", effectstest.TdAuronOracle, 2, 5)
	small := effectstest.TdCreature(g, opp.ID, "Small", 2, 2)
	big := effectstest.TdCreature(g, opp.ID, "Big", 3, 3)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(auron, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Error("a creature with power equal to Auron's was a legal target")
	}
	if got := effectstest.CountersOn(g, auron, game.CounterPlusOne); got != 1 {
		t.Fatalf("Auron has %d +1/+1 counters before the reflexive trigger, want 1", got)
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(small) {
		t.Fatal("the smaller creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(auron) })
	g.RunStateChecksForTest()
	if effectstest.FindBattlefieldByName(g, "Small") == uuid.Nil {
		t.Error("the exiled creature did not return when Auron left")
	}
}

func TestTidusBForgottenAncientGrowsOnCastsAndMovesCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ancient := effectstest.B12Push(g, me.ID, "Forgotten Ancient", "Creature — Elemental", effectstest.TdForgottenAncientOracle, 0, 3)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.TdCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.TdCastCreature(t, g, "Spell", 1, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ancient, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Ancient has %d counters after a cast, want 1", got)
	}
	effectstest.TdAddCounter(g, me.ID, ancient, game.CounterPlusOne, 2)

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AnswerNumber(t, g, me.ID, 2)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d counters, want 2 moved onto it", got)
	}
	if got := effectstest.CountersOn(g, ancient, game.CounterPlusOne); got != 1 {
		t.Errorf("the Ancient has %d counters left, want 1", got)
	}
}

func TestTidusBGattaAndLuzzuTurnDamageIntoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Gatta and Luzzu", "Legendary Creature — Human Soldier", effectstest.TdGattaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the shielded creature died")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the creature has %d +1/+1 counters, want 3", got)
	}
}

func TestTidusBLordJyscalInvestigatesAfterYouCounterACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.B12Push(g, me.ID, "Lord Jyscal Guado", "Legendary Creature — Spirit Cleric", effectstest.TdJyscalOracle, 2, 1)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	clues := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Clue" {
				n++
			}
		}
		return n
	}
	effectstest.TdAddCounter(g, me.ID, bear, game.CounterPlusOne, 1)
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if clues() != 1 {
		t.Fatalf("%d Clues after a counter this turn, want 1", clues())
	}
	// The next player's end step, with no counter that turn: nothing.
	effectstest.AdvanceToEndStepOf(t, g, (seat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if clues() != 1 {
		t.Errorf("%d Clues after a turn without counters, want still 1", clues())
	}
}

func TestTidusBRikkuStealsACounterAndMakesTheCreatureUnblockableByOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rikku := effectstest.B12Push(g, me.ID, "Rikku, Resourceful Guardian", "Legendary Creature — Human Artificer", effectstest.TdRikkuOracle, 2, 3)
	mine := effectstest.TdCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.TdCreature(g, opp.ID, "Theirs", 2, 2)
	blocker := effectstest.TdCreature(g, opp.ID, "Blocker", 2, 2)
	effectstest.TdAddCounter(g, opp.ID, theirs, game.CounterPlusOne, 1)

	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, rikku, 0, game.ActivateAbilityParams{
		Targets: append(effectstest.TdCard(theirs), effectstest.TdCard(mine)...),
	}); err != nil {
		t.Fatalf("activate Steal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, mine, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Fatalf("after Steal: mine %d, theirs %d; want 1 and 0",
			effectstest.CountersOn(g, mine, game.CounterPlusOne), effectstest.CountersOn(g, theirs, game.CounterPlusOne))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, mine); err == nil {
		t.Error("an opponent's creature blocked the creature Rikku's trigger covered")
	}
}

func TestTidusBSphereGridCountersConnectingCreaturesAndGrantsReachTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sphere Grid", "Enchantment", effectstest.TdSphereGridOracle, false)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Bear has %d counters after connecting, want 1", got)
	}
	for _, kw := range []string{"reach", "trample"} {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), kw) {
			t.Errorf("a creature with a +1/+1 counter has no %s", kw)
		}
	}
}

func TestTidusBTidusMovesACounterAndCheersOnceATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Tidus, Yuna's Guardian", "Legendary Creature — Human Warrior", effectstest.TdTidusOracle, 3, 3)
	from := effectstest.TdCreature(g, me.ID, "From", 2, 2)
	to := effectstest.TdCreature(g, me.ID, "To", 2, 2)
	effectstest.TdAddCounter(g, me.ID, from, game.CounterPlusOne, 1)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, from)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, to)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, from, game.CounterPlusOne) != 0 || effectstest.CountersOn(g, to, game.CounterPlusOne) != 1 {
		t.Fatalf("after the move: from %d, to %d; want 0 and 1",
			effectstest.CountersOn(g, from, game.CounterPlusOne), effectstest.CountersOn(g, to, game.CounterPlusOne))
	}

	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, to)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, to)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Cheer drew %d, want 1", me.Hand.Size()-hand)
	}
	if got := effectstest.CountersOn(g, to, game.CounterPlusOne); got != 2 {
		t.Errorf("after the proliferate the creature has %d counters, want 2", got)
	}
}

func TestTidusBTromellAddsACounterAndProliferatesPerEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tromell := effectstest.B12Push(g, me.ID, "Tromell, Seymour's Butler", "Legendary Creature — Elf Advisor", effectstest.TdTromellOracle, 2, 3)
	a := effectstest.TdCastCreature(t, g, "A", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	b := effectstest.TdCastCreature(t, g, "B", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, a, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, b, game.CounterPlusOne) != 1 {
		t.Fatal("a nontoken creature entered without the extra counter")
	}
	effectstest.PushToken(g, me.ID, game.Card{Name: "Token", TypeLine: "Token Creature — Test", Power: 1, Toughness: 1})

	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, tromell, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Tromell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, a)
	effectstest.AnswerProliferate(t, g, a)
	if effectstest.ProliferatePrompt(g) != nil {
		t.Error("a third proliferate was asked for two entries")
	}
	if got := effectstest.CountersOn(g, a, game.CounterPlusOne); got != 3 {
		t.Errorf("A has %d counters after two proliferates, want 3", got)
	}
}

func TestTidusBWakkaDestroysAnArtifactAndCaptainsTheTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	wakka := effectstest.B12Push(g, me.ID, "Wakka, Devoted Guardian", "Legendary Creature — Human Warrior", effectstest.TdWakkaOracle, 4, 4)
	friend := effectstest.TdCreature(g, me.ID, "Friend", 2, 2)
	relic := effectstest.B12Push(g, opp.ID, "Their Relic", "Artifact", "", 0, 0)

	effectstest.AttackWith(t, g, opp.ID, wakka)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
	if got := effectstest.CountersOn(g, wakka, game.CounterPlusOne); got != 1 {
		t.Fatalf("Wakka has %d counters, want 1", got)
	}
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, friend, game.CounterPlusOne); got != 1 {
		t.Errorf("the other creature has %d counters at the end step, want 1", got)
	}
	if got := effectstest.CountersOn(g, wakka, game.CounterPlusOne); got != 1 {
		t.Errorf("Wakka put a counter on himself (%d)", got)
	}
}
