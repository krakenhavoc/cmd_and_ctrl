package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_cast_alt_cost_test.go — #2767. A card exiled with an impulse
// permission ("you may play it this turn", Breeches) is cast from exile
// under that permission, and CR 118.9 lets the caster use any
// alternative cost the card prints instead of its mana cost: the
// permission opens a zone, it does not name a price. Overload is such a
// cost; flashback is bound to the graveyard and must stay unclaimable
// from exile.

// impulseExile puts a card on top of the active seat's library and
// exiles it with the plain "you may play it this turn" grant.
func impulseExile(t *testing.T, g *game.Game, name, typeLine, oracle string) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{}); err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})
	if !g.Exile.Contains(id) {
		t.Fatal("setup: the card is not in exile")
	}
	return me, id
}

func advanceToMainForTest(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func TestImpulseExiledRiftCanBeCastForOverload(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	me, rift := impulseExile(t, g, "Cyclonic Rift", "Instant", cyclonicRiftOracle)
	advanceToMainForTest(t, g)

	// The offer is on the table for the caster.
	var offered bool
	g.ReadSnapshot(func() {
		card, _ := g.LookupCardForEffect(rift)
		grant := g.CastPermissionForLocked(me.ID, card, game.ZoneExile)
		for _, o := range g.CastOffersForLocked(me.ID, card, game.ZoneExile, grant) {
			if o != nil && o.Key == "overload" {
				offered = true
			}
		}
	})
	if !offered {
		t.Fatal("overload is not among the offers for a Rift exiled with a play permission (#2767)")
	}

	if err := g.CastSpell(me.ID, rift, game.CastSpellParams{FromZone: "exile", AlternativeCost: "overload"}); err != nil {
		t.Fatalf("CastSpell overload from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the overloaded Rift left the opponent's permanent on the battlefield")
	}
}

// The offers bound to another zone stay bound to it: flashback is
// claimable from the graveyard only, whatever opened exile.
func TestFlashbackIsStillRefusedFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me, id := impulseExile(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracle)
	advanceToMainForTest(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", AlternativeCost: "flashback"}); err == nil {
		t.Fatal("flashback was claimed from exile")
	}
}
