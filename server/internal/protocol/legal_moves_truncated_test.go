package protocol

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// legal_moves_truncated_test.go — ADR 0122 §6.1. GameView carries
// legal_moves_truncated exactly when capLegalMoves dropped a move from
// the viewer's own list, and never on anybody else's frame.

func TestLegalMovesTruncatedIsSetExactlyWhenTheCapDrops(t *testing.T) {
	boards := map[string]func(t *testing.T) *game.Game{}
	for _, creatures := range []int{0, 3, 6, 10} {
		creatures := creatures
		boards[fmt.Sprintf("busy table, %d beasts", creatures)] = func(t *testing.T) *game.Game {
			return busyTable(t, creatures)
		}
	}
	boards["pathological"] = func(t *testing.T) *game.Game {
		g := busyTable(t, 4)
		active := g.Seats[g.Turn.ActiveSeat]
		for range 3 {
			put(g.Battlefield, active, game.Card{
				Name: "Goblin Bombardment", TypeLine: "Enchantment",
				ManaCost: "{1}{R}", OracleID: oracleGoblinBombardment,
			})
			put(active.Hand, active, game.Card{
				Name: "Lightning Bolt", TypeLine: "Instant",
				ManaCost: "{R}", OracleID: oracleLightningBolt,
			})
		}
		return g
	}
	sawTruncated, sawWhole := false, false
	for name, build := range boards {
		t.Run(name, func(t *testing.T) {
			g := build(t)
			active := g.Seats[g.Turn.ActiveSeat]
			full := legal.EnumerateFor(g, active.ID)
			v := ViewOfGameFor(g, active.ID.String())
			dropped := len(v.LegalMoves) < len(full)
			if v.LegalMovesTruncated != dropped {
				t.Errorf("legal_moves_truncated=%v, but the wire list has %d of %d moves",
					v.LegalMovesTruncated, len(v.LegalMoves), len(full))
			}
			if dropped {
				sawTruncated = true
			} else {
				sawWhole = true
			}
			// Never on another seat's frame, a spectator's or the admin's.
			for _, p := range g.Seats {
				if p.ID == active.ID {
					continue
				}
				if ViewOfGameFor(g, p.ID.String()).LegalMovesTruncated {
					t.Errorf("seat %s, which owes nothing, got the active seat's flag", p.Name)
				}
			}
			if ViewOfGameFor(g, SpectatorViewerID).LegalMovesTruncated || ViewOfGameFor(g, "").LegalMovesTruncated {
				t.Error("a seatless viewer got the flag")
			}
		})
	}
	if !sawTruncated || !sawWhole {
		t.Errorf("the boards should cover both a cut list and a whole one: cut=%v whole=%v", sawTruncated, sawWhole)
	}
}

// A list over the cap whose every move has its own key keeps them all,
// and says so.
func TestCapLegalMovesOverTheCapWithNothingToDropIsNotTruncated(t *testing.T) {
	moves := make([]LegalMoveView, 0, legalMovesWireCap+5)
	for i := 0; i < legalMovesWireCap+5; i++ {
		moves = append(moves, LegalMoveView{Type: legal.TypeCastSpell, Kind: legal.KindCast, Source: uuid.New()})
	}
	out, cut := capLegalMoves(moves)
	if cut || len(out) != len(moves) {
		t.Errorf("kept %d of %d distinct moves, truncated=%v", len(out), len(moves), cut)
	}
	moves = append(moves, LegalMoveView{Type: legal.TypeCastSpell, Kind: legal.KindCast, Source: moves[0].Source})
	if _, cut := capLegalMoves(moves); !cut {
		t.Error("an alternative was dropped and the list does not say so")
	}
}
