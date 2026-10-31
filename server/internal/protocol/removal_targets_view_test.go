package protocol

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// removal_targets_view_test.go — #4737. On a busy board a single-target
// removal spell has more legal targets than the enumerator's per-card cap
// (legal.Options.MaxExpansionPerSource, 12) lets into the move list, and
// the list is still far under the wire's 48-move cap, so
// legal_moves_truncated stays off. Two things must hold on that frame:
//
//   - the spell's legal_actions entry reports the cut, with its count,
//     which is how a seat that plays from the move list (the MCP seat)
//     learns that it must ask for the rest;
//   - the card's legal_targets, which the browser's targeting lights and
//     sends from, is the engine's whole legal set and not the capped list.

const oracleAnguishedUnmaking = "ad09b3c3-c8e7-481c-8c45-e7f234935117"

// removalTable is a four-seat table in the first seat's precombat main:
// it holds Anguished Unmaking with the mana to cast it, and each opponent
// controls perOpponent distinct artifacts.
func removalTable(t *testing.T, perOpponent int) (*game.Game, uuid.UUID, []string) {
	t.Helper()
	g := game.NewGame()
	for i := range 4 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 39 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat %d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(4737, 12))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
		p.Hand.Cards = nil
	}
	active := g.Seats[g.Turn.ActiveSeat]
	spell := put(active.Hand, active, game.Card{Name: "Anguished Unmaking", TypeLine: "Instant",
		ManaCost: "{1}{W}{B}", OracleID: oracleAnguishedUnmaking})
	active.Hand.Cards[len(active.Hand.Cards)-1].AddKnower(active.ID) // its owner sees their hand
	put(g.Battlefield, active, basicLand("Plains", "Plains"))
	put(g.Battlefield, active, basicLand("Swamp", "Swamp"))
	put(g.Battlefield, active, basicLand("Swamp", "Swamp"))
	var targets []string
	for _, p := range g.Seats {
		if p == active {
			continue
		}
		for j := range perOpponent {
			id := put(g.Battlefield, p, game.Card{Name: fmt.Sprintf("Relic %s-%d", p.Name, j), TypeLine: "Artifact", ManaCost: "{1}"})
			targets = append(targets, id.String())
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	return g, spell, targets
}

func TestARemovalSpellOverABusyBoardReportsItsCutAndKeepsItsWholeLegalSet(t *testing.T) {
	g, spell, targets := removalTable(t, 7)
	active := g.Seats[g.Turn.ActiveSeat]
	v := ViewOfGameFor(g, active.ID.String())

	if v.LegalMovesTruncated {
		t.Fatalf("%d moves: the wire cap bit, so this is not the board #4737 is about", len(v.LegalMoves))
	}
	casts := 0
	for _, m := range v.LegalMoves {
		if m.Source == spell && m.Kind == legal.KindCast {
			casts++
		}
	}
	want := legal.Cut{Source: spell, Cap: legal.CapPerSource, Omitted: len(targets) - casts}
	if casts >= len(targets) || casts == 0 {
		t.Fatalf("%d cast moves for %d legal targets; the per-card cap should cut them", casts, len(targets))
	}

	// The cut, where a seat reading the move list finds it.
	if v.LegalActions == nil || v.LegalActions.Sources[spell.String()] == nil {
		t.Fatal("the spell has no legal_actions entry")
	}
	got := v.LegalActions.Sources[spell.String()].Truncated
	if len(got) != 1 || got[0].Cap != string(want.Cap) || got[0].Omitted != want.Omitted || got[0].AtLeast {
		t.Errorf("legal_actions truncated = %+v; want %s, %d exactly", got, want.Cap, want.Omitted)
	}

	// The browser's legal set: every target, whatever the list kept.
	var card *CardView
	for i := range v.Seats {
		for j := range v.Seats[i].Hand.Cards {
			if c := &v.Seats[i].Hand.Cards[j]; c.InstanceID == spell.String() {
				card = c
			}
		}
	}
	if card == nil || card.LegalTargets == nil {
		t.Fatal("the spell in hand carries no legal_targets")
	}
	for _, id := range targets {
		if !slices.Contains(card.LegalTargets.Cards, id) {
			t.Errorf("legal_targets is missing %s, a legal target", id)
		}
	}
	if len(card.LegalTargets.Cards) != len(targets) {
		t.Errorf("legal_targets has %d cards; want the %d nonland permanents", len(card.LegalTargets.Cards), len(targets))
	}
}
