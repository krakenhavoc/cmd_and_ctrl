package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestS82DmuB14EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		citizensArrestOracle:    "Citizen's Arrest",
		temporaryLockdownOracle: "Temporary Lockdown",
		historiansBoonOracle:    "Historian's Boon",
		combatResearchOracle:    "Combat Research",
		weatherseedTreatyOracle: "The Weatherseed Treaty",
		yotiaDeclaresWarOracle:  "Yotia Declares War",
		braidsFrightfulOracle:   "Braids's Frightful Return",
		theWorldSpellOracle:     "The World Spell",
		foundingThirdPathOracle: "Founding the Third Path",
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

// The Weatherseed Treaty, chapter by chapter from chapter I: the basic land
// enters tapped, the Saproling is a 1/1, and chapter III pumps by the
// number of basic land types and grants trample.
func TestTheWeatherseedTreatyPlaysItsChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushLibraryTyped(me, "Forest", "Basic Land — Forest")

	saga := effectstest.CastReadAheadSaga(t, g, "The Weatherseed Treaty", weatherseedTreatyOracle)
	effectstest.ChooseChapter(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	forest := effectstest.FindBattlefieldByName(g, "Forest")
	if forest == uuid.Nil {
		t.Fatal("chapter I fetched no Forest")
	}
	if c, _ := g.LookupCardForEffect(forest); !c.Tapped {
		t.Error("the fetched land did not enter tapped")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	var sap *game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Saproling" {
			cc := c
			sap = &cc
		}
	}
	if sap == nil || sap.Power != 1 || sap.Toughness != 1 || !sap.IsToken() {
		t.Fatalf("chapter II Saproling = %+v", sap)
	}

	effectstest.PushPermanentForTest(g, me.ID, "Island", "", "Basic Land — Island")
	effectstest.PushPermanentForTest(g, me.ID, "Swamp", "", "Basic Land — Swamp")
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Fatalf("power after chapter III = %d, want 2+3 (Forest, Island, Swamp)", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("no trample")
	}
	if g.Battlefield.Contains(saga) {
		t.Fatal("the Saga was not sacrificed after chapter III")
	}
}

// Yotia Declares War chapter I: a 0/2 flying Thopter named Ornithopter.
func TestYotiaDeclaresWarMakesAnOrnithopter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastReadAheadSaga(t, g, "Yotia Declares War", yotiaDeclaresWarOracle)
	effectstest.ChooseChapter(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	var tok *game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ornithopter" {
			cc := c
			tok = &cc
		}
	}
	if tok == nil || !tok.IsToken() || !tok.IsArtifact() || !tok.IsCreature() || tok.Power != 0 || tok.Toughness != 2 || len(tok.Colors) != 0 ||
		!game.HasKeyword(tok, "flying") || !tok.HasSubtype("Thopter") {
		t.Fatalf("Ornithopter = %+v", tok)
	}
}

// Chapter II: tap any number of untapped artifacts; the Saga deals that
// much damage to a target creature or planeswalker.
func TestYotiaDeclaresWarTapsArtifactsForDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	a := effectstest.PushPermanentForTest(g, me.ID, "Rock A", "", "Artifact")
	b := effectstest.PushPermanentForTest(g, me.ID, "Rock B", "", "Artifact")
	tappedRock := effectstest.PushPermanentForTest(g, me.ID, "Rock C", "", "Artifact")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tappedRock) })
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 3)
	spare := effectstest.PushVanillaCreature(g, opp.ID, "Spare", 5, 5)

	effectstest.CastReadAheadSaga(t, g, "Yotia Declares War", yotiaDeclaresWarOracle)
	effectstest.ChooseChapter(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no tap prompt")
	}
	if effectstest.SlicesContainsID(pick.ChooseCards, tappedRock) {
		t.Error("a tapped artifact is offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := g.LookupCardForEffect(id); !c.Tapped {
			t.Error("a picked artifact was not tapped")
		}
	}
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("2 damage killed a creature with toughness 3")
	}
	_ = spare
}

// Tapping none does nothing; the chapter is not an error.
func TestYotiaDeclaresWarTappingNothingDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	effectstest.PushPermanentForTest(g, me.ID, "Rock A", "", "Artifact")
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 1)
	effectstest.CastReadAheadSaga(t, g, "Yotia Declares War", yotiaDeclaresWarOracle)
	effectstest.ChooseChapter(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("damage was dealt although nothing was tapped")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a reflexive trigger asked for a target after tapping nothing")
	}
}

// Chapter II kills a creature when the damage is enough.
func TestYotiaDeclaresWarDamageKills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	a := effectstest.PushPermanentForTest(g, me.ID, "Rock A", "", "Artifact")
	b := effectstest.PushPermanentForTest(g, me.ID, "Rock B", "", "Artifact")
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	effectstest.CastReadAheadSaga(t, g, "Yotia Declares War", yotiaDeclaresWarOracle)
	effectstest.ChooseChapter(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("two tapped artifacts should deal 2 damage and kill a 2/2")
	}
}

// Chapter III: an artifact becomes a 4/4 artifact creature.
func TestYotiaDeclaresWarAnimatesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	rock := effectstest.PushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
	effectstest.CastReadAheadSaga(t, g, "Yotia Declares War", yotiaDeclaresWarOracle)
	effectstest.ChooseChapter(t, g, 3)
	effectstest.PickTriggerTarget(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(rock)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatal("the artifact did not become an artifact creature")
	}
	if p := effectstest.EffectivePower(t, g, rock); p != 4 {
		t.Fatalf("power %d, want 4", p)
	}
}

// Braids's Frightful Return chapter I: sacrificing a creature makes each
// opponent discard; declining does not.
func TestBraidsFrightfulReturnChapterI(t *testing.T) {
	for _, sac := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me := g.Seats[seat]
		opp := g.Seats[(seat+1)%4]
		bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
		effectstest.HandCardForTest(opp, "Spare A", "Instant", "")
		effectstest.HandCardForTest(opp, "Spare B", "Instant", "")
		before := opp.Hand.Size()
		effectstest.CastReadAheadSaga(t, g, "Braids's Frightful Return", braidsFrightfulOracle)
		effectstest.ChooseChapter(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if pick == nil {
			t.Fatal("no sacrifice prompt")
		}
		picks := []uuid.UUID{}
		if sac {
			picks = []uuid.UUID{bear}
		}
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
			effectstest.DiscardFromHand(t, g, opp.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if sac && (g.Battlefield.Contains(bear) || opp.Hand.Size() != before-1) {
			t.Errorf("sacrificed: bear on battlefield %v, opponent hand %d want %d", g.Battlefield.Contains(bear), opp.Hand.Size(), before-1)
		}
		if !sac && (!g.Battlefield.Contains(bear) || opp.Hand.Size() != before) {
			t.Errorf("declined: bear on battlefield %v, opponent hand %d want %d", g.Battlefield.Contains(bear), opp.Hand.Size(), before)
		}
	}
}

// Chapter II returns a creature card from the graveyard to the hand.
func TestBraidsFrightfulReturnChapterII(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	creature := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "")
	spell := effectstest.PushGraveyardPermanent(me, "Dead Spell", "Sorcery", "")
	effectstest.CastReadAheadSaga(t, g, "Braids's Frightful Return", braidsFrightfulOracle)
	effectstest.ChooseChapter(t, g, 2)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: spell}); err == nil {
		t.Fatal("a sorcery card is not a legal target")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) {
		t.Fatal("the creature card is not in hand")
	}
}

// Chapter III: the opponent sacrifices a nonland nontoken permanent, or
// loses 2 life and you draw.
func TestBraidsFrightfulReturnChapterIII(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
		keep := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		effectstest.PushPermanentForTest(g, opp.ID, "Their Land", "", "Land")
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Creature — Soldier", Owner: opp.ID, Controller: opp.ID})
		effectstest.CastReadAheadSaga(t, g, "Braids's Frightful Return", braidsFrightfulOracle)
		effectstest.ChooseChapter(t, g, 3)
		pr := effectstest.LatestPickTarget(g, me.ID)
		if pr == nil {
			t.Fatal("no target prompt")
		}
		if err := g.ResolvePickTarget(pr.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
			t.Fatal("targeted myself")
		}
		if err := g.ResolvePickTarget(pr.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		return g, me, opp, keep, uuid.Nil
	}
	t.Run("sacrifices", func(t *testing.T) {
		g, me, opp, bear, _ := setup(t)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
		if pick == nil {
			t.Fatal("the opponent was not asked")
		}
		if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != bear {
			t.Fatalf("offered %v, want only the nontoken nonland creature", pick.ChooseCards)
		}
		life, hand := opp.Life, me.Hand.Size()
		if err := g.ResolveOwnPermanents(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) || opp.Life != life || me.Hand.Size() != hand {
			t.Fatalf("sacrificed: bear %v life %d->%d hand %d->%d", g.Battlefield.Contains(bear), life, opp.Life, hand, me.Hand.Size())
		}
	})
	t.Run("declines", func(t *testing.T) {
		g, me, opp, bear, _ := setup(t)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
		if pick == nil {
			t.Fatal("the opponent was not asked")
		}
		life, hand := opp.Life, me.Hand.Size()
		if err := g.ResolveOwnPermanents(pick.ID, opp.ID, nil); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(bear) || opp.Life != life-2 || me.Hand.Size() != hand+1 {
			t.Fatalf("declined: bear %v life %d->%d hand %d->%d", g.Battlefield.Contains(bear), life, opp.Life, hand, me.Hand.Size())
		}
	})
}

// With nothing to sacrifice the opponent is "not sacrificing": they lose
// life and you draw.
func TestBraidsFrightfulReturnChapterIIIWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	life, hand := opp.Life, me.Hand.Size()
	effectstest.CastReadAheadSaga(t, g, "Braids's Frightful Return", braidsFrightfulOracle)
	effectstest.ChooseChapter(t, g, 3)
	pr := effectstest.LatestPickTarget(g, me.ID)
	if pr == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pr.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Fatalf("life %d->%d hand %d->%d", life, opp.Life, hand, me.Hand.Size())
	}
}

// The World Spell chapters I and II: look at seven, take one non-Saga
// permanent card, the rest to the bottom.
func TestTheWorldSpellLooksAtSevenAndTakesANonSagaPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H", "I")
	deep := effectstest.TopLibraryCard(me, "Eighth", "Creature — Bear")
	_ = deep
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Sorcery")
	}
	saga := effectstest.PushLibraryTyped(me, "Other Saga", "Enchantment — Saga")
	want := effectstest.PushLibraryTyped(me, "Wanted", "Artifact")
	effectstest.PushLibraryTyped(me, "Spell", "Instant")
	effectstest.CastReadAheadSaga(t, g, "The World Spell", theWorldSpellOracle)
	effectstest.ChooseChapter(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no look prompt")
	}
	if effectstest.SlicesContainsID(pick.ChooseCards, saga) {
		t.Error("a Saga is offered")
	}
	if !effectstest.SlicesContainsID(pick.ChooseCards, want) {
		t.Error("the artifact is not offered")
	}
	size := me.Library.Size()
	effectstest.AnswerChooseCards(t, g, me.ID, want)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(want) {
		t.Fatal("the card is not in hand")
	}
	if me.Library.Size() != size-1 {
		t.Fatalf("library %d -> %d, want one card gone", size, me.Library.Size())
	}
	// The seven looked-at cards but one are now at the bottom: the card that
	// was eighth from the top is on top.
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.Name != "Eighth" {
		t.Errorf("top of library is %q, want the card that was eighth", top.Name)
	}
}

// Chapter III: up to two non-Saga permanent cards from hand enter.
func TestTheWorldSpellPutsUpToTwoPermanentsFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := effectstest.HandCardForTest(me, "Hand Bear", "Creature — Bear", "")
	rock := effectstest.HandCardForTest(me, "Hand Rock", "Artifact", "")
	third := effectstest.HandCardForTest(me, "Hand Land", "Land", "")
	spell := effectstest.HandCardForTest(me, "Hand Spell", "Instant", "")
	saga := effectstest.HandCardForTest(me, "Hand Saga", "Enchantment — Saga", "")
	effectstest.CastReadAheadSaga(t, g, "The World Spell", theWorldSpellOracle)
	effectstest.ChooseChapter(t, g, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	for _, bad := range []uuid.UUID{spell, saga} {
		if effectstest.SlicesContainsID(pick.ChooseCards, bad) {
			t.Errorf("%s is offered", bad)
		}
	}
	if pick.ChooseMax != 2 {
		t.Errorf("max %d, want 2", pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear, rock, third}); err == nil {
		t.Fatal("three cards accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) || g.Battlefield.Contains(third) {
		t.Fatal("wrong cards entered")
	}
}

// Founding the Third Path chapter I: a free cast of an instant or sorcery
// with mana value 1 or 2 only.
func TestFoundingTheThirdPathFreeCastIsLimited(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	mk := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		return id
	}
	good := mk("Cheap Instant", "Instant", "{1}{U}")
	pricey := mk("Pricey Instant", "Instant", "{2}{U}")
	free := mk("Free Sorcery", "Sorcery", "")
	creature := mk("Cheap Bear", "Creature — Bear", "{1}")
	effectstest.CastReadAheadSaga(t, g, "Founding the Third Path", foundingThirdPathOracle)
	effectstest.ChooseChapter(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind != game.PendingChoiceEntryReadAhead {
			offer = c
		}
	}
	if offer == nil {
		t.Fatal("no free-cast offer")
	}
	ids := offer.DiscardOptions
	if len(ids) == 0 {
		ids = offer.ChooseCards
	}
	if !effectstest.SlicesContainsID(ids, good) {
		t.Errorf("the mana value 2 instant is not offered (%+v)", offer)
	}
	for name, bad := range map[string]uuid.UUID{"pricey": pricey, "free": free, "creature": creature} {
		if effectstest.SlicesContainsID(ids, bad) {
			t.Errorf("%s is offered", name)
		}
	}
}

// Chapter II: the target player mills four.
func TestFoundingTheThirdPathMillsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	before := opp.Library.Size()
	effectstest.CastReadAheadSaga(t, g, "Founding the Third Path", foundingThirdPathOracle)
	effectstest.ChooseChapter(t, g, 2)
	pr := effectstest.LatestPickTarget(g, me.ID)
	if pr == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pr.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 4 {
		t.Fatalf("milled %d, want 4", got)
	}
}

// Chapter III: exile an instant or sorcery card from your graveyard, copy
// it, and you may cast the copy; the original stays in exile.
func TestFoundingTheThirdPathCopiesAGraveyardSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	creature := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "")
	spell := effectstest.PushGraveyardPermanent(me, "Dead Instant", "Instant", "")
	effectstest.CastReadAheadSaga(t, g, "Founding the Third Path", foundingThirdPathOracle)
	effectstest.ChooseChapter(t, g, 3)
	pr := effectstest.LatestPickTarget(g, me.ID)
	if pr == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pr.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: creature}); err == nil {
		t.Fatal("a creature card is not a legal target")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if dmuB14Zone(g, spell) != game.ZoneExile {
		t.Fatalf("the original is in %q, want exile", dmuB14Zone(g, spell))
	}
}
