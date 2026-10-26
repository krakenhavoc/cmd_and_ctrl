package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_10c_draw_done_restore_test.go — #4253 slice 10c: the `done` a
// draw-instead body continues with is data (game/draw_done.go). A body that
// stops to ask parks it on the game and puts its ID on the prompt, so the
// table is a restore point while the prompt is open, and the restored answer
// pays what the replaced draw still owed: the draws after it, and the
// instruction's "then".

// parkedDrawsOf is how many draw tails the snapshot of g writes.
func parkedDrawsOf(g *game.Game) int { return len(g.CaptureSnapshot().ParkedDraws) }

// handDiscardFor is the player's open discard from hand (a choose_cards
// prompt over cards in their hand), as opposed to a pick over a look or a
// graveyard. Nil when none is open.
func handDiscardFor(g *game.Game, p *game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceChooseCards || c.Chooser != p.ID || len(c.ChooseCards) == 0 {
			continue
		}
		inHand := true
		for _, id := range c.ChooseCards {
			if !zoneHas(p.Hand, id) {
				inHand = false
			}
		}
		if inHand {
			return c
		}
	}
	return nil
}

// Underrealm Lich replaces both draws of "draw two cards, then discard a
// card". Each look's pick is a restore point; the second is asked only once
// the first is answered on the restored table, and the discard the
// instruction owes opens after the second.
func TestUnderrealmLichPickIsARestorePointAndOwesTheRestOfTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	bfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", underrealmLichOracle)
	handBefore, gyBefore := p.Hand.Size(), p.Graveyard.Size()
	startLoot(g, p, 2, 1)

	first := chooseCardsChoiceFor(g, p.ID)
	if first == nil {
		t.Fatal("the Lich did not ask")
	}
	if parkedDrawsOf(g) != 1 {
		t.Fatalf("%d draw tails written with the first pick open, want 1", parkedDrawsOf(g))
	}
	keep := first.ChooseCards[1]

	g = restoreAtKeyedPrompt(t, g, "choose-cards/underrealm-lich-take")
	p = g.Seats[1]
	first = chooseCardsChoiceFor(g, p.ID)
	if first == nil || len(first.ChooseCards) != 3 {
		t.Fatalf("restored prompt = %+v, want a pick over three", first)
	}
	if handDiscardFor(g, p) != nil {
		t.Fatal("the restored table owed the discard before the draws were done")
	}
	answerChooseCards(t, g, p.ID, keep)
	if !zoneHas(p.Hand, keep) {
		t.Fatal("the restored answer did not put the chosen card in hand")
	}
	second := chooseCardsChoiceFor(g, p.ID)
	if second == nil {
		t.Fatal("the restored answer did not owe the second draw")
	}
	if handDiscardFor(g, p) != nil {
		t.Fatal("the discard opened before the second draw")
	}

	g = restoreAtKeyedPrompt(t, g, "choose-cards/underrealm-lich-take")
	p = g.Seats[1]
	second = chooseCardsChoiceFor(g, p.ID)
	if second == nil {
		t.Fatal("the second pick did not survive a restore")
	}
	answerChooseCards(t, g, p.ID, second.ChooseCards[0])
	if handDiscardFor(g, p) == nil {
		t.Fatal("the discard the instruction owed did not open after the second restored answer")
	}
	if p.Hand.Size() != handBefore+2 || p.Graveyard.Size() != gyBefore+4 {
		t.Errorf("hand %d (want %d), graveyard %d (want %d)", p.Hand.Size(), handBefore+2, p.Graveyard.Size(), gyBefore+4)
	}
	if parkedDrawsOf(g) != 0 {
		t.Errorf("%d draw tails left parked after both were paid", parkedDrawsOf(g))
	}
}

// Forbidden Crypt's pick is a restore point: the restored answer returns the
// card, and the next draw of a two-card draw asks again.
func TestForbiddenCryptPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	bfCardWithOracle(g, p, "Forbidden Crypt", "Enchantment", forbiddenCryptOracle)
	a := gyCardWithOracle(p, "A", "Creature — Test", "")
	b := gyCardWithOracle(p, "B", "Creature — Test", "")
	libBefore := p.Library.Size()
	drawN(g, p, 2)
	if chooseCardsChoiceFor(g, p.ID) == nil {
		t.Fatal("Forbidden Crypt did not ask")
	}

	g = restoreAtKeyedPrompt(t, g, "choose-cards/draw-instead-graveyard-return")
	p = g.Seats[1]
	pick := chooseCardsChoiceFor(g, p.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("restored prompt = %+v, want a pick over the two graveyard cards", pick)
	}
	answerChooseCards(t, g, p.ID, b)
	if !zoneHas(p.Hand, b) {
		t.Fatal("the restored answer did not return the card")
	}
	next := chooseCardsChoiceFor(g, p.ID)
	if next == nil || len(next.ChooseCards) != 1 || next.ChooseCards[0] != a {
		t.Fatalf("the second draw's prompt = %+v, want a pick over the one card left", next)
	}
	if p.Library.Size() != libBefore {
		t.Errorf("library %d, want %d: nothing is drawn", p.Library.Size(), libBefore)
	}
}

// The Necrobloom's dredge 2: after the "yes", the land pick is a restore
// point, and the restored answer mills two, returns the land and pays the
// loot's discard.
func TestNecrobloomDredgePickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	bfCardWithOracle(g, p, "The Necrobloom", "Legendary Creature — Plant", necrobloomOracle)
	forest := gyCardWithOracle(p, "Forest", "Basic Land — Forest", "")
	bear := gyCardWithOracle(p, "Bear", "Creature — Bear", "")
	libBefore := p.Library.Size()
	startLoot(g, p, 1, 1)
	answerOptional(t, g, p, true)
	if chooseCardsChoiceFor(g, p.ID) == nil {
		t.Fatal("the Necrobloom did not ask which land")
	}

	g = restoreAtKeyedPrompt(t, g, "choose-cards/necrobloom-dredge-pick")
	p = g.Seats[1]
	if handDiscardFor(g, p) != nil {
		t.Fatal("the restored table owed the discard before the dredge")
	}
	answerChooseCards(t, g, p.ID, forest)
	if !zoneHas(p.Hand, forest) || !zoneHas(p.Graveyard, bear) {
		t.Error("the land returns and the creature stays")
	}
	if p.Library.Size() != libBefore-2 {
		t.Errorf("library %d, want %d (milled two)", p.Library.Size(), libBefore-2)
	}
	if handDiscardFor(g, p) == nil {
		t.Error("the discard the loot owed did not open after the restored dredge")
	}
}

// A parked draw's tail round-trips through the snapshot with everything
// finishDrawTailLocked reads.
func TestParkedDrawTailsRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	bfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", underrealmLichOracle)
	drawN(g, p, 3)
	snap := g.CaptureSnapshot()
	if len(snap.ParkedDraws) != 1 || snap.ParkedDraws[0].Tail != 2 || snap.ParkedDraws[0].Actor != p.ID {
		t.Fatalf("parked draws = %+v, want one owing two more draws for the drawer", snap.ParkedDraws)
	}
	g = restoreRoundTrip(t, g, true)
	p = g.Seats[1]
	for i := 0; i < 3; i++ {
		pick := chooseCardsChoiceFor(g, p.ID)
		if pick == nil {
			t.Fatalf("pick %d did not open", i)
		}
		answerChooseCards(t, g, p.ID, pick.ChooseCards[0])
	}
	if chooseCardsChoiceFor(g, p.ID) != nil {
		t.Error("a fourth pick opened")
	}
	_ = uuid.Nil
}
