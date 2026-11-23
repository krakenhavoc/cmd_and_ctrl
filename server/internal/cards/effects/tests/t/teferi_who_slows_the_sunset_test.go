package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestTeferiWhoSlowsTheSunsetMinusSevenCreatesTheEmblem is CR 114.5:
// activating the −7 pays seven loyalty and puts the printed emblem in
// its controller's command zone.
func TestTeferiWhoSlowsTheSunsetMinusSevenCreatesTheEmblem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", effectstest.TeferiWhoSlowsTheSunsetOracle, 7)

	effectstest.B16Activate(t, g, me.ID, teferi, 2, game.ActivateAbilityParams{})

	if got := effectstest.LoyaltyCount(g, teferi); got != 0 {
		t.Errorf("loyalty after -7 on a 7-loyalty Teferi: got %d, want 0", got)
	}
	emblems := emblemsOfPlayer(g, me.ID)
	if len(emblems) != 1 {
		t.Fatalf("emblems = %d, want 1", len(emblems))
	}
	if emblems[0].Label != "Teferi, Who Slows the Sunset emblem" {
		t.Errorf("emblem label = %q", emblems[0].Label)
	}
}

// emblemsOfPlayer reads a player's emblems under a read lock, mirroring
// emblemsOf in emblem_test.go (a different package, so not reusable).
func emblemsOfPlayer(g *game.Game, playerID uuid.UUID) []game.EmblemView {
	var out []game.EmblemView
	g.ReadSnapshot(func() { out = g.EmblemsForPlayer(playerID) })
	return out
}

// TestTeferiWhoSlowsTheSunsetEmblemUntapsDuringEachOpponentsUntapStep
// is the untap half of the −7 seam: the emblem carries no permanent
// on the battlefield at all, and its UntapStepPermission still widens
// CR 502.3's set on every OTHER seat's untap step — seat 1's AND seat
// 2's, because "each opponent" is every other player at this table.
func TestTeferiWhoSlowsTheSunsetEmblemUntapsDuringEachOpponentsUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", effectstest.TeferiWhoSlowsTheSunsetOracle, 7)
	effectstest.B16Activate(t, g, me.ID, teferi, 2, game.ActivateAbilityParams{})

	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.TapForTest(t, g, rock)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.B16Tapped(t, g, rock) {
		t.Error("the emblem untaps all permanents I control during each opponent's untap step (seat 1)")
	}

	effectstest.TapForTest(t, g, rock)
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	if effectstest.B16Tapped(t, g, rock) {
		t.Error("the emblem untaps all permanents I control during each opponent's untap step (seat 2 too)")
	}
}

// TestTeferiWhoSlowsTheSunsetEmblemDrawsDuringEachOpponentsDrawStep is
// the draw half: the emblem's controller draws an extra card as PART
// of an opponent's draw-step turn-based action — no trigger, no stack,
// verified by checking the hand grew by the time the cursor is
// sitting IN that opponent's draw step (a trigger would still be
// waiting for a priority pass to resolve).
func TestTeferiWhoSlowsTheSunsetEmblemDrawsDuringEachOpponentsDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", effectstest.TeferiWhoSlowsTheSunsetOracle, 7)
	effectstest.B16Activate(t, g, me.ID, teferi, 2, game.ActivateAbilityParams{})

	meHandBefore := me.Hand.Size()
	oppHandBefore := opp.Hand.Size()

	effectstest.AdvanceToDrawStepOf(t, g, 1)

	if got := me.Hand.Size(); got != meHandBefore+1 {
		t.Errorf("emblem owner's hand %d -> %d, want +1 from the opponent's draw step", meHandBefore, got)
	}
	if got := opp.Hand.Size(); got != oppHandBefore+1 {
		t.Errorf("active seat's own hand %d -> %d, want +1 from its own CR 504.1 draw", oppHandBefore, got)
	}
}

// TestTeferiWhoSlowsTheSunsetPlusOneUntapsTapsAndGainsLife is the +1
// end to end: an artifact and a land I control come back untapped, a
// creature I don't control gets tapped, and I gain 2 life regardless.
func TestTeferiWhoSlowsTheSunsetPlusOneUntapsTapsAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", effectstest.TeferiWhoSlowsTheSunsetOracle, 4)
	myArtifact := effectstest.B12Permanent(g, me.ID, "My Signet", "Artifact")
	myLand := effectstest.B12Permanent(g, me.ID, "My Forest", "Land")
	theirCreature := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, myArtifact, myLand)
	lifeBefore := me.Life

	effectstest.B16Activate(t, g, me.ID, teferi, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: myArtifact},
			{Kind: game.TargetCard, ID: theirCreature},
			{Kind: game.TargetCard, ID: myLand},
		},
	})

	if effectstest.B16Tapped(t, g, myArtifact) {
		t.Error("my artifact should have untapped")
	}
	if effectstest.B16Tapped(t, g, myLand) {
		t.Error("my land should have untapped")
	}
	if !effectstest.B16Tapped(t, g, theirCreature) {
		t.Error("their creature should have tapped")
	}
	if me.Life != lifeBefore+2 {
		t.Errorf("life %d -> %d, want +2", lifeBefore, me.Life)
	}
	if got := effectstest.LoyaltyCount(g, teferi); got != 5 {
		t.Errorf("loyalty after +1 on a 4-loyalty Teferi: got %d, want 5", got)
	}
}

// TestTeferiWhoSlowsTheSunsetMinusTwoPutsOneInHandAndOrdersTheRestOnTheBottom
// is the −2 end to end (ADR 0088, #996): it looks at three cards,
// offers exactly one for the hand, and the two declined cards are
// ORDERED BY THE PLAYER onto the bottom of the library — not a random
// order, and not lost or duplicated. Mirrors
// TestImpulseTakesOneAndOrdersTheRestOnTheBottom's pattern exactly,
// since LookAtTopThenTakeOneRestOnBottom is the same composition.
func TestTeferiWhoSlowsTheSunsetMinusTwoPutsOneInHandAndOrdersTheRestOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", effectstest.TeferiWhoSlowsTheSunsetOracle, 4)
	seeded := effectstest.SeedLibrary(me, "Look1", "Look2", "Look3", "Filler1", "Filler2")
	look1, look2, look3 := seeded[0], seeded[1], seeded[2]
	libraryBefore := me.Library.Size()
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no choose-cards prompt for the -2's pick")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{look1}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(look1) || me.Hand.Size() != handBefore+1 {
		t.Fatal("the chosen card did not reach the hand")
	}

	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("the two declined cards are ordered by the player — no put_in_library prompt")
	}
	if order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 2 {
		t.Fatalf("a bottom placement over the two declined cards: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{look3, look2}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}

	if got := me.Library.Size(); got != libraryBefore-1 {
		t.Errorf("library size %d -> %d, want -1 (one card left for the hand)", libraryBefore, got)
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{look3, look2}) {
		t.Errorf("the bottom two are %v, want [Look3 Look2] top-first, in the order chosen", got)
	}
}
