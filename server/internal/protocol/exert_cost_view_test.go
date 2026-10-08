package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0130 §4: an activated row and a mana row that exert their source
// say so on the wire, and an exert never greys a row (CR 701.43b).
func TestViewStampsExertCost(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	var src uuid.UUID
	g.WithWriteLock(func() {
		src = uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: src, Name: "Exert Rows", TypeLine: "Creature — Snake Druid",
			OracleID: "00000000-0000-0000-0000-000000002048",
			Owner:    me.ID, Controller: me.ID, Power: 2, Toughness: 4,
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "plain", Cost: game.AbilityCost{Tap: true}},
				{Label: "exert", Cost: game.AbilityCost{Tap: true, Exert: true}},
			},
			ManaAbilities: []game.ManaAbilityShape{
				{TapCost: true, Produced: "{G}", Label: "{T}: Add {G}."},
				{TapCost: true, ExertCost: true, Produced: "{G}{G}", Label: "{T}, Exert this creature: Add {G}{G}."},
			},
			// Already exerted this turn: the cost can be paid again.
			NextUntapSkips: []game.UntapSkip{{Player: me.ID}},
		})
	})
	g.BumpLayerVersionForTest()

	card := frameCard(t, g, me.ID.String(), src)
	if len(card.ActivatedAbilities) != 2 {
		t.Fatalf("activated rows = %d, want 2", len(card.ActivatedAbilities))
	}
	if r := card.ActivatedAbilities[0]; r.Exert {
		t.Error("the plain row says it exerts")
	}
	if r := card.ActivatedAbilities[1]; !r.Exert || r.CantActivate != "" {
		t.Errorf("exert row = exert %v cant %q, want true and nothing", r.Exert, r.CantActivate)
	}
	if len(card.ManaAbilities) != 2 {
		t.Fatalf("mana rows = %d, want 2", len(card.ManaAbilities))
	}
	if card.ManaAbilities[0].Exert || !card.ManaAbilities[1].Exert || card.ManaAbilities[1].CantActivate != "" {
		t.Errorf("mana rows = %+v", card.ManaAbilities)
	}
}
