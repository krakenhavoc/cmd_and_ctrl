package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const chromeMoxOracle = "ec3d4466-547c-4e02-b1b5-a156ec4637e9"

// castChromeMox casts the Mox and settles it and its imprint trigger,
// stopping on the pick-from-hand prompt.
func castChromeMox(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	mox := effectstest.CastCatalogSpell(t, g, "Chrome Mox", "Artifact", chromeMoxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return mox
}

// TestChromeMoxImprintsAndTapsForThatColor is the whole card: the
// exiled card's colour is what the Mox produces.
func TestChromeMoxImprintsAndTapsForThatColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := effectstest.PushChromeMoxHandCard(me, "A Red Instant", "Instant", []string{"R"})
	mox := castChromeMox(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Chrome Mox queued no imprint prompt")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("the imprint is a \"you may\": floor = %d, want 0", pick.ChooseMin)
	}
	if !effectstest.HasID(pick.ChooseCards, bolt) {
		t.Errorf("the red instant should be a legal imprint: %+v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the imprinted card is exiled")
	}
	if me.Hand.Contains(bolt) {
		t.Error("the imprinted card left the hand")
	}

	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Chrome Mox: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "R" {
		t.Errorf("pool = %+v, want one {R} — the imprinted card's colour", me.ManaPool)
	}
}

// TestChromeMoxWithNoImprintTapsForNothing is the printed downside,
// and the reason this card is not shipped with a five-colour pipe:
// a Mox that imprinted nothing produces nothing.
func TestChromeMoxWithNoImprintTapsForNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushChromeMoxHandCard(me, "A Red Instant", "Instant", []string{"R"})
	mox := castChromeMox(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Chrome Mox queued no imprint prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the imprint: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Chrome Mox: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %+v, want empty — nothing was imprinted", me.ManaPool)
	}
}

// TestChromeMoxOffersOnlyNonartifactNonlandCards: the printed filter
// is real, and it is the only thing keeping a Mox from eating a land.
func TestChromeMoxOffersOnlyNonartifactNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushChromeMoxHandCard(me, "A Forest", "Basic Land — Forest", nil)
	rock := effectstest.PushChromeMoxHandCard(me, "An Artifact", "Artifact", nil)
	creature := effectstest.PushChromeMoxHandCard(me, "A Green Bear", "Creature — Bear", []string{"G"})
	castChromeMox(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Chrome Mox queued no imprint prompt")
	}
	if effectstest.HasID(pick.ChooseCards, forest) {
		t.Error("a land is not a legal imprint")
	}
	if effectstest.HasID(pick.ChooseCards, rock) {
		t.Error("an artifact is not a legal imprint")
	}
	if !effectstest.HasID(pick.ChooseCards, creature) {
		t.Error("a creature card is a legal imprint")
	}
}

// TestChromeMoxImprintUsesTheStack pins the #578 distinction: the
// imprint is a printed "when this enters" TRIGGER, so it goes on the
// stack and nothing is exiled until it resolves.
func TestChromeMoxImprintUsesTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	handBefore := me.Hand.Size()
	mox := effectstest.CastCatalogSpell(t, g, "Chrome Mox", "Artifact", chromeMoxOracle, nil)
	effectstest.PushChromeMoxHandCard(me, "A Red Instant", "Instant", []string{"R"})

	// One pass resolves the Mox itself; the trigger is harvested onto
	// the stack behind it and has not asked anything yet.
	for i := 0; i < 8 && effectstest.TriggerOnStack(g, mox) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.TriggerOnStack(g, mox) == nil {
		t.Fatal("the imprint must go on the stack, not happen as the Mox enters")
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("nothing is asked until the trigger resolves")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand = %d, want the Mox gone and the instant added", me.Hand.Size())
	}
}
