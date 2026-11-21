package effectstest

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
func ImpulseExile(t *testing.T, g *game.Game, name, typeLine, oracle string) (*game.Player, uuid.UUID) {
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

func AdvanceToMainForTest(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}
