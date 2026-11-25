package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b177_test.go — Ravnica Allegiance, batch 177 (#5521): +1/+1 and
// other counters. Each test plays the real catalog entry and asserts what
// its text says, and what it refuses.

const (
	s82o177Guardian = "1a49ec1c-6150-4b46-b671-b7d6f4e301ae"
	s82o177Lizrog   = "3561cd52-961f-4098-9287-58ee43970f0f"
	s82o177Scav     = "5146ae31-9578-46e3-be78-eb90e4ba48cd"
	s82o177Ruin     = "8f36e0db-eea7-4c97-98ce-4bc2d8c7f734"
	s82o177Combine  = "90abb55d-5c05-453f-b4a6-c0741dc36694"
	s82o177Familiar = "923d44a9-774d-459c-9cec-1d5b081b51f0"
	s82o177Font     = "ab44c6fe-6a66-42f0-9e75-8b25641e1f13"
	s82o177Haunt    = "e70f5520-1b9c-4351-8484-30f0dc692e01"
)

func TestS82Old177EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82o177Guardian: "Growth-Chamber Guardian",
		s82o177Lizrog:   "Galloping Lizrog",
		s82o177Scav:     "Undercity Scavenger",
		s82o177Ruin:     "Rumbling Ruin",
		s82o177Combine:  "Combine Guildmage",
		s82o177Familiar: "Biomancer's Familiar",
		s82o177Font:     "Font of Agonies",
		s82o177Haunt:    "The Haunt of Hightower",
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

// s82o177Counters is the +1/+1 counters on a battlefield permanent.
func s82o177Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

// s82o177Creature puts a creature with counters on the battlefield.
func s82o177Creature(g *game.Game, owner uuid.UUID, name string, power, toughness, counters int) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: toughness, Owner: owner, Controller: owner}
	if counters > 0 {
		c.Counters = map[string]int{game.CounterPlusOne: counters}
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func s82o177Activate(t *testing.T, g *game.Game, me uuid.UUID, src uuid.UUID, index int, targets ...uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me, src, index, game.ActivateAbilityParams{Targets: effectstest.CardRefs(targets...)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Growth-Chamber Guardian -------------------------------------------------

func TestS82Old177GuardianAdaptsThenFindsAnotherGuardian(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gg := effectstest.PushCatalogPermanent(g, me.ID, "Growth-Chamber Guardian", "Creature — Elf Crab Warrior", s82o177Guardian, false)
	other := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Growth-Chamber Guardian", TypeLine: "Creature — Elf Crab Warrior", OracleID: s82o177Guardian, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}")
	s82o177Activate(t, g, me.ID, gg, 0)
	if n := s82o177Counters(t, g, gg); n != 2 {
		t.Fatalf("adapt 2 gave %d counters, want 2", n)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no library search opened")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{other}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(other) {
		t.Error("the second Growth-Chamber Guardian did not reach the hand")
	}
}

func TestS82Old177GuardianAdaptDoesNothingWithCountersAndNeverTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gg := effectstest.PushCatalogPermanent(g, me.ID, "Growth-Chamber Guardian", "Creature — Elf Crab Warrior", s82o177Guardian, false)
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, gg); c != nil {
			c.Counters = map[string]int{game.CounterPlusOne: 1}
		}
	})
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}")
	s82o177Activate(t, g, me.ID, gg, 0)
	if n := s82o177Counters(t, g, gg); n != 1 {
		t.Errorf("adapt with a counter on it gave %d counters, want 1", n)
	}
	if effectstest.ChkB1OpenSearch(g) != nil {
		t.Error("a search opened though no counter was put on it")
	}
}

// --- Galloping Lizrog ---------------------------------------------------------

func TestS82Old177LizrogRemovesCountersAndGrowsTwiceAsMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := s82o177Creature(g, me.ID, "Bear A", 2, 2, 3)
	b := s82o177Creature(g, me.ID, "Bear B", 2, 2, 1)
	effectstest.CastAndResolveCreature(t, g, "Galloping Lizrog", "Creature — Frog Lizard", s82o177Lizrog)
	effectstest.PassPriorityAroundTable(t, g)
	pickOption(t, g, me.ID, 2) // from Bear A: remove two
	pickOption(t, g, me.ID, 1) // from Bear B: remove one
	effectstest.PassPriorityAroundTable(t, g)
	liz := effectstest.FindBattlefieldByName(g, "Galloping Lizrog")
	if s82o177Counters(t, g, a) != 1 || s82o177Counters(t, g, b) != 0 {
		t.Errorf("Bear A has %d, Bear B has %d counters, want 1 and 0", s82o177Counters(t, g, a), s82o177Counters(t, g, b))
	}
	if n := s82o177Counters(t, g, liz); n != 6 {
		t.Errorf("Lizrog has %d counters, want 6 (twice the 3 removed)", n)
	}
}

func TestS82Old177LizrogDeclinedByRemovingNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := s82o177Creature(g, me.ID, "Bear A", 2, 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Galloping Lizrog", "Creature — Frog Lizard", s82o177Lizrog)
	effectstest.PassPriorityAroundTable(t, g)
	pickOption(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	liz := effectstest.FindBattlefieldByName(g, "Galloping Lizrog")
	if s82o177Counters(t, g, a) != 2 || s82o177Counters(t, g, liz) != 0 {
		t.Errorf("counters moved though none were removed: Bear %d, Lizrog %d", s82o177Counters(t, g, a), s82o177Counters(t, g, liz))
	}
}

func TestS82Old177LizrogAsksNothingWithoutCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	s82o177Creature(g, me.ID, "Bear A", 2, 2, 0)
	effectstest.CastAndResolveCreature(t, g, "Galloping Lizrog", "Creature — Frog Lizard", s82o177Lizrog)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		t.Error("a question opened though no creature has a counter")
	}
}

// pickOption answers the open option-pick prompt with the option at index.
func pickOption(t *testing.T, g *game.Game, chooser uuid.UUID, index int) {
	t.Helper()
	c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatal("no option prompt is open")
	}
	if err := g.ResolveOptionPick(c.ID, chooser, index); err != nil {
		t.Fatalf("ResolveOptionPick(%d): %v", index, err)
	}
}

// --- Undercity Scavenger ---------------------------------------------------------

func TestS82Old177ScavengerSacrificesForCountersAndScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	fodder := s82o177Creature(g, me.ID, "Fodder", 1, 1, 0)
	effectstest.CastAndResolveCreature(t, g, "Undercity Scavenger", "Creature — Ogre Warrior", s82o177Scav)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	sc := effectstest.FindBattlefieldByName(g, "Undercity Scavenger")
	if n := s82o177Counters(t, g, sc); n != 2 {
		t.Errorf("Scavenger has %d counters, want 2", n)
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 2")
	}
}

func TestS82Old177ScavengerDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := s82o177Creature(g, me.ID, "Fodder", 1, 1, 0)
	effectstest.CastAndResolveCreature(t, g, "Undercity Scavenger", "Creature — Ogre Warrior", s82o177Scav)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPermanentPicks(t, g, me.ID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	sc := effectstest.FindBattlefieldByName(g, "Undercity Scavenger")
	if !g.Battlefield.Contains(fodder) || s82o177Counters(t, g, sc) != 0 {
		t.Error("declining the sacrifice still did something")
	}
}

// --- Rumbling Ruin ---------------------------------------------------------------

func TestS82Old177RumblingRuinStopsSmallOpposingCreaturesBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	s82o177Creature(g, me.ID, "Grown", 2, 2, 2)
	small := s82o177Creature(g, opp.ID, "Small", 2, 2, 0)
	big := s82o177Creature(g, opp.ID, "Big", 3, 3, 0)
	mine := s82o177Creature(g, me.ID, "Mine", 1, 1, 0)
	effectstest.CastAndResolveCreature(t, g, "Rumbling Ruin", "Creature — Elemental", s82o177Ruin)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraRestrictions(t, g, small)&game.CantBlock == 0 {
		t.Error("the power-2 creature can still block (2 counters)")
	}
	if effectstest.AuraRestrictions(t, g, big)&game.CantBlock != 0 {
		t.Error("the power-3 creature was stopped from blocking")
	}
	if effectstest.AuraRestrictions(t, g, mine)&game.CantBlock != 0 {
		t.Error("your own creature was stopped from blocking")
	}
}

// --- Combine Guildmage -----------------------------------------------------------

func TestS82Old177CombineGuildmageGivesCreaturesACounterThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cg := effectstest.PushCatalogPermanent(g, me.ID, "Combine Guildmage", "Creature — Merfolk Wizard", s82o177Combine, false)
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}")
	s82o177Activate(t, g, me.ID, cg, 0)
	bear := effectstest.CastAndResolveCreature(t, g, "Hill Giant", "Creature — Giant", "")
	effectstest.PassPriorityAroundTable(t, g)
	_ = bear
	giant := effectstest.FindBattlefieldByName(g, "Hill Giant")
	if n := s82o177Counters(t, g, giant); n != 1 {
		t.Errorf("a creature cast this turn entered with %d counters, want 1", n)
	}
}

func TestS82Old177CombineGuildmageMovesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cg := effectstest.PushCatalogPermanent(g, me.ID, "Combine Guildmage", "Creature — Merfolk Wizard", s82o177Combine, false)
	from := s82o177Creature(g, me.ID, "From", 2, 2, 2)
	to := s82o177Creature(g, me.ID, "To", 2, 2, 0)
	effectstest.Acrb1AddMana(t, g, me.ID, "{U}{U}")
	s82o177Activate(t, g, me.ID, cg, 1, from, to)
	if s82o177Counters(t, g, from) != 1 || s82o177Counters(t, g, to) != 1 {
		t.Errorf("From has %d, To has %d, want 1 and 1", s82o177Counters(t, g, from), s82o177Counters(t, g, to))
	}
}

func TestS82Old177CombineGuildmageRefusesTheSameCreatureTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cg := effectstest.PushCatalogPermanent(g, me.ID, "Combine Guildmage", "Creature — Merfolk Wizard", s82o177Combine, false)
	from := s82o177Creature(g, me.ID, "From", 2, 2, 2)
	effectstest.Acrb1AddMana(t, g, me.ID, "{U}{U}")
	err := g.ActivateCatalogAbility(me.ID, cg, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(from, from)})
	if err == nil {
		t.Error("one creature was accepted for both targets")
	}
}

// --- Biomancer's Familiar ----------------------------------------------------------

func TestS82Old177FamiliarCutsActivatedCostsAndLetsAnAdaptRepeat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fam := effectstest.PushCatalogPermanent(g, me.ID, "Biomancer's Familiar", "Creature — Mutant", s82o177Familiar, false)
	gg := effectstest.PushCatalogPermanent(g, me.ID, "Growth-Chamber Guardian", "Creature — Elf Crab Warrior", s82o177Guardian, false)
	// {2}{G} costs {G} with the Familiar out: one mana is enough.
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}")
	s82o177Activate(t, g, me.ID, gg, 0)
	if n := s82o177Counters(t, g, gg); n != 2 {
		t.Fatalf("adapt for the reduced cost gave %d counters, want 2", n)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false) // the Guardian's search is a "may"
	// A creature with counters adapts as though it had none, once.
	s82o177Activate(t, g, me.ID, fam, 0, gg)
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}")
	s82o177Activate(t, g, me.ID, gg, 0)
	if n := s82o177Counters(t, g, gg); n != 4 {
		t.Errorf("adapt after the Familiar gave %d counters, want 4", n)
	}
}

// --- Font of Agonies ----------------------------------------------------------------

func TestS82Old177FontCountsLifePaidThenDestroysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	font := effectstest.PushCatalogPermanent(g, me.ID, "Font of Agonies", "Enchantment", s82o177Font, false)
	victim := s82o177Creature(g, opp.ID, "Victim", 2, 2, 0)
	for _, amount := range []int{2, 2} {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: me.ID, Amount: -amount, LifePayment: true})
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	// life lost or gained that is not paid adds none
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: me.ID, Amount: -3})
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: me.ID, Amount: 3, LifePayment: true})
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: opp.ID, Amount: -3, LifePayment: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCardByID(g, font)
	if c.Counters["blood"] != 4 {
		t.Fatalf("%d blood counters, want 4", c.Counters["blood"])
	}
	effectstest.Acrb1AddMana(t, g, me.ID, "{B}{B}")
	s82o177Activate(t, g, me.ID, font, 0, victim)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	if c, _ := effectstest.BattlefieldCardByID(g, font); c.Counters["blood"] != 0 {
		t.Errorf("%d blood counters left, want 0", c.Counters["blood"])
	}
}

func TestS82Old177FontCantActivateWithThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	font := effectstest.PushCatalogPermanent(g, me.ID, "Font of Agonies", "Enchantment", s82o177Font, false)
	victim := s82o177Creature(g, opp.ID, "Victim", 2, 2, 0)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: me.ID, Amount: -3, LifePayment: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, font, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Error("the ability was activated with three blood counters")
	}
}

// --- The Haunt of Hightower ----------------------------------------------------------

func TestS82Old177HauntGrowsPerCardIntoAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	h := effectstest.PushCatalogPermanent(g, me.ID, "The Haunt of Hightower", "Legendary Creature — Vampire", s82o177Haunt, false)
	for _, name := range []string{"One", "Two"} {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	mine := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, mine); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82o177Counters(t, g, h); n != 2 {
		t.Errorf("Haunt has %d counters, want 2 (two opponent cards, not your own)", n)
	}
}
