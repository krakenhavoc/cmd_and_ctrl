package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 180 (Ravnica Allegiance: pump and combat tricks, activated
// abilities, #5521). Every test plays the real catalog entry through the
// engine.

const (
	oldB180Ram       = "3b0a97e2-e125-4f91-8581-307e3fa6b048"
	oldB180Automaton = "3d133b18-f1c7-48de-82a6-affb0514b96b"
	oldB180Glass     = "a967921c-cfd4-422c-a08b-42cdd80f89c0"
	oldB180Wight     = "b7639d86-4c7c-423f-8b13-fde3086d3cbb"
	oldB180Mockery   = "d66f7ee7-e9a1-4714-b79c-8c86c6b6d5dc"
	oldB180Pontiff   = "46a9f5af-b024-497d-9f55-98d17e2ffa76"
	oldB180Ignus     = "494b642c-b2b5-40a0-b741-dca1aa9bdb59"
	oldB180Caretaker = "5e2736e6-b812-43a4-bd79-6f6095344fe1"
	oldB180Guildmage = "838eeca8-02f9-44fa-8170-1426622adf30"
	oldB180Watchdog  = "ca2aee09-398d-4eaa-acbb-98663216185b"
)

func TestOldB180EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB180Ram: "Gatebreaker Ram", oldB180Automaton: "Dovin's Automaton",
		oldB180Glass: "Glass of the Guildpact", oldB180Wight: "Plague Wight",
		oldB180Mockery: "Macabre Mockery", oldB180Pontiff: "Pitiless Pontiff",
		oldB180Ignus: "Smelt-Ward Ignus", oldB180Caretaker: "Saruli Caretaker",
		oldB180Guildmage: "Clan Guildmage", oldB180Watchdog: "Resolute Watchdog",
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

func TestOldB180GatebreakerRamGrowsAndGainsKeywordsWithGates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ram := effectstest.B12Push(g, me.ID, "Gatebreaker Ram", "Creature — Sheep", oldB180Ram, 2, 2)
	if p, to := effectstest.EffectivePTOf(t, g, ram); p != 2 || to != 2 {
		t.Errorf("no Gates: %d/%d, want 2/2", p, to)
	}
	effectstest.B12Permanent(g, me.ID, "Gate One", "Land — Gate")
	if p, to := effectstest.EffectivePTOf(t, g, ram); p != 3 || to != 3 {
		t.Errorf("one Gate: %d/%d, want 3/3", p, to)
	}
	if effectstest.HasEffectiveKeyword(t, g, ram, "vigilance") || effectstest.HasEffectiveKeyword(t, g, ram, "trample") {
		t.Error("keywords with only one Gate")
	}
	effectstest.B12Permanent(g, g.Seats[1].ID, "Their Gate", "Land — Gate")
	effectstest.B12Permanent(g, me.ID, "Gate Two", "Land — Gate")
	if p, to := effectstest.EffectivePTOf(t, g, ram); p != 4 || to != 4 {
		t.Errorf("two Gates: %d/%d, want 4/4 (an opponent's Gate does not count)", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ram, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, ram, "trample") {
		t.Error("want vigilance and trample with two Gates")
	}
}

func TestOldB180DovinsAutomatonNeedsADovinPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	auto := effectstest.B12Push(g, me.ID, "Dovin's Automaton", "Artifact Creature — Homunculus", oldB180Automaton, 3, 3)
	effectstest.B12Permanent(g, me.ID, "Some Walker", "Legendary Planeswalker — Jace")
	if p, to := effectstest.EffectivePTOf(t, g, auto); p != 3 || to != 3 || effectstest.HasEffectiveKeyword(t, g, auto, "vigilance") {
		t.Errorf("a non-Dovin walker: %d/%d", p, to)
	}
	effectstest.B12Permanent(g, me.ID, "Dovin Walker", "Legendary Planeswalker — Dovin")
	if p, to := effectstest.EffectivePTOf(t, g, auto); p != 5 || to != 5 || !effectstest.HasEffectiveKeyword(t, g, auto, "vigilance") {
		t.Errorf("with a Dovin walker: %d/%d, want 5/5 and vigilance", p, to)
	}
}

func TestOldB180GlassOfTheGuildpactPumpsOnlyYourMulticoloredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Glass of the Guildpact", "Artifact", oldB180Glass, 0, 0)
	push := func(owner uuid.UUID, colors ...string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Colors: colors,
			Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		})
	}
	multi, mono, theirs := push(me.ID, "W", "U"), push(me.ID, "W"), push(opp.ID, "W", "U")
	for id, want := range map[uuid.UUID]int{multi: 3, mono: 2, theirs: 2} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != want {
			t.Errorf("%s: power %d, want %d", effectstest.FindCardAnywhere(t, g, id).Colors, p, want)
		}
	}
}

func TestOldB180PlagueWightShrinksEveryBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wight := effectstest.B12Push(g, me.ID, "Plague Wight", "Creature — Zombie", oldB180Wight, 2, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	chump := effectstest.B12Creature(g, opp.ID, "Chump", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, wight)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{wall, chump} {
		if err := g.DeclareBlocker(b, wight); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, wall); got != 3 {
		t.Errorf("Wall toughness = %d, want 3", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, chump); ok {
		t.Error("the 1/1 blocker survived -1/-1")
	}
}

func TestOldB180MacabreMockeryStealsPumpsAndSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.B17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	mine := effectstest.B17GraveyardCard(me, "My Bear", "Creature — Bear", "{1}{G}")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Macabre Mockery", "Instant", oldB180Mockery, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("targeted a creature card in my own graveyard")
	}
	effectstest.CastCatalogSpell(t, g, "Macabre Mockery", "Instant", oldB180Mockery, effectstest.CardRefs(dead))
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the creature did not enter under my control: %v", ok)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, dead); ok {
		t.Error("the creature was not sacrificed at the beginning of the end step")
	}
}

func TestOldB180PitilessPontiffSacrificesAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pontiff := effectstest.B12Push(g, me.ID, "Pitiless Pontiff", "Creature — Vampire Cleric", oldB180Pontiff, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, pontiff, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{pontiff}}); err == nil {
		t.Fatal("the Pontiff sacrificed itself")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, pontiff, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, pontiff, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, pontiff, "indestructible") {
		t.Error("want deathtouch and indestructible")
	}
}

func TestOldB180SmeltWardIgnusStealsASmallCreatureAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ignus := effectstest.B12Push(g, me.ID, "Smelt-Ward Ignus", "Creature — Elemental", oldB180Ignus, 2, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	effectstest.FloatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, ignus, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(big)}); err == nil {
		t.Fatal("targeted a creature with power 4")
	}
	effectstest.FloatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, ignus, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, small)
	if !ok || c.Controller != me.ID {
		t.Fatal("control did not change")
	}
	if g.Battlefield.Contains(ignus) {
		t.Error("Ignus was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, small, "haste") {
		t.Error("no haste")
	}
}

func TestOldB180SmeltWardIgnusIsSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	ignus := effectstest.B12Push(g, me.ID, "Smelt-Ward Ignus", "Creature — Elemental", oldB180Ignus, 2, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	effectstest.FloatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, ignus, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestOldB180SaruliCaretakerTapsItselfAndACreatureForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	caretaker := effectstest.B12Push(g, me.ID, "Saruli Caretaker", "Creature — Dryad", oldB180Caretaker, 0, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, caretaker, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("paid with no other creature to tap")
	}
	effectstest.ActivateManaFor(t, g, me.ID, caretaker, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}})
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	effectstest.AssertTapped(t, g, true, caretaker, bear)
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool = %v, want one mana", me.ManaPool)
	}
}

func TestOldB180ClanGuildmageStopsABlockerAndAnimatesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	guild := effectstest.B12Push(g, me.ID, "Clan Guildmage", "Creature — Human Shaman", oldB180Guildmage, 2, 2)
	other := effectstest.B12Push(g, me.ID, "Clan Guildmage", "Creature — Human Shaman", oldB180Guildmage, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	effectstest.FloatForTest(g, me, "RC")
	if err := g.ActivateCatalogAbility(me.ID, guild, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(blocker)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "GCC")
	if err := g.ActivateCatalogAbility(me.ID, other, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.EffectivePTOf(t, g, land); p != 4 || to != 4 {
		t.Errorf("the land is %d/%d, want 4/4", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land has no haste")
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land left")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the creature that could not block blocked")
	}
}

func TestOldB180ResoluteWatchdogGrantsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dog := effectstest.B12Push(g, me.ID, "Resolute Watchdog", "Creature — Dog", oldB180Watchdog, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dog, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("targeted an opponent's creature")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dog, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dog) {
		t.Error("the Watchdog was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the target is not indestructible")
	}
}
