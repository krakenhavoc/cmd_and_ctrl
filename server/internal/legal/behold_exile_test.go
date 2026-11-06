package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// behold_exile_test.go — #4687: the Champions' mandatory "behold a Kithkin
// and exile it". Every cast names one card on reveal_ids; the seat exiles
// a hand card before a permanent and a token before a nontoken permanent,
// and is offered no cast with nothing to behold. Every move is one the
// engine accepts (#544).

const oracleChampionOfTheClachan = "4a220e0c-e463-4264-9a96-658cf842b82e"

func clachanSetup(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 4, "Plains")
	champ := handCard(seat, game.Card{Name: "Champion of the Clachan", TypeLine: "Creature — Kithkin Knight", ManaCost: "{3}{W}", OracleID: oracleChampionOfTheClachan})
	return g, seat, champ
}

func clachanReveals(t *testing.T, moves []legal.Move, champ uuid.UUID) [][]string {
	t.Helper()
	var out [][]string
	for _, m := range beholdMovesOf(t, moves, champ) {
		out = append(out, m.Reveal)
	}
	return out
}

func TestBeholdExileExilesAHandCardBeforeAPermanent(t *testing.T) {
	g, seat, champ := clachanSetup(t)
	battlefieldCard(g, seat, game.Card{Name: "Kithkin Guard", TypeLine: "Creature — Kithkin Soldier", ManaCost: "{W}", Power: 1, Toughness: 1})
	inHand := handCard(seat, game.Card{Name: "Kithkin Card", TypeLine: "Creature — Kithkin", ManaCost: "{1}{W}"})
	moves := legal.EnumerateFor(g, seat.ID)
	got := clachanReveals(t, moves, champ)
	if len(got) == 0 {
		t.Fatal("no cast of the Champion was offered")
	}
	for _, r := range got {
		if len(r) != 1 || r[0] != inHand.String() {
			t.Fatalf("reveal_ids = %v, want the hand card %s", r, inHand)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestBeholdExileExilesATokenBeforeANontokenPermanent(t *testing.T) {
	g, seat, champ := clachanSetup(t)
	battlefieldCard(g, seat, game.Card{Name: "Kithkin Lord", TypeLine: "Creature — Kithkin", ManaCost: "{2}{W}", Power: 3, Toughness: 3})
	token := battlefieldCard(g, seat, game.Card{Name: "Kithkin Token", TypeLine: "Token Creature — Kithkin", Power: 1, Toughness: 1})
	moves := legal.EnumerateFor(g, seat.ID)
	got := clachanReveals(t, moves, champ)
	if len(got) == 0 {
		t.Fatal("no cast of the Champion was offered")
	}
	for _, r := range got {
		if len(r) != 1 || r[0] != token.String() {
			t.Fatalf("reveal_ids = %v, want the token %s", r, token)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestBeholdExileNotOfferedWithNothingToBehold(t *testing.T) {
	g, seat, champ := clachanSetup(t)
	battlefieldCard(g, seat, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	if got := clachanReveals(t, legal.EnumerateFor(g, seat.ID), champ); len(got) != 0 {
		t.Fatalf("casts offered with no Kithkin to behold: %v", got)
	}
}
