package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_exile_view_test.go — #3192, ADR 0066's second 2026-10-13
// amendment. A conditional grant is public and stays on the card, but
// while its condition is false the holder gets no cast surface and the
// grant says so, so the client withholds the button. Once the condition
// holds, both read the same engine answer the cast path reads.

func TestConditionalExileGrantIsDormantUntilItsConditionHolds(t *testing.T) {
	g, me, opp := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(me.ID, "Held Bolt", "Instant", "{R}"),
		game.CastPermission{
			Player: me.ID, Duration: game.WhileInZoneDuration(),
			Condition: &game.PermissionCondition{Key: game.CondControlSubtype, Subtype: "Kavu"},
		})

	// Dormant: the grant is on the card for every viewer, flagged unmet,
	// and nobody gets a cast surface.
	for _, viewer := range []string{me.ID.String(), opp.ID.String()} {
		c := stripCard(t, g, viewer, id)
		if c.ExilePlay == nil {
			t.Fatalf("viewer %s: a dormant grant left the card", viewer)
		}
		if !c.ExilePlay.ConditionUnmet || c.ExilePlay.Condition == "" {
			t.Errorf("viewer %s: exile_play = %+v, want the printed clause and condition_unmet", viewer, c.ExilePlay)
		}
		if c.CastableHere || len(c.CastPrices) > 0 {
			t.Errorf("viewer %s: castable_here=%v cast_prices=%+v while the condition is false", viewer, c.CastableHere, c.CastPrices)
		}
	}

	// A Kavu arrives: the same grant is now live.
	kavu := game.NewCard("Kavu", me.ID)
	kavu.TypeLine = "Creature — Kavu"
	g.Battlefield.PushTop(kavu)
	c := stripCard(t, g, me.ID.String(), id)
	if c.ExilePlay == nil || c.ExilePlay.ConditionUnmet {
		t.Fatalf("exile_play = %+v, want the condition met", c.ExilePlay)
	}
	if !c.CastableHere {
		t.Error("the condition holds and the card is not castable")
	}
}
