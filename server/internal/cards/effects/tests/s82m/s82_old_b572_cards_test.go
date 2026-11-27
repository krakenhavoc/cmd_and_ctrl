package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b572_cards_test.go — Guildpact's life gain and loss, pump, damage and
// enters-and-dies cards, OLD batch 572 (#5599). Each test plays the real
// catalog entry.

const (
	oldB572GhostCouncilOracle = "2ba16e4d-2de7-4906-801a-4ba2b90b5050"
	oldB572OgreOracle         = "38872730-0fd1-414e-8015-29d1bc649913"
	oldB572SoulsOracle        = "5a12e989-d6bc-441e-88dd-779e68a9fee3"
	oldB572GathererOracle     = "3b19054c-ba33-44b4-b328-bb8ea3f52dab"
	oldB572LeylineOracle      = "971f7b30-0088-4e12-bc1a-262acc7f5f90"
	oldB572EarthSurgeOracle   = "fb8fae53-3b06-41d8-baa0-638820a4b550"
	oldB572InfernoOracle      = "6783c46d-e4e5-4b43-9d97-6eae8aab4cb9"
	oldB572FeedbackOracle     = "a01071c5-aca5-4faa-8a8a-30b50b642f42"
	oldB572TiborOracle        = "ee6d6f2f-68d5-466a-86c6-20fdb360e629"
	oldB572OrderOracle        = "66418c65-a15f-489e-8a2d-6cee1b5d977c"
	oldB572SwallowerOracle    = "b345c413-df5a-4dae-8b92-5fab54a097fc"
)

func TestOldB572CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB572GhostCouncilOracle: "Ghost Council of Orzhova",
		oldB572OgreOracle:         "Poisonbelly Ogre",
		oldB572SoulsOracle:        "Souls of the Faultless",
		oldB572GathererOracle:     "Gatherer of Graces",
		oldB572LeylineOracle:      "Leyline of the Meek",
		oldB572EarthSurgeOracle:   "Earth Surge",
		oldB572InfernoOracle:      "Living Inferno",
		oldB572FeedbackOracle:     "Parallectric Feedback",
		oldB572TiborOracle:        "Tibor and Lumia",
		oldB572OrderOracle:        "Order of the Stars",
		oldB572SwallowerOracle:    "Sky Swallower",
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

func oldB572PT(t *testing.T, g *game.Game, id uuid.UUID, what string, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("%s is %d/%d, want %d/%d", what, gp, gt, p, tough)
	}
}

func oldB572OnBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

func oldB572Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if len(g.PendingChoices) > 0 {
			return
		}
	}
}

// --- Ghost Council of Orzhova ----------------------------------------------

func TestOldB572GhostCouncilDrainsOnEntryAndBlinksBySacrificing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	myLife, oppLife := me.Life, opp.Life
	council := effectstest.CastHoldCreature(t, g, "Ghost Council of Orzhova", "Legendary Creature — Spirit", oldB572GhostCouncilOracle, 4, 4)
	if c := effectstest.FirstPendingTarget(g); c != nil {
		if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	oldB572Settle(t, g)
	if me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Fatalf("entry drain: me %d→%d, opp %d→%d", myLife, me.Life, oppLife, opp.Life)
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, council, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB572OnBattlefield(g, council) {
		t.Fatal("the Council was not exiled")
	}
	if oldB572OnBattlefield(g, fodder) {
		t.Error("the sacrificed creature is still there")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	back := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == oldB572GhostCouncilOracle && c.Controller == me.ID {
				back = true
			}
		}
	})
	if !back {
		t.Error("the Council did not return at the next end step")
	}
}

// --- Poisonbelly Ogre ------------------------------------------------------

func TestOldB572PoisonbellyOgreDrainsTheControllerOfEachEnteringCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Poisonbelly Ogre", "Creature — Ogre Warrior", oldB572OgreOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	myLife, oppLife := me.Life, opp.Life
	mine := effectstest.B12Push(g, me.ID, "Entering Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: mine, Actor: me.ID}) })
	oldB572Settle(t, g)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, Actor: opp.ID}) })
	oldB572Settle(t, g)
	if me.Life != myLife-1 {
		t.Errorf("my life %d→%d, want 1 lost for my own creature", myLife, me.Life)
	}
	if opp.Life != oppLife-1 {
		t.Errorf("their life %d→%d, want 1 lost for their creature", oppLife, opp.Life)
	}
}

// --- Souls of the Faultless ---------------------------------------------------

func TestOldB572SoulsOfTheFaultlessGainsAndDrainsTheAttackerForCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	souls := effectstest.B12Push(g, opp.ID, "Souls of the Faultless", "Creature — Spirit", oldB572SoulsOracle, 0, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, souls, "defender") {
		t.Error("no defender")
	}
	myLife, oppLife := me.Life, opp.Life
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{souls: atk})
	effectstest.Pr7bDamageStep(t, g)
	oldB572Settle(t, g)
	if opp.Life != oppLife+3 {
		t.Errorf("Souls' controller life %d→%d, want +3", oppLife, opp.Life)
	}
	if me.Life != myLife-3 {
		t.Errorf("attacking player life %d→%d, want -3", myLife, me.Life)
	}
}

func TestOldB572SoulsOfTheFaultlessIgnoresNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	souls := effectstest.B12Push(g, opp.ID, "Souls of the Faultless", "Creature — Spirit", oldB572SoulsOracle, 0, 4)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Goblin", 1, 1)
	myLife, oppLife := me.Life, opp.Life
	effectstest.Damage(t, g, pinger, souls, 2)
	oldB572Settle(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Errorf("noncombat damage triggered it: me %d→%d, opp %d→%d", myLife, me.Life, oppLife, opp.Life)
	}
}

// --- Gatherer of Graces ------------------------------------------------------

func TestOldB572GathererOfGracesGrowsPerAuraAndSacrificesAnAuraToRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gath := effectstest.B12Push(g, me.ID, "Gatherer of Graces", "Creature — Human Druid", oldB572GathererOracle, 1, 2)
	other := effectstest.AuraBear(g, me.ID)
	aura1 := effectstest.AuraCast(t, g, "Quicksilver Dagger", "5c77bf76-bb0f-4e44-babe-0e7b7b555e1f", gath)
	oldB572PT(t, g, gath, "the Gatherer with one Aura", 2, 3)
	// An Aura on another creature does not count for the Gatherer.
	aura2 := effectstest.AuraCast(t, g, "Quicksilver Dagger", "5c77bf76-bb0f-4e44-babe-0e7b7b555e1f", other)
	oldB572PT(t, g, gath, "the Gatherer with an Aura on another creature too", 2, 3)
	_ = aura2
	// Sacrificing an Aura (any Aura you control) regenerates.
	if err := g.ActivateCatalogAbility(me.ID, gath, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{aura2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB572OnBattlefield(g, aura2) {
		t.Error("the Aura was not sacrificed")
	}
	if !oldB572OnBattlefield(g, aura1) {
		t.Error("the wrong Aura left")
	}
	effectstest.Damage(t, g, other, gath, 10)
	g.RunStateChecksForTest()
	if !oldB572OnBattlefield(g, gath) {
		t.Error("the Gatherer was not regenerated")
	}
}

// --- Leyline of the Meek ---------------------------------------------------------

func TestOldB572LeylineOfTheMeekPumpsEveryCreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Leyline of the Meek", "Enchantment", oldB572LeylineOracle, false)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	card := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB572PT(t, g, mine, "my token", 2, 2)
	oldB572PT(t, g, theirs, "their token", 2, 2)
	oldB572PT(t, g, card, "a nontoken creature", 2, 2)
}

// --- Earth Surge ---------------------------------------------------------------------

func TestOldB572EarthSurgePumpsOnlyLandCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Earth Surge", "Enchantment", oldB572EarthSurgeOracle, false)
	landCreature := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Manland", TypeLine: "Land Creature — Elemental",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	plainLand := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB572PT(t, g, landCreature, "an opponent's land creature", 4, 4)
	oldB572PT(t, g, bear, "a nonland creature", 2, 2)
	if p, tough := effectstest.PtOf(t, g, plainLand); p != 0 || tough != 0 {
		t.Errorf("a plain land is %d/%d, want 0/0 (not a creature)", p, tough)
	}
}

// --- Living Inferno -------------------------------------------------------------------

func TestOldB572LivingInfernoDividesItsPowerAndTakesTheDamageBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inferno := effectstest.B12Push(g, me.ID, "Living Inferno", "Creature — Elemental", oldB572InfernoOracle, 8, 5)
	a := effectstest.B12Creature(g, opp.ID, "Target A", "Creature — Bear", 1, 3)
	b := effectstest.B12Creature(g, opp.ID, "Target B", "Creature — Bear", 2, 6)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.ActivateCatalogAbility(me.ID, inferno, 0, game.ActivateAbilityParams{Strict: true, Targets: targets,
		Distribution: map[uuid.UUID]int{a: 3, b: 5}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if oldB572OnBattlefield(g, a) {
		t.Error("the creature dealt 3 (toughness 3) lived")
	}
	if !oldB572OnBattlefield(g, b) {
		t.Error("the creature dealt 5 (toughness 6) died")
	}
	if !oldB572OnBattlefield(g, inferno) {
		t.Fatal("the Inferno died to 1+2 damage (toughness 5)")
	}
	if d := effectstest.DamageMarkedOn(g, inferno); d != 3 {
		t.Errorf("the Inferno has %d damage, want 3 (1 from A and 2 from B, A's dead or not)", d)
	}
}

func TestOldB572LivingInfernoTakesEachTargetsPowerBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inferno := effectstest.B12Push(g, me.ID, "Living Inferno", "Creature — Elemental", oldB572InfernoOracle, 8, 5)
	a := effectstest.B12Creature(g, opp.ID, "Target A", "Creature — Bear", 2, 9)
	if err := g.ActivateCatalogAbility(me.ID, inferno, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}, Distribution: map[uuid.UUID]int{a: 8}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.DamageMarkedOn(g, inferno); d != 2 {
		t.Errorf("the Inferno has %d damage, want 2 (the target's power)", d)
	}
	if d := effectstest.DamageMarkedOn(g, a); d != 8 {
		t.Errorf("the target has %d damage, want the whole 8", d)
	}
}

// --- Parallectric Feedback ----------------------------------------------------------------

func TestOldB572ParallectricFeedbackDamagesTheSpellsControllerForItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastCatalogSpell(t, g, "Costly Spell", "Sorcery", "", nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == spell {
				g.Stack.Cards[i].ManaCost = "{3}{R}"
			}
		}
	})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Parallectric Feedback", "Instant", oldB572FeedbackOracle, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("the spell's controller life %d→%d, want 4 damage", life, me.Life)
	}
}

// --- Tibor and Lumia ------------------------------------------------------------------------

func TestOldB572TiborAndLumiaTriggersByColor(t *testing.T) {
	for _, tc := range []struct {
		color string
		want  int
	}{{"U", 1}, {"R", 1}, {"G", 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		tibor := effectstest.B12Push(g, me.ID, "Tibor and Lumia", "Legendary Creature — Human Wizard", oldB572TiborOracle, 3, 3)
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Colored Spell", TypeLine: "Instant", Colors: []string{tc.color}, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", tc.color, err)
		}
		n := effectstest.TriggersOnStackFrom(g, tibor)
		if effectstest.FirstPendingTarget(g) != nil {
			n++ // a targeted trigger asks for its target before it goes on the stack
		}
		if n != tc.want {
			t.Errorf("a %s spell put %d Tibor and Lumia triggers on the stack, want %d", tc.color, n, tc.want)
		}
	}
}

func TestOldB572TiborAndLumiaRedSpellPingsEveryCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tibor := effectstest.B12Push(g, me.ID, "Tibor and Lumia", "Legendary Creature — Human Wizard", oldB572TiborOracle, 3, 3)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Red Spell", TypeLine: "Instant", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.DamageMarkedOn(g, walker); d != 1 {
		t.Errorf("a creature without flying has %d damage, want 1", d)
	}
	if d := effectstest.DamageMarkedOn(g, tibor); d != 1 {
		t.Errorf("Tibor and Lumia has %d damage, want 1 (it has no flying)", d)
	}
	if d := effectstest.DamageMarkedOn(g, flyer); d != 0 {
		t.Errorf("a creature with flying has %d damage, want 0", d)
	}
}

// --- Order of the Stars -----------------------------------------------------------------------

func TestOldB572OrderOfTheStarsHasProtectionFromTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	order := effectstest.CastHoldCreature(t, g, "Order of the Stars", "Creature — Human Cleric", oldB572OrderOracle, 0, 1)
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceColor {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatalf("no colour prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveColorChoice(prompt.ID, me.ID, "R"); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, order, "protection from red") {
		t.Error("no protection from the chosen colour")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, order, "protection from white") {
		t.Error("protection from a colour that was not chosen")
	}
}

// --- Sky Swallower ------------------------------------------------------------------------------

func TestOldB572SkySwallowerGivesTheOpponentEveryOtherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	swallower := effectstest.CastHoldCreature(t, g, "Sky Swallower", "Creature — Leviathan", oldB572SwallowerOracle, 8, 8)
	if c := effectstest.FirstPendingTarget(g); c != nil {
		if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	oldB572Settle(t, g)
	for name, id := range map[string]uuid.UUID{"creature": bear, "land": land} {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c == nil || c.Controller != opp.ID {
			t.Errorf("my %s was not given to the opponent", name)
		}
	}
	if c := effectstest.FindBattlefieldCardForTest(g, swallower); c == nil || c.Controller != me.ID {
		t.Error("the Swallower itself changed hands")
	}
	if c := effectstest.FindBattlefieldCardForTest(g, theirs); c == nil || c.Controller != opp.ID {
		t.Error("the opponent's own creature changed hands")
	}
}
