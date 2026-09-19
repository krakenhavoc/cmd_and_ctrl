package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_any_kind_view_test.go — the wire half of #943, an
// ANY-KIND removal split across permanents (Tekuthal).
//
// What is pinned is that the view needs NOTHING new for it. A
// CounterCostOptionView has always been a permanent plus the kinds on
// it that could pay, so the shape the client reads — among, no printed
// kind, one row per (permanent, kind) — is already the answer, and the
// picker's stepper per row is the kind choice. The projection's job is
// to list every kind on every matched permanent, down to a floor of
// one, because a permanent holding a single counter is a legal part of
// an among payment.

func TestCounterCostViewAnyKindAmongListsEveryKindOnEveryPermanent(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatCounterCostSource(g, owner, &game.CounterRemovalCost{
		N: 3, Among: true,
		From: &game.TargetSpec{
			Mode:  "permanent",
			Label: "other artifacts, creatures, and planeswalkers you control",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.Name != "Counter Cost Source" &&
					(c.IsArtifact() || c.IsCreature() || c.IsPlaneswalker())
			},
			Min: 1, Max: 1,
		},
	})
	bear := seatCreature(g, owner, "Bear", 2, false, false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bear {
			g.Battlefield.Cards[i].Counters = map[string]int{
				game.CounterPlusOne: 2, game.CounterShield: 1,
			}
		}
	}
	walker := seatCostWalker(g, owner, 1)
	seatCreature(g, owner, "Uncountered Bear", 2, false, false) // no counters: not an option
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.CounterCostN != 3 || !ab.CounterCostAmong || ab.CounterCostKind != "" {
		t.Errorf("shape: n=%d among=%v kind=%q, want 3 / true / empty",
			ab.CounterCostN, ab.CounterCostAmong, ab.CounterCostKind)
	}
	if ab.CounterCostSelf || ab.CounterCostVariable {
		t.Errorf("self=%v variable=%v, want both false", ab.CounterCostSelf, ab.CounterCostVariable)
	}
	if ab.CounterCostLabel != "other artifacts, creatures, and planeswalkers you control" {
		t.Errorf("counter_cost_label = %q", ab.CounterCostLabel)
	}
	if len(ab.CounterCostOptions) != 2 {
		t.Fatalf("counter_cost_options = %+v, want the Bear and the walker", ab.CounterCostOptions)
	}
	// Most counters first, and the Bear offers BOTH its kinds — the
	// two rows the picker turns into two steppers.
	if ab.CounterCostOptions[0].CardID != bear.String() {
		t.Errorf("options not most-counters-first: %+v", ab.CounterCostOptions)
	}
	k := ab.CounterCostOptions[0].Kinds
	if len(k) != 2 || k[0].Kind != game.CounterPlusOne || k[0].Count != 2 ||
		k[1].Kind != game.CounterShield || k[1].Count != 1 {
		t.Errorf("kinds on the Bear = %+v, want +1/+1 (2) then shield (1)", k)
	}
	// A single loyalty counter is a legal PART of the payment, so the
	// walker is offered even though it cannot pay three on its own.
	if ab.CounterCostOptions[1].CardID != walker.String() {
		t.Fatalf("options = %+v, want the one-loyalty walker offered", ab.CounterCostOptions)
	}
	if k := ab.CounterCostOptions[1].Kinds; len(k) != 1 || k[0].Kind != game.CounterLoyalty || k[0].Count != 1 {
		t.Errorf("kinds on the walker = %+v, want loyalty ×1", k)
	}
}
