package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sasaya_orochi_ascendant_test.go — #3250: "Reveal your hand" as an
// activated ability's cost (game.AbilityCost.RevealHand) and the flip
// card it pays for, down the real import road.

func sasayaRow() cards.Card {
	return flipRow(sasayaOracleID, "{1}{G}{G}", []string{"G"}, nil,
		cards.CardFace{Name: "Sasaya, Orochi Ascendant", TypeLine: "Legendary Creature — Snake Monk",
			OracleText: "Reveal your hand: If you have seven or more land cards in your hand, flip Sasaya.",
			Power:      "2", Toughness: "3"},
		cards.CardFace{Name: "Sasaya's Essence", TypeLine: "Legendary Enchantment",
			OracleText: "Whenever a land you control is tapped for mana, add an additional one mana of any type that land produced for each other land you control with the same name as it."})
}

// emptyHand clears the seat's hand.
func emptyHand(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
}

// sasayaHand gives the seat `lands` Forests and `others` Bears.
func sasayaHand(p *game.Player, lands, others int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < lands; i++ {
		ids = append(ids, ccHandCard(p, "Forest", "Basic Land — Forest", ""))
	}
	for i := 0; i < others; i++ {
		ids = append(ids, ccHandCard(p, "Bear", "Creature — Bear", "{1}{G}"))
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
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, sasayaRow(), me)
	emptyHand(g, me)
	hand := sasayaHand(me, 7, 2)
	if handKnownTo(g, me, opp) {
		t.Fatal("the opponent knew the hand before the reveal")
	}

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	wantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)
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
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, sasayaRow(), me)
	emptyHand(g, me)
	sasayaHand(me, 6, 3)

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	wantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
	if !handKnownTo(g, me, opp) {
		t.Error("the hand was not revealed to the opponent")
	}
	// It can be activated again: nothing was spent but the reveal.
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	wantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
}

// CR 601.2h: the reveal is the cost, so the opponent sees the hand
// while the ability is still on the stack, and a land put into the
// hand in response counts when the ability resolves (CR 608.2c).
func TestSasayaRevealsAtActivationNotResolution(t *testing.T) {
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, sasayaRow(), me)
	emptyHand(g, me)
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
	wantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")

	ccHandCard(me, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	wantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)
}

// An empty hand can still pay the cost (CR 118.3); there is nothing to
// show and nothing to flip for.
func TestSasayaActivatesWithAnEmptyHand(t *testing.T) {
	g, me, _ := flipMain(t)
	id := flipEntered(t, g, sasayaRow(), me)
	emptyHand(g, me)

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	wantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
	if got := revealEventCount(g); got != 0 {
		t.Errorf("%d reveal events for an empty hand, want 0", got)
	}
}

// Restoring a snapshot taken before the activation takes the reveal
// back: the opponent does not know the hand, and doing it again flips.
func TestSasayaRevealSurvivesUndo(t *testing.T) {
	g, me, _ := flipMain(t)
	id := flipEntered(t, g, sasayaRow(), me)
	emptyHand(g, me)
	sasayaHand(me, 7, 0)
	snapshot := g.Clone()

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	wantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)

	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wantUnflipped(t, g, id, "Sasaya, Orochi Ascendant")
	if handKnownTo(g, me, opp) {
		t.Error("the opponent still knows the hand after the undo")
	}
	if revealEventCount(g) != 0 {
		t.Error("the undone reveal is still in the event log")
	}
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	wantFlipped(t, g, id, "Sasaya's Essence", []string{"G"}, 3)
}

// --- Sasaya's Essence -------------------------------------------------

func flippedSasaya(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := flipEntered(t, g, sasayaRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	return id
}

func pushNamedLand(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	return b31Push(g, p.ID, name, typeLine, "", "", 0, 0)
}

// Each OTHER land with the tapped land's name adds one more mana of a
// type it produced.
func TestSasayasEssenceAddsOneManaPerOtherLandWithTheSameName(t *testing.T) {
	g, me, _ := flipMain(t)
	flippedSasaya(t, g, me)
	forests := []uuid.UUID{
		pushNamedLand(g, me, "Forest", "Basic Land — Forest"),
		pushNamedLand(g, me, "Forest", "Basic Land — Forest"),
		pushNamedLand(g, me, "Forest", "Basic Land — Forest"),
	}
	pushNamedLand(g, me, "Island", "Basic Land — Island")

	tapForMana(t, g, me.ID, forests[0])
	if got := sortedPool(me); !slices.Equal(got, []string{"G", "G", "G"}) {
		t.Errorf("pool = %v, want {G}{G}{G}: the land's own mana and one per other Forest", got)
	}
}

// A land with no namesake adds nothing extra, and a namesake the
// opponent controls does not count.
func TestSasayasEssenceIgnoresUniqueLandsAndOpposingNamesakes(t *testing.T) {
	g, me, opp := flipMain(t)
	flippedSasaya(t, g, me)
	island := pushNamedLand(g, me, "Island", "Basic Land — Island")
	pushNamedLand(g, opp, "Island", "Basic Land — Island")
	pushNamedLand(g, me, "Forest", "Basic Land — Forest")

	tapForMana(t, g, me.ID, island)
	if got := sortedPool(me); !slices.Equal(got, []string{"U"}) {
		t.Errorf("pool = %v, want only the Island's {U}", got)
	}
}

// Unflipped, Sasaya adds nothing.
func TestSasayaWithoutTheFlipAddsNoExtraMana(t *testing.T) {
	g, me, _ := flipMain(t)
	flipEntered(t, g, sasayaRow(), me)
	f1 := pushNamedLand(g, me, "Forest", "Basic Land — Forest")
	pushNamedLand(g, me, "Forest", "Basic Land — Forest")

	tapForMana(t, g, me.ID, f1)
	if got := sortedPool(me); !slices.Equal(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
}
