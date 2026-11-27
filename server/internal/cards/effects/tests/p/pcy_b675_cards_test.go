package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pcy_b675_cards_test.go — Prophecy, S82 OLD batch 675 (#5617): Bog Glider,
// Infernal Genesis, Elephant Resurgence, Heightened Awareness, Dual Nature,
// Copper-Leaf Angel, Rethink, Forgotten Harvest, Celestial Convergence and
// Wintermoon Mesa, each through the real cast, activation and trigger paths.
// (Withdraw was already on develop; Rhystic Cave is skipped, see the commit.)

const (
	pcy675Glider      = "6bff9d3a-a98d-4060-8409-4b81ce5aab15"
	pcy675Genesis     = "c3ce3473-e3bc-4b7c-ae57-957e6c40dc7a"
	pcy675Elephant    = "e2d54c59-7722-4520-adca-cd02a0871ab2"
	pcy675Awareness   = "a69a0720-35fb-450d-8d48-8e51a14ffa49"
	pcy675DualNature  = "7a8fcf6f-3bda-485c-b960-62a37112f0eb"
	pcy675Angel       = "efab5a30-74ee-4959-b7cd-2a5d25d18327"
	pcy675Rethink     = "ede61683-fa78-4103-8248-e0e64162ed49"
	pcy675Harvest     = "9e5de2ed-e8a1-4d52-832f-c4a1bfdf2ae5"
	pcy675Convergence = "92f0677f-64d8-4c83-abad-ffa2a37f88c1"
	pcy675Mesa        = "a4a6f95e-856c-4eb5-82ba-b2406be22b23"
)

func TestPcy675EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		pcy675Glider: "Bog Glider", pcy675Genesis: "Infernal Genesis", pcy675Elephant: "Elephant Resurgence",
		pcy675Awareness: "Heightened Awareness", pcy675DualNature: "Dual Nature", pcy675Angel: "Copper-Leaf Angel",
		pcy675Rethink: "Rethink", pcy675Harvest: "Forgotten Harvest", pcy675Convergence: "Celestial Convergence",
		pcy675Mesa: "Wintermoon Mesa",
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

func pcy675Lands(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for range n {
		ids = append(ids, effectstest.SeedLand(g, owner, "Swamp", "Basic Land — Swamp", ""))
	}
	return ids
}

func pcy675Tokens(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name && c.Controller == controller {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

func TestPcy675FlyingKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Bog Glider", pcy675Glider, "Creature — Human Mercenary"},
		{"Copper-Leaf Angel", pcy675Angel, "Artifact Creature — Angel"},
	} {
		id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, false)
		if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
			t.Errorf("%s does not have flying", tc.name)
		}
	}
}

// --- Bog Glider ---

func TestPcy675BogGliderSacrificesALandToFetchAMercenaryWithManaValueTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	glider := effectstest.B12Push(g, me.ID, "Bog Glider", "Creature — Human Mercenary", pcy675Glider, 1, 1)
	land := pcy675Lands(g, me.ID, 1)[0]
	cheap, dear, other := uuid.New(), uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: cheap, Name: "Cheap Mercenary", TypeLine: "Creature — Human Mercenary", ManaCost: "{1}{B}", Power: 1, Toughness: 1},
		{InstanceID: dear, Name: "Dear Mercenary", TypeLine: "Creature — Human Mercenary", ManaCost: "{2}{B}", Power: 2, Toughness: 2},
		{InstanceID: other, Name: "Cheap Soldier", TypeLine: "Creature — Human Soldier", ManaCost: "{B}", Power: 1, Toughness: 1},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	if err := g.ActivateCatalogAbility(me.ID, glider, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no land to sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, glider, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) || !effectstest.TappedOnBattlefield(t, g, glider) {
		t.Error("the land was not sacrificed, or the Glider not tapped, as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.SearchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not asked to search")
	}
	if !effectstest.SameIDs(pick.SearchCards, []uuid.UUID{cheap}) {
		t.Fatalf("offered %v, want only the cheap Mercenary", pick.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the Mercenary was not put onto the battlefield")
	}
}

// --- Copper-Leaf Angel ---

func TestPcy675CopperLeafAngelSacrificesXLandsForXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	angel := effectstest.B12Push(g, me.ID, "Copper-Leaf Angel", "Artifact Creature — Angel", pcy675Angel, 2, 2)
	lands := pcy675Lands(g, me.ID, 3)
	if err := g.ActivateCatalogAbility(me.ID, angel, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, SacrificeIDs: lands[:1]}); err == nil {
		t.Fatal("X = 2 was paid with one land")
	}
	if err := g.ActivateCatalogAbility(me.ID, angel, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, SacrificeIDs: lands[:2]}); err != nil {
		t.Fatalf("activate with X = 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, angel); p != 4 || tough != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", p, tough)
	}
	if g.Battlefield.Contains(lands[0]) || g.Battlefield.Contains(lands[1]) || !g.Battlefield.Contains(lands[2]) {
		t.Error("exactly the two named lands should be sacrificed")
	}
}

// --- Rethink ---

func TestPcy675RethinkCountersUnlessItsControllerPaysItsManaValue(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		spell := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
		effectstest.B06AddMana(me, "G", "G", "G")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast the target spell: %v", err)
		}
		effectstest.CastCatalogSpell(t, g, "Rethink", "Instant", pcy675Rethink, effectstest.CardTarget(spell))
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.FillPool(me, 3)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if g.Stack.Contains(spell) != pay {
			t.Errorf("paid=%v: the spell is still on the stack = %v", pay, g.Stack.Contains(spell))
		}
		if pay && me.ManaPool.Total() != 0 {
			t.Errorf("%d mana left after paying {3}", me.ManaPool.Total())
		}
	}
}

// --- Heightened Awareness ---

func TestPcy675HeightenedAwarenessDiscardsYourHandAsItEntersAndDrawsAnExtraCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	for range 3 {
		effectstest.PushHandCardOfType(me, "Card", "Instant")
	}
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Heightened Awareness", TypeLine: "Enchantment", OracleID: pcy675Awareness}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 || me.Graveyard.Size() < 3 {
		t.Fatalf("hand %d, graveyard %d: the hand was not discarded", me.Hand.Size(), me.Graveyard.Size())
	}
	pcy675BeginDrawStep(t, g, opp.ID)
	if me.Hand.Size() != 0 {
		t.Error("an opponent's draw step drew you a card")
	}
	pcy675BeginDrawStep(t, g, me.ID)
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d after your draw step trigger, want 1", me.Hand.Size())
	}
}

// --- Celestial Convergence ---

func TestPcy675CelestialConvergenceCountsDownAndTheHighestLifeWins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.TdCastFromHand(t, g, game.Card{Name: "Celestial Convergence", TypeLine: "Enchantment", OracleID: pcy675Convergence}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	omens := func() int {
		c, _ := g.LookupCardForEffect(id)
		return c.Counters["omen"]
	}
	if got := omens(); got != 7 {
		t.Fatalf("it entered with %d omen counters, want 7", got)
	}
	effectstest.BeginUpkeep(t, g, g.Seats[1].ID)
	if got := omens(); got != 7 {
		t.Errorf("an opponent's upkeep changed the counters to %d", got)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	if got := omens(); got != 6 {
		t.Errorf("one upkeep left %d omen counters, want 6", got)
	}
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: id, Kind: "omen", N: -5}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: id}))
	})
	for i, p := range g.Seats {
		p.Life = 20 - i
	}
	me.Life = 20
	g.Seats[1].Life = 25
	effectstest.BeginUpkeep(t, g, me.ID)
	if g.Outcome == nil || g.Outcome.Winner != g.Seats[1].ID {
		t.Errorf("outcome %+v, want the player with 25 life to win", g.Outcome)
	}
}

func TestPcy675CelestialConvergenceTiedForHighestIsADraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.TdCastFromHand(t, g, game.Card{Name: "Celestial Convergence", TypeLine: "Enchantment", OracleID: pcy675Convergence}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: id, Kind: "omen", N: -6}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: id}))
	})
	for _, p := range g.Seats {
		p.Life = 10
	}
	g.Seats[1].Life, g.Seats[2].Life = 30, 30
	effectstest.BeginUpkeep(t, g, me.ID)
	if g.Outcome == nil || g.Outcome.Kind != game.OutcomeDraw {
		t.Errorf("outcome %+v, want a draw", g.Outcome)
	}
}

// --- Forgotten Harvest ---

func TestPcy675ForgottenHarvestExilesALandCardForAPlusOnePlusOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Forgotten Harvest", "Enchantment", pcy675Harvest, false)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	land := effectstest.GraveyardCardOf(g, me, "Forest", "Basic Land — Forest", "", "")
	nonland := effectstest.GraveyardCardOf(g, me, "Spell", "Sorcery", "{G}", "")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) || !me.Graveyard.Contains(nonland) {
		t.Error("exactly the land card should be exiled")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
}

func TestPcy675ForgottenHarvestDeclinedPutsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Forgotten Harvest", "Enchantment", pcy675Harvest, false)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	land := effectstest.GraveyardCardOf(g, me, "Forest", "Basic Land — Forest", "", "")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(land) {
		t.Error("a declined land card was exiled")
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the target is %d/x after declining, want 2", p)
	}
}

// --- Wintermoon Mesa ---

func TestPcy675WintermoonMesaEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Wintermoon Mesa", pcy675Mesa)
	if !effectstest.IsTapped(g, id) {
		t.Error("Wintermoon Mesa entered untapped")
	}
}

func pcy675BeginDrawStep(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginDrawStep, Actor: who}) })
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Infernal Genesis ---

func TestPcy675InfernalGenesisMillsAndMakesMinionsEqualToTheManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Infernal Genesis", "Enchantment", pcy675Genesis, false)
	for _, p := range []*game.Player{me, opp} {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", ManaCost: "{2}{B}", Owner: p.ID, Controller: p.ID})
	}
	gy := me.Graveyard.Size()
	effectstest.BeginUpkeep(t, g, me.ID)
	if got := len(pcy675Tokens(g, me.ID, "Minion")); got != 3 {
		t.Errorf("my upkeep made %d Minions, want 3", got)
	}
	if me.Graveyard.Size() != gy+1 {
		t.Errorf("graveyard %d → %d, want one card milled", gy, me.Graveyard.Size())
	}
	if len(pcy675Tokens(g, opp.ID, "Minion")) != 0 {
		t.Error("the opponent got Minions on my upkeep")
	}
	effectstest.BeginUpkeep(t, g, opp.ID)
	if got := len(pcy675Tokens(g, opp.ID, "Minion")); got != 3 {
		t.Errorf("an opponent's upkeep made them %d Minions, want 3", got)
	}
}

func TestPcy675InfernalGenesisMakesNothingForALandOrAnEmptyLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Infernal Genesis", "Enchantment", pcy675Genesis, false)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	opp.Library.Cards = nil
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.BeginUpkeep(t, g, opp.ID)
	if len(pcy675Tokens(g, me.ID, "Minion"))+len(pcy675Tokens(g, opp.ID, "Minion")) != 0 {
		t.Error("Minions were made from a land card or an empty library")
	}
}

// --- Elephant Resurgence ---

func TestPcy675ElephantResurgenceMakesEachPlayerAnElephantSizedByTheirOwnGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	for range 2 {
		effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	}
	effectstest.GraveyardCardOf(g, opp, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.GraveyardCardOf(g, opp, "Spell", "Sorcery", "{G}", "")
	effectstest.CastCatalogSpell(t, g, "Elephant Resurgence", "Sorcery", pcy675Elephant, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mine, theirs := pcy675Tokens(g, me.ID, "Elephant"), pcy675Tokens(g, opp.ID, "Elephant")
	if len(mine) != 1 || len(theirs) != 1 {
		t.Fatalf("Elephants: mine %d, theirs %d, want one each", len(mine), len(theirs))
	}
	if p, tough := effectstest.PtOf(t, g, mine[0]); p != 2 || tough != 2 {
		t.Errorf("my Elephant is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs[0]); p != 1 || tough != 1 {
		t.Errorf("their Elephant is %d/%d, want 1/1", p, tough)
	}
	if len(pcy675Tokens(g, third.ID, "Elephant")) != 0 {
		t.Error("a third player with no creature cards kept a 0/0 Elephant")
	}
	// The size is live: another creature card in my graveyard grows it.
	pcy675Destroy(t, g, effectstest.B12Creature(g, me.ID, "Fallen", "Creature — Bear", 1, 1))
	if p, _ := effectstest.PtOf(t, g, mine[0]); p != 3 {
		t.Errorf("my Elephant is %d/x after a third creature card, want 3", p)
	}
}

// --- Dual Nature ---

func pcy675Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (effects.DestroyTarget{Target: id}).Apply(effects.NewContext(g, &game.StackItem{Controller: g.Seats[0].ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestPcy675DualNatureCopiesNontokenCreaturesAndExilesTheTokensWhenTheOriginalLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Dual Nature", "Enchantment", pcy675DualNature, false)
	bear := effectstest.TdCastFromHand(t, g, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	copies := pcy675Tokens(g, me.ID, "Bear")
	if len(copies) != 1 {
		t.Fatalf("%d token copies, want one (the token must not copy itself)", len(copies))
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the original left the battlefield")
	}
	pcy675Destroy(t, g, bear)
	if got := len(pcy675Tokens(g, me.ID, "Bear")); got != 0 {
		t.Errorf("%d Bear tokens remain after the original left, want none", got)
	}
}

func TestPcy675DualNatureLeavingExilesTheTokensItMade(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dual := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Dual Nature", "Enchantment", pcy675DualNature, false)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(pcy675Tokens(g, me.ID, "Bear")) != 1 {
		t.Fatal("Dual Nature made no token")
	}
	pcy675Destroy(t, g, dual)
	if got := len(pcy675Tokens(g, me.ID, "Bear")); got != 0 {
		t.Errorf("%d token copies remain after Dual Nature left, want none", got)
	}
}

func TestPcy675DualNatureIgnoresTokensEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Dual Nature", "Enchantment", pcy675DualNature, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := (effects.CreateToken{Controller: me.ID, Template: effects.TokenCard("1/1 white Soldier"), N: 1}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(pcy675Tokens(g, me.ID, "Soldier")); got != 1 {
		t.Errorf("%d Soldier tokens, want 1: a token entering must not be copied", got)
	}
}
