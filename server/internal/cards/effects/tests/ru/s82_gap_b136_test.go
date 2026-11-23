package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b136_test.go — Dominaria Remastered GAP batch 136 (#5948). Each
// test plays the real catalog entry and asserts what its text says.

const (
	s82B136LynxOracle       = "b5187725-9a39-433c-ae09-eccd6111574c"
	s82B136TreefolkOracle   = "b7efcb42-aa52-4d13-8c7c-b2db2dd51afd"
	s82B136SeerOracle       = "8ba2f559-f9bd-48ef-9fef-6dffe768f4ed"
	s82B136GamekeeperOracle = "c6921cfe-48f9-4378-b145-d9b430c2aaf6"
	s82B136KamahlOracle     = "8b21146d-3e92-46e7-b5d9-8240fba5088e"
	s82B136PrimarchOracle   = "09537e9f-6c8e-4849-9d35-d4f3950a64d7"
	s82B136RestorerOracle   = "ea845951-0158-4646-acd9-3b0b28da40b1"
	s82B136CollectorOracle  = "4e1bc03c-e131-4422-beba-0a45a26fdf43"
	s82B136BobcatOracle     = "f0787397-8edb-4794-8ee1-22f97a057be8"
	s82B136InvokerOracle    = "9d277206-9156-4dc9-bdd2-9d9219787735"
)

func TestS82GapB136EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82B136LynxOracle:       "Spectral Lynx",
		s82B136TreefolkOracle:   "Deadwood Treefolk",
		s82B136SeerOracle:       "Fa'adiyah Seer",
		s82B136GamekeeperOracle: "Gamekeeper",
		s82B136KamahlOracle:     "Kamahl, Fist of Krosa",
		s82B136PrimarchOracle:   "Kavu Primarch",
		s82B136RestorerOracle:   "Krosan Restorer",
		s82B136CollectorOracle:  "Nut Collector",
		s82B136BobcatOracle:     "Penumbra Bobcat",
		s82B136InvokerOracle:    "Stonewood Invoker",
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

func TestS82GapB136SpectralLynxRegeneratesForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lynx := effectstest.B12Push(g, me.ID, "Spectral Lynx", "Creature — Cat Spirit", s82B136LynxOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, lynx, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("regenerated for no mana")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, lynx, 0, game.ActivateAbilityParams{}, "B")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(lynx); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(lynx) {
		t.Error("the regenerated Lynx was destroyed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, lynx, game.ProtectionFromColor("G")) {
		t.Error("the Lynx has no protection from green")
	}
}

func TestS82GapB136DeadwoodTreefolkReturnsAnotherCreatureOnEnteringAndLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	first := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	second := effectstest.PushGraveyardCardTyped(me, "Dead Elk", "Creature — Elk")
	treefolk := effectstest.DftABCast(t, g, me, game.Card{Name: "Deadwood Treefolk", TypeLine: "Creature — Treefolk",
		OracleID: s82B136TreefolkOracle, Power: 3, Toughness: 6})
	effectstest.PickCard(t, g, me.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, first) {
		t.Fatal("the entering trigger did not return the chosen creature card")
	}
	if got := effectstest.Pr8Counters(g, treefolk, game.CounterTime); got != 3 {
		t.Errorf("%d time counters, want 3 (vanishing 3)", got)
	}
	effectstest.KillCreature(t, g, me.ID, treefolk)
	g.RunStateChecksForTest()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the leaves trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, treefolk) {
		t.Error("the Treefolk itself was offered as 'another' target")
	}
	effectstest.PickCard(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, second) {
		t.Error("the leaves trigger did not return the chosen creature card")
	}
}

func TestS82GapB136FaadiyahSeerKeepsALandAndDiscardsAnythingElse(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		kept           bool
	}{
		{"land", "Basic Land — Forest", true},
		{"nonland", "Sorcery", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			seer := effectstest.B12Push(g, me.ID, "Fa'adiyah Seer", "Creature — Human Shaman", s82B136SeerOracle, 1, 1)
			top := effectstest.SeedLibraryTop(me, "Drawn", tc.typeLine, "")
			effectstest.AdvanceToMain(t, g)
			effectstest.RavB2MustActivate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{})
			if got := effectstest.HandHasID(me, top); got != tc.kept {
				t.Errorf("drawn card in hand = %v, want %v", got, tc.kept)
			}
			if got := me.Graveyard.Contains(top); got == tc.kept {
				t.Errorf("drawn card in graveyard = %v, want %v", got, !tc.kept)
			}
		})
	}
}

func TestS82GapB136GamekeeperExilesItselfAndPutsTheFirstCreatureIntoPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keeper := effectstest.B12Push(g, me.ID, "Gamekeeper", "Creature — Elf", s82B136GamekeeperOracle, 2, 2)
	me.Library.Cards = nil
	deep := effectstest.SeedLibraryTop(me, "Deep Bear", "Creature — Bear", "{1}{G}")
	hit := effectstest.SeedLibraryTop(me, "Hit Elk", "Creature — Elk", "{2}")
	skipOne := effectstest.SeedLibraryTop(me, "Skipped Sorcery", "Sorcery", "{1}")
	skipTwo := effectstest.SeedLibraryTop(me, "Skipped Land", "Basic Land — Forest", "")
	effectstest.KillCreature(t, g, me.ID, keeper)
	g.RunStateChecksForTest()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(keeper) {
		t.Error("the Gamekeeper was not exiled")
	}
	if !g.Battlefield.Contains(hit) {
		t.Error("the first creature card did not enter the battlefield")
	}
	if !me.Graveyard.Contains(skipOne) || !me.Graveyard.Contains(skipTwo) {
		t.Error("the other revealed cards did not go to the graveyard")
	}
	if !me.Library.Contains(deep) {
		t.Error("a card below the revealed ones left the library")
	}
}

func TestS82GapB136GamekeeperDeclinedStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keeper := effectstest.B12Push(g, me.ID, "Gamekeeper", "Creature — Elf", s82B136GamekeeperOracle, 2, 2)
	hit := effectstest.SeedLibraryTop(me, "Hit Elk", "Creature — Elk", "{2}")
	effectstest.KillCreature(t, g, me.ID, keeper)
	g.RunStateChecksForTest()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(keeper) || g.Battlefield.Contains(hit) {
		t.Error("declining the exile still revealed cards")
	}
}

func TestS82GapB136KamahlAnimatesALandAndPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kamahl := effectstest.B12Push(g, me.ID, "Kamahl, Fist of Krosa", "Legendary Creature — Human Druid", s82B136KamahlOracle, 4, 3)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, kamahl, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(forest)}, "G")
	c := effectstest.FindBattlefieldCardForTest(g, forest)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("the Forest is creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if p, tg := effectstest.EffectivePower(t, g, forest), effectstest.EffectiveToughness(t, g, forest); p != 1 || tg != 1 {
		t.Errorf("the animated Forest is %d/%d, want 1/1", p, tg)
	}
	effectstest.RavB2MustActivate(t, g, me.ID, kamahl, 1, game.ActivateAbilityParams{}, "G", "G", "G", "G", "G")
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("my Bear has power %d, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Ox has power %d, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("my Bear has no trample")
	}
}

func TestS82GapB136KavuPrimarchEntersWithFourCountersOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		pool     int
		want     int
	}{
		{"unkicked", nil, 0, 0},
		{"kicked", []int{0}, 4, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			var elves []uuid.UUID
			for i := 0; i < 4; i++ {
				elves = append(elves, effectstest.B31Push(g, me.ID, "Elf", "Creature — Elf", "", "{G}", 1, 1, "G"))
			}
			card := game.Card{Name: "Kavu Primarch", TypeLine: "Creature — Kavu", ManaCost: "{3}{G}",
				OracleID: s82B136PrimarchOracle, Power: 3, Toughness: 3}
			if err := effectstest.SocB19CastWithPool(g, me, card, "G", tc.pool, game.CastSpellParams{TapIDs: elves, OptionalCosts: tc.optional}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			var id uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Kavu Primarch" {
					id = c.InstanceID
				}
			}
			if id == uuid.Nil {
				t.Fatal("the Primarch did not enter")
			}
			if n := effectstest.Pr8Counters(g, id, game.CounterPlusOne); n != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", n, tc.want)
			}
			for _, e := range elves {
				if !effectstest.TappedForTest(t, g, e) {
					t.Error("a convoking creature was not tapped")
				}
			}
		})
	}
}

func TestS82GapB136KrosanRestorerUntapsOneLandAndThreeWithThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	restorer := effectstest.B12Push(g, me.ID, "Krosan Restorer", "Creature — Human Druid", s82B136RestorerOracle, 1, 2)
	var lands []uuid.UUID
	for i := 0; i < 4; i++ {
		l := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		g.WithWriteLock(func() { effectstest.B16Tap(g, l) })
		lands = append(lands, l)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, restorer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lands[0])})
	if effectstest.TappedForTest(t, g, lands[0]) {
		t.Error("the targeted land is still tapped")
	}
	effectstest.Untap(g, restorer)
	if err := effectstest.RavB2Activate(t, g, me.ID, restorer, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lands[1], lands[2], lands[3])}); err == nil {
		t.Fatal("the threshold ability activated with an empty graveyard")
	}
	for i := 0; i < 7; i++ {
		effectstest.PushGraveyardCardForTest(me, "Filler")
	}
	effectstest.Untap(g, restorer)
	effectstest.RavB2MustActivate(t, g, me.ID, restorer, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lands[1], lands[2], lands[3])})
	for _, l := range lands[1:] {
		if effectstest.TappedForTest(t, g, l) {
			t.Errorf("land %s is still tapped", l)
		}
	}
}

func TestS82GapB136NutCollectorMakesSquirrelsAndPumpsAllSquirrelsAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	collector := effectstest.B12Push(g, me.ID, "Nut Collector", "Creature — Human Druid", s82B136CollectorOracle, 1, 1)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Squirrel", "Creature — Squirrel", 1, 1)
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Fatalf("a Squirrel has power %d before threshold, want 1", p)
	}
	effectstest.BlbMPFillGraveyard(g, me, 7)
	if p := effectstest.EffectivePower(t, g, theirs); p != 3 {
		t.Errorf("an opponent's Squirrel has power %d at threshold, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, collector); p != 1 {
		t.Errorf("the Collector has power %d, want 1: it is no Squirrel", p)
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Squirrel"); n != 1 {
		t.Errorf("%d Squirrel tokens, want 1", n)
	}
}

func TestS82GapB136PenumbraBobcatLeavesACatWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bobcat := effectstest.B12Push(g, me.ID, "Penumbra Bobcat", "Creature — Cat", s82B136BobcatOracle, 2, 1)
	effectstest.KillCreature(t, g, me.ID, bobcat)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && effects.IsToken(c) && c.Name == "Cat" {
			id = c.InstanceID
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Cat tokens, want 1", n)
	}
	if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tg != 1 {
		t.Errorf("the Cat is %d/%d, want 2/1", p, tg)
	}
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if !c.HasColor("B") {
		t.Error("the Cat token is not black")
	}
}

func TestS82GapB136StonewoodInvokerGetsPlusFiveForEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	invoker := effectstest.B12Push(g, me.ID, "Stonewood Invoker", "Creature — Elf Mutant", s82B136InvokerOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, invoker, 0, game.ActivateAbilityParams{}, "G", "G", "G"); err == nil {
		t.Fatal("activated for three mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.RavB2MustActivate(t, g, me.ID, invoker, 0, game.ActivateAbilityParams{}, "G", "G", "G", "G", "G", "G", "G", "G")
	if p, tg := effectstest.EffectivePower(t, g, invoker), effectstest.EffectiveToughness(t, g, invoker); p != 7 || tg != 7 {
		t.Errorf("the Invoker is %d/%d, want 7/7", p, tg)
	}
}
