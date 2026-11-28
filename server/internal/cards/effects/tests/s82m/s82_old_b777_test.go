package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b777_test.go — Weatherlight, batch 777 (#5629): enters and dies
// triggers, card draw and library. Each test plays the real catalog entry.

const (
	oldB777Alabaster = "2392a41a-59d3-4749-be94-4d9df0af9c4c"
	oldB777Benefact  = "78c1d6fe-38e4-40c4-9252-87ff272247ad"
	oldB777Fallow    = "7b414f70-a9b1-42c3-9655-6aa520d3b19b"
	oldB777Griffin   = "7fe9d30f-c778-4769-a3ae-3cb48cd23e07"
	oldB777Elephant  = "bed7ac55-fe40-46d0-bc22-1c8d11f41459"
	oldB777Sentinel  = "272ab5d5-7edb-4457-8a09-cabe2325c881"
	oldB777Abeyance  = "6300de53-6e71-4f0e-87e1-b0acd25c59a8"
	oldB777Tome      = "78d2a3fd-864c-49f3-b531-c71dc3b35977"
	oldB777Flux      = "b8cff5f1-6871-49d2-89a7-24870308aadb"
	oldB777Blood     = "cd55a003-da0d-4522-b08a-d0731b1050c6"

	oldB777BearType = "Creature — Bear"
)

func TestOldB777EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB777Alabaster: "Alabaster Dragon", oldB777Benefact: "Noble Benefactor",
		oldB777Fallow: "Fallow Wurm", oldB777Griffin: "Mistmoon Griffin",
		oldB777Elephant: "Rogue Elephant", oldB777Sentinel: "Llanowar Sentinel",
		oldB777Abeyance: "Abeyance", oldB777Tome: "Thran Tome",
		oldB777Flux: "Flux", oldB777Blood: "Boiling Blood",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB777InZone(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func oldB777Die(t *testing.T, g *game.Game, owner *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
}

func TestOldB777AlabasterDragonFliesAndShufflesItselfAwayWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	dragon := effectstest.B12Push(g, me.ID, "Alabaster Dragon", "Creature — Dragon", oldB777Alabaster, 4, 4)
	effectstest.AssertKeywords(t, g, dragon, "flying")
	oldB777Die(t, g, me, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB777InZone(me.Library, dragon) || oldB777InZone(me.Graveyard, dragon) {
		t.Errorf("the Dragon did not shuffle into its owner's library: library %v, graveyard %v",
			oldB777InZone(me.Library, dragon), oldB777InZone(me.Graveyard, dragon))
	}
}

func TestOldB777NobleBenefactorLetsEachPlayerSearchForAnyCardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	want := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		ids := effectstest.SeedSearchLibrary(p, game.Card{Name: "Filler", TypeLine: "Sorcery"}, game.Card{Name: "Prize", TypeLine: "Artifact"})
		want[p.ID] = ids[1]
	}
	me := g.Seats[0]
	benefactor := effectstest.B12Push(g, me.ID, "Noble Benefactor", "Creature — Human Cleric", oldB777Benefact, 2, 2)
	oldB777Die(t, g, me, benefactor)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if i == 1 {
			effectstest.AnswerSearchFailToFind(t, g, p.ID)
			continue
		}
		effectstest.AnswerSearchByID(t, g, p.ID, want[p.ID])
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		inHand := p.Hand.Contains(want[p.ID])
		if i == 1 && inHand {
			t.Error("seat 1 declined but the card reached its hand")
		}
		if i != 1 && !inHand {
			t.Errorf("seat %d: the searched card is not in hand", i)
		}
	}
}

func oldB777CastFallow(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Fallow Wurm", "Creature — Wurm", oldB777Fallow, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB777FallowWurmStaysWhenYouDiscardALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell, land := uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	id := oldB777CastFallow(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Wurm was sacrificed although a land card was discarded")
	}
	if !oldB777InZone(me.Graveyard, land) || oldB777InZone(me.Graveyard, spell) {
		t.Error("the land card was not the card discarded")
	}
}

func TestOldB777FallowWurmIsSacrificedWithoutALandCardInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	id := oldB777CastFallow(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Wurm stayed with no land card in hand")
	}
	if oldB777InZone(me.Graveyard, spell) {
		t.Error("a nonland card was discarded")
	}
}

func TestOldB777FallowWurmDeclinedDiscardSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	id := oldB777CastFallow(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Wurm stayed although no land card was discarded")
	}
	if oldB777InZone(me.Graveyard, land) {
		t.Error("the land card was discarded anyway")
	}
}

func TestOldB777MistmoonGriffinFliesExilesItselfAndReturnsTheTopCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	older := effectstest.PushGraveyardCardTyped(me, "Older Bear", oldB777BearType)
	top := effectstest.PushGraveyardCardTyped(me, "Top Bear", oldB777BearType)
	effectstest.PushGraveyardCardTyped(me, "Spell on top", "Sorcery")
	griffin := effectstest.B12Push(g, me.ID, "Mistmoon Griffin", "Creature — Griffin", oldB777Griffin, 2, 2)
	effectstest.AssertKeywords(t, g, griffin, "flying")
	oldB777Die(t, g, me, griffin)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, griffin) || oldB777InZone(me.Graveyard, griffin) {
		t.Error("the Griffin did not exile itself")
	}
	if !effectstest.OnBattlefieldByID(g, top) {
		t.Error("the top creature card of the graveyard did not return")
	}
	if effectstest.OnBattlefieldByID(g, older) || !oldB777InZone(me.Graveyard, older) {
		t.Error("a creature card other than the top one returned")
	}
}

func TestOldB777MistmoonGriffinWithNoOtherCreatureCardReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
	griffin := effectstest.B12Push(g, me.ID, "Mistmoon Griffin", "Creature — Griffin", oldB777Griffin, 2, 2)
	oldB777Die(t, g, me, griffin)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, griffin) || !effectstest.InExile(g, griffin) {
		t.Error("the Griffin should be exiled and gone")
	}
}

func oldB777CastElephant(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Rogue Elephant", "Creature — Elephant", oldB777Elephant, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB777RogueElephantSacrificesAForestToStay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	plains := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	id := oldB777CastElephant(t, g)
	c := effectstest.Sweep05OptionKey(t, g, me.ID, "option-pick/s82-old-b777-sacrifice-this-unless-you-sacrifice-a-forest")
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("pick the Forest: %v", err)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, forest) || effectstest.HasID(pick.ChooseCards, plains) {
		t.Fatalf("the sacrifice must offer the Forest only: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !g.Battlefield.Contains(id) {
		t.Error("the Forest should be gone and the Elephant kept")
	}
}

func TestOldB777RogueElephantSacrificesItselfWithNoForestOrWhenDeclined(t *testing.T) {
	t.Run("no Forest", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
		id := oldB777CastElephant(t, g)
		if g.Battlefield.Contains(id) {
			t.Error("the Elephant stayed with no Forest to sacrifice")
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		forest := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		id := oldB777CastElephant(t, g)
		c := effectstest.Sweep05OptionKey(t, g, me.ID, "option-pick/s82-old-b777-sacrifice-this-unless-you-sacrifice-a-forest")
		if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
			t.Fatalf("pick the Elephant: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) || !g.Battlefield.Contains(forest) {
			t.Error("the Elephant should be gone and the Forest kept")
		}
	})
}

func TestOldB777LlanowarSentinelPaysToFetchAnotherOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Llanowar Sentinel", TypeLine: "Creature — Elf", OracleID: oldB777Sentinel, Power: 2, Toughness: 3},
		game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid"},
	)
	first := effectstest.CastCatalogSpell(t, g, "Llanowar Sentinel", "Creature — Elf", oldB777Sentinel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || effectstest.SearchOptionNamed(g, c, "Llanowar Elves") != uuid.Nil {
		t.Fatalf("the search must offer only cards named Llanowar Sentinel: %+v", c)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, first) || !effectstest.OnBattlefieldByID(g, ids[0]) {
		t.Error("both Sentinels should be on the battlefield")
	}
	if !me.Library.Contains(ids[1]) {
		t.Error("the other card should stay in the library")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want {1}{G} spent", me.ManaPool.Total())
	}
}

func TestOldB777LlanowarSentinelDecliningTheOfferSearchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Llanowar Sentinel", TypeLine: "Creature — Elf", OracleID: oldB777Sentinel, Power: 2, Toughness: 3})
	effectstest.CastCatalogSpell(t, g, "Llanowar Sentinel", "Creature — Elf", oldB777Sentinel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil || effectstest.OnBattlefieldByID(g, ids[0]) {
		t.Error("a search ran although the {1}{G} was not paid")
	}
}

func TestOldB777AbeyanceStopsTheTargetCastingInstantsAndSorceriesAndActivatingNonManaAbilitiesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	instant := effectstest.PushHandCardWithManaCost(opp, "Their Instant", "Instant", "{1}")
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	thirdInstant := effectstest.PushHandCardWithManaCost(third, "Third Instant", "Instant", "{1}")
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Abeyance", "Instant", oldB777Abeyance, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("the caster's hand went %d -> %d, want one card drawn", handBefore, got)
	}
	effectstest.FillPool(opp, 3)
	if err := g.CastSpell(opp.ID, instant, game.CastSpellParams{FromZone: "hand", Strict: true}); err == nil {
		t.Error("the target cast an instant")
	}
	effectstest.FillPool(third, 3)
	if err := g.CastSpell(third.ID, thirdInstant, game.CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Errorf("a third player could not cast an instant: %v", err)
	}
	if err := g.ActivateManaAbility(opp.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("a mana ability was refused: %v", err)
	}
	oppTome := effectstest.PushArtifactToBattlefieldForTest(g, opp.ID, "Thran Tome", "Artifact — Book", oldB777Tome)
	thirdTome := effectstest.PushArtifactToBattlefieldForTest(g, third.ID, "Thran Tome", "Artifact — Book", oldB777Tome)
	effectstest.FillPool(opp, 5)
	if err := g.ActivateCatalogAbility(opp.ID, oppTome, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("the target activated an ability that isn't a mana ability")
	}
	effectstest.FillPool(third, 5)
	if err := g.ActivateCatalogAbility(third.ID, thirdTome, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Errorf("a third player could not activate an ability: %v", err)
	}
}

func TestOldB777FluxEachPlayerDiscardsAnyNumberThenDrawsThatManyAndYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var hands = map[uuid.UUID][]uuid.UUID{}
	for _, p := range g.Seats {
		for i := 0; i < 3; i++ {
			hands[p.ID] = append(hands[p.ID], effectstest.PushHandCardWithManaCost(p, "Card", "Sorcery", "{1}"))
		}
		effectstest.SeedLibrary(p, "L1", "L2", "L3", "L4", "L5", "L6")
	}
	sizes := map[uuid.UUID]int{}
	effectstest.CastCatalogSpell(t, g, "Flux", "Sorcery", oldB777Flux, nil)
	for _, p := range g.Seats {
		sizes[p.ID] = p.Hand.Size()
	}
	effectstest.PassPriorityAroundTable(t, g)
	discards := map[uuid.UUID]int{}
	for i, p := range g.Seats {
		n := i % 3 // 0, 1, 2, 0 cards
		discards[p.ID] = n
		if n == 0 {
			effectstest.AnswerDiscard(t, g, p.ID)
			continue
		}
		effectstest.AnswerDiscard(t, g, p.ID, hands[p.ID][:n]...)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := sizes[p.ID] // discards n, draws n
		if p.ID == me.ID {
			want++ // the caster's own "Draw a card"
		}
		if p.Hand.Size() != want {
			t.Errorf("seat %s: hand %d, want %d", p.ID, p.Hand.Size(), want)
		}
		wantGrave := discards[p.ID]
		if p.ID == me.ID {
			wantGrave++ // Flux itself
		}
		if got := len(p.Graveyard.Cards); got != wantGrave {
			t.Errorf("seat %s: %d cards in the graveyard, want %d", p.ID, got, wantGrave)
		}
	}
}

func TestOldB777BoilingBloodForcesTheTargetToAttackAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB777BearType, 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Boiling Blood", "Instant", oldB777Blood, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasRequirementRecord(g) {
		t.Error("the target has no must-attack requirement")
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want one card drawn", handBefore, got)
	}
}

func TestOldB777ThranTomeRevealsThreeOpponentBinsOneAndYouDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedLibrary(me, "Deep1", "Deep2", "Deep3", "R1", "R2", "R3")
	tome := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Thran Tome", "Artifact — Book", oldB777Tome)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	effectstest.Pr7Activate(t, g, me.ID, tome, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	handBefore := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil || len(pick.ChooseCards) != 3 {
		t.Fatalf("the opponent should choose among the three revealed cards: %+v", pick)
	}
	chosen := pick.ChooseCards[0]
	effectstest.AnswerChooseCards(t, g, opp.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB777InZone(me.Graveyard, chosen) {
		t.Error("the chosen card did not go to the controller's graveyard")
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d -> %d, want two cards drawn", handBefore, got)
	}
	if !effectstest.PermanentTapped(g, tome) {
		t.Error("the Tome was not tapped to pay")
	}
}
