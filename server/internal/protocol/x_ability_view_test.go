package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_ability_view_test.go — the wire half of {X} on an activated
// ability. The client cannot open an X picker for a prompt it cannot
// see, and it must not open one for an ability that has no X.

func seatXAbilityPermanent(g *game.Game) (uuid.UUID, uuid.UUID) {
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Vault",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000000bb",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "{2}: do a fixed thing", Cost: game.AbilityCost{Mana: "{2}"}},
			{Label: "{X}{X}, {T}: do X things", Cost: game.AbilityCost{Mana: "{X}{X}", Tap: true}},
			{Label: "{X}, {T}: do X things, X can't be 0", Cost: game.AbilityCost{Mana: "{X}", Tap: true, MinX: 1}},
		},
	})
	return id, owner
}

func TestActivatedAbilityViewCarriesTheXPrompt(t *testing.T) {
	g := buildActiveGame(t)
	id, _ := seatXAbilityPermanent(g)
	var c CardView
	for _, v := range ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == id.String() {
			c = v
		}
	}
	if len(c.ActivatedAbilities) != 3 {
		t.Fatalf("got %d abilities on the wire, want 3", len(c.ActivatedAbilities))
	}
	fixed, double, floored := c.ActivatedAbilities[0], c.ActivatedAbilities[1], c.ActivatedAbilities[2]

	if fixed.DemandsX || fixed.MinX != 0 || fixed.XSlots != 0 {
		t.Errorf("a fixed cost must not ask for an X: %+v", fixed)
	}
	if !double.DemandsX {
		t.Error(`"{X}{X}" should set demands_x`)
	}
	if double.XSlots != 2 {
		t.Errorf(`"{X}{X}" x_slots: got %d, want 2 — the picker prices X at twice the announcement`, double.XSlots)
	}
	if double.MinX != 0 {
		t.Errorf("no printed floor should ship min_x 0, got %d", double.MinX)
	}
	if !floored.DemandsX || floored.XSlots != 1 {
		t.Errorf(`"{X}" should be one slot: %+v`, floored)
	}
	if floored.MinX != 1 {
		t.Errorf(`"X can't be 0" min_x: got %d, want 1`, floored.MinX)
	}
}
