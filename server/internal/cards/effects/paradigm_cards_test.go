package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paradigm_cards_test.go — the paradigm cards (#1866) over the engine's
// keyword (game/paradigm.go): Decorum Dissertation end to end, and
// Echocasting Symposium's copy choosing its own targets.

const decorumDissertationOracle = "f4c57454-d36f-4f38-b7ac-76477a516c3e"

// paradigmOfferOn walks to seat's next precombat main and passes until
// the paradigm trigger has resolved and a copy is on offer.
func paradigmOfferOn(t *testing.T, g *game.Game, seat int, name string) game.Card {
	t.Helper()
	b39NextTurnOf(t, g, seat)
	for i := 0; i < 12 && len(cardCopies(g)) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return copyOf(t, g, name)
}

// The spell resolves as Sign in Blood, exiles itself, and from then on
// each of its caster's precombat main phases offers a free copy, which
// picks its own target.
func TestDecorumDissertationDrawsThenOffersAFreeCopyEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	oppHand, oppLife := len(opp.Hand.Cards), opp.Life

	id := castCatalogSpell(t, g, "Decorum Dissertation", "Sorcery — Lesson", decorumDissertationOracle,
		[]game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)

	if got := len(opp.Hand.Cards) - oppHand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
	if opp.Life != oppLife-2 {
		t.Errorf("the target's life %d, want %d", opp.Life, oppLife-2)
	}
	if !inExile(g, id) || me.Graveyard.Contains(id) {
		t.Fatal("Decorum Dissertation was not exiled by its paradigm")
	}

	cp := paradigmOfferOn(t, g, seat, "Decorum Dissertation")
	myHand, myLife := len(me.Hand.Cards), me.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(me.ID)); err != nil {
		t.Fatalf("free cast of the copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - myHand; got != 2 {
		t.Errorf("the copy's target drew %d, want 2", got)
	}
	if me.Life != myLife-2 {
		t.Errorf("the copy's target's life %d, want %d", me.Life, myLife-2)
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the copy resolved", n)
	}
	if !inExile(g, id) {
		t.Error("the original left exile")
	}

	// And again on the next turn: the trigger lasts the rest of the game.
	paradigmOfferOn(t, g, seat, "Decorum Dissertation")
}

// The copy is a new spell with new targets: this turn's token can go to
// a different player than the original's.
func TestEchocastingSymposiumCopyChoosesItsOwnTargets(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	mine := b12Push(g, me.ID, "Big Wurm", "Creature — Wurm", "", 6, 6)

	id := castCatalogSpell(t, g, "Echocasting Symposium", "Sorcery — Lesson", echocastingSymposiumOracle,
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: mine, Slot: 1},
		})
	passPriorityAroundTable(t, g)
	if !inExile(g, id) {
		t.Fatal("Echocasting Symposium was not exiled by its paradigm")
	}

	cp := paradigmOfferOn(t, g, seat, "Echocasting Symposium")
	if err := castCopy(g, me.ID, cp.InstanceID, true,
		game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 1},
	); err != nil {
		t.Fatalf("free cast of the copy: %v", err)
	}
	passPriorityAroundTable(t, g)

	owners := map[string]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Big Wurm" && IsToken(c) {
			switch c.Controller {
			case me.ID:
				owners["me"]++
			case opp.ID:
				owners["opp"]++
			}
		}
	}
	if owners["me"] != 1 || owners["opp"] != 1 {
		t.Errorf("Big Wurm tokens %v, want one for the original's target and one for the copy's", owners)
	}
}
