package legal_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cant_spend_mana_test.go — #6227: the affordability probe this package
// asks is the one the engine's payment asks (#544), so "You can't spend
// mana to cast this spell" (CR 106.1, CR 601.2h) reaches the move list
// with no code of its own. A cast the delve exiles (CR 702.66a) can
// cover whole is offered; one that would need mana is not, however many
// untapped lands the seat has.
//
// This package enumerates no TAP payments (see castMoves), so a convoke
// spell is offered only when the mana alone could pay for it. With the
// clause that is never, which is why the Hogaak-shaped card below is
// offered no cast at all — conservative in the direction that never
// advertises a move the engine would refuse. Offering a convoke payment
// is a seam of its own, and the engine casts such a spell fine
// (game/cant_spend_mana_test.go).

const (
	cantSpendManaConvokeOracle = "test-6227-legal-convoke-delve"
	cantSpendManaDelveOracle   = "test-6227-legal-delve-only"
)

// cantSpendManaCatalog points the clause at `flagged` oracle IDs and
// gives both test cards convoke and delve, chaining to the real catalog
// for every other card.
func cantSpendManaCatalog(t *testing.T, flagged ...string) {
	t.Helper()
	prevTap, prevDelve, prevNoMana := game.CatalogTapPermanentsCost, game.CatalogDelve, game.CatalogCantSpendMana
	mine := map[string]bool{cantSpendManaConvokeOracle: true, cantSpendManaDelveOracle: true}
	game.CatalogTapPermanentsCost = func(key string) *game.TapPermanentsCost {
		if key == cantSpendManaConvokeOracle {
			return &game.TapPermanentsCost{
				Key: "convoke", ColorClause: true,
				Spec: &game.TargetSpec{Zones: []game.ZoneKind{game.ZoneBattlefield}},
			}
		}
		if prevTap != nil {
			return prevTap(key)
		}
		return nil
	}
	game.CatalogDelve = func(key string) bool {
		return mine[key] || (prevDelve != nil && prevDelve(key))
	}
	flaggedSet := make(map[string]bool, len(flagged))
	for _, k := range flagged {
		flaggedSet[k] = true
	}
	game.CatalogCantSpendMana = func(key string) bool { return flaggedSet[key] }
	t.Cleanup(func() {
		game.CatalogTapPermanentsCost, game.CatalogDelve, game.CatalogCantSpendMana = prevTap, prevDelve, prevNoMana
	})
}

// cantSpendManaTable seeds a main phase with `swamps` Swamps, `forests`
// Forests, `fuel` cards in the graveyard and five creatures that could
// convoke {B}{B}{G}{G}{G}.
func cantSpendManaTable(t *testing.T, swamps, forests, fuel int) (*game.Game, *game.Player) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, swamps, "Swamp")
	basicLands(g, seat, forests, "Forest")
	for i := 0; i < fuel; i++ {
		graveyardCard(seat, game.Card{Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant", ManaCost: "{1}"})
	}
	for i, colour := range []string{"B", "B", "G", "G", "G"} {
		battlefieldCard(g, seat, game.Card{
			Name: fmt.Sprintf("Convoker %d", i), TypeLine: "Creature — Bear",
			Colors: []string{colour}, Power: 1, Toughness: 1,
		})
	}
	return g, seat
}

func cantSpendManaCastMoves(moves []legal.Move, src uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Source == src && m.Kind == legal.KindCast {
			out = append(out, m)
		}
	}
	return out
}

func cantSpendManaDelveIDs(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		DelveIDs []string `json:"delve_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %s: %v", m.Params, err)
	}
	return p.DelveIDs
}

// Four Swamps and three Forests can pay {2}{B}{B}{G}{G}{G} — and the
// clause says no mana may, so the move list offers no cast, which is
// what the engine would do with one (CR 601.2h).
func TestCantSpendManaOffersNoCastLandsWouldHavePaid(t *testing.T) {
	cantSpendManaCatalog(t, cantSpendManaConvokeOracle)
	g, seat := cantSpendManaTable(t, 4, 3, 0)
	spell := handCard(seat, game.Card{
		Name: "Test Necropolis", TypeLine: "Creature — Avatar",
		ManaCost: "{2}{B}{B}{G}{G}{G}", OracleID: cantSpendManaConvokeOracle,
	})
	moves := legal.EnumerateFor(g, seat.ID)
	if got := cantSpendManaCastMoves(moves, spell); len(got) != 0 {
		t.Fatalf("%d cast(s) offered for a spell no mana may pay for", len(got))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The control: the same board and the same card without the clause. The
// lands pay, the cast is offered, and the engine accepts it — so it is
// the clause that takes the move off the list and nothing else about the
// fixture.
func TestWithoutTheClauseTheLandsPayAndTheCastIsOffered(t *testing.T) {
	cantSpendManaCatalog(t)
	g, seat := cantSpendManaTable(t, 4, 3, 0)
	spell := handCard(seat, game.Card{
		Name: "Test Necropolis", TypeLine: "Creature — Avatar",
		ManaCost: "{2}{B}{B}{G}{G}{G}", OracleID: cantSpendManaConvokeOracle,
	})
	moves := legal.EnumerateFor(g, seat.ID)
	if got := cantSpendManaCastMoves(moves, spell); len(got) == 0 {
		t.Fatal("no cast offered for a spell the lands can pay for")
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The positive direction the probe can reach today: a {3} spell with the
// clause and delve, and three cards in the graveyard. Delve covers the
// whole of the mana, so the cast IS offered — and it names the three
// cards, because nothing else could pay.
func TestCantSpendManaOffersTheCastDelveCanPayWhole(t *testing.T) {
	cantSpendManaCatalog(t, cantSpendManaDelveOracle)
	g, seat := cantSpendManaTable(t, 0, 0, 3)
	spell := handCard(seat, game.Card{
		Name: "Test Delver", TypeLine: "Sorcery",
		ManaCost: "{3}", OracleID: cantSpendManaDelveOracle,
	})
	moves := legal.EnumerateFor(g, seat.ID)
	got := cantSpendManaCastMoves(moves, spell)
	if len(got) == 0 {
		t.Fatal("no cast offered for a spell the delve cards can pay for whole")
	}
	for _, m := range got {
		if n := len(cantSpendManaDelveIDs(t, m)); n != 3 {
			t.Errorf("cast offered exiling %d card(s), want all three", n)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Two cards in the graveyard leave one generic mana of the {3} owed, and
// seven untapped lands may not pay it: no cast is offered.
func TestCantSpendManaOffersNoCastDelveLeavesShort(t *testing.T) {
	cantSpendManaCatalog(t, cantSpendManaDelveOracle)
	g, seat := cantSpendManaTable(t, 4, 3, 2)
	spell := handCard(seat, game.Card{
		Name: "Test Delver", TypeLine: "Sorcery",
		ManaCost: "{3}", OracleID: cantSpendManaDelveOracle,
	})
	moves := legal.EnumerateFor(g, seat.ID)
	if got := cantSpendManaCastMoves(moves, spell); len(got) != 0 {
		t.Fatalf("%d cast(s) offered for a spell the graveyard cannot pay for whole", len(got))
	}
	dispatchAll(t, g, seat.ID, moves)
}
