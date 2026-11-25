package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #6270: "Exile a creature card from your graveyard and pay its mana
// cost" (Back from the Brink). The enumerator names a card the seat can
// pay for, and prices the move with it, so the move it offers is one the
// engine accepts (#544); with none affordable there is no move.
func TestExileAndPayItsManaCostNamesAnAffordableCard(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	brink := battlefieldCard(g, active, game.Card{
		Name: "Test Brink", TypeLine: "Enchantment",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Exile a creature card from your graveyard and pay its mana cost: do nothing",
			Cost: game.AbilityCost{
				ExileCards:         &game.ExileCost{N: 1, Label: "a creature card", From: game.ZoneGraveyard, Match: func(c game.Card) bool { return c.IsCreature() }},
				ManaFromExiledCard: true,
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	cheap := uuid.New()
	g.WithWriteLock(func() {
		active.Graveyard.Cards = nil
		active.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Huge", TypeLine: "Creature — Wurm", ManaCost: "{9}{G}", Owner: active.ID, Controller: active.ID})
		active.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{G}", Owner: active.ID, Controller: active.ID})
	})
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { active.ManaPool.AddMana(game.ManaToken{Color: "G"}) })

	acts := activationsOf(legal.EnumerateFor(g, active.ID), brink)
	if len(acts) != 1 {
		t.Fatalf("want one move, got %v", labels(acts))
	}
	var p struct {
		ExileIDs []string `json:"exile_ids"`
	}
	if err := json.Unmarshal(acts[0].Params, &p); err != nil {
		t.Fatal(err)
	}
	if len(p.ExileIDs) != 1 || p.ExileIDs[0] != cheap.String() {
		t.Fatalf("exile_ids = %v, want the {G} card the seat can pay for", p.ExileIDs)
	}
}
