package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teferi_who_slows_the_sunset_test.go — #1315's proof card. The −7 is
// the point: an emblem widening two turn-based actions
// (untap.go / draw_step.go) rather than triggering, which is the seam
// #1315 opened. The +1 and −2 are exercised too, since the card
// registers all three abilities, but the emblem is what this issue is
// about.

const teferiWhoSlowsTheSunsetOracle = "98c389ed-b960-42e2-9c76-a062f77c9a78"

// TestTeferiWhoSlowsTheSunsetMinusSevenCreatesTheEmblem is CR 114.5:
// activating the −7 pays seven loyalty and puts the printed emblem in
// its controller's command zone.
func TestTeferiWhoSlowsTheSunsetMinusSevenCreatesTheEmblem(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", teferiWhoSlowsTheSunsetOracle, 7)

	b16Activate(t, g, me.ID, teferi, 2, game.ActivateAbilityParams{})

	if got := loyaltyCount(g, teferi); got != 0 {
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
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", teferiWhoSlowsTheSunsetOracle, 7)
	b16Activate(t, g, me.ID, teferi, 2, game.ActivateAbilityParams{})

	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	tapForTest(t, g, rock)

	advanceToUpkeepOf(t, g, 1)
	if b16Tapped(t, g, rock) {
		t.Error("the emblem untaps all permanents I control during each opponent's untap step (seat 1)")
	}

	tapForTest(t, g, rock)
	advanceToUpkeepOf(t, g, 2)
	if b16Tapped(t, g, rock) {
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
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", teferiWhoSlowsTheSunsetOracle, 7)
	b16Activate(t, g, me.ID, teferi, 2, game.ActivateAbilityParams{})

	meHandBefore := me.Hand.Size()
	oppHandBefore := opp.Hand.Size()

	advanceToDrawStepOf(t, g, 1)

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
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", teferiWhoSlowsTheSunsetOracle, 4)
	myArtifact := b12Permanent(g, me.ID, "My Signet", "Artifact")
	myLand := b12Permanent(g, me.ID, "My Forest", "Land")
	theirCreature := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tapForTest(t, g, myArtifact, myLand)
	lifeBefore := me.Life

	b16Activate(t, g, me.ID, teferi, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: myArtifact},
			{Kind: game.TargetCard, ID: theirCreature},
			{Kind: game.TargetCard, ID: myLand},
		},
	})

	if b16Tapped(t, g, myArtifact) {
		t.Error("my artifact should have untapped")
	}
	if b16Tapped(t, g, myLand) {
		t.Error("my land should have untapped")
	}
	if !b16Tapped(t, g, theirCreature) {
		t.Error("their creature should have tapped")
	}
	if me.Life != lifeBefore+2 {
		t.Errorf("life %d -> %d, want +2", lifeBefore, me.Life)
	}
	if got := loyaltyCount(g, teferi); got != 5 {
		t.Errorf("loyalty after +1 on a 4-loyalty Teferi: got %d, want 5", got)
	}
}

// TestTeferiWhoSlowsTheSunsetMinusTwoPutsOneInHandAndRestOnBottom is
// the −2: it looks at three cards, offers exactly one for the hand,
// and the two declined cards are conserved (library size unchanged),
// not lost or duplicated.
func TestTeferiWhoSlowsTheSunsetMinusTwoPutsOneInHandAndRestOnBottom(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", teferiWhoSlowsTheSunsetOracle, 4)
	seeded := seedLibrary(me, "Look1", "Look2", "Look3", "Filler1", "Filler2")
	libraryBefore := me.Library.Size()
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -2: %v", err)
	}
	passPriorityAroundTable(t, g)
	choice := chooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no choose-cards prompt for the -2's pick")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{seeded[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size %d -> %d, want +1", handBefore, got)
	}
	if got := me.Library.Size(); got != libraryBefore-1 {
		t.Errorf("library size %d -> %d, want -1 (one card left for the hand)", libraryBefore, got)
	}
	if !me.Hand.Contains(seeded[0]) {
		t.Error("the chosen card did not reach the hand")
	}
	for _, id := range seeded[1:3] {
		if !me.Library.Contains(id) {
			t.Errorf("declined card %s is missing from the library — it must go to the bottom, not vanish", id)
		}
	}
}
