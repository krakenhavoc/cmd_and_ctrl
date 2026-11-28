package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b768_cards_test.go — S82 OLD batch 768 (#5628, Tempest: enters and
// dies triggers). Each test plays the real catalog entry.

const (
	oldB768Greven      = "08fa1bee-cedf-4a29-9b02-cbdcc0926f8f"
	oldB768Dracoplasm  = "3155c833-8170-4166-8f47-c3c1fad1575a"
	oldB768Angel       = "3415a199-05ac-4627-be97-f842eb1415a3"
	oldB768RathiDragon = "6a180f07-7afa-4605-9c58-4521a5f3fcc1"
	oldB768WildWurm    = "ba14d185-9432-4a67-8bfb-544c2cebe71d"
	oldB768Pallimud    = "cad0f865-62e0-47ba-84ae-8e97cb52bbd1"
)

func TestOldB768EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB768Greven:      "Commander Greven il-Vec",
		oldB768Dracoplasm:  "Dracoplasm",
		oldB768Angel:       "Avenging Angel",
		oldB768RathiDragon: "Rathi Dragon",
		oldB768WildWurm:    "Wild Wurm",
		oldB768Pallimud:    "Pallimud",
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

func TestOldB768GrevenHasFearAndSacrificesACreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	greven := effectstest.CastCatalogSpell(t, g, "Commander Greven il-Vec", "Legendary Creature — Phyrexian Human Warrior", oldB768Greven, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(greven) {
		t.Error("the trigger did not sacrifice the creature chosen")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, greven, "fear") {
		t.Error("Greven has no fear")
	}
}

func TestOldB768GrevenAloneSacrificesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	greven := effectstest.CastCatalogSpell(t, g, "Commander Greven il-Vec", "Legendary Creature — Phyrexian Human Warrior", oldB768Greven, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, greven)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(greven) {
		t.Error("Greven, the only creature, was not sacrificed")
	}
}

func TestOldB768AvengingAngelMayGoOnTopOfItsOwnersLibrary(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		angel := effectstest.B12Push(g, me.ID, "Avenging Angel", "Creature — Angel", oldB768Angel, 3, 3)
		effectstest.CastFromHandForTest(t, g, me, "Lightning Bolt", "Instant", "", effectstest.LightningBoltOracle, game.CastSpellParams{Targets: effectstest.CardRefs(angel)})
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
		effectstest.PassPriorityAroundTable(t, g)
		want := game.ZoneGraveyard
		if accept {
			want = game.ZoneLibrary
		}
		if got := effectstest.ZoneOf(g, angel); got != want {
			t.Errorf("accept=%v: the Angel is in %s, want %s", accept, got, want)
		}
		if accept && effectstest.LibraryTopIDs(me, 1)[0] != angel {
			t.Error("the Angel is not on top of the library")
		}
	}
}

func TestOldB768WildWurmReturnsToHandOnALostFlip(t *testing.T) {
	sawKeep, sawLose := false, false
	for i := 0; i < 80 && !(sawKeep && sawLose); i++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		wurm := effectstest.CastCatalogSpell(t, g, "Wild Wurm", "Creature — Wurm", oldB768WildWurm, nil)
		effectstest.PassPriorityAroundTable(t, g)
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("no coin call as the Wurm enters")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		effectstest.PassPriorityAroundTable(t, g)
		if won {
			sawKeep = true
			if !g.Battlefield.Contains(wurm) {
				t.Error("won the flip but the Wurm returned")
			}
		} else {
			sawLose = true
			if g.Battlefield.Contains(wurm) || !me.Hand.Contains(wurm) {
				t.Error("lost the flip but the Wurm did not return to its owner's hand")
			}
		}
	}
	if !sawKeep || !sawLose {
		t.Errorf("never saw both outcomes of the flip (keep %v, lose %v)", sawKeep, sawLose)
	}
}

func oldB768Mountains(g *game.Game, me *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}))
	}
	return out
}

func TestOldB768RathiDragonIsSacrificedWithoutTwoMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB768Mountains(g, me, 1)
	dragon := effectstest.CastCatalogSpell(t, g, "Rathi Dragon", "Creature — Dragon", oldB768RathiDragon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) {
		t.Error("Rathi Dragon stayed with only one Mountain")
	}
	if n := effectstest.NamedOnBattlefield(g, "Mountain", me.ID); n != 1 {
		t.Errorf("%d Mountains left, want the one untouched", n)
	}
}

func TestOldB768RathiDragonStaysForTwoSacrificedMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountains := oldB768Mountains(g, me, 3)
	dragon := effectstest.CastCatalogSpell(t, g, "Rathi Dragon", "Creature — Dragon", oldB768RathiDragon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: mountains[:2]})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Fatal("the Dragon was sacrificed after two Mountains were")
	}
	if g.Battlefield.Contains(mountains[0]) || g.Battlefield.Contains(mountains[1]) || !g.Battlefield.Contains(mountains[2]) {
		t.Error("the wrong Mountains were sacrificed")
	}
}

func TestOldB768RathiDragonMaySacrificeItselfInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountains := oldB768Mountains(g, me, 2)
	dragon := effectstest.CastCatalogSpell(t, g, "Rathi Dragon", "Creature — Dragon", oldB768RathiDragon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) || !g.Battlefield.Contains(mountains[0]) || !g.Battlefield.Contains(mountains[1]) {
		t.Error("choosing the Dragon should sacrifice only the Dragon")
	}
}

func TestOldB768PallimudCountsTheChosenPlayersTappedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	chosen := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	land := func(p *game.Player, tapped bool) uuid.UUID {
		return effectstest.PushTestPermanent(g, p.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island", Tapped: tapped})
	}
	land(chosen, true)
	land(chosen, true)
	untapped := land(chosen, false)
	land(other, true)
	land(me, true)
	effectstest.AdvanceToMain(t, g)
	pallimud := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: pallimud, Name: "Pallimud", TypeLine: "Creature — Beast", OracleID: oldB768Pallimud,
			Power: 0, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, pallimud, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Pallimud: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.PtOf(t, g, pallimud); p != 0 || th != 3 {
		t.Errorf("before a player is chosen Pallimud is %d/%d, want 0/3", p, th)
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, chosen)
	if p, th := effectstest.PtOf(t, g, pallimud); p != 2 || th != 3 {
		t.Errorf("Pallimud is %d/%d, want 2/3 (two tapped lands the chosen player controls)", p, th)
	}
	if err := g.TapCard(untapped, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if p, _ := effectstest.PtOf(t, g, pallimud); p != 3 {
		t.Errorf("after another land taps Pallimud has power %d, want 3", p)
	}
}

func TestOldB768DracoplasmBecomesTheTotalOfTheCreaturesSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	b := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 1)
	keep := effectstest.B12Creature(g, me.ID, "Kept", "Creature — Bear", 7, 7)
	id := effectstest.CastCatalogSpell(t, g, "Dracoplasm", "Creature — Shapeshifter", oldB768Dracoplasm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Dracoplasm did not enter")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(keep) {
		t.Error("the wrong creatures were sacrificed")
	}
	if p, th := effectstest.PtOf(t, g, id); p != 6 || th != 4 {
		t.Errorf("Dracoplasm is %d/%d, want 6/4", p, th)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("Dracoplasm lost flying")
	}
	// {R}: +1/+0 until end of turn, on top of the base size.
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.PtOf(t, g, id); p != 7 || th != 4 {
		t.Errorf("after the pump Dracoplasm is %d/%d, want 7/4", p, th)
	}
}

func TestOldB768DracoplasmReadsLastKnownSizeOfPumpedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, a).Counters = map[string]int{game.CounterPlusOne: 2}
	})
	id := effectstest.CastCatalogSpell(t, g, "Dracoplasm", "Creature — Shapeshifter", oldB768Dracoplasm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.PtOf(t, g, id); p != 3 || th != 3 {
		t.Errorf("Dracoplasm is %d/%d, want 3/3 (the +1/+1 counters count)", p, th)
	}
}

func TestOldB768DracoplasmSacrificingNothingLeavesItAZeroZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Kept", "Creature — Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Dracoplasm", "Creature — Shapeshifter", oldB768Dracoplasm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	// A 0/0 dies as a state-based action (CR 704.5f); the seeded card has no
	// printing behind it, so the engine may leave it standing, but it is never
	// anything but 0/0.
	if g.Battlefield.Contains(id) {
		if p, th := effectstest.PtOf(t, g, id); p != 0 || th != 0 {
			t.Errorf("Dracoplasm that sacrificed nothing is %d/%d, want 0/0", p, th)
		}
	}
}
