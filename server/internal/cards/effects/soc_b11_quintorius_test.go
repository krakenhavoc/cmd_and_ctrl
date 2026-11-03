package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b11_quintorius_test.go — the two Quintorius cards of Secrets of
// Strixhaven Commander batch 11 that make a Spirit when cards leave a
// graveyard (#4834).

// --- Quintorius, Field Historian ---

func TestSocB11FieldHistorianPumpsOnlyYourSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := b12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", socB11FieldHistorianOracle, 2, 4)
	mySpirit := b12Push(g, me.ID, "My Spirit", "Creature — Spirit", "", 3, 2)
	theirSpirit := b12Push(g, opp.ID, "Their Spirit", "Creature — Spirit", "", 3, 2)
	myBear := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{historian, "Quintorius", 2}, {mySpirit, "my Spirit", 4}, {theirSpirit, "their Spirit", 3}, {myBear, "my Bear", 2}} {
		if got := effectivePower(t, g, tc.id); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
	if got := effectiveToughness(t, g, mySpirit); got != 2 {
		t.Errorf("my Spirit's toughness is %d, want 2 (+1/+0 only)", got)
	}
}

func TestSocB11FieldHistorianMakesASpiritWhenACardLeavesYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := b12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", socB11FieldHistorianOracle, 2, 4)
	mine := socB11GraveyardCreature(me, "Dead Bear")
	theirs := socB11GraveyardCreature(opp, "Their Dead Bear")

	// A card leaving ANOTHER player's graveyard makes nothing.
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneHand)
	if triggerOnStack(g, historian) != nil {
		t.Fatal("triggered for a card leaving an opponent's graveyard")
	}
	// Any card of yours, to any zone, makes a 3/2 red and white Spirit.
	socB11Move(t, g, me.ID, mine, game.ZoneGraveyard, game.ZoneHand)
	if triggerOnStack(g, historian) == nil {
		t.Fatal("no trigger for a card leaving your graveyard")
	}
	passPriorityAroundTable(t, g)
	spirits := socB11Named(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirit tokens, want 1", len(spirits))
	}
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(spirits[0])
		if !c.IsToken() || c.Controller != me.ID || c.Power != 3 || c.Toughness != 2 || len(c.Colors) != 2 {
			t.Errorf("token %+v is not a 3/2 red and white Spirit of yours", c)
		}
	})
	// The lord pumps the token it made: 4/2.
	if got := effectivePower(t, g, spirits[0]); got != 4 {
		t.Errorf("the new Spirit has power %d, want 4", got)
	}
}

func TestSocB11FieldHistorianMakesOneSpiritForCardsLeavingAsOneEvent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", socB11FieldHistorianOracle, 2, 4)
	a := socB11GraveyardCreature(me, "Dead One")
	b := socB11GraveyardCreature(me, "Dead Two")
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{a, b}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := len(socB11Named(g, "Spirit")); got != 1 {
		t.Errorf("%d Spirits for two cards leaving as one event, want 1", got)
	}
}

func TestSocB11FieldHistorianMakesOneSpiritPerSeparateEvent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", socB11FieldHistorianOracle, 2, 4)
	a := socB11GraveyardCreature(me, "Dead One")
	b := socB11GraveyardCreature(me, "Dead Two")
	socB11Move(t, g, me.ID, a, game.ZoneGraveyard, game.ZoneHand)
	passPriorityAroundTable(t, g)
	socB11Move(t, g, me.ID, b, game.ZoneGraveyard, game.ZoneHand)
	passPriorityAroundTable(t, g)
	if got := len(socB11Named(g, "Spirit")); got != 2 {
		t.Errorf("%d Spirits for two separate events, want 2", got)
	}
}

// --- Quintorius, History Chaser ---

func socB11Chaser(g *game.Game, me *game.Player, loyalty int) uuid.UUID {
	q := pushCatalogPermanent(g, me.ID, "Quintorius, History Chaser", "Legendary Planeswalker — Quintorius", socB11HistoryChaserOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, loyalty) })
	return q
}

func TestSocB11HistoryChaserMakesASpiritWhenACardLeavesYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	q := socB11Chaser(g, me, 5)
	dead := socB11GraveyardCreature(me, "Dead Bear")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneExile)
	if triggerOnStack(g, q) == nil {
		t.Fatal("no trigger for a card leaving your graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := len(socB11Named(g, "Spirit")); got != 1 {
		t.Errorf("%d Spirits, want 1", got)
	}
}

func TestSocB11HistoryChaserPlusOneDiscardsDrawsTwoAndMills(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	q := socB11Chaser(g, me, 5)
	pitch := game.NewCard("Pitch", me.ID)
	me.Hand.PushTop(pitch)
	seedLibrary(me, "A", "B", "C", "D")
	hand, library, graveyard := handCount(g, me), me.Library.Size(), me.Graveyard.Size()

	b16Activate(t, g, me.ID, q, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no optional discard prompt")
	}
	answerDiscard(t, g, me.ID, pitch.InstanceID)
	passPriorityAroundTable(t, g)
	if got := handCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d (discard one, draw two)", got, hand+1)
	}
	if got := me.Library.Size(); got != library-3 {
		t.Errorf("library %d, want %d (two drawn, one milled)", got, library-3)
	}
	if got := me.Graveyard.Size(); got != graveyard+2 {
		t.Errorf("graveyard %d, want %d (the discard and the milled card)", got, graveyard+2)
	}
	if got := loyaltyOf(g, q); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestSocB11HistoryChaserPlusOneDecliningDrawsAndMillsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	q := socB11Chaser(g, me, 5)
	me.Hand.PushTop(game.NewCard("Keep", me.ID))
	seedLibrary(me, "A", "B", "C", "D")
	hand, library := handCount(g, me), me.Library.Size()
	b16Activate(t, g, me.ID, q, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if handCount(g, me) != hand || me.Library.Size() != library {
		t.Errorf("hand %d library %d after declining, want %d and %d", handCount(g, me), me.Library.Size(), hand, library)
	}
}

func TestSocB11HistoryChaserMinusFourGivesYourSpiritsDoubleStrikeAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	q := socB11Chaser(g, me, 5)
	mine := b12Push(g, me.ID, "My Spirit", "Creature — Spirit", "", 3, 2)
	theirs := b12Push(g, opp.ID, "Their Spirit", "Creature — Spirit", "", 3, 2)
	bear := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	b16Activate(t, g, me.ID, q, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"double strike", "vigilance"} {
		if !hasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my Spirit lacks %s", kw)
		}
		if hasEffectiveKeyword(t, g, theirs, kw) || hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a creature that is not my Spirit got %s", kw)
		}
	}
	if got := loyaltyOf(g, q); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
	// Until end of turn.
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if hasEffectiveKeyword(t, g, mine, "double strike") {
		t.Error("double strike outlived the turn")
	}
}

func TestSocB11HistoryChaserMinusFourNeedsFourLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	q := socB11Chaser(g, me, 3)
	if err := g.ActivateCatalogAbility(me.ID, q, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("−4 activated with 3 loyalty")
	}
}
