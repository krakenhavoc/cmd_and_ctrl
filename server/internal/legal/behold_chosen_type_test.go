package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// behold_chosen_type_test.go — #5406: "you may choose a creature type and
// behold two creatures of that type" (Celestial Reunion). The beholding
// cast names the type on behold_type and two creatures of it on
// reveal_ids, is offered only when the seat has two of some type, and
// every move is one the engine accepts through the wire (#544).

const oracleCelestialReunion = "ba0a815c-1dfe-47c1-890f-496c670124f3"

type chosenBeholdMove struct {
	Optional []int
	Reveal   []string
	Type     string
}

func chosenBeholdMovesOf(t *testing.T, moves []legal.Move, src uuid.UUID) []chosenBeholdMove {
	t.Helper()
	var out []chosenBeholdMove
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p struct {
			OptionalCosts []int    `json:"optional_costs"`
			RevealIDs     []string `json:"reveal_ids"`
			BeholdType    string   `json:"behold_type"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		out = append(out, chosenBeholdMove{Optional: p.OptionalCosts, Reveal: p.RevealIDs, Type: p.BeholdType})
	}
	return out
}

// Two Elves (one on the table, one in hand) and a lone Goblin: the
// beholding form names Elf and both Elves, the permanent first; the plain
// form names neither a type nor a card. Every move dispatches.
func TestChosenTypeBeholdNamesTheTypeAndTwoCreatures(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Forest")
	elfHand := handCard(seat, game.Card{Name: "Elf Friend", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"})
	handCard(seat, game.Card{Name: "Goblin Pal", TypeLine: "Creature — Goblin", ManaCost: "{R}"})
	elfTable := battlefieldCard(g, seat, game.Card{Name: "Elf Scout", TypeLine: "Creature — Elf Scout", Power: 1, Toughness: 1})
	reunion := handCard(seat, game.Card{Name: "Celestial Reunion", TypeLine: "Sorcery", ManaCost: "{X}{G}", OracleID: oracleCelestialReunion})

	moves := legal.EnumerateFor(g, seat.ID)
	got := chosenBeholdMovesOf(t, moves, reunion)
	var with, without int
	for _, m := range got {
		if len(m.Optional) == 0 {
			without++
			if m.Type != "" || len(m.Reveal) != 0 {
				t.Errorf("plain move = %+v, want no type and no cards", m)
			}
			continue
		}
		with++
		if m.Type != "Elf" || len(m.Reveal) != 2 || m.Reveal[0] != elfTable.String() || m.Reveal[1] != elfHand.String() {
			t.Errorf("beholding move = %+v, want Elf naming %s then %s", m, elfTable, elfHand)
		}
	}
	if with == 0 || without == 0 {
		t.Fatalf("moves = %+v, want a beholding and a plain form", got)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// One creature of each type: the beholding form is not offered.
func TestChosenTypeBeholdNotOfferedWithOneOfEachType(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Forest")
	handCard(seat, game.Card{Name: "Elf Friend", TypeLine: "Creature — Elf", ManaCost: "{G}"})
	handCard(seat, game.Card{Name: "Elf Rite", TypeLine: "Kindred Instant — Elf", ManaCost: "{G}"})
	battlefieldCard(g, seat, game.Card{Name: "Goblin Pal", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	reunion := handCard(seat, game.Card{Name: "Celestial Reunion", TypeLine: "Sorcery", ManaCost: "{X}{G}", OracleID: oracleCelestialReunion})

	moves := legal.EnumerateFor(g, seat.ID)
	got := chosenBeholdMovesOf(t, moves, reunion)
	if len(got) == 0 {
		t.Fatal("no cast of Celestial Reunion offered")
	}
	for _, m := range got {
		if len(m.Optional) != 0 || m.Type != "" {
			t.Errorf("move = %+v, want only the plain form", m)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}
