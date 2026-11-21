package testsp

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paradigmOfferOn walks to seat's next precombat main and passes until
// the paradigm trigger has resolved and a copy is on offer.
func paradigmOfferOn(t *testing.T, g *game.Game, seat int, name string) game.Card {
	t.Helper()
	effectstest.B39NextTurnOf(t, g, seat)
	for i := 0; i < 12 && len(effectstest.CardCopies(g)) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return effectstest.CopyOf(t, g, name)
}

// The spell resolves as Sign in Blood, exiles itself, and from then on
// each of its caster's precombat main phases offers a free copy, which
// picks its own target.
func TestDecorumDissertationDrawsThenOffersAFreeCopyEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	oppHand, oppLife := len(opp.Hand.Cards), opp.Life

	id := effectstest.CastCatalogSpell(t, g, "Decorum Dissertation", "Sorcery — Lesson", effectstest.DecorumDissertationOracle,
		[]game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(opp.Hand.Cards) - oppHand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
	if opp.Life != oppLife-2 {
		t.Errorf("the target's life %d, want %d", opp.Life, oppLife-2)
	}
	if !effectstest.InExile(g, id) || me.Graveyard.Contains(id) {
		t.Fatal("Decorum Dissertation was not exiled by its paradigm")
	}

	cp := paradigmOfferOn(t, g, seat, "Decorum Dissertation")
	myHand, myLife := len(me.Hand.Cards), me.Life
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(me.ID)); err != nil {
		t.Fatalf("free cast of the copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - myHand; got != 2 {
		t.Errorf("the copy's target drew %d, want 2", got)
	}
	if me.Life != myLife-2 {
		t.Errorf("the copy's target's life %d, want %d", me.Life, myLife-2)
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the copy resolved", n)
	}
	if !effectstest.InExile(g, id) {
		t.Error("the original left exile")
	}

	// And again on the next turn: the trigger lasts the rest of the game.
	paradigmOfferOn(t, g, seat, "Decorum Dissertation")
}

// The copy is a new spell with new targets: this turn's token can go to
// a different player than the original's.
func TestEchocastingSymposiumCopyChoosesItsOwnTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	mine := effectstest.B12Push(g, me.ID, "Big Wurm", "Creature — Wurm", "", 6, 6)

	id := effectstest.CastCatalogSpell(t, g, "Echocasting Symposium", "Sorcery — Lesson", effectstest.EchocastingSymposiumOracle,
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: mine, Slot: 1},
		})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, id) {
		t.Fatal("Echocasting Symposium was not exiled by its paradigm")
	}

	cp := paradigmOfferOn(t, g, seat, "Echocasting Symposium")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true,
		game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 1},
	); err != nil {
		t.Fatalf("free cast of the copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	owners := map[string]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Big Wurm" && effects.IsToken(c) {
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
