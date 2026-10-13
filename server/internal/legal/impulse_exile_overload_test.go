package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// impulse_exile_overload_test.go — #2767. A card exiled under a plain
// "you may play it this turn" permission (Breeches) is cast for its
// printed cost or for any alternative cost it prints that is not bound
// to another zone (CR 118.9). The enumerator walks the engine's own
// offer list, so the bot sees overload from exile; a flashback offer
// stays out.
func TestEnumeratorOffersOverloadFromAnImpulseExile(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 7)
	battlefieldCard(g, other, creature("Bear", "{1}{G}", 2, 2))

	rift := game.NewCard("Cyclonic Rift", seat.ID)
	rift.TypeLine = "Instant"
	rift.ManaCost = "{1}{U}"
	rift.OracleID = oracleCyclonicRift
	seat.Library.PushTop(rift)
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(seat.ID, seat.ID, 1, game.CastPermission{}); err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})

	moves := legal.EnumerateFor(g, seat.ID)
	got := offerKeys(t, moves, rift.InstanceID)
	if got[""] == 0 {
		t.Errorf("the printed cast from exile disappeared: %v", labels(moves))
	}
	if got["overload"] != 1 {
		t.Errorf("want one overload cast from exile, offers = %v: %v", got, labels(moves))
	}
	for _, p := range castPayloadsOf(t, moves, rift.InstanceID) {
		if p.FromZone != "exile" {
			t.Errorf("cast from %q, want exile", p.FromZone)
		}
		if p.AlternativeCost == "overload" && len(p.Targets) != 0 {
			t.Errorf("an overloaded cast announced %d targets — the clause is deleted", len(p.Targets))
		}
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, rift.InstanceID))
}

func TestEnumeratorWithholdsFlashbackFromAnImpulseExile(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 4, "Mountain")

	loot := game.NewCard("Faithless Looting", seat.ID)
	loot.TypeLine = "Sorcery"
	loot.ManaCost = "{R}"
	loot.OracleID = oracleFaithlessLooting
	seat.Library.PushTop(loot)
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(seat.ID, seat.ID, 1, game.CastPermission{}); err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})

	moves := legal.EnumerateFor(g, seat.ID)
	got := offerKeys(t, moves, loot.InstanceID)
	if got["flashback"] != 0 {
		t.Errorf("flashback offered from exile: %v", got)
	}
	if got[""] == 0 {
		t.Errorf("the printed cast from exile is not offered: %v", got)
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, loot.InstanceID))
}
