package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b185_test.go — Guilds of Ravnica leftovers (S82 older sets, batch 185).

const (
	s82Old185EgoOracle        = "1206b672-e075-45ba-97d4-7c3012be062d"
	s82Old185GuideOracle      = "1eb4c001-1115-4b99-bbbf-703eb551a381"
	s82Old185GuildmageOracle  = "257cc8e9-55b4-4510-9d1f-5df8c1ae2061"
	s82Old185FlowerOracle     = "2f8243a4-f347-4e18-9094-570b34df903b"
	s82Old185SummitOracle     = "38a3e144-93ad-497b-9b29-7040b6f337e1"
	s82Old185AgentOracle      = "9a293c87-ad40-4c6c-9fa4-9a8b5f081ac9"
	s82Old185DispersalOracle  = "a81b7629-9db9-4351-8d0e-c45a7a6167d4"
	s82Old185StoneglareOracle = "c39e2046-173a-44b6-8e93-b22ede35485d"
	s82Old185DevourerOracle   = "fb6d24cb-4e79-4d0c-b32f-47d5f709de1b"
)

func TestS82Old185CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82Old185EgoOracle:           "Unmoored Ego",
		s82Old185GuideOracle:         "District Guide",
		s82Old185GuildmageOracle:     "House Guildmage",
		s82Old185FlowerOracle:        "Flower",
		s82Old185FlowerOracle + "#1": "Flourish",
		s82Old185SummitOracle:        "Guild Summit",
		s82Old185AgentOracle:         "Darkblade Agent",
		s82Old185DispersalOracle:     "Ral's Dispersal",
		s82Old185StoneglareOracle:    "Vraska's Stoneglare",
		s82Old185DevourerOracle:      "Book Devourer",
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

func TestS82Old185UnmooredEgoExilesNamedCardsAndTheyDrawForTheHandOne(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	x.Put("lib 2", x.Opp.Library, "Lightning Bolt", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Unmoored Ego", "Sorcery", s82Old185EgoOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.Me.ID}}); err == nil {
		t.Fatal("Unmoored Ego targeted its caster; the text says target opponent")
	}
	hand := x.Opp.Hand.Size()
	effectstest.CastCatalogSpell(t, x.G, "Unmoored Ego", "Sorcery", s82Old185EgoOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: x.Opp.ID}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.ActNameFor(t, x.G, x.Me.ID, "Lightning Bolt")
	c := effectstest.SearchChoiceFor(x.G, x.Me.ID)
	if c == nil || c.SearchMax != 4 {
		t.Fatalf("the search takes up to four, got %+v", c)
	}
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["gy"], x.Cards["gy 2"], x.Cards["hand"], x.Cards["lib"])
	x.Exiled(t, "gy", "gy 2", "hand", "lib")
	x.NotExiled(t, "lib 2", "gy other", "hand other", "lib other")
	effectstest.PassPriorityAroundTable(t, x.G)
	if got := x.Opp.Hand.Size(); got != hand {
		t.Errorf("their hand %d → %d, want unchanged: one card exiled from it, one drawn", hand, got)
	}
}

func TestS82Old185DistrictGuideFetchesABasicLandOrGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Cradle", TypeLine: "Legendary Land"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Dimir Gate", TypeLine: "Land — Gate"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "District Guide", "Creature — Elf Scout", s82Old185GuideOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Cradle") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a card that is neither a basic land nor a Gate was offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Dimir Gate") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Error("a Gate and a basic land should both be offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Dimir Gate")
	if !effectstest.B02bHandHasNamed(me, "Dimir Gate") {
		t.Error("the Gate did not reach the hand")
	}
}

func TestS82Old185HouseGuildmageFreezesACreatureForOneUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "House Guildmage", "Creature — Human Wizard", s82Old185GuildmageOracle, 2, 2)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{U}{U}")
	effectstest.B16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, mage) {
		t.Error("the Guildmage taps to pay")
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	effectstest.AssertTurnOf(t, g, 0, false, "my next turn")
	if !effectstest.TappedNow(g, bear) {
		t.Error("its controller's next untap step untapped it")
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	if effectstest.TappedNow(g, bear) {
		t.Error("it stayed tapped through a second untap step")
	}
}

func TestS82Old185HouseGuildmageSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "House Guildmage", "Creature — Human Wizard", s82Old185GuildmageOracle, 2, 2)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.B16Activate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EoeQZSurveil(t, g, me.ID, 2)
	effectstest.EoeQZKeepOnTop(t, g, me.ID, c)
}

func TestS82Old185FlowerFetchesAForestOrPlainsAndFlourishPumpsTheTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Taiga", TypeLine: "Land — Mountain Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
	effectstest.CastCatalogSpell(t, g, "Flower", "Sorcery", s82Old185FlowerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Island") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Taiga") != uuid.Nil {
		t.Error("only a basic Forest or Plains should be offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains")
	if !effectstest.B02bHandHasNamed(me, "Plains") {
		t.Error("the Plains did not reach the hand")
	}

	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Flourish", "Sorcery", s82Old185FlowerOracle+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("bear power %d, want 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Errorf("bear toughness %d, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opposing creature has power %d, want 2", got)
	}
}

func TestS82Old185GuildSummitTapsGatesToDrawAndDrawsOnAGateEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gateA := effectstest.B12Permanent(g, me.ID, "Gate A", "Land — Gate")
	gateB := effectstest.B12Permanent(g, me.ID, "Gate B", "Land — Gate")
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Guild Summit", "Enchantment", s82Old185SummitOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("not asked which Gates to tap")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("min %d: tapping Gates is optional", pick.ChooseMin)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{gateA, gateB}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.B16Tapped(t, g, gateA) || !effectstest.B16Tapped(t, g, gateB) {
		t.Error("the Gates should be tapped")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2 (two Gates tapped)", hand, got)
	}
	before := me.Hand.Size()
	effectstest.B31PlayLand(t, g, "Dimir Gate", "Land — Gate", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d: the Gate left the hand and one card was drawn, want +1 net", before, got)
	}
}

func TestS82Old185DarkbladeAgentNeedsASurveilThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	agent := effectstest.B12Push(g, me.ID, "Darkblade Agent", "Creature — Human Assassin", s82Old185AgentOracle, 2, 3)
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, agent, "deathtouch") {
		t.Error("deathtouch before any surveil")
	}
	effectstest.DealCombatDamageToPlayer(g, agent, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(top) {
		t.Error("drew without having surveilled")
	}
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventSurveil, Actor: me.ID, LookedAt: 1}) })
	if !effectstest.HasEffectiveKeyword(t, g, agent, "deathtouch") {
		t.Error("no deathtouch after surveilling")
	}
	effectstest.DealCombatDamageToPlayer(g, agent, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("did not draw after surveilling")
	}
}

func TestS82Old185RalsDispersalBouncesAndMayFetchRal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	ral := effectstest.SeedSearchLibrary(me, game.Card{Name: "Ral, Caller of Storms", TypeLine: "Legendary Planeswalker — Ral"})[0]
	effectstest.CastCatalogSpell(t, g, "Ral's Dispersal", "Instant", s82Old185DispersalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Ral, Caller of Storms")
	if !me.Hand.Contains(ral) {
		t.Error("Ral did not reach the hand")
	}
}

func TestS82Old185VraskasStoneglareDestroysGainsToughnessAndMayFetchVraska(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 7, "W")
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Vraska, Regal Gorgon", TypeLine: "Legendary Planeswalker — Vraska"})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Vraska's Stoneglare", "Sorcery", s82Old185StoneglareOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: wall}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the creature survived")
	}
	if me.Life != life+7 {
		t.Errorf("life %d → %d, want +7 (its toughness)", life, me.Life)
	}
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Vraska, Regal Gorgon")
	if !effectstest.B02bHandHasNamed(me, "Vraska, Regal Gorgon") {
		t.Error("Vraska did not reach the hand")
	}
}

func TestS82Old185BookDevourerDiscardsTheHandAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	devourer := effectstest.B12Push(g, me.ID, "Book Devourer", "Creature — Beast", s82Old185DevourerOracle, 4, 5)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	var inHand []uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		for i := 0; i < 2; i++ {
			id := uuid.New()
			inHand = append(inHand, id)
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Old", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, devourer, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	for _, id := range inHand {
		if !me.Graveyard.Contains(id) {
			t.Error("a card from the old hand was not discarded")
		}
	}
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2 (drew as many as were discarded)", got)
	}
}
