package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b819_cards_test.go — Alliances, damage, token makers, enters and
// dies triggers, batch 819 (#5634, S82 OLD). Every test plays the real
// catalog entry through the engine.

const (
	oldB819WarBeast   = "0c93c418-e10b-4a6d-bddb-3b786bb6a928"
	oldB819DeathSpark = "4b23a89c-89ba-41ad-b894-ec02221689eb"
	oldB819Horror     = "a4c15223-784f-43a7-8626-30f96ccfdcaf"
	oldB819Sphere     = "4230c7a3-4591-411e-afab-d2bfb58446de"
	oldB819HomeGuard  = "5f7ae611-a78d-4535-9d76-755001bda34c"
	oldB819Starfish   = "663d6c61-e269-4f2f-9d8f-49c4c23f05d8"
	oldB819Phelddagr  = "d0e99535-0ea0-4aa5-99f5-6c5255b27c38"
	oldB819SolGrail   = "0396ef40-3774-4684-9971-160aaccf6ac6"
	oldB819Horde      = "fd2a94d3-901d-4b46-aa9b-a30ed6dea6bd"
	oldB819Bookworms  = "ff3eeed9-6801-4cab-a97a-72f5f28acb24"
)

func TestOldB819CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB819WarBeast: "Phyrexian War Beast", oldB819DeathSpark: "Death Spark", oldB819Horror: "Krovikan Horror",
		oldB819Sphere: "Phantasmal Sphere", oldB819HomeGuard: "Kjeldoran Home Guard", oldB819Starfish: "Spiny Starfish",
		oldB819Phelddagr: "Phelddagrif", oldB819SolGrail: "Sol Grail", oldB819Horde: "Balduvian Horde",
		oldB819Bookworms: "Insidious Bookworms",
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

func TestOldB819WarBeastCostsALandAndALifePointWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.B12Push(g, me.ID, "Phyrexian War Beast", "Artifact Creature — Phyrexian Beast", oldB819WarBeast, 3, 4)
	a := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	b := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	life := me.Life
	effectstest.Destroy(t, g, beast)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Errorf("the chosen land should be sacrificed and only it (Forest %v, Island %v)", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, me.Life)
	}
}

func TestOldB819DeathSparkPingsAndComesBackUnderACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Death Spark", "Instant", oldB819DeathSpark, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, opp.Life)
	}
	t.Run("a creature card directly above it", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		spark := effectstest.GraveyardCardOf(g, me, "Death Spark", "Instant", "{R}", oldB819DeathSpark)
		effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
		effectstest.BeginUpkeep(t, g, me.ID)
		oldB819Mana(me, "R")
		effectstest.AnswerPayUnless(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(spark) || me.Graveyard.Contains(spark) {
			t.Error("paying {1} should return Death Spark to hand")
		}
	})
	t.Run("a noncreature card directly above it", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		spark := effectstest.GraveyardCardOf(g, me, "Death Spark", "Instant", "{R}", oldB819DeathSpark)
		effectstest.GraveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
		effectstest.BeginUpkeep(t, g, me.ID)
		if effectstest.HasPayUnlessFor(g, me.ID) || !me.Graveyard.Contains(spark) {
			t.Error("no creature card above it, so nothing triggers")
		}
	})
	t.Run("a creature card below it, not above", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
		spark := effectstest.GraveyardCardOf(g, me, "Death Spark", "Instant", "{R}", oldB819DeathSpark)
		effectstest.BeginUpkeep(t, g, me.ID)
		if effectstest.HasPayUnlessFor(g, me.ID) || !me.Graveyard.Contains(spark) {
			t.Error("the creature card is under it, so nothing triggers")
		}
	})
	t.Run("declining the payment leaves it", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		spark := effectstest.GraveyardCardOf(g, me, "Death Spark", "Instant", "{R}", oldB819DeathSpark)
		effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
		effectstest.BeginUpkeep(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(spark) {
			t.Error("an unpaid Death Spark stays in the graveyard")
		}
	})
}

func TestOldB819KrovikanHorrorReturnsAtAnyEndStepAndSacrificesForDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horror := effectstest.GraveyardCardOf(g, me, "Krovikan Horror", "Creature — Horror Spirit", "{3}{B}", oldB819Horror)
	effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: opp.ID}) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(horror) {
		t.Fatal("Krovikan Horror did not return to hand at an opponent's end step")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	pit := effectstest.PushCatalogPermanent(g2, me2.ID, "Krovikan Horror", "Creature — Horror Spirit", oldB819Horror, false)
	fodder := effectstest.B12Creature(g2, me2.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g2)
	oldB819Mana(me2, "C")
	life := opp2.Life
	if err := g2.ActivateCatalogAbility(me2.ID, pit, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp2.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(fodder) || opp2.Life != life-1 {
		t.Errorf("fodder on battlefield %v, life %d → %d: want it sacrificed and 1 damage", g2.Battlefield.Contains(fodder), life, opp2.Life)
	}
}

func TestOldB819PhantasmalSphereGrowsPaysAndLeavesAnOrbForAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := effectstest.B12Push(g, me.ID, "Phantasmal Sphere", "Creature — Illusion", oldB819Sphere, 0, 1)
	effectstest.BeginUpkeep(t, g, me.ID)
	if effectstest.CountersOn(g, sphere, game.CounterPlusOne) != 1 {
		t.Fatalf("counters %d, want 1", effectstest.CountersOn(g, sphere, game.CounterPlusOne))
	}
	oldB819Mana(me, "U")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sphere) {
		t.Fatal("a paid upkeep sacrificed it")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.RemoraUpkeepPrompt(t, g, me.ID); p.PayCost != "{1}{1}" {
		t.Errorf("second upkeep costs %q, want {1}{1}", p.PayCost)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sphere) {
		t.Fatal("an unpaid Sphere was not sacrificed")
	}
	// target opponent: a player pick, then the token
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	var orb uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Orb" && c.Controller == opp.ID {
			orb = c.InstanceID
		}
	}
	if orb == uuid.Nil {
		t.Fatal("the opponent has no Orb")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, orb); p != 2 || tough != 2 || !effectstest.HasKeywordOnBattlefield(t, g, orb, "flying") {
		t.Errorf("the Orb is %d/%d, flying %v, want a 2/2 flier (two counters as it left)", p, tough, effectstest.HasKeywordOnBattlefield(t, g, orb, "flying"))
	}
}

func TestOldB819KjeldoranHomeGuardShrinksAndLeavesADeserterAfterFighting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Kjeldoran Home Guard", "Creature — Human Soldier", oldB819HomeGuard, 1, 6)
	idle := effectstest.B12Push(g, me.ID, "Idle Guard", "Creature — Human Soldier", oldB819HomeGuard, 1, 6)
	effectstest.DeclareAttack(t, g, opp.ID, guard)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, guard, "-0/-1") != 1 {
		t.Errorf("-0/-1 counters on the attacker: %d, want 1", effectstest.CountersOn(g, guard, "-0/-1"))
	}
	if effectstest.CountersOn(g, idle, "-0/-1") != 0 {
		t.Errorf("a creature that did not fight has %d counters", effectstest.CountersOn(g, idle, "-0/-1"))
	}
	if n := effectstest.CountOnBattlefield(g, "Deserter", me.ID); n != 1 {
		t.Errorf("Deserter tokens: %d, want exactly 1", n)
	}
}

func TestOldB819SpinyStarfishMakesAStarfishForEachRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fish := effectstest.B12Push(g, me.ID, "Spiny Starfish", "Creature — Starfish", oldB819Starfish, 0, 1)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		oldB819Mana(me, "U")
		if err := g.ActivateCatalogAbility(me.ID, fish, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("regenerate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.DestroyIt(t, g, fish)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(fish) {
		t.Fatal("the Starfish did not regenerate")
	}
	effectstest.AdvanceToStepOnThisTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Starfish", me.ID); n != 2 {
		t.Errorf("Starfish tokens: %d, want 2 (one for each regeneration)", n)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B12Push(g2, me2.ID, "Spiny Starfish", "Creature — Starfish", oldB819Starfish, 0, 1)
	effectstest.AdvanceToMain(t, g2)
	effectstest.AdvanceToStepOnThisTurn(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.CountOnBattlefield(g2, "Starfish", me2.ID); n != 0 {
		t.Errorf("a Starfish that never regenerated made %d tokens", n)
	}
}

func TestOldB819PhelddagrifGivesTheOpponentWhatItPrints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	p := effectstest.B12Push(g, me.ID, "Phelddagrif", "Legendary Creature — Phelddagrif", oldB819Phelddagr, 4, 4)
	effectstest.AdvanceToMain(t, g)
	target := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	oldB819Mana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, p, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("{G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Hippo", opp.ID); n != 1 || !effectstest.HasKeywordOnBattlefield(t, g, p, "trample") {
		t.Errorf("Hippos %d, trample %v", n, effectstest.HasKeywordOnBattlefield(t, g, p, "trample"))
	}
	life := opp.Life
	oldB819Mana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, p, 1, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+2 || !effectstest.HasKeywordOnBattlefield(t, g, p, "flying") {
		t.Errorf("life %d → %d, flying %v", life, opp.Life, effectstest.HasKeywordOnBattlefield(t, g, p, "flying"))
	}
	hand := opp.Hand.Size()
	oldB819Mana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, p, 2, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("{U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(p) || !me.Hand.Contains(p) {
		t.Error("Phelddagrif did not return to its owner's hand")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the opponent's hand %d → %d, want one drawn card", hand, opp.Hand.Size())
	}
}

func TestOldB819SolGrailAddsTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	grail := effectstest.HandCardFull(me, "Sol Grail", "Artifact", "{3}", oldB819SolGrail, nil)
	if err := g.CastSpell(me.ID, grail, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(grail) {
		t.Fatal("Sol Grail did not enter")
	}
	if err := g.ActivateManaAbility(me.ID, grail, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "R" {
		t.Errorf("mana pool %v, want one red", me.ManaPool)
	}
}

func TestOldB819BalduvianHordeKeepsItselfOnlyByDiscardingAtRandom(t *testing.T) {
	t.Run("discarding", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.HandCardFull(me, "Spare", "Instant", "{U}", "", []string{"U"})
		hand := me.Hand.Size()
		horde := effectstest.HandCardFull(me, "Balduvian Horde", "Creature — Human Barbarian", "{2}{R}{R}", oldB819Horde, []string{"R"})
		g.WithWriteLock(func() {})
		if err := g.CastSpell(me.ID, horde, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(horde) || me.Hand.Size() != hand-1 {
			t.Errorf("on battlefield %v, hand %d → %d: want it kept and one card discarded", g.Battlefield.Contains(horde), hand, me.Hand.Size())
		}
	})
	t.Run("declining", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		horde := effectstest.HandCardFull(me, "Balduvian Horde", "Creature — Human Barbarian", "{2}{R}{R}", oldB819Horde, []string{"R"})
		if err := g.CastSpell(me.ID, horde, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(horde) {
			t.Error("an undiscarded Horde stays")
		}
	})
}

func TestOldB819InsidiousBookwormsMakeAPlayerDiscardAtRandomForAPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worm := effectstest.B12Push(g, me.ID, "Insidious Bookworms", "Creature — Worm", oldB819Bookworms, 1, 1)
	hand := opp.Hand.Size()
	effectstest.Destroy(t, g, worm)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB819Mana(me, "B", "B")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("the opponent's hand %d → %d, want one discarded", hand, opp.Hand.Size())
	}
}

func oldB819Mana(p *game.Player, colors ...string) {
	p.ManaPool.AddMana(effectstest.ManaTokens(colors...)...)
}
