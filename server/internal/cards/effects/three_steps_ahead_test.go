package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// three_steps_ahead_test.go — the Spree proof card (CR 702.172a, ADR
// 0065's 2026-09-23 amendment, issue #1330). Each bullet on its own,
// then all three together — the point of #764's per-mode targets
// carried one seam further: a mode's own COST composes the same way
// its own TARGET does.

const threeStepsAheadOracle = "282dfeaa-6243-4f92-838a-5cb54fa85184"

func TestThreeStepsAheadCountersASpellAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Their Bolt", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	castModal(t, g, "Three Steps Ahead", "Instant", threeStepsAheadOracle,
		[]int{0}, []game.TargetRef{modeRef(game.TargetCard, bolt, 0, 0)})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("choosing only the counter bullet should have countered the spell")
	}
}

func TestThreeStepsAheadCopiesYourOwnPermanentAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	before := g.Battlefield.Size()

	castModal(t, g, "Three Steps Ahead", "Instant", threeStepsAheadOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, mine, 0, 0)})
	passPriorityAroundTable(t, g)

	if got := g.Battlefield.Size(); got != before+1 {
		t.Fatalf("the token-copy bullet should have added one permanent: %d, want %d", got, before+1)
	}
}

func TestThreeStepsAheadDrawsTwoThenDiscardsOneAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	handBefore := me.Hand.Size()

	castModal(t, g, "Three Steps Ahead", "Instant", threeStepsAheadOracle,
		[]int{2}, nil)
	passPriorityAroundTable(t, g)
	if c := discardChoiceFor(g, me.ID); c == nil {
		t.Fatal("draw two, then discard a card should have queued the caster's own discard choice")
	} else if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("discard exactly one: bounds [%d,%d], want [1,1]", c.ChooseMin, c.ChooseMax)
	}
	discardFromHand(t, g, me.ID)

	// handBefore was taken before the spell itself was pushed to
	// hand and cast, so the cast is a wash; +2 drawn, -1 discarded
	// nets +1 against it.
	if got, want := me.Hand.Size(), handBefore+1; got != want {
		t.Errorf("hand size after draw-two-discard-one = %d, want %d", got, want)
	}
	if me.Graveyard.Size() == 0 {
		t.Error("the discarded card should be in the graveyard")
	}
}

// All three bullets, each with its own target (or none), each
// composing with the others exactly as Kolaghan's Command's do — the
// #764 machinery this card leans on unmodified. This is also the
// card-level half of the cost: three bullets chosen means {1}{U}+{3}
// +{2} joins the printed {U}, which cast_price_test.go's engine-level
// case proves the arithmetic for; this proves the three EFFECTS all
// actually ran.
func TestThreeStepsAheadAllThreeBulletsCompose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Their Bolt", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	bfBefore := g.Battlefield.Size()
	handBefore := me.Hand.Size()

	castModal(t, g, "Three Steps Ahead", "Instant", threeStepsAheadOracle,
		[]int{0, 1, 2},
		[]game.TargetRef{
			modeRef(game.TargetCard, bolt, 0, 0), // occurrence 0 = counter
			modeRef(game.TargetCard, mine, 1, 0), // occurrence 1 = token copy
			// occurrence 2 (draw-two-discard-one) is untargeted.
		})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the counter bullet did not counter the targeted spell")
	}
	if got := g.Battlefield.Size(); got != bfBefore+1 {
		t.Errorf("the token-copy bullet should have added one permanent: %d, want %d", got, bfBefore+1)
	}
	discardFromHand(t, g, me.ID)
	if got, want := me.Hand.Size(), handBefore+1; got != want {
		t.Errorf("hand size after all three bullets = %d, want %d (the cast is a wash; +2 drawn, -1 discarded)", got, want)
	}
}
