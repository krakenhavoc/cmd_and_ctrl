package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b61_cards_test.go — GAP batch 61 (#5948, S82): Fallout
// artifacts and creatures. Each test plays the real catalog entry.

const (
	gapb61AlphaDeathclawOracle  = "0a7d0d1f-a5ac-4db9-860c-9fcbe3fd9e71"
	gapb61FrankHorriganOracle   = "aedcde3b-3ed5-443c-a81c-03d785153e55"
	gapb61BoomerScrapperOracle  = "3292d795-4fdc-4e42-959a-dd5e8591c5cc"
	gapb61MisterGutsyOracle     = "09763d9a-056f-49db-a72b-1c7325075789"
	gapb61EDEOracle             = "1b1bbad2-13b3-4716-85ef-3956268e23ad"
	gapb61ButchOracle           = "20f92504-d03c-437e-9814-b25ee384b3ba"
	gapb61LaniusOracle          = "c6b5ca62-e5ee-4636-b12e-7c1d455e8384"
	gapb61FormalwearOracle      = "c5782f70-05c6-4df9-812d-7ecbe2dbaaff"
	gapb61PrydwenOracle         = "f6ea8085-5ba9-43b6-994d-04be35c5793b"
	gapb61YoungDeathclawsOracle = "f24fc213-6341-482c-a056-946d0f1dd76b"
	gapb61CaesarOracle          = "8e62d05a-6efd-4764-bca8-97895e0cb613"
)

func TestGapb61CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb61AlphaDeathclawOracle:  "Alpha Deathclaw",
		gapb61FrankHorriganOracle:   "Agent Frank Horrigan",
		gapb61BoomerScrapperOracle:  "Boomer Scrapper",
		gapb61MisterGutsyOracle:     "Mister Gutsy",
		gapb61EDEOracle:             "ED-E, Lonesome Eyebot",
		gapb61ButchOracle:           "Butch DeLoria, Tunnel Snake",
		gapb61LaniusOracle:          "Legate Lanius, Caesar's Ace",
		gapb61FormalwearOracle:      "Pre-War Formalwear",
		gapb61PrydwenOracle:         "The Prydwen, Steel Flagship",
		gapb61YoungDeathclawsOracle: "Young Deathclaws",
		gapb61CaesarOracle:          "Caesar, Legion's Emperor",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func TestGapb61AlphaDeathclawDestroysAPermanentWhenItBecomesMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	claw := effectstest.PushMonster(g, me.ID, "Alpha Deathclaw", gapb61AlphaDeathclawOracle, 6, 6)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{B}{G}")
	effectstest.ActivateMonstrosity(t, g, me.ID, claw, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the targeted permanent survived the becomes-monstrous trigger")
	}
	if n := effectstest.CountersOn(g, claw, game.CounterPlusOne); n != 4 {
		t.Errorf("Alpha Deathclaw has %d +1/+1 counters, want 4", n)
	}
	if !effectstest.MonstrousNow(g, claw) {
		t.Error("Alpha Deathclaw is not monstrous")
	}
}

func TestGapb61AlphaDeathclawDestroysAPermanentWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Alpha Deathclaw", TypeLine: "Creature — Lizard Mutant",
		OracleID: gapb61AlphaDeathclawOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{B}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the targeted permanent survived the enters trigger")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("Alpha Deathclaw did not enter")
	}
}

func TestGapb61FrankHorriganIsIndestructibleOnlyOnceItAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frank := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Agent Frank Horrigan", TypeLine: "Legendary Creature — Mutant Warrior",
		OracleID: gapb61FrankHorriganOracle, Keywords: []string{"trample"}, Power: 8, Toughness: 6,
		Owner: me.ID, Controller: me.ID,
	})
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, frank), "indestructible") {
		t.Fatal("indestructible before it attacked")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, frank)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, frank), "indestructible") {
		t.Errorf("not indestructible after attacking: %v", effectstest.EffectiveAbilities(t, g, frank))
	}
}

func TestGapb61FrankHorriganProliferatesTwiceWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frank := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Agent Frank Horrigan", TypeLine: "Legendary Creature — Mutant Warrior",
		OracleID: gapb61FrankHorriganOracle, Keywords: []string{"trample"}, Power: 8, Toughness: 6,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.AddCounter(frank, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(frank, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	// Two proliferates, each asked of the controller: take Horrigan
	// both times.
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ProliferatePrompt(g) == nil {
		t.Fatal("no first proliferate prompt")
	}
	effectstest.AnswerProliferate(t, g, frank)
	if effectstest.ProliferatePrompt(g) == nil {
		t.Fatal("no second proliferate prompt")
	}
	effectstest.AnswerProliferate(t, g, frank)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, frank, game.CounterPlusOne); n != 3 {
		t.Errorf("Horrigan has %d +1/+1 counters after proliferating twice from 1, want 3", n)
	}
}

// gapb61Named lists the battlefield permanents with a name under a player.
func gapb61Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// gapb61Creature pushes a catalog creature onto the battlefield.
func gapb61Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- Boomer Scrapper ---------------------------------------------------------

func TestGapb61BoomerScrapperLosesLifeAndMakesJunkWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Boomer Scrapper", TypeLine: "Creature — Human Soldier",
		OracleID: gapb61BoomerScrapperOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{B}{R}")
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	junk := gapb61Named(g, me.ID, "Junk")
	if len(junk) != 1 {
		t.Fatalf("%d Junk tokens, want 1", len(junk))
	}
	// A token you control leaving puts a counter on it.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(junk[0]) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 1 {
		t.Errorf("Boomer Scrapper has %d +1/+1 counters after a token left, want 1", n)
	}
}

func TestGapb61BoomerScrapperIgnoresAnOpponentsToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boomer := gapb61Creature(g, me.ID, "Boomer Scrapper", "Creature — Human Soldier", gapb61BoomerScrapperOracle, 1, 1)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, boomer, game.CounterPlusOne); n != 0 {
		t.Errorf("a counter for an opponent's token: %d", n)
	}
}

// --- Mister Gutsy ------------------------------------------------------------

func TestGapb61MisterGutsyGrowsOnAuraAndEquipmentSpellsAndLeavesJunk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gutsy := gapb61Creature(g, me.ID, "Mister Gutsy", "Artifact Creature — Robot Soldier", gapb61MisterGutsyOracle, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Test Blade", "Artifact — Equipment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Test Boon", "Enchantment — Aura", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, gutsy, game.CounterPlusOne); n != 2 {
		t.Fatalf("Gutsy has %d counters after an Equipment, an Aura and a creature spell, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gutsy) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb61Named(g, me.ID, "Junk")); n != 2 {
		t.Errorf("%d Junk tokens after dying with two counters, want 2", n)
	}
}

func TestGapb61MisterGutsyDyingWithNoCountersMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gutsy := gapb61Creature(g, me.ID, "Mister Gutsy", "Artifact Creature — Robot Soldier", gapb61MisterGutsyOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gutsy) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb61Named(g, me.ID, "Junk")); n != 0 {
		t.Errorf("%d Junk tokens from a Gutsy with no counters", n)
	}
}

// --- ED-E, Lonesome Eyebot -----------------------------------------------------

func TestGapb61EDEGetsAQuestCounterOnlyWhenMoreCreaturesAttackThanItHas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ede := gapb61Creature(g, me.ID, "ED-E, Lonesome Eyebot", "Legendary Artifact Creature — Robot", gapb61EDEOracle, 2, 1)
	bear := gapb61Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, ede)
	if n := effectstest.CountersOn(g, ede, "quest"); n != 1 {
		t.Fatalf("1 attacker vs 0 quest counters: %d counters, want 1", n)
	}
	_ = bear
}

func TestGapb61EDESacrificeDrawsOnePlusOnePerQuestCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ede := gapb61Creature(g, me.ID, "ED-E, Lonesome Eyebot", "Legendary Artifact Creature — Robot", gapb61EDEOracle, 2, 1)
	if err := g.AddCounter(ede, "quest", 2); err != nil {
		t.Fatal(err)
	}
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ede, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ede) {
		t.Error("ED-E was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards with two quest counters, want 3", got)
	}
}

// --- Butch DeLoria, Tunnel Snake ---------------------------------------------

func TestGapb61ButchGrowsForEachOtherRogueOrSnakeWhenHeAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butch := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Butch DeLoria, Tunnel Snake", TypeLine: "Legendary Creature — Human Rogue",
		OracleID: gapb61ButchOracle, Keywords: []string{"menace"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	gapb61Creature(g, me.ID, "Rogue", "Creature — Human Rogue", "", 1, 1)
	gapb61Creature(g, me.ID, "Snake", "Creature — Snake", "", 1, 1)
	gapb61Creature(g, me.ID, "Rogue Snake", "Creature — Rogue Snake", "", 1, 1)
	gapb61Creature(g, me.ID, "Bear", "Creature — Bear", "", 1, 1)
	gapb61Creature(g, opp.ID, "Their Rogue", "Creature — Rogue", "", 1, 1)
	effectstest.Pr7bAttack(t, g, opp.ID, butch)
	if p, tt := effectstest.PtOf(t, g, butch); p != 5 || tt != 5 {
		t.Errorf("Butch is %d/%d with three other Rogues and/or Snakes, want 5/5", p, tt)
	}
}

func TestGapb61ButchGivesAMenaceCounterAndTheRogueType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	butch := gapb61Creature(g, me.ID, "Butch DeLoria, Tunnel Snake", "Legendary Creature — Human Rogue", gapb61ButchOracle, 2, 2)
	bear := gapb61Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{B}")
	// Not itself: "another target creature".
	if err := g.ActivateCatalogAbility(me.ID, butch, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(butch), Strict: true}); err == nil {
		t.Fatal("Butch targeted himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, butch, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterMenace); n != 1 {
		t.Errorf("Bear has %d menace counters, want 1", n)
	}
	var isRogue, isBear bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bear {
				isRogue, isBear = c.HasSubtypeWord("Rogue"), c.HasSubtypeWord("Bear")
			}
		}
	})
	if !isRogue || !isBear {
		t.Errorf("Bear is a Rogue: %v, still a Bear: %v; want both", isRogue, isBear)
	}
}

// --- Legate Lanius, Caesar's Ace ---------------------------------------------

func TestGapb61LaniusDecimatesATenthRoundedUpAndGrowsPerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	var theirs []uuid.UUID
	for i := 0; i < 11; i++ {
		theirs = append(theirs, gapb61Creature(g, opp.ID, "Their Bear", "Creature — Bear", "", 1, 1))
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Legate Lanius, Caesar's Ace", TypeLine: "Legendary Creature — Human Soldier",
		OracleID: gapb61LaniusOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{B}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// 11 creatures: a tenth, rounded up, is 2. Each is the opponent's choice.
	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked to sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs[0], theirs[1])
	effectstest.PassPriorityAroundTable(t, g)
	left := 0
	for _, c := range theirs {
		if g.Battlefield.Contains(c) {
			left++
		}
	}
	if left != 9 {
		t.Errorf("%d of their 11 creatures remain, want 9", left)
	}
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("Lanius has %d +1/+1 counters after two opponent sacrifices, want 2", n)
	}
}

func TestGapb61EDENoQuestCounterWhenNoMoreAttackThanItHas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ede := gapb61Creature(g, me.ID, "ED-E, Lonesome Eyebot", "Legendary Artifact Creature — Robot", gapb61EDEOracle, 2, 1)
	if err := g.AddCounter(ede, "quest", 1); err != nil {
		t.Fatal(err)
	}
	effectstest.Pr7bAttack(t, g, opp.ID, ede)
	if n := effectstest.CountersOn(g, ede, "quest"); n != 1 {
		t.Errorf("1 attacker vs 1 quest counter: %d counters, want 1 (the intervening if fails)", n)
	}
}

func TestGapb61EDEGetsOneQuestCounterPerAttackWhateverTheNumberOfAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ede := gapb61Creature(g, me.ID, "ED-E, Lonesome Eyebot", "Legendary Artifact Creature — Robot", gapb61EDEOracle, 2, 1)
	a := gapb61Creature(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := gapb61Creature(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, ede, a, b)
	if n := effectstest.CountersOn(g, ede, "quest"); n != 1 {
		t.Errorf("three attackers: %d quest counters, want exactly 1 (once per declaration)", n)
	}
}

// --- Pre-War Formalwear ----------------------------------------------------------

func TestGapb61FormalwearReturnsACheapCreatureAndAttachesToIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.GfDead(t, g, me, "Cheap Bear", "Creature — Bear", "{1}{G}")
	pricey := effectstest.GfDead(t, g, me, "Big Wurm", "Creature — Wurm", "{4}{G}")
	effectstest.CastCatalogSpell(t, g, "Pre-War Formalwear", "Artifact — Equipment", gapb61FormalwearOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, pricey) {
			t.Error("a creature with mana value 5 was offered")
		}
		effectstest.PickCard(t, g, me.ID, cheap)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if !me.Graveyard.Contains(pricey) {
		t.Error("the expensive creature left the graveyard")
	}
	var host uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == gapb61FormalwearOracle {
				host = c.AttachedTo.ID
			}
		}
	})
	if host != cheap {
		t.Fatalf("Formalwear is attached to %v, want the returned creature %v", host, cheap)
	}
	if p, tt := effectstest.PtOf(t, g, cheap); p != 4 || tt != 4 {
		t.Errorf("equipped Bear is %d/%d, want 4/4", p, tt)
	}
	effectstest.AssertKeywords(t, g, cheap, "vigilance")
}

// --- The Prydwen, Steel Flagship ----------------------------------------------------

func TestGapb61PrydwenMakesAKnightForAnotherNontokenArtifactAndItIsBigThatTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	prydwen := effectstest.PushCatalogPermanent(g, me.ID, "The Prydwen, Steel Flagship", "Legendary Artifact — Vehicle", gapb61PrydwenOracle, false)
	_ = prydwen
	effectstest.CastCatalogSpell(t, g, "Test Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	knights := gapb61Named(g, me.ID, "Human Knight")
	if len(knights) != 1 {
		t.Fatalf("%d Human Knights after another artifact entered, want 1", len(knights))
	}
	if p, tt := effectstest.PtOf(t, g, knights[0]); p != 4 || tt != 4 {
		t.Errorf("the Knight is %d/%d the turn an artifact entered, want 4/4", p, tt)
	}
	// An artifact creature counts; a token artifact does not.
	effectstest.CastCatalogSpell(t, g, "Test Golem", "Artifact Creature — Golem", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb61Named(g, me.ID, "Human Knight")); n != 2 {
		t.Errorf("%d Knights after a second artifact, want 2", n)
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.JunkToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb61Named(g, me.ID, "Human Knight")); n != 2 {
		t.Errorf("%d Knights after a Junk token entered, want still 2", n)
	}
}

func TestGapb61HumanKnightIsTwoByTwoWhenNoArtifactEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var knight uuid.UUID
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.GapB61HumanKnightToken(), 1)
	})
	knights := gapb61Named(g, me.ID, "Human Knight")
	if len(knights) != 1 {
		t.Fatalf("%d Knights", len(knights))
	}
	knight = knights[0]
	if p, tt := effectstest.PtOf(t, g, knight); p != 2 || tt != 2 {
		t.Errorf("a Knight with no artifact entered this turn is %d/%d, want 2/2", p, tt)
	}
}

// --- Young Deathclaws ----------------------------------------------------------------

func TestGapb61YoungDeathclawsGivesEachCreatureCardScavengeAtItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gapb61Creature(g, me.ID, "Young Deathclaws", "Creature — Lizard Mutant", gapb61YoungDeathclawsOracle, 4, 2)
	bear := gapb61Creature(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	dead := effectstest.GfDead(t, g, me, "Dead Troll", "Creature — Troll", "{1}{B}{B}")
	instant := effectstest.GfDead(t, g, me, "Dead Spell", "Instant", "{1}{B}")
	theirs := effectstest.GfDead(t, g, opp, "Their Dead", "Creature — Troll", "{B}")
	for name, id := range map[string]uuid.UUID{"an instant": instant, "an opponent's creature": theirs} {
		if _, ok := effectstest.GfEncoreCost(g, id); ok {
			t.Errorf("%s was given scavenge", name)
		}
	}
	if got, ok := effectstest.GfEncoreCost(g, dead); !ok || got != "{1}{B}{B}" {
		t.Fatalf("the creature card's scavenge cost is %q (has a row: %v), want its mana cost {1}{B}{B}", got, ok)
	}
	// Generic mana cannot pay a coloured cost.
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("scavenge was paid with colourless mana for a {1}{B}{B} cost")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	effectstest.FloatForTest(g, me, "CBB")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("scavenge for {1}{B}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) {
		t.Error("scavenge did not exile the card")
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("the target has %d +1/+1 counters, want the card's power 2", n)
	}
}

// --- Caesar, Legion's Emperor --------------------------------------------------------

func caesarBoard(t *testing.T) (g *game.Game, me, opp *game.Player, caesar, fodder uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	caesar = gapb61Creature(g, me.ID, "Caesar, Legion's Emperor", "Legendary Creature — Human Soldier", gapb61CaesarOracle, 4, 4)
	fodder = gapb61Creature(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	return g, me, opp, caesar, fodder
}

func TestGapb61CaesarSacrificesThenDrawsAndDamagesByCreatureTokens(t *testing.T) {
	g, me, opp, caesar, fodder := caesarBoard(t)
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	hand, life, oppLife := me.Hand.Size(), me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, caesar)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if opp.Life != oppLife-2 {
		t.Errorf("the opponent took %d, want 2 (two creature tokens)", oppLife-opp.Life)
	}
}

func TestGapb61CaesarMakesTwoHastySoldiersTappedAndAttacking(t *testing.T) {
	g, me, opp, caesar, fodder := caesarBoard(t)
	effectstest.DeclareAttack(t, g, opp.ID, caesar)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	soldiers := gapb61Named(g, me.ID, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("%d Soldiers, want 2", len(soldiers))
	}
	g.ReadSnapshot(func() {
		for _, id := range soldiers {
			c, _ := g.LookupCardForEffect(id)
			if !c.Tapped || c.AttackingTarget == uuid.Nil {
				t.Errorf("a Soldier is tapped=%v attacking=%v, want both", c.Tapped, c.AttackingTarget != uuid.Nil)
			}
		}
	})
}

func TestGapb61CaesarDecliningTheSacrificeDoesNothing(t *testing.T) {
	g, me, opp, caesar, fodder := caesarBoard(t)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, caesar)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("a declined sacrifice still asked for modes")
	}
	if !g.Battlefield.Contains(fodder) || me.Hand.Size() != hand {
		t.Error("something happened without the sacrifice")
	}
}
