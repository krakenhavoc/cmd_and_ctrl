package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b322_cards_test.go — Magic Origins, batch 322 (S82 older sets,
// #5549). Every test plays the real catalog entry through the engine.

const (
	oldB322GaeasRevenge = "26f6cc1e-d85e-45a3-a1ed-7f09d7cfc7df"
	oldB322Nightmare    = "375932e6-1b3e-48dc-8154-9b664c3add34"
	oldB322Sentinel     = "572b1aeb-d2a0-433d-b167-86e683c2127d"
	oldB322Nissa        = "35754a21-9fba-4370-a254-292918a777ba"
	oldB322Jace         = "594f6881-c059-46f8-aa4e-7151d502de73"
	oldB322Chandra      = "6f6b8e14-c247-4dfd-8b39-38c7a5f2f354"
	oldB322AegisAngel   = "5902e074-7154-4fc7-8776-b87e8b1925e7"
	oldB322NivixBarrier = "aa4cb5fb-e81c-410f-8986-645671b0a1d3"
	oldB322Alhammarret  = "c2531360-61f8-4edc-886b-408811a0639d"
	oldB322DeepSeaTerr  = "36775412-0280-4acc-bae8-3ea5d4e341be"
	oldB322Willbreaker  = "7c365041-806f-42d6-a2ff-40e20dd8e9f0"
	oldB322GiantGrowth  = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
	oldB322Bolt         = "4457ed35-7c10-48c8-9776-456485fdf070"
	oldB322Bear         = "Creature — Bear"
)

func TestOldB322EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB322GaeasRevenge, oldB322Nightmare, oldB322Sentinel, oldB322Nissa, oldB322Nissa + "#1",
		oldB322Jace, oldB322Jace + "#1", oldB322Chandra, oldB322Chandra + "#1", oldB322AegisAngel,
		oldB322NivixBarrier, oldB322Alhammarret, oldB322DeepSeaTerr, oldB322Willbreaker,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB322Row(oracle, front, frontType, frontCost, back, backType, power, toughness, loyalty string, colors []string) cards.Card {
	row := effectstest.TransformRow(oracle, front, frontType, frontCost, back, backType, "", "", colors)
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = power, toughness
	row.CardFaces[1].Loyalty = loyalty
	return row
}

func oldB322NissaRow() cards.Card {
	return oldB322Row(oldB322Nissa, "Nissa, Vastwood Seer", "Legendary Creature — Elf Scout", "{2}{G}",
		"Nissa, Sage Animist", "Legendary Planeswalker — Nissa", "2", "2", "3", []string{"G"})
}

func oldB322JaceRow() cards.Card {
	return oldB322Row(oldB322Jace, "Jace, Vryn's Prodigy", "Legendary Creature — Human Wizard", "{1}{U}",
		"Jace, Telepath Unbound", "Legendary Planeswalker — Jace", "0", "2", "5", []string{"U"})
}

func oldB322ChandraRow() cards.Card {
	return oldB322Row(oldB322Chandra, "Chandra, Fire of Kaladesh", "Legendary Creature — Human Shaman", "{1}{R}{R}",
		"Chandra, Roaring Flame", "Legendary Planeswalker — Chandra", "2", "2", "4", []string{"R"})
}

// oldB322Flipped is the transformed permanent of `oracle` that `p` controls.
func oldB322Flipped(g *game.Game, p *game.Player, oracle string) (uuid.UUID, bool) {
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == oracle && c.Controller == p.ID && effects.TransformedPermanent(c) {
				id = c.InstanceID
			}
		}
	})
	return id, id != uuid.Nil
}

// oldB322Walker puts a front face onto the battlefield and turns it over,
// returning the planeswalker with `loyalty` counters on it.
func oldB322Walker(t *testing.T, g *game.Game, row cards.Card, p *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	front := effectstest.ImportToBattlefield(t, g, row, p)
	var err error
	g.WithWriteLock(func() { err = g.ExileAndReturnTransformedForEffect(front, uuid.Nil) })
	if err != nil {
		t.Fatal(err)
	}
	id, ok := oldB322Flipped(g, p, row.OracleID.String())
	if !ok {
		t.Fatal("the front face did not come back transformed")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Counters[game.CounterLoyalty] = loyalty })
	return id
}

func TestOldB322GaeasRevengeRefusesNongreenSpellsButTakesGreenOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	revenge := effectstest.B12Push(g, me.ID, "Gaea's Revenge", "Creature — Elemental", oldB322GaeasRevenge, 8, 5)
	effectstest.AdvanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: revenge}}
	for _, p := range []*game.Player{me, opp} {
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oldB322Bolt, Owner: p.ID, Controller: p.ID})
		})
		effectstest.FloatForTest(g, p, "R")
		if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: target}); err == nil {
			t.Errorf("a nongreen spell from %v could target it", p.Name)
		}
	}
	green := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: green, Name: "Giant Growth", TypeLine: "Instant", ManaCost: "{G}", Colors: []string{"G"}, OracleID: oldB322GiantGrowth, Owner: opp.ID, Controller: opp.ID})
	})
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.FloatForTest(g, opp, "G")
	if err := g.CastSpell(opp.ID, green, game.CastSpellParams{Strict: true, Targets: target}); err != nil {
		t.Errorf("a green spell could not target it: %v", err)
	}
}

func TestOldB322GaeasRevengeCantBeCountered(t *testing.T) {
	spec, _ := effects.Lookup(oldB322GaeasRevenge)
	if !spec.CantBeCountered {
		t.Error("Gaea's Revenge can be countered")
	}
	if !effectstest.ContainsString(spec.PrintedKeywords, "haste") {
		t.Error("Gaea's Revenge has no haste")
	}
}

func TestOldB322NightmareCountsYourSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	nightmare := effectstest.B12Push(g, me.ID, "Nightmare", "Creature — Nightmare Horse", oldB322Nightmare, 3, 3)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", "", false)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp", "", false)
	if p, tough := effectstest.EffectivePTOf(t, g, nightmare); p != 3 || tough != 3 {
		t.Errorf("with three Swamps it is %d/%d, want 3/3", p, tough)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", "", false)
	if p, tough := effectstest.EffectivePTOf(t, g, nightmare); p != 4 || tough != 4 {
		t.Errorf("with four Swamps it is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB322SentinelTapsACreatureOfTheOpponentWhoseTurnItIs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sentinel of the Eternal Watch", "Creature — Giant Soldier", oldB322Sentinel, 4, 6)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB322Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 2, 2)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the opponent's creature was not tapped")
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("my own creature was tapped")
	}
}

func TestOldB322SentinelIsQuietOnYourOwnTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sentinel of the Eternal Watch", "Creature — Giant Soldier", oldB322Sentinel, 4, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 2, 2)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, theirs) {
		t.Error("it tapped a creature on its controller's own turn")
	}
}

func TestOldB322AegisAngelGrantsIndestructibleWhileYouControlIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 2, 2)
	angel := effectstest.RfdCast(t, g, "Aegis Angel", "Creature — Angel", oldB322AegisAngel, 5, 5)
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, foe), "indestructible") {
		t.Fatal("the target did not gain indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(angel) })
	g.RunStateChecksForTest()
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, foe), "indestructible") {
		t.Error("the target kept indestructible after the Angel left")
	}
}

func TestOldB322AegisAngelCantTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	angel := effectstest.RfdCast(t, g, "Aegis Angel", "Creature — Angel", oldB322AegisAngel, 5, 5)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: angel}); err == nil {
			t.Fatal("the Angel targeted itself")
		}
	}
}

func TestOldB322NivixBarrierShrinksAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Their Attacker", oldB322Bear, 5, 5)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	barrier := effectstest.CastFromHandForTest(t, g, opp, "Nivix Barrier", "Creature — Illusion Wall", "", oldB322NivixBarrier, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, opp.ID) != nil {
		effectstest.PickCard(t, g, opp.ID, attacker)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, attacker); p != 1 {
		t.Errorf("the attacker has %d power, want 1", p)
	}
	if effectstest.ZoneOf(g, barrier) != game.ZoneBattlefield {
		t.Error("the Barrier did not enter")
	}
}

func TestOldB322DeepSeaTerrorNeedsSevenCardsInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	terror := effectstest.B12Push(g, me.ID, "Deep-Sea Terror", "Creature — Serpent", oldB322DeepSeaTerr, 6, 6)
	for i := 0; i < 6; i++ {
		effectstest.PushGraveyardCardTyped(me, "Dead Card", "Sorcery")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(terror, opp.ID); err == nil {
		t.Fatal("it attacked with six cards in the graveyard")
	}
	effectstest.PushGraveyardCardTyped(me, "Dead Card", "Sorcery")
	if err := g.DeclareAttacker(terror, opp.ID); err != nil {
		t.Fatalf("it could not attack with seven cards in the graveyard: %v", err)
	}
}

func TestOldB322WillbreakerStealsATargetedCreatureWhileYouControlIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	breaker := effectstest.B12Push(g, me.ID, "Willbreaker", "Creature — Human Wizard", oldB322Willbreaker, 2, 3)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 2, 9)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB322Bear, 2, 9)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB322Bolt, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.E2Card(t, g, mine); c.Controller != me.ID {
		t.Error("it took a creature I already controlled")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB322Bolt, []game.TargetRef{{Kind: game.TargetCard, ID: foe}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.E2Card(t, g, foe); c.Controller != me.ID {
		t.Fatalf("the targeted creature is controlled by %v, want me", c.Controller)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(breaker) })
	g.RunStateChecksForTest()
	if c := effectstest.E2Card(t, g, foe); c.Controller != opp.ID {
		t.Error("the creature did not go home when Willbreaker left")
	}
}

func TestOldB322WillbreakerIgnoresAnOpponentsTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.B12Push(g, me.ID, "Willbreaker", "Creature — Human Wizard", oldB322Willbreaker, 2, 3)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 2, 9)
	other := effectstest.B12Creature(g, g.Seats[2].ID, "Third Bear", oldB322Bear, 2, 9)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.CastFromHandForTest(t, g, opp, "Lightning Bolt", "Instant", "{R}", oldB322Bolt,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.E2Card(t, g, other); c.Controller == me.ID {
		t.Error("a spell an opponent controls made me take a creature")
	}
	_ = foe
}

func TestOldB322AlhammarretNamesARevealedNonlandCardAndBarsItForOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for _, s := range g.Seats[1:] {
			s.Hand.Cards = nil
		}
	})
	bolt := effectstest.PushHandCardOfType(opp, "Lightning Bolt", "Instant")
	g.WithWriteLock(func() {
		for i := range opp.Hand.Cards {
			if opp.Hand.Cards[i].InstanceID == bolt {
				opp.Hand.Cards[i].OracleID = oldB322Bolt
				opp.Hand.Cards[i].ManaCost = "{R}"
			}
		}
	})
	effectstest.PushHandCardOfType(opp, "Forest", "Basic Land — Forest")
	opt := effectstest.PushHandCardOfType(opp, "Opt", "Instant")
	effectstest.RfdCast(t, g, "Alhammarret, High Arbiter", "Legendary Creature — Sphinx", oldB322Alhammarret, 5, 5)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no name was asked for")
	}
	labels := map[string]bool{}
	for _, o := range pick.PickOptions {
		labels[o.Label] = true
	}
	if !labels["Lightning Bolt"] || !labels["Opt"] || labels["Forest"] || len(labels) != 2 {
		t.Fatalf("the names offered are %v, want the two nonland cards", labels)
	}
	for i, o := range pick.PickOptions {
		if o.Label == "Lightning Bolt" {
			effectstest.AnswerOptionPick(t, g, me.ID, i)
		}
	}
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.FloatForTest(g, opp, "R")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("an opponent cast the chosen name")
	}
	if err := g.CastSpell(opp.ID, opt, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("an opponent could not cast another name: %v", err)
	}
}

func TestOldB322AlhammarretBarsNothingAfterItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for _, s := range g.Seats[1:] {
			s.Hand.Cards = nil
		}
	})
	bolt := effectstest.PushHandCardOfType(opp, "Lightning Bolt", "Instant")
	g.WithWriteLock(func() {
		for i := range opp.Hand.Cards {
			if opp.Hand.Cards[i].InstanceID == bolt {
				opp.Hand.Cards[i].OracleID = oldB322Bolt
				opp.Hand.Cards[i].ManaCost = "{R}"
			}
		}
	})
	angel := effectstest.RfdCast(t, g, "Alhammarret, High Arbiter", "Legendary Creature — Sphinx", oldB322Alhammarret, 5, 5)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(angel) })
	g.RunStateChecksForTest()
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.FloatForTest(g, opp, "R")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Errorf("the name is still barred after Alhammarret left: %v", err)
	}
}

// oldB322LandEnters puts a land onto the battlefield from the hand through the
// entry pipeline, so "whenever a land you control enters" sees it.
func oldB322LandEnters(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	id := effectstest.PushHandCardOfType(p, "Mountain", "Basic Land — Mountain")
	var err error
	g.WithWriteLock(func() { _, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{}) })
	if err != nil {
		t.Fatal(err)
	}
}

func TestOldB322NissaFrontSearchesForAForestAndFlipsOnTheSeventhLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.PushLibraryTyped(me, "Forest", "Basic Land — Forest")
	effectstest.PushLibraryTyped(me, "Island", "Basic Land — Island")
	for i := 0; i < 5; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	}
	nissa := effectstest.ImportToBattlefield(t, g, oldB322NissaRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Fatal("the Forest was not put into my hand")
	}
	oldB322LandEnters(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := oldB322Flipped(g, me, oldB322Nissa); ok {
		t.Fatal("it flipped with six lands")
	}
	oldB322LandEnters(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	id, ok := oldB322Flipped(g, me, oldB322Nissa)
	if !ok {
		t.Fatal("it did not flip with seven lands")
	}
	if effectstest.E2Card(t, g, id).Name != "Nissa, Sage Animist" || effectstest.LoyaltyOf(g, id) != 3 {
		t.Errorf("the back face is %q with %d loyalty", effectstest.E2Card(t, g, id).Name, effectstest.LoyaltyOf(g, id))
	}
	_ = nissa
}

func TestOldB322NissaSageAnimistPlusOneRevealsLandOntoBattlefieldOtherwiseHand(t *testing.T) {
	for _, land := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		walker := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Sage Animist", oldB322Nissa+"#1", 3)
		typeLine := "Instant"
		if land {
			typeLine = "Basic Land — Forest"
		}
		top := effectstest.SeedLibraryTop(me, "Top Card", typeLine, "")
		effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if land && effectstest.ZoneOf(g, top) != game.ZoneBattlefield {
			t.Errorf("a revealed land is in %v, want the battlefield", effectstest.ZoneOf(g, top))
		}
		if !land && !me.Hand.Contains(top) {
			t.Error("a revealed nonland card was not put into my hand")
		}
		if effectstest.LoyaltyOf(g, walker) != 4 {
			t.Errorf("loyalty %d, want 4", effectstest.LoyaltyOf(g, walker))
		}
	}
}

func TestOldB322NissaSageAnimistMinusTwoMakesAshaya(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Sage Animist", oldB322Nissa+"#1", 3)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Ashaya, the Awoken World" && c.IsToken() {
			n++
			if p, tough := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 4 || tough != 4 || !c.IsLegendary() {
				t.Errorf("Ashaya is %d/%d legendary=%v, want a legendary 4/4", p, tough, c.IsLegendary())
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Ashaya tokens, want 1", n)
	}
}

func TestOldB322NissaSageAnimistMinusSevenAnimatesUpToSixLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Sage Animist", oldB322Nissa+"#1", 7)
	var lands []game.TargetRef
	var first uuid.UUID
	for i := 0; i < 3; i++ {
		id := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		if i == 0 {
			first = id
		}
		g.WithWriteLock(func() {
			for j := range g.Battlefield.Cards {
				if g.Battlefield.Cards[j].InstanceID == id {
					g.Battlefield.Cards[j].Tapped = true
				}
			}
		})
		lands = append(lands, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{Targets: lands})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, first) {
		t.Error("the land was not untapped")
	}
	c := effectstest.E2Card(t, g, first)
	if p, tough := effectstest.EffectivePTOf(t, g, first); p != 6 || tough != 6 || !c.IsLand() || !c.IsCreature() {
		t.Errorf("the land is %d/%d land=%v creature=%v, want a 6/6 land creature", p, tough, c.IsLand(), c.IsCreature())
	}
	g.WithWriteLock(func() { g.Turn.Step = game.StepCleanup })
	if p, _ := effectstest.EffectivePTOf(t, g, first); p != 6 {
		t.Error("the animation ended with the turn; it has no duration")
	}
}

func TestOldB322JaceLootsAndFlipsWithFiveCardsInTheGraveyard(t *testing.T) {
	for _, five := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		jace := effectstest.ImportToBattlefield(t, g, oldB322JaceRow(), me)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, jace).SummonedThisTurn = false })
		n := 3
		if five {
			n = 4
		}
		for i := 0; i < n; i++ {
			effectstest.PushGraveyardCardTyped(me, "Dead Card", "Sorcery")
		}
		effectstest.SeedLibrary(me, "A", "B", "C")
		hand := me.Hand.Size()
		effectstest.B16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 {
			t.Fatalf("five=%v: hand %d, want %d after the draw", five, me.Hand.Size(), hand+1)
		}
		pick := effectstest.DiscardChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatalf("five=%v: no discard was asked", five)
		}
		effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
		effectstest.PassPriorityAroundTable(t, g)
		_, flipped := oldB322Flipped(g, me, oldB322Jace)
		if flipped != five {
			t.Errorf("five=%v: flipped = %v (the graveyard has %d cards after the discard)", five, flipped, len(me.Graveyard.Cards))
		}
	}
}

func TestOldB322JaceTelepathUnboundPlusOneShrinksUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Jace, Telepath Unbound", oldB322Jace+"#1", 5)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 4, 4)
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, foe); p != 2 {
		t.Errorf("the target has %d power, want 2", p)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, foe); p != 2 {
		t.Errorf("it wore off on the opponent's turn: %d power", p)
	}
}

func TestOldB322JaceTelepathUnboundMinusThreeLetsYouCastAnInstantFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Jace, Telepath Unbound", oldB322Jace+"#1", 5)
	bolt := effectstest.PushCatalogGraveyardCard(me, "Lightning Bolt", "Instant", oldB322Bolt, 0, 0)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}})
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
	if effectstest.ZoneOf(g, bolt) != game.ZoneExile {
		t.Errorf("the spell went to %v, want exile", effectstest.ZoneOf(g, bolt))
	}
}

func TestOldB322JaceTelepathUnboundEmblemMillsAnOpponentOnEachCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Jace, Telepath Unbound", oldB322Jace+"#1", 9)
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("%d emblems, want 1", me.Emblems.Size())
	}
	before := opp.Library.Size()
	effectstest.CastFromHandForTest(t, g, me, "Opt", "Instant", "{U}", "", game.CastSpellParams{})
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 5 {
		t.Errorf("the opponent milled %d, want 5", got)
	}
}

func TestOldB322ChandraFrontPingsAndFlipsAtThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	chandra := effectstest.ImportToBattlefield(t, g, oldB322ChandraRow(), me)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, chandra).SummonedThisTurn = false })
	life := opp.Life
	for i := 0; i < 3; i++ {
		effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
		effectstest.PassPriorityAroundTable(t, g)
		_, flipped := oldB322Flipped(g, me, oldB322Chandra)
		if flipped != (i == 2) {
			t.Fatalf("after ping %d: flipped = %v", i+1, flipped)
		}
		if i < 2 {
			g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, chandra).Tapped = false })
		}
	}
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
}

func TestOldB322ChandraFrontUntapsWhenYouCastARedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	chandra := effectstest.ImportToBattlefield(t, g, oldB322ChandraRow(), me)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, chandra)
		c.SummonedThisTurn = false
		c.Tapped = true
	})
	effectstest.CastFromHandForTest(t, g, me, "Lightning Bolt", "Instant", "{R}", oldB322Bolt,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, chandra) {
		t.Error("Chandra stayed tapped after a red spell")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, chandra).Tapped = true })
	effectstest.CastFromHandForTest(t, g, me, "Giant Growth", "Instant", "{G}", oldB322GiantGrowth,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: chandra}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, chandra) {
		t.Error("Chandra untapped for a green spell")
	}
}

func TestOldB322ChandraRoaringFlamePlusOneHitsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Roaring Flame", oldB322Chandra+"#1", 4)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 2, 2)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the +1 left the opponent at %d, want %d", opp.Life, life-2)
	}
	if effectstest.ZoneOf(g, foe) != game.ZoneBattlefield {
		t.Error("the +1 hurt a creature")
	}
}

func TestOldB322ChandraRoaringFlameMinusTwoKillsATwoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Roaring Flame", oldB322Chandra+"#1", 4)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB322Bear, 2, 2)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, foe) == game.ZoneBattlefield {
		t.Error("the −2 did not kill a 2/2")
	}
}

func TestOldB322ChandraRoaringFlameUltimateGivesEachDamagedOpponentAnEmblem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	walker := oldB322Walker(t, g, oldB322ChandraRow(), me, 7)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		lives[p.ID] = p.Life
	}
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != lives[p.ID]-6 {
			t.Errorf("%s is at %d, want %d", p.Name, p.Life, lives[p.ID]-6)
		}
		if p.Emblems.Size() != 1 {
			t.Errorf("%s has %d emblems, want 1", p.Name, p.Emblems.Size())
		}
	}
	if me.Emblems.Size() != 0 {
		t.Error("Chandra's controller got an emblem")
	}
	opp := g.Seats[1]
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lives[opp.ID]-6-3 {
		t.Errorf("after its upkeep the opponent is at %d, want %d", opp.Life, lives[opp.ID]-9)
	}
}
