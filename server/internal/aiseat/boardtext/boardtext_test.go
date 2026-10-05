package boardtext_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func view() *protocol.GameView {
	v := &protocol.GameView{}
	v.Turn.Seq = 5
	v.Turn.Number = 3
	v.Turn.Step = "precombat_main"
	v.Turn.ActiveSeat = 0
	v.Turn.PriorityHolder = 0
	v.Seats = []protocol.PlayerView{{ID: "a", Name: "Ann", Seat: 0, Life: 40}, {ID: "b", Name: "Bo", Seat: 1, Life: 38}}
	v.Seats[0].Hand.Count = 1
	v.Seats[0].Hand.Cards = []protocol.CardView{{Name: "Odd Spell", ManaCost: "{1}{R}", TypeLine: "Sorcery", Unimplemented: true}}
	v.Battlefield.Cards = []protocol.CardView{{Name: "Odd Bear", Controller: "a", TypeLine: "Creature", Power: 2, Toughness: 2, Unimplemented: true}}
	return v
}

func TestRenderBotFlavourIsTerse(t *testing.T) {
	got := boardtext.Render(view(), "a", boardtext.Options{})
	for _, want := range []string{
		"TURN 5 (round 3) — precombat main — active player: Ann (YOU) — priority: Ann (YOU)\n",
		"Ann (YOU) — 40 life, 1 cards in hand, 0 in library\n",
		"  battlefield: Odd Bear 2/2 (unimplemented)\n",
		"  your hand: Odd Spell {1}{R} — Sorcery (unimplemented)\n",
		"Bo — 38 life, 0 cards in hand, 0 in library\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, boardtext.UnimplementedNote) {
		t.Error("the note must stay off unless asked for")
	}
}

func TestRenderNoteUnimplementedForTheAgent(t *testing.T) {
	got := boardtext.Render(view(), "a", boardtext.Options{NoteUnimplemented: true})
	for _, want := range []string{
		"Odd Bear 2/2 (" + boardtext.UnimplementedNote + ")",
		"Odd Spell {1}{R} — Sorcery (" + boardtext.UnimplementedNote + ")",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}

func TestRenderCapsZones(t *testing.T) {
	v := view()
	for i := 0; i < 5; i++ {
		v.Seats[0].Graveyard.Cards = append(v.Seats[0].Graveyard.Cards, protocol.CardView{Name: "G"})
	}
	got := boardtext.Render(v, "a", boardtext.Options{MaxZoneCards: 2})
	if !strings.Contains(got, "  graveyard: G, G, …\n") {
		t.Errorf("cap not applied:\n%s", got)
	}
}

// #2279: Number is the round; the log counts turns (Seq). The line prints
// both, labelled, and a view with no turn sequence says ROUND, not TURN.
func TestRenderLabelsTheRoundAndTheTurn(t *testing.T) {
	v := view()
	if got := boardtext.Render(v, "a", boardtext.Options{}); !strings.HasPrefix(got, "TURN 5 (round 3) — ") {
		t.Errorf("turn line: %q", strings.SplitN(got, "\n", 2)[0])
	}
	v.Turn.Seq = 0
	if got := boardtext.Render(v, "a", boardtext.Options{}); !strings.HasPrefix(got, "ROUND 3 — ") {
		t.Errorf("no sequence: %q", strings.SplitN(got, "\n", 2)[0])
	}
}
