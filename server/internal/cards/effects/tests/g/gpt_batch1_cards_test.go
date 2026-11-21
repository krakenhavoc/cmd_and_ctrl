package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gpt_batch1_cards_test.go — Guildpact, batch 1 (#5599). Each test plays
// the real catalog entry and asserts what its text says.

const (
	gptB1NocturnusOracle   = "b79c4066-66a9-4a53-8716-23f7a340a85f"
	gptB1AgentOracle       = "632ccf7e-eaf4-42a6-9069-6ad915d03a5b"
	gptB1AngelOracle       = "02520741-d7f0-46e6-acbe-010b65fc24c0"
	gptB1WurmOracle        = "fdcfa49a-70ed-48a4-b424-527a1698fd85"
	gptB1MagemarkOracle    = "f6889f57-2965-4a82-8556-fd91c6227058"
	gptB1ProwlerOracle     = "655a584c-5280-4317-b707-101b7e31dcb1"
	gptB1BorborygmosOracle = "9ab0820b-4f5f-47b3-8fa9-43f802aed275"
	gptB1BloodscaleOracle  = "a9be4be0-71fe-41d9-8f30-351385753fc9"
	gptB1ShamanOracle      = "8593e899-8e6f-4d40-96ab-98788b3d9486"
	gptB1CastigateOracle   = "28a972a6-cf67-4ee3-aa91-f7b2549f6c48"
	gptB1CausticRainOracle = "8df1cbbc-87fc-4948-b224-9175b9f396d9"
	gptB1VortexOracle      = "c68b4003-e539-4b88-acc9-e95c100173c0"
	gptB1CrashOracle       = "5d36e5ee-1e23-48fe-ad13-081cc153342b"
	gptB1CryptOracle       = "01d4c859-8b15-4567-a2c4-70a30c709999"
)

func TestGptB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gptB1NocturnusOracle:   "Abyssal Nocturnus",
		gptB1AgentOracle:       "Agent of Masks",
		gptB1AngelOracle:       "Angel of Despair",
		gptB1WurmOracle:        "Battering Wurm",
		gptB1MagemarkOracle:    "Beastmaster's Magemark",
		gptB1ProwlerOracle:     "Bloodscale Prowler",
		gptB1BorborygmosOracle: "Borborygmos",
		gptB1BloodscaleOracle:  "Burning-Tree Bloodscale",
		gptB1ShamanOracle:      "Burning-Tree Shaman",
		gptB1CastigateOracle:   "Castigate",
		gptB1CausticRainOracle: "Caustic Rain",
		gptB1VortexOracle:      "Cerebral Vortex",
		gptB1CrashOracle:       "Crash Landing",
		gptB1CryptOracle:       "Cryptwailing",
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

// --- Abyssal Nocturnus -----------------------------------------------------------

func TestGptB1AbyssalNocturnusGrowsAndGainsFearPerOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	noct := effectstest.B27Push(g, me.ID, "Abyssal Nocturnus", "Creature — Horror", gptB1NocturnusOracle, "{1}{B}{B}", 2, 2, "B")
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, me)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, noct); p != 2 || tough != 2 {
		t.Fatalf("after its controller discarded, the Nocturnus is %d/%d, want 2/2", p, tough)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, noct, "fear") {
		t.Fatal("the controller's own discard gave it fear")
	}

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, noct); p != 6 || tough != 6 {
		t.Errorf("after two opponent discards, the Nocturnus is %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, noct, "fear") {
		t.Error("the Nocturnus has no fear")
	}
}

// --- Agent of Masks ----------------------------------------------------------------

func TestGptB1AgentOfMasksDrainsEachOpponentAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.B27Push(g, owner.ID, "Agent of Masks", "Creature — Human Advisor", gptB1AgentOracle, "{3}{W}{B}", 2, 3, "W", "B")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID == owner.ID {
			continue
		}
		if p.Life != before[p.ID]-1 {
			t.Errorf("opponent %s life %d, want %d", p.ID, p.Life, before[p.ID]-1)
		}
	}
	if want := before[owner.ID] + len(g.Seats) - 1; owner.Life != want {
		t.Errorf("the Agent's controller has %d life, want %d", owner.Life, want)
	}
}

// --- Angel of Despair ---------------------------------------------------------------

func TestGptB1AngelOfDespairDestroysTargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	angel := effectstest.CastHoldCreature(t, g, "Angel of Despair", "Creature — Angel", gptB1AngelOracle, 5, 5)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: relic}, true)
	if g.Battlefield.Contains(relic) {
		t.Error("the targeted artifact survived")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, angel, "flying") {
		t.Error("the Angel has no flying")
	}
}

// --- bloodthirst: Bloodscale Prowler, Battering Wurm, Burning-Tree Bloodscale ----

func TestGptB1BloodthirstNeedsAnOpponentDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	prowler := effectstest.CastHoldCreature(t, g, "Bloodscale Prowler", "Creature — Lizard Warrior", gptB1ProwlerOracle, 3, 1)
	if p, tough := effectstest.PtOf(t, g, prowler); p != 3 || tough != 1 {
		t.Errorf("with no opponent damaged, the Prowler is %d/%d, want 3/1", p, tough)
	}
}

func TestGptB1BloodthirstIgnoresDamageToItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	effectstest.GptB1DamagePlayer(g, me.ID, 3)
	prowler := effectstest.CastHoldCreature(t, g, "Bloodscale Prowler", "Creature — Lizard Warrior", gptB1ProwlerOracle, 3, 1)
	if n := effectstest.CounterCount(g, prowler, game.CounterPlusOne); n != 0 {
		t.Errorf("damage to its own controller put %d counters on the Prowler", n)
	}
}

func TestGptB1BloodthirstCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, t                   int
	}{
		{"Bloodscale Prowler", "Creature — Lizard Warrior", gptB1ProwlerOracle, 3, 1},
		{"Battering Wurm", "Creature — Wurm", gptB1WurmOracle, 4, 3},
		{"Burning-Tree Bloodscale", "Creature — Lizard Berserker", gptB1BloodscaleOracle, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, opp := effectstest.GptB1Seats(g)
			effectstest.GptB1DamagePlayer(g, opp.ID, 1)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.t)
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 1 {
				t.Errorf("%d +1/+1 counters, want 1", n)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.p+1 || tough != tc.t+1 {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.p+1, tc.t+1)
			}
		})
	}
}

func TestGptB1BatteringWurmCantBeBlockedBySmallerCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	effectstest.GptB1DamagePlayer(g, opp.ID, 2)
	wurm := effectstest.CastHoldCreature(t, g, "Battering Wurm", "Creature — Wurm", gptB1WurmOracle, 4, 3)
	four := effectstest.B16Creature(g, opp.ID, "Four", "Creature — Ogre", 4, 4)
	five := effectstest.B16Creature(g, opp.ID, "Five", "Creature — Giant", 5, 5)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		get := func(id uuid.UUID) game.Card { c, _ := g.LookupCardForEffect(id); return c }
		w, f4, f5 := get(wurm), get(four), get(five)
		if r := g.BlockPairRefusalLocked(&w, &f4); r.Legal() {
			t.Error("a 4-power creature blocked the 5-power Wurm")
		}
		if r := g.BlockPairRefusalLocked(&w, &f5); !r.Legal() {
			t.Errorf("a 5-power creature may not block the 5-power Wurm: %q", r.Reason)
		}
	})
}

// --- Burning-Tree Bloodscale ---------------------------------------------------------

func TestGptB1BurningTreeBloodscaleStopsOneBlockerAndLuresAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	scale := effectstest.B12Push(g, me.ID, "Burning-Tree Bloodscale", "Creature — Lizard Berserker", gptB1BloodscaleOracle, 2, 2)
	kept := effectstest.B16Creature(g, opp.ID, "Kept Back", "Creature — Bear", 2, 2)
	lured := effectstest.B16Creature(g, opp.ID, "Lured", "Creature — Bear", 2, 2)

	effectstest.RavB2MustActivate(t, g, me.ID, scale, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(kept)}, "C", "C", "R")
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		s, _ := g.LookupCardForEffect(scale)
		k, _ := g.LookupCardForEffect(kept)
		l, _ := g.LookupCardForEffect(lured)
		if r := g.BlockPairRefusalLocked(&s, &k); r.Legal() {
			t.Error("the targeted creature may still block the Bloodscale")
		}
		if r := g.BlockPairRefusalLocked(&s, &l); !r.Legal() {
			t.Errorf("an untargeted creature may not block the Bloodscale: %q", r.Reason)
		}
	})

	effectstest.RavB2MustActivate(t, g, me.ID, scale, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lured)}, "C", "C", "G")
	recs := effectstest.BlocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != scale {
		t.Fatalf("records = %+v, want one naming the Bloodscale", recs)
	}
}

// --- Burning-Tree Shaman --------------------------------------------------------------

func TestGptB1BurningTreeShamanBurnsEveryActivatorButNotManaAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Burning-Tree Shaman", "Creature — Centaur Shaman", gptB1ShamanOracle, 3, 4, "R", "G")
	myKrenko := effectstest.ActPushKrenko(g, me.ID)
	theirKrenko := effectstest.ActPushKrenko(g, them.ID)
	theirElves := effectstest.B21Push(g, them.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.ActLlanowarElvesOracle, 1, 1, "G")
	mine, theirs := me.Life, them.Life

	if err := g.ActivateManaAbility(them.ID, theirElves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("their Llanowar Elves: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != theirs {
		t.Errorf("a mana ability cost its activator %d life", theirs-them.Life)
	}

	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("their Krenko: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != theirs-1 {
		t.Errorf("after an opponent's activation: %d, want %d", them.Life, theirs-1)
	}
	if me.Life != mine {
		t.Errorf("the Shaman's controller took damage for an opponent's activation")
	}

	if err := g.ActivateCatalogAbility(me.ID, myKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("own Krenko: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine-1 {
		t.Errorf("after its controller's own activation: %d, want %d", me.Life, mine-1)
	}
}

// --- Beastmaster's Magemark ------------------------------------------------------------

func TestGptB1BeastmastersMagemarkPumpsEnchantedCreaturesAndBlockedOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Beastmaster's Magemark", effectstest.AuraTypeLine, gptB1MagemarkOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("the enchanted Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, plain); p != 2 || tough != 2 {
		t.Fatalf("an unenchanted creature is %d/%d, want 2/2", p, tough)
	}

	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Wall", 0, 5)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Wall", 0, 5)
	c := effectstest.B12Creature(g, opp.ID, "Blocker C", "Creature — Wall", 0, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, bear, plain)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{a: bear, b: bear, c: plain})
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the enchanted Bear blocked by two is %d/%d, want 5/5", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, plain); p != 2 || tough != 2 {
		t.Errorf("the blocked unenchanted creature is %d/%d, want 2/2", p, tough)
	}
}

// --- Borborygmos ------------------------------------------------------------------------

func TestGptB1BorborygmosCountersEachCreatureOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	borb := effectstest.B12Push(g, me.ID, "Borborygmos", "Legendary Creature — Cyclops", gptB1BorborygmosOracle, 6, 7)
	other := effectstest.B12Creature(g, me.ID, "Stay Home", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, borb, "trample") {
		t.Error("Borborygmos has no trample")
	}
	life := opp.Life
	effectstest.Pr7bAttack(t, g, opp.ID, borb)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Fatalf("defender life %d, want %d", opp.Life, life-6)
	}
	for _, id := range []uuid.UUID{borb, other} {
		if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("%d counters on a creature its controller controls, want 1", n)
		}
	}
	if n := effectstest.CounterCount(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's creature got %d counters", n)
	}
}

// --- Castigate ---------------------------------------------------------------------------

func TestGptB1CastigateExilesTheChosenNonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	ids := effectstest.RevealHand(opp,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	effectstest.CastAt(t, g, "Castigate", "Sorcery", gptB1CastigateOracle, effectstest.PbPlayer(opp.ID))
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[0]}); err == nil {
		t.Fatal("the land was accepted as a nonland card")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[1]) {
		t.Error("the chosen card was not exiled")
	}
	if !opp.Hand.Contains(ids[0]) {
		t.Error("the land left the hand")
	}
}

func TestGptB1CastigateTargetsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Castigate", "Sorcery", gptB1CastigateOracle, effectstest.PlayerRefs(me.ID)); err == nil {
		t.Error("Castigate targeted its caster")
	}
}

// --- Caustic Rain ---------------------------------------------------------------------------

func TestGptB1CausticRainExilesTargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Caustic Rain", "Sorcery", gptB1CausticRainOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Caustic Rain targeted a creature")
	}
	effectstest.CastAt(t, g, "Caustic Rain", "Sorcery", gptB1CausticRainOracle, game.TargetRef{Kind: game.TargetCard, ID: land})
	if !g.Exile.Contains(land) {
		t.Error("the land was not exiled")
	}
}

// --- Cerebral Vortex ------------------------------------------------------------------------

func TestGptB1CerebralVortexDamagesByEveryCardDrawnThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	effectstest.SeedLibrary(opp, "One", "Two", "Three", "Four")
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.CastAt(t, g, "Cerebral Vortex", "Instant", gptB1VortexOracle, effectstest.PbPlayer(opp.ID))
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Fatalf("the target drew %d, want 2", got)
	}
	if opp.Life != life-2 {
		t.Fatalf("after drawing two this turn, life %d, want %d", opp.Life, life-2)
	}
	effectstest.CastAt(t, g, "Cerebral Vortex", "Instant", gptB1VortexOracle, effectstest.PbPlayer(opp.ID))
	if opp.Life != life-2-4 {
		t.Errorf("after drawing four this turn, life %d, want %d", opp.Life, life-6)
	}
}

// --- Crash Landing --------------------------------------------------------------------------

func TestGptB1CrashLandingGroundsAndHitsByForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	flyer := effectstest.TypedCreature(g, opp.ID, "Their Bird", "Creature — Bird", "", 2, 4, "flying")
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Crash Landing", "Instant", gptB1CrashOracle, effectstest.CardRefs(ground)); err == nil {
		t.Fatal("Crash Landing targeted a creature without flying")
	}
	effectstest.CastAt(t, g, "Crash Landing", "Instant", gptB1CrashOracle, game.TargetRef{Kind: game.TargetCard, ID: flyer})
	if effectstest.HasKeywordOnBattlefield(t, g, flyer, "flying") {
		t.Error("the Bird kept flying")
	}
	c, ok := effectstest.BattlefieldCard(g, flyer)
	if !ok {
		t.Fatal("the 2/4 Bird died to two Forests' worth of damage")
	}
	if c.DamageMarked != 2 {
		t.Errorf("damage marked %d, want 2 (the Forests its caster controls)", c.DamageMarked)
	}
}

// --- Cryptwailing ---------------------------------------------------------------------------

func TestGptB1CryptwailingExilesTwoCreatureCardsForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crypt := effectstest.B12Push(g, me.ID, "Cryptwailing", "Enchantment", gptB1CryptOracle, 0, 0)
	a := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	b := effectstest.PushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf")
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	effectstest.PushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	params := func(ids ...uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), ExileIDs: ids}
	}
	// The harness starts in the draw step, where a sorcery-speed row is shut.
	if err := effectstest.RavB2Activate(t, g, me.ID, crypt, 0, params(a, b), "C"); err == nil {
		t.Fatal("activated a sorcery-speed ability in the draw step")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, crypt, 0, params(a, spell), "C"); err == nil {
		t.Fatal("a sorcery card paid for \"two creature cards\"")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.RavB2MustActivate(t, g, me.ID, crypt, 0, params(a, b), "C")
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the two creature cards were not exiled")
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Errorf("the target owes %d discards, want 1", n)
	}
}
