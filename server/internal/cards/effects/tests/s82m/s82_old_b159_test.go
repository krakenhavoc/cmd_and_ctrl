package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b159_test.go — Modern Horizons, batch 159 (#5519): each test plays
// the real catalog entry through the real cast, activation and trigger paths.

const (
	oldB159AnuridOracle    = "0239b661-80be-4b52-b3b7-fca229192b58"
	oldB159SnuffOracle     = "19f3ec61-f334-4708-ad93-04ebe15c1798"
	oldB159GambitOracle    = "33bb4aff-6476-4787-b3fe-028b1b1bd1bc"
	oldB159MageOracle      = "3f278970-201d-4a93-b038-cc7218f2ac57"
	oldB159GuideOracle     = "55b12241-2495-45e6-b618-3746d5521291"
	oldB159TrademageOracle = "6c94a255-c031-48d7-aa07-1189ad3da3b6"
	oldB159DjinnOracle     = "da0042a7-89a1-4500-b689-b0ea7cab5ed8"
	oldB159AstrolabeOracle = "ec463bde-dadf-4044-8e41-71338fd4d62f"
)

func TestOldB159CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB159AnuridOracle:    "Excavating Anurid",
		oldB159SnuffOracle:     "Spell Snuff",
		oldB159GambitOracle:    "Geomancer's Gambit",
		oldB159MageOracle:      "Pondering Mage",
		oldB159GuideOracle:     "Alpine Guide",
		oldB159TrademageOracle: "Bazaar Trademage",
		oldB159DjinnOracle:     "Thundering Djinn",
		oldB159AstrolabeOracle: "Arcum's Astrolabe",
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

func oldB159Land(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, name, typeLine, "", false)
}

// --- Excavating Anurid ---

func TestOldB159AnuridMaySacrificeALandToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB159Land(g, me.ID, "Forest", "Basic Land — Forest")
	hand := me.Hand.Size()
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Excavating Anurid", TypeLine: "Creature — Frog Beast", OracleID: oldB159AnuridOracle, Power: 4, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestOldB159AnuridDecliningTheSacrificeDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB159Land(g, me.ID, "Forest", "Basic Land — Forest")
	hand := me.Hand.Size()
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Excavating Anurid", TypeLine: "Creature — Frog Beast", OracleID: oldB159AnuridOracle, Power: 4, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || me.Hand.Size() != hand {
		t.Error("declining the sacrifice still sacrificed a land or drew")
	}
}

func TestOldB159AnuridThresholdGrowsAndGainsVigilanceAtSevenCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	anurid := effectstest.B12Push(g, me.ID, "Excavating Anurid", "Creature — Frog Beast", oldB159AnuridOracle, 4, 4)
	effectstest.BlbMPFillGraveyard(g, me, 6)
	if p, tt := effectstest.PtOf(t, g, anurid); p != 4 || tt != 4 || effectstest.HasEffectiveKeyword(t, g, anurid, "vigilance") {
		t.Errorf("with six cards the Anurid is %d/%d (vigilance %v), want a plain 4/4", p, tt, effectstest.HasEffectiveKeyword(t, g, anurid, "vigilance"))
	}
	effectstest.BlbMPFillGraveyard(g, me, 1)
	if p, tt := effectstest.PtOf(t, g, anurid); p != 5 || tt != 5 || !effectstest.HasEffectiveKeyword(t, g, anurid, "vigilance") {
		t.Errorf("with seven cards the Anurid is %d/%d, want a 5/5 with vigilance", p, tt)
	}
}

// --- Thundering Djinn ---

func TestOldB159DjinnDealsDamageEqualToCardsDrawnThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	djinn := effectstest.B12Push(g, me.ID, "Thundering Djinn", "Creature — Djinn", oldB159DjinnOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 3); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	if drawn < 3 {
		t.Fatalf("the turn tally holds %d draws after drawing three", drawn)
	}
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, djinn)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-drawn {
		t.Errorf("opponent life %d -> %d, want %d damage for %d cards drawn this turn", life, opp.Life, drawn, drawn)
	}
}

func TestOldB159DjinnCountsOnlyYourOwnDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	djinn := effectstest.B12Push(g, me.ID, "Thundering Djinn", "Creature — Djinn", oldB159DjinnOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	before := g.TurnTallyFor(me.ID).CardsDrawn
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 5); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, djinn)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-before {
		t.Errorf("opponent life %d -> %d, want %d (only my own %d draws count)", life, opp.Life, before, before)
	}
}

// --- Bazaar Trademage ---

func TestOldB159TrademageDrawsTwoThenDiscardsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	var held []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		held = append(held, id)
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Held", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Bazaar Trademage", TypeLine: "Creature — Human Wizard", OracleID: oldB159TrademageOracle, Power: 3, Toughness: 4, Keywords: []string{"flying"}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 4 {
		t.Fatalf("hand is %d after drawing two, want 4", got)
	}
	effectstest.AnswerDiscard(t, g, me.ID, held[0], held[1], me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand is %d after the discard, want 1", got)
	}
	if !me.Graveyard.Contains(held[0]) || !me.Graveyard.Contains(held[1]) {
		t.Error("the chosen cards were not discarded")
	}
}

// --- Spell Snuff ---

func TestOldB159SpellSnuffCountersAndDrawsOnlyAtFiveOrLessLife(t *testing.T) {
	for _, tc := range []struct {
		life, draws int
	}{{6, 0}, {5, 1}, {2, 1}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		me.Life = tc.life
		bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Spell", "Instant", "", "{R}", nil)
		hand := me.Hand.Size()
		snuff := effectstest.B13HandCard(me, "Spell Snuff", "Instant", oldB159SnuffOracle)
		if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Graveyard.Contains(bolt) {
			t.Errorf("life %d: the spell was not countered", tc.life)
		}
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("life %d: net hand change %d, want %d", tc.life, got, tc.draws)
		}
	}
}

// --- Geomancer's Gambit ---

func TestOldB159GeomancersGambitDestroysALandAndItsControllerMaySearchForAnUntappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := oldB159Land(g, opp.ID, "Their Forest", "Basic Land — Forest")
	basic := effectstest.PushLibraryCard(opp, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Geomancer's Gambit", "Sorcery", oldB159GambitOracle, effectstest.B16TargetCard(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(basic) {
		t.Fatal("the land's controller did not put a basic onto the battlefield")
	}
	if effectstest.IsTapped(g, basic) {
		t.Error("the basic entered tapped")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the caster's hand %d -> %d, want the spell cast and a card drawn (net one)", hand, me.Hand.Size())
	}
}

func TestOldB159GeomancersGambitCantTargetANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Geomancer's Gambit", "Sorcery", oldB159GambitOracle, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("the spell targeted a creature")
	}
}

// --- Pondering Mage ---

func TestOldB159PonderingMageLooksReordersAndDrawsTheNewTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushLibraryCard(me, game.Card{Name: "Bottom Of Three", TypeLine: "Sorcery"})
	b := effectstest.PushLibraryCard(me, game.Card{Name: "Middle Of Three", TypeLine: "Sorcery"})
	c := effectstest.PushLibraryCard(me, game.Card{Name: "Top Of Three", TypeLine: "Sorcery"})
	hand := me.Hand.Size()
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Pondering Mage", TypeLine: "Creature — Human Wizard", OracleID: oldB159MageOracle, Power: 3, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	look := effectstest.LookAtTopChoiceFor(g, me.ID)
	if look == nil {
		t.Fatal("no look-at-the-top prompt")
	}
	if me.Hand.Size() != hand {
		t.Fatal("a card was drawn before the order and shuffle were settled")
	}
	// Put them back with the old bottom card on top.
	if err := g.ResolveLookAtTop(look.ID, me.ID, []uuid.UUID{a, b, c}); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want one card drawn after the shuffle question", hand, me.Hand.Size())
	}
	if !me.Hand.Contains(a) && !me.Hand.Contains(b) && !me.Hand.Contains(c) {
		t.Error("the card drawn is not one of the three that were looked at")
	}
}

func TestOldB159PonderingMageShufflingStillDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Pondering Mage", TypeLine: "Creature — Human Wizard", OracleID: oldB159MageOracle, Power: 3, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	look := effectstest.LookAtTopChoiceFor(g, me.ID)
	if look == nil {
		t.Fatal("no look-at-the-top prompt")
	}
	if err := g.ResolveLookAtTop(look.ID, me.ID, look.ScryCards); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn after shuffling", hand, me.Hand.Size())
	}
}

// --- Arcum's Astrolabe ---

func TestOldB159AstrolabeDrawsOnEntryAndMakesAnyColourForOneMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	astrolabe := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Arcum's Astrolabe", TypeLine: "Snow Artifact", OracleID: oldB159AstrolabeOracle})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn on entry", hand, me.Hand.Size())
	}
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, astrolabe, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want 1 (one paid, one made)", me.ManaPool.Total())
	}
}

func TestOldB159AstrolabeNeedsOneMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	astrolabe := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Arcum's Astrolabe", TypeLine: "Snow Artifact", OracleID: oldB159AstrolabeOracle})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, astrolabe, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the Astrolabe made mana with no mana to pay for it")
	}
}

// --- Alpine Guide ---

func TestOldB159AlpineGuideFetchesAnyMountainTappedAndMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mountain := effectstest.PushLibraryCard(me, game.Card{Name: "Sulfurous Springs Pro", TypeLine: "Land — Mountain Island"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	guide := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Alpine Guide", TypeLine: "Snow Creature — Human Scout", OracleID: oldB159GuideOracle, Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		if effectstest.SearchOptionNamed(g, c, "Plains") != uuid.Nil {
			t.Error("a Plains was offered for a Mountain search")
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, "Sulfurous Springs Pro")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mountain) || !effectstest.IsTapped(g, mountain) {
		t.Fatal("the Mountain card did not enter the battlefield tapped")
	}
	_ = guide
}

func TestOldB159AlpineGuideMustAttackEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Alpine Guide", "Snow Creature — Human Scout", oldB159GuideOracle, 3, 3)
	if !effectstest.FaMustAttack(t, g) {
		t.Error("the Guide let its controller pass the declare attackers step without attacking")
	}
}

func TestOldB159AlpineGuideLeavingMakesYouSacrificeAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guide := effectstest.B12Push(g, me.ID, "Alpine Guide", "Snow Creature — Human Scout", oldB159GuideOracle, 3, 3)
	m1 := oldB159Land(g, me.ID, "Mountain", "Basic Land — Mountain")
	forest := oldB159Land(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.P1Destroy(t, g, guide)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, m1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(m1) {
		t.Error("the Mountain was not sacrificed")
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("a land that is not a Mountain was sacrificed")
	}
}
