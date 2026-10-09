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

// #2698 (CR 701.60): a suspected creature says so, beside the menace
// its ability list carries, so a model seat is told why it cannot block.
func TestRenderMarksASuspectedCreature(t *testing.T) {
	v := view()
	v.Battlefield.Cards = []protocol.CardView{{
		Name: "Barbed Servitor", Controller: "a", TypeLine: "Artifact Creature", Power: 2, Toughness: 2,
		Suspected: true, Abilities: []string{"indestructible", "menace"},
	}}
	got := boardtext.Render(v, "a", boardtext.Options{})
	if want := "  battlefield: Barbed Servitor 2/2 (suspected, indestructible menace)\n"; !strings.Contains(got, want) {
		t.Errorf("missing %q in\n%s", want, got)
	}
}

// ADR 0129 §7: each seat's non-zero player counters follow its pool, in
// name order, so a model seat sees the energy it can pay.
func TestRenderPrintsPlayerCounters(t *testing.T) {
	v := view()
	v.Seats[0].ManaPool = []string{"{G}"}
	v.Seats[0].Counters = map[string]int{"poison": 2, "energy": 4, "rad": 0}
	got := boardtext.Render(v, "a", boardtext.Options{})
	if want := "Ann (YOU) — 40 life, 1 cards in hand, 0 in library, mana pool {G}, 4 energy, 2 poison\n"; !strings.Contains(got, want) {
		t.Errorf("missing %q in\n%s", want, got)
	}
	if !strings.Contains(got, "Bo — 38 life, 0 cards in hand, 0 in library\n") {
		t.Errorf("a seat with no counters grew a counters phrase:\n%s", got)
	}
}

// ADR 0138 §8: a seat with speed says so after its counters, and max
// speed is named; a seat with none says nothing.
func TestRenderPrintsSpeed(t *testing.T) {
	v := view()
	v.Seats[0].Speed = 3
	v.Seats[1].Speed = 4
	got := boardtext.Render(v, "a", boardtext.Options{})
	if want := "Ann (YOU) — 40 life, 1 cards in hand, 0 in library, speed 3\n"; !strings.Contains(got, want) {
		t.Errorf("missing %q in\n%s", want, got)
	}
	if want := "Bo — 38 life, 0 cards in hand, 0 in library, speed 4 (max speed)\n"; !strings.Contains(got, want) {
		t.Errorf("missing %q in\n%s", want, got)
	}
	v.Seats[0].Speed, v.Seats[1].Speed = 0, 0
	if got := boardtext.Render(v, "a", boardtext.Options{}); strings.Contains(got, "speed") {
		t.Errorf("a seat with no speed grew a speed phrase:\n%s", got)
	}
}

// ADR 0129 §3: an energy payment the seat owes says how much, and a
// pay_amount prompt its bounds, the card's threshold and the unit.
func TestRenderEnergyPrompts(t *testing.T) {
	v := view()
	two := 2
	v.PendingChoices = []protocol.PendingChoiceView{
		{ID: "p1", Kind: "pay_unless", Chooser: "a", Reason: "Thriving Rhino — pay {E}{E}?", Count: 1, PayEnergy: &two},
		{ID: "p2", Kind: "pay_amount", Chooser: "a", Reason: "Harnessed Lightning", Count: 1,
			PayAmount: &protocol.PayAmountView{Min: 0, Max: 5, Goal: 3, Unit: "damage"}},
	}
	got := boardtext.Render(v, "a", boardtext.Options{})
	for _, want := range []string{
		"YOU OWE A CHOICE: pay_unless — Thriving Rhino — pay {E}{E}? (pay 2 energy) (choose 1)\n",
		"YOU OWE A CHOICE: pay_amount — Harnessed Lightning (pay nothing, or 1 to 5 energy; 3 reaches the card's threshold; one energy is one point of damage)\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
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

// #2390: CR 903.9b's commander question says when the commander is
// headed for a hand, which its Reason does not.
func TestRenderSaysWhereABouncedCommanderGoes(t *testing.T) {
	const note = "(if you say no it goes to your hand, where you can cast it without the commander tax)"
	for _, hand := range []bool{true, false} {
		v := view()
		v.PendingChoices = []protocol.PendingChoiceView{{
			ID: "c", Kind: "optional_replacement", Chooser: "a", Count: 1,
			Reason: "Send commander to command zone instead?", PlayableFromZone: hand,
		}}
		got := boardtext.Render(v, "a", boardtext.Options{})
		if strings.Contains(got, note) != hand {
			t.Errorf("headed for a hand %v: the note's presence is wrong in\n%s", hand, got)
		}
	}
}
