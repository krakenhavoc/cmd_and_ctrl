package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// budget_cast_view_test.go — #2017. The client's cast offers read the same
// gate as the cast path (ADR 0033 §1), so a card over what is left of a
// shared "total mana value N or less" budget must stop being offered the
// moment a cast spends it, and not before.

func TestBudgetedFreeCastOffersFollowTheRunningTotal(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	mk := func(name, cost string) game.Card {
		c := game.NewCard(name, me.ID)
		c.TypeLine = "Artifact"
		c.ManaCost = cost
		c.KnownBy = seen
		g.Exile.PushTop(c)
		return c
	}
	six, five, four := mk("Six", "{6}"), mk("Five", "{5}"), mk("Four", "{4}")
	budget := 10
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:          me.ID,
			Zone:            game.ZoneExile,
			Duration:        g.UntilEndOfTurnDuration(),
			Cost:            "{0}",
			Timing:          game.TimingFlash,
			CastOnly:        true,
			ManaValueBudget: &budget,
			LapseOnPass:     game.LapseStaysInExile,
			Label:           "Cast any number of spells with total mana value 10 or less",
		}, []game.Card{six, five, four})
	})

	offered := func(id uuid.UUID) bool {
		v := ViewOfGameFor(g, me.ID.String())
		for i := range v.Exile.Cards {
			c := &v.Exile.Cards[i]
			if c.InstanceID == id.String() {
				return c.ExilePlay != nil && c.CastableHere
			}
		}
		t.Fatalf("card %v is not in the exile view", id)
		return false
	}
	for name, id := range map[string]uuid.UUID{"Six": six.InstanceID, "Five": five.InstanceID, "Four": four.InstanceID} {
		if !offered(id) {
			t.Errorf("%s fits the total of 10 and is not offered", name)
		}
	}
	if err := g.CastSpell(me.ID, six.InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the 6: %v", err)
	}
	if offered(five.InstanceID) {
		t.Error("the 5 is offered with 4 left")
	}
	if !offered(four.InstanceID) {
		t.Error("the 4 is not offered with 4 left")
	}
}
