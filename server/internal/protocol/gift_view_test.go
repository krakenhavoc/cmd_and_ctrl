package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gift_view_test.go — the VIEW half of ADR 0089 (#1267, CR 702.174).
// The client can only offer the promise, name its recipients and walk
// the promised target clause if the offer carries them, and a
// responder can only see a promise on the stack if the stack item
// does.

const viewGiftOracle = "test-view-gift"

func stubGiftCard(t *testing.T) {
	t.Helper()
	stubViewOptionalCosts(t, viewGiftOracle, []game.AdditionalCost{{
		Optional: true, Key: game.GiftKey, ChoosesOpponent: true, Label: "Gift a card",
		Targets: &game.TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1},
	}})
}

func pushGiftCard(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Gift Thing", TypeLine: "Sorcery", OracleID: viewGiftOracle,
			ManaCost: "{1}", Owner: owner.ID, Controller: owner.ID,
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
	})
	return id
}

func TestGiftOfferNamesItsRecipientsAndItsClause(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stubGiftCard(t)
	id := pushGiftCard(g, me)

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if len(c.OptionalCosts) != 1 {
		t.Fatalf("optional_costs = %+v, want the gift offer", c.OptionalCosts)
	}
	offer := c.OptionalCosts[0]
	if !offer.ChoosesOpponent || offer.Key != game.GiftKey {
		t.Errorf("offer = %+v, want a gift that chooses an opponent", offer)
	}
	if len(offer.OpponentOptions) != 1 || offer.OpponentOptions[0] != opp.ID.String() {
		t.Errorf("opponent_options = %v, want just %s (never the caster)", offer.OpponentOptions, opp.ID)
	}
	if offer.TargetMode != "player" || offer.LegalTargets == nil || len(offer.LegalTargets.Players) == 0 {
		t.Errorf("the promised clause must ride the offer: mode %q, legal %+v", offer.TargetMode, offer.LegalTargets)
	}

	// Nobody left to promise it to: still a gift offer, with no
	// recipients, so the client greys the toggle rather than hide it.
	g.WithWriteLock(func() { opp.Eliminated = true })
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if got := c.OptionalCosts[0]; !got.ChoosesOpponent || len(got.OpponentOptions) != 0 {
		t.Errorf("with no opponent left: %+v, want a gift offer with no recipients", got)
	}
}

func TestGiftPromiseIsShownOnTheStack(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stubGiftCard(t)
	id := pushGiftCard(g, me)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		OptionalCosts: []int{0}, GiftOpponent: opp.ID,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Public: the opponent sees the promise too.
	for _, viewer := range []*game.Player{me, opp} {
		v := ViewOfGameFor(g, viewer.ID.String())
		found := false
		for _, it := range v.StackItems {
			if it.ID == id.String() {
				found = true
				if it.GiftTo != opp.ID.String() {
					t.Errorf("viewer %s: gift_to = %q, want %s", viewer.Name, it.GiftTo, opp.ID)
				}
			}
		}
		if !found {
			t.Fatalf("viewer %s: the spell is not on the stack view", viewer.Name)
		}
	}
}
