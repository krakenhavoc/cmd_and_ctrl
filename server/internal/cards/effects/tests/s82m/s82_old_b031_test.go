package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 31 (S82, #5501): Modern Horizons 2 — Tourach, Dread Cantor,
// Glinting Creeper, Caprichrome, Steel Dromedary, Graceful Restoration,
// Arcus Acolyte, Flametongue Yearling, Flay Essence, Monoskelion, Breathless
// Knight and Suspend. Tizerus Charger is skipped (see the commit).

const (
	oldB031Tourach    = "0c952586-e522-4982-bebd-96f1db73053d"
	oldB031Creeper    = "0e252f94-7731-45e1-bdbb-1bd11441577e"
	oldB031Caprichrom = "27831489-b444-4b7b-91db-feb54281bca9"
	oldB031Dromedary  = "40fa5656-354f-47f9-ba8d-40291ee5b40e"
	oldB031Graceful   = "59bfeacc-f80b-49fc-8a35-f7e1ad905161"
	oldB031Acolyte    = "628f2deb-519e-4810-a7e5-e31316e9a4e0"
	oldB031Yearling   = "8181f131-626b-4f9a-812c-9fddaae58025"
	oldB031FlayEss    = "b233a9e4-0b4c-41e8-a5b8-77824241966b"
	oldB031Monoskel   = "c16d79d9-c159-421f-90a2-13313a1de5f6"
	oldB031Knight     = "e15fb040-2e11-478e-87f1-55675d614f8e"
	oldB031Suspend    = "fef9425d-82d5-470b-9cbc-8c03a001099e"
)

func TestOldB031CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB031Tourach:    "Tourach, Dread Cantor",
		oldB031Creeper:    "Glinting Creeper",
		oldB031Caprichrom: "Caprichrome",
		oldB031Dromedary:  "Steel Dromedary",
		oldB031Graceful:   "Graceful Restoration",
		oldB031Acolyte:    "Arcus Acolyte",
		oldB031Yearling:   "Flametongue Yearling",
		oldB031FlayEss:    "Flay Essence",
		oldB031Monoskel:   "Monoskelion",
		oldB031Knight:     "Breathless Knight",
		oldB031Suspend:    "Suspend",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// oldB031Counter adds n (or removes -n) counters of a kind through the engine.
func oldB031Counter(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, kind, n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

func oldB031Card(name, typeLine, cost, oracle string, p, tough int) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Layout: "normal", Power: p, Toughness: tough}
}

// --- Tourach, Dread Cantor -------------------------------------------------

func TestOldB031TourachGrowsWhenAnOpponentDiscardsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tourach := effectstest.B12Push(g, me.ID, "Tourach, Dread Cantor", "Legendary Creature — Human Cleric", oldB031Tourach, 2, 1)
	effectstest.HandCardFull(me, "My Fodder", "Sorcery", "{1}", "", nil)
	effectstest.B30DiscardOne(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, tourach, game.CounterPlusOne); n != 0 {
		t.Fatalf("my own discard gave %d counters", n)
	}
	effectstest.HandCardFull(opp, "Their Fodder", "Sorcery", "{1}", "", nil)
	effectstest.B30DiscardOne(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, tourach, game.CounterPlusOne); n != 1 {
		t.Errorf("an opponent's discard gave %d counters, want 1", n)
	}
}

func TestOldB031TourachKickedMakesAnOpponentDiscardTwoAtRandom(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		opp.Hand.Cards = nil
		for i := 0; i < 4; i++ {
			effectstest.HandCardFull(opp, "Card", "Sorcery", "{1}", "", nil)
		}
		params := game.CastSpellParams{}
		if kicked {
			params.OptionalCosts = []int{0}
		}
		id, err := effectstest.Msc17Cast(t, g, me, oldB031Card("Tourach, Dread Cantor", "Legendary Creature — Human Cleric", "{1}{B}", oldB031Tourach, 2, 1), params)
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if kicked {
			effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
		}
		want := 4
		if kicked {
			want = 2
		}
		if got := opp.Hand.Size(); got != want {
			t.Errorf("kicked=%v: the opponent holds %d cards, want %d", kicked, got, want)
		}
		wantCounters := 0
		if kicked {
			wantCounters = 2 // the two discards each grow Tourach
		}
		if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != wantCounters {
			t.Errorf("kicked=%v: %d counters on Tourach, want %d", kicked, n, wantCounters)
		}
	}
}

func TestOldB031TourachHasProtectionFromWhite(t *testing.T) {
	spec, _ := effects.Lookup(oldB031Tourach)
	found := false
	for _, k := range spec.PrintedKeywords {
		if k == "protection from white" {
			found = true
		}
	}
	if !found {
		t.Errorf("printed keywords %v lack protection from white", spec.PrintedKeywords)
	}
}

// --- Glinting Creeper ------------------------------------------------------

func TestOldB031GlintingCreeperEntersWithTwoCountersPerColorSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "GWUBB")
	id := effectstest.CastFromHandForTest(t, g, me, "Glinting Creeper", "Creature — Plant", "{4}{G}", oldB031Creeper, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	// G, W, U and B were spent: four colours, two counters each.
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 8 {
		t.Errorf("%d counters, want 8", n)
	}
}

func TestOldB031GlintingCreeperCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	creeper := effectstest.B12Push(g, me.ID, "Glinting Creeper", "Creature — Plant", oldB031Creeper, 0, 0)
	oldB031Counter(t, g, creeper, game.CounterPlusOne, 4)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 3, 3)
	can := func(blocker uuid.UUID) (ok bool) {
		g.WithWriteLock(func() {
			ok = g.CanBlockLocked(effectstest.FindBattlefieldCardByID(g, creeper), effectstest.FindBattlefieldCardByID(g, blocker))
		})
		return ok
	}
	if can(small) {
		t.Error("a power-2 creature can block the Creeper")
	}
	if !can(big) {
		t.Error("a power-3 creature cannot block the Creeper")
	}
}

// --- Caprichrome -----------------------------------------------------------

func TestOldB031CaprichromeDevoursArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	bear := effectstest.PushDevourFodder(g, me.ID, "Goblin")
	id := effectstest.CastDevourer(t, g, "Caprichrome", oldB031Caprichrom)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, rock)
	if got := effectstest.DevourPlusOnes(g, id); got != 1 {
		t.Fatalf("counters = %d, want 1 (devour artifact 1, one artifact)", got)
	}
	if g.Battlefield.Contains(rock) || !g.Battlefield.Contains(bear) {
		t.Errorf("rock on battlefield %v, goblin on battlefield %v: only the artifact is devoured", g.Battlefield.Contains(rock), g.Battlefield.Contains(bear))
	}
}

// --- Steel Dromedary -------------------------------------------------------

func TestOldB031SteelDromedaryEntersTappedWithCountersAndStaysTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id, err := effectstest.Msc17Cast(t, g, me, oldB031Card("Steel Dromedary", "Artifact Creature — Camel", "{3}", oldB031Dromedary, 2, 2), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !c.Tapped || c.Counters[game.CounterPlusOne] != 2 {
		t.Fatalf("Dromedary on battlefield %v, tapped %v, %d counters; want tapped with 2", ok, c.Tapped, c.Counters[game.CounterPlusOne])
	}
	oldB031NextUpkeep(t, g, me.ID, id)
	if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
		t.Error("the Dromedary untapped with +1/+1 counters on it")
	}
	oldB031Counter(t, g, id, game.CounterPlusOne, -2)
	oldB031NextUpkeep(t, g, me.ID, id)
	if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
		t.Error("the Dromedary stayed tapped with no counter")
	}
}

// oldB031NextUpkeep walks to the owner's next upkeep, declining the Dromedary's
// combat trigger on the way.
func oldB031NextUpkeep(t *testing.T, g *game.Game, owner, drom uuid.UUID) {
	t.Helper()
	seat := 0
	for i, p := range g.Seats {
		if p.ID == owner {
			seat = i
		}
	}
	left := false
	for i := 0; i < 400; i++ {
		if g.Turn.ActiveSeat != seat {
			left = true
		}
		if left && g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepUpkeep {
			return
		}
		if effectstest.LatestPickTarget(g, owner) != nil {
			effectstest.PickTriggerTarget(t, g, owner, drom)
			continue
		}
		if effectstest.LatestTriggerPromptFor(g, owner) != nil {
			effectstest.AnswerLatestTriggerPrompt(t, g, owner, false)
			continue
		}
		if g.Stack.Size() > 0 || len(g.PendingTriggers) > 0 {
			effectstest.PassPriorityAroundTable(t, g)
			continue
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached the owner's next upkeep")
}

func TestOldB031SteelDromedaryMovesACounterAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	drom := effectstest.B12Push(g, me.ID, "Steel Dromedary", "Artifact Creature — Camel", oldB031Dromedary, 2, 2)
	oldB031Counter(t, g, drom, game.CounterPlusOne, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, drom, game.CounterPlusOne); n != 1 {
		t.Errorf("the Dromedary has %d counters, want 1", n)
	}
	if n := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the target has %d counters, want 1", n)
	}
}

// --- Graceful Restoration --------------------------------------------------

func TestOldB031GracefulRestorationFirstModeReturnsOneWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.PushGraveyardPermanent(me, "Big Bear", "Creature — Bear", "{4}")
	g.WithWriteLock(func() {
		me.Graveyard.Cards[len(me.Graveyard.Cards)-1].Power, me.Graveyard.Cards[len(me.Graveyard.Cards)-1].Toughness = 5, 5
	})
	effectstest.CastModal(t, g, "Graceful Restoration", "Sorcery", oldB031Graceful, []int{0}, effectstest.CardTargets(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Fatal("the creature did not return")
	}
	if n := effectstest.B12Counter(t, g, big, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want the one additional +1/+1 counter", n)
	}
}

func TestOldB031GracefulRestorationSecondModeReturnsTwoSmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushGraveyardPermanent(me, "Small A", "Creature — Bear", "{1}")
	b := effectstest.PushGraveyardPermanent(me, "Small B", "Creature — Bear", "{1}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 2, 2
		}
	})
	effectstest.CastModal(t, g, "Graceful Restoration", "Sorcery", oldB031Graceful, []int{1}, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Errorf("returned: A %v, B %v", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
	if n := effectstest.B12Counter(t, g, a, game.CounterPlusOne); n != 0 {
		t.Errorf("the second mode added %d counters", n)
	}
}

func TestOldB031GracefulRestorationSecondModeRefusesPowerThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.PushGraveyardPermanent(me, "Big", "Creature — Bear", "{4}")
	g.WithWriteLock(func() { me.Graveyard.Cards[len(me.Graveyard.Cards)-1].Power = 3 })
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Graceful Restoration", TypeLine: "Sorcery", OracleID: oldB031Graceful, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardTargets(big)})
	if err == nil {
		t.Error("a power-3 creature was targeted by the power-2-or-less mode")
	}
}

// --- Arcus Acolyte ---------------------------------------------------------

func TestOldB031ArcusAcolyteGivesOutlastToOtherCreaturesWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Arcus Acolyte", "Creature — Human Cleric Archer", oldB031Acolyte, 2, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, bear, 0, "{G}", nil); err != nil {
		t.Fatalf("outlast on another creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bear has %d counters, want 1", n)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Error("outlast did not tap the creature")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, bear, 0, "{G}", nil); err == nil {
		t.Error("a creature with a +1/+1 counter still has outlast")
	}
}

func TestOldB031ArcusAcolyteDoesNotGrantOutlastToOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Arcus Acolyte", "Creature — Human Cleric Archer", oldB031Acolyte, 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("an opponent's creature has outlast")
	}
}

// --- Flametongue Yearling --------------------------------------------------

func TestOldB031FlametongueYearlingDealsItsPowerWithAKickCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	id, err := effectstest.Msc17Cast(t, g, me, oldB031Card("Flametongue Yearling", "Creature — Kavu", "{R}{R}", oldB031Yearling, 2, 1), game.CastSpellParams{OptionalCosts: []int{0}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters, want one per kick (1)", n)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage (power 3 with the counter) did not kill the 3/3")
	}
}

func TestOldB031FlametongueYearlingUnkickedDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	if _, err := effectstest.Msc17Cast(t, g, me, oldB031Card("Flametongue Yearling", "Creature — Kavu", "{R}{R}", oldB031Yearling, 2, 1), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.DamageMarked != 2 {
		t.Errorf("victim on battlefield %v with %d damage, want it alive with 2", ok, c.DamageMarked)
	}
}

// --- Flay Essence ----------------------------------------------------------

func TestOldB031FlayEssenceExilesAndGainsLifePerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Counter Bear", 2, 2)
	oldB031Counter(t, g, victim, game.CounterPlusOne, 2)
	oldB031Counter(t, g, victim, game.CounterStun, 1)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Flay Essence", "Sorcery", oldB031FlayEss, effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not exiled")
	}
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life, want 3 (two +1/+1 counters and one stun)", got)
	}
}

func TestOldB031FlayEssenceExilesAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.GuffbWalker(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test", "", 4)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Flay Essence", "Sorcery", oldB031FlayEss, effectstest.CardTargets(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker was not exiled")
	}
	if got := me.Life - life; got != 4 {
		t.Errorf("gained %d life, want 4 (its loyalty counters)", got)
	}
}

// --- Monoskelion -----------------------------------------------------------

func TestOldB031MonoskelionEntersWithACounterAndPingsByRemovingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id, err := effectstest.Msc17Cast(t, g, me, oldB031Card("Monoskelion", "Artifact Creature — Construct", "{2}", oldB031Monoskel, 1, 1), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters on entering, want 1", n)
	}
	life := opp.Life
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}", []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters after the cost, want 0", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 1 {
		t.Errorf("the opponent lost %d life, want 1", got)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}", []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err == nil {
		t.Error("activated with no counter to remove")
	}
}

// --- Breathless Knight -----------------------------------------------------

func TestOldB031BreathlessKnightGrowsOnlyForCreaturesFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	knight := effectstest.B12Push(g, me.ID, "Breathless Knight", "Creature — Spirit Knight", oldB031Knight, 2, 2)
	effectstest.EnterFromHand(t, g, me.ID, "Hand Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, knight, game.CounterPlusOne); n != 0 {
		t.Fatalf("a creature from hand gave %d counters", n)
	}
	dead := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{2}")
	effectstest.B30Reanimate(t, g, dead, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, knight, game.CounterPlusOne); n != 1 {
		t.Errorf("a creature returned from a graveyard gave %d counters, want 1", n)
	}
}

// --- Suspend ---------------------------------------------------------------

func TestOldB031SuspendExilesWithTwoTimeCountersAndGrantsSuspend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Suspend", "Instant", oldB031Suspend, effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is still on the battlefield")
	}
	c := effectstest.CardInExile(t, g, victim)
	if c.Counters[game.CounterTime] != 2 || !c.SuspendGranted {
		t.Errorf("exiled with %d time counters, suspend granted %v; want 2 and true", c.Counters[game.CounterTime], c.SuspendGranted)
	}
}
