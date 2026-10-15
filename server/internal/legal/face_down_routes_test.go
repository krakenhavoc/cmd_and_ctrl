package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ADR 0189 decision 4, enumerator half: a manifested card whose card
// also prints megamorph offers turn_face_up twice, at the same price,
// and each move names its route so the engine can tell them apart.

const megamorphMovesOracle = "legal-megamorph-oracle"

func withMegamorphCast(t *testing.T, faceUp string) {
	t.Helper()
	prev := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != megamorphMovesOracle {
			return nil
		}
		return []game.AlternativeCost{{
			Key:      "megamorph",
			Label:    "Megamorph — cast face down {3}",
			ManaCost: "{3}",
			FaceDown: &game.FaceDownCast{Kind: game.FaceDownMorphed, FaceUpCost: faceUp, FaceUpCounter: true},
		}}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prev })
}

func manifestedDenProtector(t *testing.T, g *game.Game, owner *game.Player) uuid.UUID {
	t.Helper()
	id := battlefieldCard(g, owner, game.Card{
		Name: "Den Protector", TypeLine: "Creature — Human Warrior", ManaCost: "{1}{G}",
		OracleID: megamorphMovesOracle, Power: 2, Toughness: 1,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SetFaceDown(game.FaceDownManifested)
			}
		}
	})
	return id
}

func TestManifestedMegamorphEnumeratesOneMovePerRoute(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withMegamorphCast(t, "{1}{G}")
	advanceTo(t, g, game.StepPrecombatMain)
	id := manifestedDenProtector(t, g, active)
	basics(g, active, "Forest", 2)

	moves := legal.EnumerateFor(g, active.ID)
	acts := specialActionsOf(moves, id)
	if len(acts) != 2 {
		t.Fatalf("want two turn-face-up moves, got %v", labels(acts))
	}
	if acts[0].Label == acts[1].Label {
		t.Errorf("the two moves share a label %q", acts[0].Label)
	}
	var routes []string
	for _, m := range acts {
		var p struct {
			Route string `json:"route"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		routes = append(routes, p.Route)
	}
	if strings.Join(routes, ",") != "morph,mana_cost" {
		t.Errorf("routes = %v, want [morph mana_cost]", routes)
	}
	if !strings.Contains(acts[0].Label, "megamorph") || !strings.Contains(acts[1].Label, "mana cost") {
		t.Errorf("labels do not name their routes: %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)

	// Each move turns it up its own way: the megamorph one adds the
	// counter and the mana-cost one does not.
	for i, wantCounters := range []int{1, 0} {
		clone := g.Clone()
		if err := actions.Dispatch(clone, actions.Action{
			Type: actions.Type(acts[i].Type), Player: acts[i].Player, Caller: active.ID, Params: acts[i].Params,
		}); err != nil {
			t.Fatalf("move %d: %v", i, err)
		}
		var got int
		clone.ReadSnapshot(func() {
			for _, c := range clone.Battlefield.Cards {
				if c.InstanceID == id {
					got = c.Counters["+1/+1"]
				}
			}
		})
		if got != wantCounters {
			t.Errorf("move %d (%s): +1/+1 counters = %d, want %d", i, acts[i].Label, got, wantCounters)
		}
	}
}

// A card with one way up sends no route: its params are what they
// always were.
func TestASingleTurnFaceUpOfferSendsNoRoute(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withMorphCast(t, "{1}{U}")
	advanceTo(t, g, game.StepPrecombatMain)
	down := battlefieldCard(g, active, morphCreature())
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == down {
				g.Battlefield.Cards[i].SetFaceDown(game.FaceDownMorphed)
			}
		}
	})
	basics(g, active, "Island", 2)
	acts := specialActionsOf(legal.EnumerateFor(g, active.ID), down)
	if len(acts) != 1 {
		t.Fatalf("want one move, got %v", labels(acts))
	}
	if strings.Contains(string(acts[0].Params), "route") {
		t.Errorf("a lone offer sent a route: %s", acts[0].Params)
	}
}
