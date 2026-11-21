package testss82v

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sasayaHand gives the seat `lands` Forests and `others` Bears.
func sasayaHand(p *game.Player, lands, others int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < lands; i++ {
		ids = append(ids, effectstest.CcHandCard(p, "Forest", "Basic Land — Forest", ""))
	}
	for i := 0; i < others; i++ {
		ids = append(ids, effectstest.CcHandCard(p, "Bear", "Creature — Bear", "{1}{G}"))
	}
	return ids
}

// handKnownTo reports whether every card in p's hand is known to seat.
func handKnownTo(g *game.Game, p, seat *game.Player) bool {
	all := true
	g.ReadSnapshot(func() {
		for i := range p.Hand.Cards {
			if !p.Hand.Cards[i].IsKnownTo(seat.ID) {
				all = false
			}
		}
	})
	return all
}

func revealEventCount(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventRevealCards {
				n++
			}
		}
	})
	return n
}

// Seven land cards in hand flip Sasaya; the whole hand is shown to the
// table and stays in the hand.
func TestSasayaFlipsWithSevenLandsAndRevealsTheWholeHand(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, effectstest.SasayaRow(), me)
	effectstest.EmptyHand(g, me)
	hand := sasayaHand(me, 7, 2)
	if handKnownTo(g, me, opp) {
		t.Fatal("the opponent knew the hand before the reveal")
	}

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	effectstest.WantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)
	if me.Hand.Size() != len(hand) {
		t.Errorf("hand = %d, want %d: revealing moves nothing (CR 701.20b)", me.Hand.Size(), len(hand))
	}
	for _, seat := range g.Seats {
		if !handKnownTo(g, me, seat) {
			t.Errorf("seat %s does not know the whole revealed hand", seat.Name)
		}
	}
	if got := revealEventCount(g); got != len(hand) {
		t.Errorf("%d reveal events, want one per card (%d)", got, len(hand))
	}
}

// Six lands is a reveal and no flip: the hand is shown regardless, and
// the count is taken as the ability resolves.
func TestSasayaRevealsButDoesNotFlipBelowSevenLands(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, effectstest.SasayaRow(), me)
	effectstest.EmptyHand(g, me)
	sasayaHand(me, 6, 3)

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	effectstest.WantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
	if !handKnownTo(g, me, opp) {
		t.Error("the hand was not revealed to the opponent")
	}
	// It can be activated again: nothing was spent but the reveal.
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	effectstest.WantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
}

// CR 601.2h: the reveal is the cost, so the opponent sees the hand
// while the ability is still on the stack, and a land put into the
// hand in response counts when the ability resolves (CR 608.2c).
func TestSasayaRevealsAtActivationNotResolution(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, effectstest.SasayaRow(), me)
	effectstest.EmptyHand(g, me)
	sasayaHand(me, 6, 0)

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !handKnownTo(g, me, opp) {
		t.Fatal("the hand was not shown by the time the ability was on the stack")
	}
	if got := revealEventCount(g); got != 6 {
		t.Errorf("%d reveal events at activation, want 6", got)
	}
	effectstest.WantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")

	effectstest.CcHandCard(me, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)
}

// An empty hand can still pay the cost (CR 118.3); there is nothing to
// show and nothing to flip for.
func TestSasayaActivatesWithAnEmptyHand(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, effectstest.SasayaRow(), me)
	effectstest.EmptyHand(g, me)

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	effectstest.WantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
	if got := revealEventCount(g); got != 0 {
		t.Errorf("%d reveal events for an empty hand, want 0", got)
	}
}

// Restoring a snapshot taken before the activation takes the reveal
// back: the opponent does not know the hand, and doing it again flips.
func TestSasayaRevealSurvivesUndo(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, effectstest.SasayaRow(), me)
	effectstest.EmptyHand(g, me)
	sasayaHand(me, 7, 0)
	snapshot := g.Clone()

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	effectstest.WantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)

	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.WantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
	if handKnownTo(g, me, opp) {
		t.Error("the opponent still knows the hand after the undo")
	}
	if revealEventCount(g) != 0 {
		t.Error("the undone reveal is still in the event log")
	}
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	effectstest.WantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)
}

// Each OTHER land with the tapped land's name adds one more mana of a
// type it produced.
func TestSasayasEssenceAddsOneManaPerOtherLandWithTheSameName(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	effectstest.FlippedSasaya(t, g, me)
	forests := []uuid.UUID{
		effectstest.PushNamedLand(g, me, "Forest", "Basic Land — Forest"),
		effectstest.PushNamedLand(g, me, "Forest", "Basic Land — Forest"),
		effectstest.PushNamedLand(g, me, "Forest", "Basic Land — Forest"),
	}
	effectstest.PushNamedLand(g, me, "Island", "Basic Land — Island")

	effectstest.TapForMana(t, g, me.ID, forests[0])
	if got := effectstest.SortedPool(me); !slices.Equal(got, []string{"G", "G", "G"}) {
		t.Errorf("pool = %v, want {G}{G}{G}: the land's own mana and one per other Forest", got)
	}
}

// A land with no namesake adds nothing extra, and a namesake the
// opponent controls does not count.
func TestSasayasEssenceIgnoresUniqueLandsAndOpposingNamesakes(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	effectstest.FlippedSasaya(t, g, me)
	island := effectstest.PushNamedLand(g, me, "Island", "Basic Land — Island")
	effectstest.PushNamedLand(g, opp, "Island", "Basic Land — Island")
	effectstest.PushNamedLand(g, me, "Forest", "Basic Land — Forest")

	effectstest.TapForMana(t, g, me.ID, island)
	if got := effectstest.SortedPool(me); !slices.Equal(got, []string{"U"}) {
		t.Errorf("pool = %v, want only the Island's {U}", got)
	}
}

// Unflipped, Sasaya adds nothing.
func TestSasayaWithoutTheFlipAddsNoExtraMana(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	effectstest.FlipEntered(t, g, effectstest.SasayaRow(), me)
	f1 := effectstest.PushNamedLand(g, me, "Forest", "Basic Land — Forest")
	effectstest.PushNamedLand(g, me, "Forest", "Basic Land — Forest")

	effectstest.TapForMana(t, g, me.ID, f1)
	if got := effectstest.SortedPool(me); !slices.Equal(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
}
