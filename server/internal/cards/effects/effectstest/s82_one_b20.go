package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82Walker puts a planeswalker onto the battlefield as an ordinary entry.
func S82Walker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Planeswalker — Test", OracleID: oracle,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.RunStateChecksForTest()
	return id
}

func S82Activate(t *testing.T, g *game.Game, who, walker uuid.UUID, idx int, p game.ActivateAbilityParams) {
	t.Helper()
	p.Strict = true
	if err := g.ActivateCatalogAbility(who, walker, idx, p); err != nil {
		t.Fatalf("activate %d: %v", idx, err)
	}
}

func S82TargetCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// AnswerChooseCardsOrDiscard answers the open hand-discard prompt with the
// first card offered.
func AnswerChooseCardsOrDiscard(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == p.ID {
			t.Logf("prompt kind %v", c.Kind)
			opts := c.DiscardOptions
			if len(opts) == 0 {
				opts = c.ChooseCards
			}
			if len(opts) == 0 {
				opts = []uuid.UUID{p.Hand.Cards[0].InstanceID}
			}
			if c.Kind == game.PendingChoiceChooseCards {
				if err := g.ResolveChooseCards(c.ID, p.ID, opts[:1]); err != nil {
					t.Fatalf("discard: %v", err)
				}
				return
			}
			if err := g.ResolvePendingChoice(c.ID, p.ID, opts[:1]); err != nil {
				t.Fatalf("discard: %v", err)
			}
			return
		}
	}
	t.Fatal("no discard prompt is open")
}
