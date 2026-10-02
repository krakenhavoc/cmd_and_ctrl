package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// land_play_gate_test.go — ADR 0109 §4 (#1895): the enumerator asks the
// engine's own land-play gate, so a bot is never offered a land that
// "players can't play lands" refuses (ADR 0033 §1), and the move it IS
// offered is one the engine accepts. Real catalog cards, so the gate the
// engine reads is the one a card file declared.

const (
	oracleTerritorialDispute1895 = "a1785817-f17b-471b-a63b-866e7972df1f"
	oracleCityInABottle1895      = "a83f25e3-4d84-4c9b-ab12-19b8d326e459"
)

// landMovesFor lists the land-play moves seat is offered for card.
func landMovesFor(g *game.Game, seat, card uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind == legal.KindLand && m.Source == card {
			out = append(out, m)
		}
	}
	return out
}

func TestNoLandPlayIsOfferedUnderTerritorialDispute(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	forest := handCard(active, basic("Forest", "Forest"))
	dispute := battlefieldCard(g, active, game.Card{
		Name: "Territorial Dispute", TypeLine: "Enchantment", OracleID: oracleTerritorialDispute1895,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	// An extra drop does not lift a "can't" (CR 101.2).
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(active.ID, 2) })

	if got := landMovesFor(g, active.ID, forest); len(got) != 0 {
		t.Fatalf("a land play was offered under Territorial Dispute: %v", labels(got))
	}

	// The same table with the Dispute gone offers it, and the engine
	// accepts what is offered.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(dispute) })
	moves := landMovesFor(g, active.ID, forest)
	if len(moves) != 1 {
		t.Fatalf("the land play was not offered once the Dispute left: %v", labels(legal.EnumerateFor(g, active.ID)))
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestNoLandPlayIsOfferedUnderATurnBan(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	forest := handCard(active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)
	if len(landMovesFor(g, active.ID, forest)) != 1 {
		t.Fatal("setup: the land play is not offered before the ban")
	}
	g.WithWriteLock(func() { g.CantPlayLandsThisTurnForEffect(uuid.Nil, active.ID, "Turf Wound") })
	if got := landMovesFor(g, active.ID, forest); len(got) != 0 {
		t.Fatalf("a land play was offered under a this-turn ban: %v", labels(got))
	}
}

func TestOnlyTheNamedLandsAreRefusedUnderCityInABottle(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	forest := handCard(active, basic("Forest", "Forest"))
	desert := handCard(active, game.Card{Name: "Desert", TypeLine: "Land"})
	battlefieldCard(g, active, game.Card{Name: "City in a Bottle", TypeLine: "Artifact", OracleID: oracleCityInABottle1895})
	advanceTo(t, g, game.StepPrecombatMain)

	if len(landMovesFor(g, active.ID, forest)) != 1 {
		t.Error("a land with another name was refused under City in a Bottle")
	}
	if got := landMovesFor(g, active.ID, desert); len(got) != 0 {
		t.Errorf("Desert was offered under City in a Bottle: %v", labels(got))
	}
}
