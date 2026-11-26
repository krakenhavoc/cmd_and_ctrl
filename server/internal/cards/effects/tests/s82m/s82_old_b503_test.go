package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b503_test.go — Alara Reborn, older-sets batch 503 (#5586). Every test
// plays the real catalog entry through the engine.

const (
	oldB503Predatory = "24547885-2e2c-4118-9d5e-4058471f289b"
	oldB503Bant      = "67ceffa4-2fdb-499c-88cd-49fb5eb9be59"
	oldB503Broodmom  = "a9a6c2ba-30ba-4e45-9ea0-ea9833342069"
	oldB503Bloom     = "aacec73f-9c60-4c3f-b323-6261eab342af"
	oldB503Reaver    = "687abe1c-b2fc-49e0-bd36-3db3dc3a5039"
	oldB503Naya      = "703215df-dee2-4b82-9f32-256e1410543c"
	oldB503Godtrack  = "9be3e610-f6bf-4e79-87ca-79dde27945d0"
	oldB503Griffin   = "dd4b9e5a-b3ab-4b70-b200-228c8add9f52"
	oldB503Firewild  = "312a5a75-9b68-4e4b-97fe-3ab644e3e493"
	oldB503Veinfire  = "3f65d1f8-3e46-4e3d-b087-5496a4106e4e"
	oldB503Wildfield = "b8141489-00fa-4aba-85d0-4c883196fece"
)

func TestOldB503EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB503Predatory: "Predatory Advantage", oldB503Bant: "Bant Sojourners", oldB503Broodmom: "Dragon Broodmother",
		oldB503Bloom: "Morbid Bloom", oldB503Reaver: "Lightning Reaver", oldB503Naya: "Naya Sojourners",
		oldB503Godtrack: "Godtracker of Jund", oldB503Griffin: "Retaliator Griffin", oldB503Firewild: "Firewild Borderpost",
		oldB503Veinfire: "Veinfire Borderpost", oldB503Wildfield: "Wildfield Borderpost",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Predatory Advantage ------------------------------------------------------

func oldB503OpponentsEnd(t *testing.T, creatureSpell, noncreatureSpell bool) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Predatory Advantage", "Enchantment", oldB503Predatory, false)
	effectstest.AdvanceToMainOf(t, g, 1)
	tally := game.CastTally{}
	if creatureSpell {
		tally.Total++
	}
	if noncreatureSpell {
		tally.Total++
		tally.Noncreature++
	}
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
		}
		g.SpellsCastThisTurn[opp.ID] = tally
	})
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestOldB503PredatoryAdvantageMakesALizardWhenTheOpponentCastNoCreatureSpell(t *testing.T) {
	g, me := oldB503OpponentsEnd(t, false, true) // a noncreature spell does not count
	if n := effectstest.TokensNamedControlledBy(g, "Lizard", me.ID); n != 1 {
		t.Errorf("%d Lizard tokens under my control, want 1", n)
	}
}

func TestOldB503PredatoryAdvantageMakesNothingAfterACreatureSpell(t *testing.T) {
	g, me := oldB503OpponentsEnd(t, true, false)
	if n := effectstest.TokensNamedControlledBy(g, "Lizard", me.ID); n != 0 {
		t.Errorf("%d Lizard tokens, want 0 after the opponent cast a creature spell", n)
	}
}

func TestOldB503PredatoryAdvantageIgnoresMyOwnEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Predatory Advantage", "Enchantment", oldB503Predatory, false)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Lizard", me.ID); n != 0 {
		t.Errorf("%d Lizard tokens at my own end step, want 0", n)
	}
}

// --- Bant Sojourners ----------------------------------------------------------

func TestOldB503BantCycledMayCreateASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CycleFromHand(t, g, "Bant Sojourners", "Creature — Human Soldier", oldB503Bant, "{C}{C}{W}")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Soldier", me.ID); n != 1 {
		t.Errorf("%d Soldier tokens, want 1", n)
	}
}

func TestOldB503BantDiesMayCreateASoldierOrDeclines(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bant := effectstest.PushCatalogPermanent(g, me.ID, "Bant Sojourners", "Creature — Human Soldier", oldB503Bant, false)
		effectstest.KillCreature(t, g, me.ID, bant)
		effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, yes)
		want := 0
		if yes {
			want = 1
		}
		if n := effectstest.TokensNamedControlledBy(g, "Soldier", me.ID); n != want {
			t.Errorf("yes=%v: %d Soldier tokens, want %d", yes, n, want)
		}
	}
}

// --- Naya Sojourners ----------------------------------------------------------

func TestOldB503NayaCycledAndDiesPutACounterOnATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CycleFromHand(t, g, "Naya Sojourners", "Creature — Elf Shaman", oldB503Naya, "{C}{C}{G}")
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("cycled: %d counters, want 1", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	bear2 := effectstest.B12Creature(g2, me2.ID, "My Bear", "Creature — Bear", 2, 2)
	naya := effectstest.PushCatalogPermanent(g2, me2.ID, "Naya Sojourners", "Creature — Elf Shaman", oldB503Naya, false)
	effectstest.KillCreature(t, g2, me2.ID, naya)
	effectstest.AssistCardSettle(t, g2, me2.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear2}, true)
	if got := effectstest.PlusOnes(g2, bear2); got != 1 {
		t.Errorf("dies: %d counters, want 1 (and only one trigger)", got)
	}
}

// --- Godtracker of Jund -------------------------------------------------------

func oldB503BigToken(g *game.Game, owner uuid.UUID, power int) {
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(owner, game.Card{
			Name: "Big", TypeLine: "Token Creature — Giant", Power: power, Toughness: power,
		}, 1)
	})
}

func TestOldB503GodtrackerGrowsForAPowerFiveCreatureOnly(t *testing.T) {
	for _, tc := range []struct{ power, want int }{{5, 1}, {4, 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		god := effectstest.B12Push(g, me.ID, "Godtracker of Jund", "Creature — Elf Shaman", oldB503Godtrack, 2, 2)
		oldB503BigToken(g, me.ID, tc.power)
		effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
		if got := effectstest.PlusOnes(g, god); got != tc.want {
			t.Errorf("power %d: %d counters, want %d", tc.power, got, tc.want)
		}
	}
}

func TestOldB503GodtrackerIgnoresAnOpponentsBigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	god := effectstest.B12Push(g, me.ID, "Godtracker of Jund", "Creature — Elf Shaman", oldB503Godtrack, 2, 2)
	oldB503BigToken(g, opp.ID, 6)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := effectstest.PlusOnes(g, god); got != 0 {
		t.Errorf("%d counters for an opponent's creature, want 0", got)
	}
}

// --- Retaliator Griffin -------------------------------------------------------

func TestOldB503GriffinTakesCountersEqualToDamageFromAnOpponentsSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	griffin := effectstest.B12Push(g, me.ID, "Retaliator Griffin", "Creature — Griffin", oldB503Griffin, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7Hit(t, g, theirs, me.ID, 3)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := effectstest.PlusOnes(g, griffin); got != 3 {
		t.Errorf("%d counters, want 3", got)
	}
}

func TestOldB503GriffinIgnoresMyOwnSourcesAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	griffin := effectstest.B12Push(g, me.ID, "Retaliator Griffin", "Creature — Griffin", oldB503Griffin, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7Hit(t, g, mine, me.ID, 2)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := effectstest.PlusOnes(g, griffin); got != 0 {
		t.Errorf("my own source gave %d counters, want 0", got)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7Hit(t, g, theirs, me.ID, 2)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	if got := effectstest.PlusOnes(g, griffin); got != 0 {
		t.Errorf("declined, yet %d counters", got)
	}
}

// --- Lightning Reaver ---------------------------------------------------------

func TestOldB503ReaverGrowsChargeCountersAndBurnsEachOpponentAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	reaver := effectstest.B12Push(g, me.ID, "Lightning Reaver", "Creature — Zombie Beast", oldB503Reaver, 3, 3)
	effectstest.DealCombatDamageToPlayer(g, reaver, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, reaver, game.CounterCharge); got != 1 {
		t.Fatalf("%d charge counters after combat damage, want 1", got)
	}
	effectstest.AddCountersForTest(g, reaver, game.CounterCharge, 1)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 2
		if p.ID == me.ID {
			want = lives[i]
		}
		if p.Life != want {
			t.Errorf("%s: life %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestOldB503ReaverWithNoChargeCountersDoesNothingAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Lightning Reaver", "Creature — Zombie Beast", oldB503Reaver, 3, 3)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != lives[i] {
			t.Errorf("%s lost life with no charge counters", p.Name)
		}
	}
}

// --- Morbid Bloom -------------------------------------------------------------

func TestOldB503BloomExilesTheCreatureCardAndMakesSaprolingsEqualToItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Tough Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	effectstest.AssistCardCastX(t, g, "Morbid Bloom", "Sorcery", oldB503Bloom, "{4}{B}{G}",
		[]string{"B", "G", "C", "C", "C", "C"}, 0, effectstest.CardRefs(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, dead) {
		t.Error("the creature card was not exiled")
	}
	if n := effectstest.TokensNamedControlledBy(g, "Saproling", me.ID); n != 4 {
		t.Errorf("%d Saprolings, want 4 (the exiled card's toughness)", n)
	}
}

func TestOldB503BloomRefusesANoncreatureCardInAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	_, err := effectstest.AssistCardCastXErr(t, g, "Morbid Bloom", "Sorcery", oldB503Bloom, "{4}{B}{G}",
		[]string{"B", "G", "C", "C", "C", "C"}, 0, effectstest.CardRefs(spell))
	if err == nil {
		t.Error("a sorcery card was accepted as the target")
	}
}

// --- Dragon Broodmother -------------------------------------------------------

func TestOldB503BroodmotherMakesADragonEachUpkeepAndDevoursOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Dragon Broodmother", "Creature — Dragon", oldB503Broodmom, false)
	fodder := effectstest.PushDevourFodder(g, me.ID, "Fodder A")
	effectstest.AdvanceToUpkeepOf(t, g, 1) // each upkeep: an opponent's counts
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID) == nil {
		t.Fatal("the Dragon token did not ask what to devour")
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder)
	var dragon *game.Card
	for i := range g.Battlefield.Cards {
		if c := g.Battlefield.Cards[i]; c.Name == "Dragon" && c.IsToken() {
			dragon = &g.Battlefield.Cards[i]
		}
	}
	if dragon == nil {
		t.Fatal("no Dragon token on the battlefield")
	}
	if dragon.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("the Dragon has %d +1/+1 counters, want 2 (devoured one creature)", dragon.Counters[game.CounterPlusOne])
	}
	if !effectstest.P1g2HasKeyword(g, dragon.InstanceID, "flying") {
		t.Error("the Dragon token has no flying")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the devoured creature is still on the battlefield")
	}
}

// --- Borderposts --------------------------------------------------------------

func TestOldB503BorderpostsEnterTappedAndTapForEitherColour(t *testing.T) {
	for _, tc := range []struct{ oracle, name, a, b string }{
		{oldB503Firewild, "Firewild Borderpost", "R", "G"},
		{oldB503Veinfire, "Veinfire Borderpost", "B", "R"},
		{oldB503Wildfield, "Wildfield Borderpost", "G", "W"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		id := effectstest.CastFromHandForTest(t, g, me, tc.name, "Artifact", "{1}{"+tc.a+"}{"+tc.b+"}", tc.oracle, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Fatalf("%s did not resolve", tc.name)
		}
		if !effectstest.TappedOnBattlefield(t, g, id) {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		effectstest.UntapForTest(g, id)
		effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{Colors: []string{tc.b}})
		if me.ManaPool.Total() != 1 {
			t.Errorf("%s: pool %d, want 1 mana", tc.name, me.ManaPool.Total())
		}
	}
}

func TestOldB503BorderpostAlternativeCostPaysOneAndReturnsABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	post := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: post, Name: "Firewild Borderpost", TypeLine: "Artifact", ManaCost: "{1}{R}{G}", OracleID: oldB503Firewild, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C")
	if err := g.CastSpell(me.ID, post, game.CastSpellParams{Strict: true, AlternativeCost: "borderpost", AltCostIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("alternative cast: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("the basic land did not return to its owner's hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(post) {
		t.Error("the Borderpost did not resolve")
	}
}

func TestOldB503BorderpostAlternativeCostRefusesANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, me.ID, "Nonbasic Land", "Land")
	post := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: post, Name: "Firewild Borderpost", TypeLine: "Artifact", ManaCost: "{1}{R}{G}", OracleID: oldB503Firewild, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C")
	if err := g.CastSpell(me.ID, post, game.CastSpellParams{Strict: true, AlternativeCost: "borderpost", AltCostIDs: []uuid.UUID{land}}); err == nil {
		t.Error("a nonbasic land was accepted for the alternative cost")
	}
}
