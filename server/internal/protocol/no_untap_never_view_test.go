package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4574: "can't become untapped" projects as never, together with static,
// for a lock and for a flagged static; an ordinary hold is static only.
func TestNoUntapViewProjectsCantBecomeUntappedAsNever(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	sourceID, lockedID, heldID, frozenID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: sourceID, Name: "Source", Owner: owner.ID, Controller: owner.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: lockedID, Name: "Locked", Owner: other.ID, Controller: other.ID, Tapped: true})
		g.Battlefield.PushTop(game.Card{InstanceID: heldID, Name: "Held", Owner: other.ID, Controller: other.ID, Tapped: true})
		g.Battlefield.PushTop(game.Card{InstanceID: frozenID, Name: "Frozen", Owner: other.ID, Controller: other.ID, Tapped: true})
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Aura", OracleID: "cant-untap-source", Owner: owner.ID, Controller: owner.ID})
		d, ok := g.ForAsLongAsYouControlDuration(sourceID, owner.ID)
		if !ok {
			t.Fatal("duration did not start")
		}
		if err := g.CantUntapForEffect(lockedID, d); err != nil {
			t.Fatal(err)
		}
		if err := g.HoldUntappedForEffect(heldID, d); err != nil {
			t.Fatal(err)
		}
	})
	prev := game.CatalogUntapStepRestrictions
	game.CatalogUntapStepRestrictions = func(key string) []game.UntapStepRestriction {
		if key != "cant-untap-source" {
			return nil
		}
		return []game.UntapStepRestriction{{
			CantBecomeUntapped: true,
			Restricts:          func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.InstanceID == frozenID },
		}}
	}
	t.Cleanup(func() { game.CatalogUntapStepRestrictions = prev })

	v := ViewOfGameFor(g, owner.ID.String())
	for id, want := range map[uuid.UUID]bool{lockedID: true, heldID: false, frozenID: true} {
		var got *NoUntapView
		for _, card := range v.Battlefield.Cards {
			if card.InstanceID == id.String() {
				got = card.NoUntap
			}
		}
		if got == nil || !got.Static || got.Never != want {
			t.Errorf("%s no_untap = %+v, want static with never=%v", id, got, want)
		}
	}
}
