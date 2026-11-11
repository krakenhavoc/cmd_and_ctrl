package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s14YumaOracle        = "b8a0b967-3b5b-46f5-830b-5c8f1037784a"
	s14HazezonOracle     = "66b16cb3-c2b1-447c-a4b0-a67326c99f20"
	s14SandScoutOracle   = "ca9f9017-c952-4e89-b284-02396e6be0af"
	s14RegrowthOracle    = "c4d05f64-d1bb-4725-81ba-baf737f3066c"
	s14ChanterOracle     = "7811b50d-af76-43cd-8753-43cf65fbefa3"
	s14ProspectingOracle = "76dc9452-4e81-4152-bb4d-cdec6db767ed"
	s14TitaniaOracle     = "46c3a69a-4984-40b0-905f-4b0762c7ad72"
	s14RumbleweedOracle  = "9cfc01a7-e291-4388-bd09-94846d72e291"
	s14DesertOracle      = "195107ad-879d-4b02-a44a-a3ba70fedf88"
)

func s14Mill(t *testing.T, g *game.Game, me *game.Player, name, typeLine string) {
	t.Helper()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
	})
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Errorf("mill: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
}

func s14Graveyard(g *game.Game, me *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
	})
	return id
}

func s14Pool(me *game.Player, colors ...string) {
	for _, c := range colors {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func TestCmdr2S14RumbleweedCostsLessPerLandCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		s14Graveyard(g, me, "Forest", "Basic Land — Forest")
	}
	s14Graveyard(g, me, "Bear", "Creature — Bear")
	advanceToMain(t, g)
	// {10}{G} less 3 = {7}{G}: eight mana.
	s14Pool(me, "G", "C", "C", "C", "C", "C", "C", "C")
	castFromHandForTest(t, g, me, "Rumbleweed", "Creature — Plant Elemental", "{10}{G}", s14RumbleweedOracle, game.CastSpellParams{Strict: true})
}

func TestCmdr2S14RumbleweedWithoutLandsIsRefusedAtEightMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	s14Pool(me, "G", "C", "C", "C", "C", "C", "C", "C")
	if _, err := castFromHandErr(g, me, "Rumbleweed", "Creature — Plant Elemental", "{10}{G}", s14RumbleweedOracle, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("eight mana cannot pay {10}{G} with an empty graveyard")
	}
}

func TestCmdr2S14RumbleweedEntersPumpsOthersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Wolf", "Creature — Wolf", 2, 2)
	advanceToMain(t, g)
	weed := castCatalogSpell(t, g, "Rumbleweed", "Creature — Plant Elemental", s14RumbleweedOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(weed) {
		t.Fatal("Rumbleweed should be on the battlefield")
	}
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("your other creature power %d, want 5", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("your other creature should have trample")
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature power %d, want 2", got)
	}
	if got := effectivePower(t, g, weed); got != 0 {
		t.Errorf("Rumbleweed power %d, want its own 0 here (not pumped by itself)", got)
	}
}

func TestCmdr2S14TitaniaMakesAnElementalPerForestAndPlaysForestsFromGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Titania, Nature's Force", "Legendary Creature — Elemental", s14TitaniaOracle, 6, 6)
	forest := s14Graveyard(g, me, "Forest", "Basic Land — Forest")
	plains := s14Graveyard(g, me, "Plains", "Basic Land — Plains")
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, plains, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true}); err == nil {
		t.Fatal("Titania only lets you play Forests from your graveyard")
	}
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true}); err != nil {
		t.Fatalf("play Forest from graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Elemental", me.ID); n != 1 {
		t.Errorf("Elemental tokens = %d, want 1", n)
	}
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Elemental", me.ID); n != 1 {
		t.Errorf("a Plains entering made a token (%d)", n)
	}
}

func TestCmdr2S14HazezonMakesTwoSandWarriorsPerDesertAndPlaysDesertsFromGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Hazezon, Shaper of Sand", "Legendary Creature — Human Warrior", s14HazezonOracle, 3, 3)
	desert := s14Graveyard(g, me, "Desert", "Land — Desert")
	forest := s14Graveyard(g, me, "Forest", "Basic Land — Forest")
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true}); err == nil {
		t.Fatal("Hazezon only lets you play Desert lands from your graveyard")
	}
	if err := g.CastSpell(me.ID, desert, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true}); err != nil {
		t.Fatalf("play Desert from graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Sand Warrior", me.ID); n != 2 {
		t.Errorf("Sand Warriors = %d, want 2", n)
	}
	b12Permanent(g, g.Seats[1].ID, "Their Desert", "Land — Desert")
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Sand Warrior", me.ID); n != 2 {
		t.Errorf("an opponent's Desert made tokens (%d)", n)
	}
}

func TestCmdr2S14HazezonIsUnblockableWhileTheDefenderControlsADesert(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	haz := b12Push(g, me.ID, "Hazezon, Shaper of Sand", "Legendary Creature — Human Warrior", s14HazezonOracle, 3, 3)
	b12Permanent(g, opp.ID, "Their Desert", "Land — Desert")
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, haz)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, haz); err == nil {
		t.Fatal("desertwalk: Hazezon can't be blocked while the defender controls a Desert")
	}
}

func TestCmdr2S14SandScoutMakesOneSandWarriorPerTurnForLandCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Sand Scout", "Creature — Human Scout", s14SandScoutOracle, 2, 2)
	s14Mill(t, g, me, "Shock", "Instant")
	if n := tokensNamedControlledBy(g, "Sand Warrior", me.ID); n != 0 {
		t.Fatalf("a nonland card made a token (%d)", n)
	}
	s14Mill(t, g, me, "Forest", "Basic Land — Forest")
	s14Mill(t, g, me, "Island", "Basic Land — Island")
	if n := tokensNamedControlledBy(g, "Sand Warrior", me.ID); n != 1 {
		t.Errorf("Sand Warriors = %d, want 1 (once each turn)", n)
	}
}

func TestCmdr2S14SandScoutFetchesATappedDesertOnlyWhenBehindOnLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	desert := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: desert, Name: "Desert", TypeLine: "Land — Desert", Owner: me.ID, Controller: me.ID})
	})
	b12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	castCatalogSpell(t, g, "Sand Scout", "Creature — Human Scout", s14SandScoutOracle, nil)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID, desert)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCardByID(g, desert)
	if !ok {
		t.Fatal("the Desert should have been put onto the battlefield")
	}
	if !c.Tapped {
		t.Error("the Desert should enter tapped")
	}
}

func TestCmdr2S14SandScoutDoesNothingWhenNotBehind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	desert := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: desert, Name: "Desert", TypeLine: "Land — Desert", Owner: me.ID, Controller: me.ID})
	})
	castCatalogSpell(t, g, "Sand Scout", "Creature — Human Scout", s14SandScoutOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(desert) {
		t.Error("no opponent controls more lands: nothing is fetched")
	}
}

func TestCmdr2S14YumaMakesAPlantWarriorForADesertCardNotAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Yuma, Proud Protector", "Legendary Creature — Human Ranger", s14YumaOracle, 6, 6)
	s14Mill(t, g, me, "Forest", "Basic Land — Forest")
	if n := tokensNamedControlledBy(g, "Plant Warrior", me.ID); n != 0 {
		t.Fatalf("a Forest card made a token (%d)", n)
	}
	s14Mill(t, g, me, "Desert", "Land — Desert")
	s14Mill(t, g, me, "Desert", "Land — Desert")
	if n := tokensNamedControlledBy(g, "Plant Warrior", me.ID); n != 2 {
		t.Errorf("Plant Warriors = %d, want 2 (one per Desert card)", n)
	}
}

func TestCmdr2S14YumaSacrificesALandToDrawWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	yuma := b12Push(g, me.ID, "Yuma, Proud Protector", "Legendary Creature — Human Ranger", s14YumaOracle, 6, 6)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	before := me.Hand.Size()
	declareAttack(t, g, opp.ID, yuma)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, land)
	if g.Battlefield.Contains(land) {
		t.Error("the land should have been sacrificed")
	}
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want +1 card", before, me.Hand.Size())
	}
}

func TestCmdr2S14VengefulRegrowthReturnsLandsTappedAndMakesThatManyPlantWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := s14Graveyard(g, me, "Forest", "Basic Land — Forest")
	b := s14Graveyard(g, me, "Island", "Basic Land — Island")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Vengeful Regrowth", "Sorcery", s14RegrowthOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := battlefieldCardByID(g, id)
		if !ok || !c.Tapped {
			t.Errorf("land %s should be on the battlefield tapped", id)
		}
	}
	if n := tokensNamedControlledBy(g, "Plant Warrior", me.ID); n != 2 {
		t.Errorf("Plant Warriors = %d, want 2", n)
	}
}

func TestCmdr2S14CataclysmicProspectingDamagesEveryCreatureAndPaysTreasureForDesertMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, g.Seats[1].ID, "Ogre", "Creature — Ogre", 4, 4)
	toMainForCost(t, g)
	d1 := pushCatalogPermanent(g, me.ID, "Desert", "Land — Desert", s14DesertOracle, false)
	d2 := pushCatalogPermanent(g, me.ID, "Desert", "Land — Desert", s14DesertOracle, false)
	b28TapForMana(t, g, me.ID, d1, "")
	b28TapForMana(t, g, me.ID, d2, "")
	b06AddMana(me, "R", "R")
	id := handCardFull(me, "Cataclysmic Prospecting", "Sorcery", "{X}{R}{R}", s14ProspectingOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 should have died to 2 damage")
	}
	if !g.Battlefield.Contains(ogre) {
		t.Error("the 4/4 should survive 2 damage")
	}
	n, tapped := countTokensNamed(g, me.ID, "Treasure")
	if n != 2 || tapped != 2 {
		t.Errorf("Treasures = %d (%d tapped), want 2 tapped", n, tapped)
	}
}

func TestCmdr2S14CataclysmicProspectingWithoutDesertManaMakesNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	b06AddMana(me, "R", "R", "C")
	id := handCardFull(me, "Cataclysmic Prospecting", "Sorcery", "{X}{R}{R}", s14ProspectingOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("Treasures = %d, want 0", n)
	}
}

func TestCmdr2S14DuneChanterMakesLandsAndOffBattlefieldLandCardsDeserts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	inHand := uuid.New()
	inGrave := s14Graveyard(g, me, "Island", "Basic Land — Island")
	bear := s14Graveyard(g, me, "Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	})
	b12Push(g, me.ID, "Dune Chanter", "Creature — Plant Druid", s14ChanterOracle, 2, 3)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	has := func(id uuid.UUID, want bool, what string) {
		t.Helper()
		got := zoneLayerCard(t, g, id).HasSubtype("Desert")
		if got != want {
			t.Errorf("%s is a Desert = %v, want %v", what, got, want)
		}
	}
	has(forest, true, "a land you control")
	has(theirs, false, "an opponent's land")
	has(inHand, true, "a land card in your hand")
	has(inGrave, true, "a land card in your graveyard")
	has(bear, false, "a nonland card in your graveyard")
}

func TestCmdr2S14DuneChanterLandsTapForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Push(g, me.ID, "Dune Chanter", "Creature — Plant Druid", s14ChanterOracle, 2, 3)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	abs := game.ManaAbilitiesForCard(gaLayered(t, g, forest))
	found := len(abs) == 2 && abs[1].Produced == "{W|U|B|R|G}"
	if !found {
		t.Error("a land you control should have \"{T}: Add one mana of any color\"")
	}
}

func TestCmdr2S14DuneChanterMillsTwoAndGainsLifePerLandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	chanter := b12Push(g, me.ID, "Dune Chanter", "Creature — Plant Druid", s14ChanterOracle, 2, 3)
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for _, c := range []game.Card{
			{Name: "Deep", TypeLine: "Instant"},
			{Name: "Top Spell", TypeLine: "Instant"},
			{Name: "Top Land", TypeLine: "Basic Land — Forest"},
		} {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Library.PushTop(c)
		}
	})
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, chanter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 (one land card milled)", before, me.Life)
	}
}
