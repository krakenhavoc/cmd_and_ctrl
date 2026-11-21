package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b11_quintorius_test.go — the two Quintorius cards of Secrets of
// Strixhaven Commander batch 11 that make a Spirit when cards leave a
// graveyard (#4834).

// --- Quintorius, Field Historian ---

func TestSocB11FieldHistorianPumpsOnlyYourSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := effectstest.B12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", effectstest.SocB11FieldHistorianOracle, 2, 4)
	mySpirit := effectstest.B12Push(g, me.ID, "My Spirit", "Creature — Spirit", "", 3, 2)
	theirSpirit := effectstest.B12Push(g, opp.ID, "Their Spirit", "Creature — Spirit", "", 3, 2)
	myBear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{historian, "Quintorius", 2}, {mySpirit, "my Spirit", 4}, {theirSpirit, "their Spirit", 3}, {myBear, "my Bear", 2}} {
		if got := effectstest.EffectivePower(t, g, tc.id); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
	if got := effectstest.EffectiveToughness(t, g, mySpirit); got != 2 {
		t.Errorf("my Spirit's toughness is %d, want 2 (+1/+0 only)", got)
	}
}

func TestSocB11FieldHistorianMakesASpiritWhenACardLeavesYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := effectstest.B12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", effectstest.SocB11FieldHistorianOracle, 2, 4)
	mine := effectstest.SocB11GraveyardCreature(me, "Dead Bear")
	theirs := effectstest.SocB11GraveyardCreature(opp, "Their Dead Bear")

	// A card leaving ANOTHER player's graveyard makes nothing.
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneHand)
	if effectstest.TriggerOnStack(g, historian) != nil {
		t.Fatal("triggered for a card leaving an opponent's graveyard")
	}
	// Any card of yours, to any zone, makes a 3/2 red and white Spirit.
	socB11Move(t, g, me.ID, mine, game.ZoneGraveyard, game.ZoneHand)
	if effectstest.TriggerOnStack(g, historian) == nil {
		t.Fatal("no trigger for a card leaving your graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	if got := effectstest.EffectivePower(t, g, spirits[0]); got != 4 {
		t.Errorf("the new Spirit has power %d, want 4", got)
	}
}

func TestSocB11FieldHistorianMakesOneSpiritForCardsLeavingAsOneEvent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", effectstest.SocB11FieldHistorianOracle, 2, 4)
	a := effectstest.SocB11GraveyardCreature(me, "Dead One")
	b := effectstest.SocB11GraveyardCreature(me, "Dead Two")
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{a, b}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(socB11Named(g, "Spirit")); got != 1 {
		t.Errorf("%d Spirits for two cards leaving as one event, want 1", got)
	}
}

func TestSocB11FieldHistorianMakesOneSpiritPerSeparateEvent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Quintorius, Field Historian", "Legendary Creature — Elephant Cleric", effectstest.SocB11FieldHistorianOracle, 2, 4)
	a := effectstest.SocB11GraveyardCreature(me, "Dead One")
	b := effectstest.SocB11GraveyardCreature(me, "Dead Two")
	socB11Move(t, g, me.ID, a, game.ZoneGraveyard, game.ZoneHand)
	effectstest.PassPriorityAroundTable(t, g)
	socB11Move(t, g, me.ID, b, game.ZoneGraveyard, game.ZoneHand)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(socB11Named(g, "Spirit")); got != 2 {
		t.Errorf("%d Spirits for two separate events, want 2", got)
	}
}

func TestSocB11HistoryChaserMakesASpiritWhenACardLeavesYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	q := effectstest.SocB11Chaser(g, me, 5)
	dead := effectstest.SocB11GraveyardCreature(me, "Dead Bear")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneExile)
	if effectstest.TriggerOnStack(g, q) == nil {
		t.Fatal("no trigger for a card leaving your graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(socB11Named(g, "Spirit")); got != 1 {
		t.Errorf("%d Spirits, want 1", got)
	}
}

func TestSocB11HistoryChaserPlusOneDiscardsDrawsTwoAndMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	q := effectstest.SocB11Chaser(g, me, 5)
	pitch := game.NewCard("Pitch", me.ID)
	me.Hand.PushTop(pitch)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand, library, graveyard := effectstest.HandCount(g, me), me.Library.Size(), me.Graveyard.Size()

	effectstest.B16Activate(t, g, me.ID, q, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no optional discard prompt")
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitch.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d (discard one, draw two)", got, hand+1)
	}
	if got := me.Library.Size(); got != library-3 {
		t.Errorf("library %d, want %d (two drawn, one milled)", got, library-3)
	}
	if got := me.Graveyard.Size(); got != graveyard+2 {
		t.Errorf("graveyard %d, want %d (the discard and the milled card)", got, graveyard+2)
	}
	if got := effectstest.LoyaltyOf(g, q); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestSocB11HistoryChaserPlusOneDecliningDrawsAndMillsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	q := effectstest.SocB11Chaser(g, me, 5)
	me.Hand.PushTop(game.NewCard("Keep", me.ID))
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand, library := effectstest.HandCount(g, me), me.Library.Size()
	effectstest.B16Activate(t, g, me.ID, q, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandCount(g, me) != hand || me.Library.Size() != library {
		t.Errorf("hand %d library %d after declining, want %d and %d", effectstest.HandCount(g, me), me.Library.Size(), hand, library)
	}
}

func TestSocB11HistoryChaserMinusFourGivesYourSpiritsDoubleStrikeAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	q := effectstest.SocB11Chaser(g, me, 5)
	mine := effectstest.B12Push(g, me.ID, "My Spirit", "Creature — Spirit", "", 3, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Spirit", "Creature — Spirit", "", 3, 2)
	bear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.B16Activate(t, g, me.ID, q, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"double strike", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my Spirit lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) || effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a creature that is not my Spirit got %s", kw)
		}
	}
	if got := effectstest.LoyaltyOf(g, q); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
	// Until end of turn.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if effectstest.HasEffectiveKeyword(t, g, mine, "double strike") {
		t.Error("double strike outlived the turn")
	}
}

func TestSocB11HistoryChaserMinusFourNeedsFourLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	q := effectstest.SocB11Chaser(g, me, 3)
	if err := g.ActivateCatalogAbility(me.ID, q, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("−4 activated with 3 loyalty")
	}
}
