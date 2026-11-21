package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b12_test.go — the Phyrexia: All Will Be One oil cards of S82
// batch 12 (#5479), on the real engine.

const (
	oneB12Archfiend  = "9188031e-1c24-4981-a3d2-461a2bf917b6"
	oneB12Spinoderm  = "482ec736-7d4a-49c6-a757-e134123329ca"
	oneB12Wellspring = "56a55f45-2cc6-407d-b212-5619c98456d9"
	oneB12Tablet     = "6eca2ac9-16ba-4a11-b6ee-3caf2d7d8254"
	oneB12Forge      = "7c4b2203-a937-4d91-8e39-2283fe686b45"
	oneB12Sylex      = "38347ab3-ae43-4d10-b665-8520ace0eb76"
	oneB12Reservoir  = "9d1c5987-3572-4bf9-b01c-eb6e98a7da77"
	oneB12Migloz     = "85eb890e-a5e1-48f8-b813-e8424d18b28a"
	oneB12Factory    = "c5b0d98b-9743-488e-a3be-07043d6ad78a"
)

func TestOneB12CardsAreRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oneB12Archfiend: "Archfiend of the Dross", oneB12Spinoderm: "Evolved Spinoderm",
		oneB12Wellspring: "Norn's Wellspring", oneB12Tablet: "Tablet of Compleation",
		oneB12Forge: "Urabrask's Forge", oneB12Sylex: "The Filigree Sylex",
		oneB12Reservoir: "Churning Reservoir", oneB12Migloz: "Migloz, Maze Crusher",
		oneB12Factory: "Vulshok Factory",
	} {
		s, ok := effects.Lookup(id)
		if !ok || s.Name != name || s.Completeness != effects.CompletenessFull || len(s.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, s.Name)
		}
	}
}

func oneB12Oil(g *game.Game, id uuid.UUID) int { return effectstest.CounterCount(g, id, "oil") }

func oneB12UpkeepResolves(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.FadingToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOneB12ArchfiendLosesTheGameWhenTheLastOilCounterGoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B38CastCreature(t, g, "Archfiend of the Dross", "Creature — Phyrexian Demon", oneB12Archfiend, 6, 6)
	if got := oneB12Oil(g, id); got != 4 {
		t.Fatalf("entered with %d oil counters, want 4", got)
	}
	for _, want := range []int{3, 2, 1} {
		oneB12UpkeepResolves(t, g)
		if got := oneB12Oil(g, id); got != want {
			t.Fatalf("oil counters %d, want %d", got, want)
		}
		if me.Eliminated {
			t.Fatal("lost the game with oil counters left")
		}
	}
	oneB12UpkeepResolves(t, g)
	if !me.Eliminated {
		t.Error("the controller did not lose the game with no oil counters on the Archfiend")
	}
}

func TestOneB12ArchfiendDrainsTheControllerOfADyingOpposingCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B38CastCreature(t, g, "Archfiend of the Dross", "Creature — Phyrexian Demon", oneB12Archfiend, 6, 6)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("my own creature dying changed life: me %d→%d opp %d→%d", myLife, me.Life, oppLife, opp.Life)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, oppLife-2)
	}
	if me.Life != myLife {
		t.Errorf("my life changed to %d", me.Life)
	}
}

func TestOneB12SpinodermHexproofAboveTwoTrampleAtTwoOrFewerThenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B38CastCreature(t, g, "Evolved Spinoderm", "Creature — Phyrexian Beast", oneB12Spinoderm, 5, 5)
	if got := oneB12Oil(g, id); got != 4 {
		t.Fatalf("entered with %d oil counters, want 4", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "hexproof") || effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("with four oil counters: want hexproof and no trample")
	}
	oneB12UpkeepResolves(t, g) // 3
	if !effectstest.HasEffectiveKeyword(t, g, id, "hexproof") || effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("with three oil counters: want hexproof and no trample")
	}
	oneB12UpkeepResolves(t, g) // 2
	if effectstest.HasEffectiveKeyword(t, g, id, "hexproof") || !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("with two oil counters: want trample and no hexproof")
	}
	oneB12UpkeepResolves(t, g) // 1
	oneB12UpkeepResolves(t, g) // 0 → sacrificed
	if _, here := effectstest.BattlefieldCard(g, id); here {
		t.Fatal("the Spinoderm survived with no oil counters")
	}
	if !effectstest.InGraveyard(me, id) {
		t.Error("the Spinoderm is not in its owner's graveyard")
	}
}

func TestOneB12WellspringScriesAndCountsOnlyYourCreaturesAndPaysTwoForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CeSeedLibrary(g, me.ID, 10)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Norn's Wellspring", "Artifact", oneB12Wellspring, false)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oneB12Oil(g, id); got != 0 {
		t.Fatalf("an opponent's creature died and the Wellspring has %d oil counters", got)
	}
	for i := 0; i < 2; i++ {
		mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 1, 1)
		effectstest.KillCreature(t, g, me.ID, mine)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerScryKeepAll(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := oneB12Oil(g, id); got != 2 {
		t.Fatalf("oil counters %d, want 2", got)
	}
	effectstest.B06AddMana(me, "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if got := oneB12Oil(g, id); got != 0 {
		t.Errorf("oil counters %d after paying two, want 0", got)
	}
	effectstest.B22Untap(g, id)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no oil counters to remove")
	}
}

func TestOneB12TabletGatesItsManaAndItsCardOnOilCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Tablet of Compleation", "Artifact", oneB12Tablet, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("tapped for {C} with no oil counters")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("drew with no oil counters")
	}
	effectstest.PutCounters(t, g, id, "oil", 1)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("tapped for {C} with one oil counter")
	}
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}) // {T}: oil counter
	if got := oneB12Oil(g, id); got != 2 {
		t.Fatalf("oil counters %d, want 2", got)
	}
	effectstest.B22Untap(g, id)
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C} with two oil counters: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	effectstest.B22Untap(g, id)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("drew with two oil counters")
	}
	effectstest.PutCounters(t, g, id, "oil", 3)
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if got := oneB12Oil(g, id); got != 5 {
		t.Errorf("drawing removes no counters: %d, want 5", got)
	}
}

func TestOneB12ForgeMakesAGrowingHasteTramplerAndSacrificesItAtEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Urabrask's Forge", "Artifact", oneB12Forge, false)
	for turn := 1; turn <= 2; turn++ {
		for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepBeginCombat && effectstest.FadingTriggerWaiting(g, id)); i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := oneB12Oil(g, id); got != turn {
			t.Fatalf("turn %d: oil counters %d", turn, got)
		}
		var horror uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Phyrexian Horror" {
				horror = c.InstanceID
				if c.Power != turn || c.Toughness != 1 {
					t.Fatalf("Horror is %d/%d, want %d/1", c.Power, c.Toughness, turn)
				}
			}
		}
		if horror == uuid.Nil {
			t.Fatalf("turn %d: no Horror token", turn)
		}
		if !effectstest.HasEffectiveKeyword(t, g, horror, "trample") || !effectstest.HasEffectiveKeyword(t, g, horror, "haste") {
			t.Error("the Horror lacks trample or haste")
		}
		for i := 0; i < 400 && g.Turn.Step != game.StepEnd; i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if _, here := effectstest.BattlefieldCard(g, horror); here {
			t.Fatalf("turn %d: the Horror was not sacrificed at the end step", turn)
		}
		for i := 0; i < 400 && g.Turn.ActiveSeat == 0; i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
	}
}

func TestOneB12SylexSweepsByOilAndRefusesTheBlastWithoutTenCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "The Filigree Sylex", "Legendary Artifact", oneB12Sylex, false)
	effectstest.PutCounters(t, g, id, "oil", 1)
	// Mana value is read from the printed cost: seed permanents with known values.
	token := effectstest.PushVanillaCreature(g, opp.ID, "Token Bear", 2, 2) // mana value 0
	land := effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}) // {T}: oil
	if got := oneB12Oil(g, id); got != 2 {
		t.Fatalf("oil counters %d, want 2", got)
	}
	effectstest.B22Untap(g, id)
	if err := g.ActivateCatalogAbility(me.ID, id, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("dealt 10 damage with only two oil counters")
	}
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{})
	if _, here := effectstest.BattlefieldCard(g, id); here {
		t.Error("the Sylex was not sacrificed")
	}
	if _, here := effectstest.BattlefieldCard(g, land); !here {
		t.Error("a land was destroyed")
	}
	if _, here := effectstest.BattlefieldCard(g, token); !here {
		t.Error("a mana value 0 creature was destroyed by a sweep for mana value 2")
	}
}

func TestOneB12SylexDealsTenDamageByRemovingTenOilFromAmongYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "The Filigree Sylex", "Legendary Artifact", oneB12Sylex, false)
	other := effectstest.PushVanillaCreature(g, me.ID, "Oily Bear", 2, 2)
	effectstest.PutCounters(t, g, id, "oil", 4)
	effectstest.PutCounters(t, g, other, "oil", 6)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, id, 2, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{id, other},
		CounterCounts:    []int{4, 6},
		Targets:          []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-10 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-10)
	}
	if _, here := effectstest.BattlefieldCard(g, id); here {
		t.Error("the Sylex was not sacrificed")
	}
	if got := oneB12Oil(g, other); got != 0 {
		t.Errorf("the bear kept %d oil counters", got)
	}
}

func TestOneB12ReservoirGatesTheGoblinOnAnOilRemovalOrAnOilPermanentDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Churning Reservoir", "Artifact", oneB12Reservoir, false)
	tryGoblin := func() error {
		effectstest.B06AddMana(me, "C", "C")
		return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{})
	}
	if err := tryGoblin(); err == nil {
		t.Fatal("made a Goblin with no oil event this turn")
	}
	// Oil placed (not removed) does not open the gate.
	bear := effectstest.PushVanillaCreature(g, me.ID, "Oily Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, "oil", 2)
	if err := tryGoblin(); err == nil {
		t.Fatal("made a Goblin after only PLACING oil counters")
	}
	// An oil counter removed from an opponent's permanent does not count.
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PutCounters(t, g, theirs, "oil", 1)
	if err := g.AddCounter(theirs, "oil", -1); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := tryGoblin(); err == nil {
		t.Fatal("made a Goblin after an oil counter came off an opponent's permanent")
	}
	// A counter removed from a permanent I control does.
	if err := g.AddCounter(bear, "oil", -1); err != nil {
		t.Fatalf("remove: %v", err)
	}
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Goblin" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Phyrexian Goblins, want 1", n)
	}
}

func TestOneB12ReservoirCountsAnyPermanentWithOilPutIntoAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Churning Reservoir", "Artifact", oneB12Reservoir, false)
	plain := effectstest.PushVanillaCreature(g, opp.ID, "Plain Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, plain)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a permanent with no oil dying opened the gate")
	}
	oily := effectstest.PushVanillaCreature(g, opp.ID, "Oily Bear", 2, 2)
	effectstest.PutCounters(t, g, oily, "oil", 3)
	effectstest.KillCreature(t, g, opp.ID, oily)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	found := false
	for _, c := range g.Battlefield.Cards {
		found = found || c.Name == "Phyrexian Goblin"
	}
	if !found {
		t.Error("no Goblin after an opposing permanent with oil was put into a graveyard")
	}
}

func TestOneB12ReservoirUpkeepPutsOilOnAnotherNontokenArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Churning Reservoir", "Artifact", oneB12Reservoir, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.FadingToUpkeepOfSeat0(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oneB12Oil(g, bear); got != 1 {
		t.Errorf("the bear has %d oil counters, want 1", got)
	}
	if got := oneB12Oil(g, id); got != 0 {
		t.Errorf("the Reservoir targeted itself: %d oil counters", got)
	}
}

func TestOneB12MiglozPaysItsCountersUpFront(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B38CastCreature(t, g, "Migloz, Maze Crusher", "Legendary Creature — Phyrexian Beast", oneB12Migloz, 4, 4)
	if got := oneB12Oil(g, id); got != 5 {
		t.Fatalf("entered with %d oil counters, want 5", got)
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
		t.Error("Migloz lacks vigilance or menace")
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{})
	if p, tt := effectstest.B38PT(t, g, id); p != 6 || tt != 6 {
		t.Errorf("Migloz is %d/%d, want 6/6", p, tt)
	}
	if got := oneB12Oil(g, id); got != 2 {
		t.Fatalf("oil counters %d, want 2", got)
	}
	effectstest.B06AddMana(me, "C", "C", "C")
	target := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Opposing Relic", "Artifact", "", false)
	if err := g.ActivateCatalogAbility(me.ID, id, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Fatal("paid three counters with two on Migloz")
	}
	effectstest.PutCounters(t, g, id, "oil", 1)
	effectstest.B16Activate(t, g, me.ID, id, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if _, here := effectstest.BattlefieldCard(g, target); here {
		t.Error("the target artifact survived")
	}
	if got := oneB12Oil(g, id); got != 0 {
		t.Errorf("oil counters %d, want 0", got)
	}
}

func TestOneB12FactoryBanksChargeAndSacrificesIntoAHastyGolemAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Vulshok Factory", "Artifact", oneB12Factory, false)
	for i := 0; i < 2; i++ {
		me.ManaPool.EmptyPool()
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap for {R}: %v", err)
		}
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
			t.Fatalf("pool %v, want [R]", got)
		}
		effectstest.B22Untap(g, id)
	}
	if got := effectstest.CounterCount(g, id, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters %d, want 2", got)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "C", "C")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var golem uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Golem" {
			golem = c.InstanceID
		}
	}
	if golem == uuid.Nil {
		t.Fatal("no Golem token")
	}
	if p, tt := effectstest.B38PT(t, g, golem); p != 2 || tt != 2 {
		t.Errorf("Golem is %d/%d, want 2/2", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, golem, "haste") {
		t.Error("the Golem lacks haste")
	}
	if _, here := effectstest.BattlefieldCard(g, id); here {
		t.Error("the Factory was not sacrificed")
	}
}

func TestOneB12FactoryCannotMakeAGolemOutsideSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Vulshok Factory", "Artifact", oneB12Factory, false)
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 1, 1)
	effectstest.B06AddMana(me, "R", "C", "C")
	for i := 0; i < 400 && g.Turn.Step != game.StepBeginCombat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated the sacrifice ability during combat")
	}
}
