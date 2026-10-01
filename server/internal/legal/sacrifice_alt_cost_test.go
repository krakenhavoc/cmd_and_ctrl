package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// sacrifice_alt_cost_test.go — #1727: an alternative cost whose price
// is a sacrifice. The enumerator takes the payment from the engine's
// own candidate walk (AltCostCandidatesLocked) and its count from
// CardPaymentCount, the same two it reads for escape and Daze, so a
// flashed-back Dread Return is a move with three of the seat's own
// creatures named — and dispatchAll proves the engine accepts it.

const (
	oracleDreadReturn = "352b64d2-2ae5-44ee-a64f-94932ef545d3" // flashback—sacrifice three creatures
	oracleFireblast   = "9dd7f27a-e862-47d7-9158-034cf4d353b8" // sacrifice two Mountains instead
	oracleThalia      = "9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee" // noncreature spells cost {1} more
)

// dreadReturnTable seats a Dread Return in the active seat's
// graveyard beside a creature card for it to return, gives the seat
// `creatures` creatures of its own (the LAST a token, so payment order
// — tokens first — differs from board order), and puts an opponent's
// creature on the board, which must never be named.
func dreadReturnTable(t *testing.T, creatures int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	dread := graveyardCard(seat, game.Card{
		Name: "Dread Return", TypeLine: "Sorcery", ManaCost: "{2}{B}{B}", OracleID: oracleDreadReturn,
	})
	graveyardCard(seat, creature("Buried Wurm", "{6}{G}", 6, 6))
	var mine []uuid.UUID
	for i := 0; i < creatures; i++ {
		c := creature("Fodder", "{2}", 2, 2)
		if i == creatures-1 {
			c = game.Card{Name: "Goblin Token", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1}
		}
		mine = append(mine, battlefieldCard(g, seat, c))
	}
	theirs := battlefieldCard(g, other, creature("Their Bear", "{1}{G}", 2, 2))
	return g, seat, dread, mine, theirs
}

func TestEnumeratorOffersASacrificeFlashbackAndNamesTheCreatures(t *testing.T) {
	g, seat, dread, mine, theirs := dreadReturnTable(t, 4)

	moves := legal.EnumerateFor(g, seat.ID)
	var flashbacks []castPayload
	for _, p := range castPayloadsOf(t, moves, dread) {
		if p.AlternativeCost == "flashback" && p.FromZone == "graveyard" {
			flashbacks = append(flashbacks, p)
		}
	}
	if len(flashbacks) == 0 {
		t.Fatalf("no Dread Return flashback offered with four creatures to sacrifice: %v", labels(moves))
	}
	isMine := map[string]bool{}
	for _, id := range mine {
		isMine[id.String()] = true
	}
	for _, fb := range flashbacks {
		if len(fb.AltCostIDs) != 3 {
			t.Fatalf("the flashback names %d permanents, the cost is three: %+v", len(fb.AltCostIDs), fb)
		}
		named := map[string]bool{}
		for _, id := range fb.AltCostIDs {
			if named[id] {
				t.Errorf("the flashback named a creature twice: %v", fb.AltCostIDs)
			}
			named[id] = true
			if id == theirs.String() {
				t.Error("the flashback named an opponent's creature — CR 701.21a, you sacrifice your own")
			}
			if !isMine[id] {
				t.Errorf("the flashback named %s, which is not one of the seat's creatures", id)
			}
		}
		if len(fb.Targets) != 1 {
			t.Errorf("the flashback targets %d cards, want the one creature card it returns", len(fb.Targets))
		}
	}
	// Payment order (#747): the token goes first, although it is the
	// last of the seat's creatures on the board.
	token := mine[len(mine)-1]
	if got := flashbacks[0].AltCostIDs[0]; got != token.String() {
		t.Errorf("first payment leads with %s, want the token %s", got, token)
	}
	if !hasLabel(moves, "Cast Dread Return from graveyard (Flashback—Sacrifice three creatures") {
		t.Errorf("the flashback label does not name the price: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Two creatures cannot pay "sacrifice three", and an opponent's do not
// count: the offer is filtered out where the rule lives.
func TestSacrificeFlashbackIsNotOfferedWithTooFewCreatures(t *testing.T) {
	g, seat, dread, _, _ := dreadReturnTable(t, 2)

	moves := legal.EnumerateFor(g, seat.ID)
	if got := castMovesFor(moves, dread); len(got) != 0 {
		t.Errorf("offered %v with two creatures for a three-creature cost", labels(got))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Fireblast's sacrifice is paid from HAND, with no zone binding: two
// Mountains, and only Mountains.
func TestEnumeratorOffersFireblastsSacrifice(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mountains := []uuid.UUID{
		battlefieldCard(g, seat, basic("Mountain A", "Mountain")),
		battlefieldCard(g, seat, basic("Mountain B", "Mountain")),
	}
	battlefieldCard(g, seat, basic("Island", "Island"))
	fireblast := handCard(seat, game.Card{
		Name: "Fireblast", TypeLine: "Instant", ManaCost: "{4}{R}{R}", OracleID: oracleFireblast,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	var sacs []castPayload
	for _, p := range castPayloadsOf(t, moves, fireblast) {
		if p.AlternativeCost == "sacrifice" {
			sacs = append(sacs, p)
		}
		if p.AlternativeCost == "" {
			t.Errorf("Fireblast offered at its printed {4}{R}{R} off three lands: %+v", p)
		}
	}
	if len(sacs) == 0 {
		t.Fatalf("no Fireblast sacrifice cast offered with two Mountains: %v", labels(moves))
	}
	want := map[string]bool{mountains[0].String(): true, mountains[1].String(): true}
	for _, p := range sacs {
		if len(p.AltCostIDs) != 2 || !want[p.AltCostIDs[0]] || !want[p.AltCostIDs[1]] {
			t.Errorf("Fireblast names %v, want the two Mountains", p.AltCostIDs)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The auto-tap half (#1727, game.CastAutoTapExclusions). Under Thalia
// a flashed-back Dread Return owes {1}. With two creatures and an
// Eldrazi Spawn and no other mana, the Spawn would have to be both
// the {1} and one of the three sacrifices: not a move. With a Swamp
// beside it, it is.
func TestASacrificeFlashbackIsNotOfferedWhenItsSacrificeIsAlsoItsMana(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, other, game.Card{
		Name: "Thalia, Guardian of Thraben", TypeLine: "Legendary Creature — Human Soldier",
		ManaCost: "{1}{W}", Power: 2, Toughness: 1, OracleID: oracleThalia,
	})
	dread := graveyardCard(seat, game.Card{
		Name: "Dread Return", TypeLine: "Sorcery", ManaCost: "{2}{B}{B}", OracleID: oracleDreadReturn,
	})
	graveyardCard(seat, creature("Buried Wurm", "{6}{G}", 6, 6))
	battlefieldCard(g, seat, creature("Fodder A", "{2}", 2, 2))
	battlefieldCard(g, seat, creature("Fodder B", "{2}", 2, 2))
	battlefieldCard(g, seat, spawnCard())

	moves := legal.EnumerateFor(g, seat.ID)
	if got := castMovesFor(moves, dread); len(got) != 0 {
		t.Errorf("offered %v — the Spawn cannot be both a sacrifice and Thalia's {1}", labels(got))
	}
	dispatchAll(t, g, seat.ID, moves)

	battlefieldCard(g, seat, basic("Swamp", "Swamp"))
	moves = legal.EnumerateFor(g, seat.ID)
	if got := castMovesFor(moves, dread); len(got) == 0 {
		t.Fatalf("a Swamp pays Thalia's {1}, but no flashback was offered: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}
