package legal_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// face_down_moves_test.go — the enumerator half of #1194 / ADR 0082.
//
// #544's invariant, twice over: offer exactly the moves the engine
// accepts, and nothing it would refuse. Both halves of the seam have
// a way to get that wrong that nothing else in this package would
// catch:
//
//   - the CAST walk reads the card's targets, modes and optional
//     costs out of the catalog, and a morph's cast has none of them
//     (CR 708.2a) — so the walk has to stamp the face-down state the
//     engine stamps, or it offers an announcement CastSpell rejects.
//   - the SPECIAL-ACTION walk has only ever looked in a hand, and the
//     turn-face-up action's card is on the battlefield.

const morphMovesOracle = "legal-morph-oracle"

// withMorphCast wires a morph declaration for one test.
func withMorphCast(t *testing.T, faceUp string) {
	t.Helper()
	prev := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != morphMovesOracle {
			return nil
		}
		return []game.AlternativeCost{{
			Key:      "morph",
			Label:    "Morph — cast face down {3}",
			ManaCost: "{3}",
			FaceDown: &game.FaceDownCast{Kind: game.FaceDownMorphed, FaceUpCost: faceUp},
		}}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prev })
}

func morphCreature() game.Card {
	return game.Card{
		Name:      "Willbender",
		TypeLine:  "Creature — Human Wizard",
		ManaCost:  "{1}{U}",
		OracleID:  morphMovesOracle,
		Power:     1,
		Toughness: 2,
	}
}

// withCastTargetMode wires the S13.1 free-form announce-time target
// prompt for one card.
//
// It stands in here for EVERY clause the card prints: the cast walk
// reads the target mode, the target spec, the modes and the optional
// costs through one `CatalogKey(card)`, so a walk that reads the card
// face up gets all four wrong at once and this is the one of them the
// walk refuses outright on (a card the catalog marks targeted the
// S13.1 way is not enumerable at all). CR 708.2a says a face-down
// spell has none of them.
func withCastTargetMode(t *testing.T, oracle, mode string) {
	t.Helper()
	prev := game.CatalogTargetMode
	game.CatalogTargetMode = func(id string) string {
		if id == oracle {
			return mode
		}
		return ""
	}
	t.Cleanup(func() { game.CatalogTargetMode = prev })
}

// TestTheFaceDownCastIsOfferedBesideThePrintedOne: a morph creature is
// two casts, not one — {1}{U} face up and {3} face down — and every
// one of them is a move the dispatcher accepts.
//
// The card also prints an announce-time target clause, which is what
// makes the face-down half load-bearing: face up it is unenumerable
// the S13.1 way, and face down there is no clause at all, so a walk
// that failed to stamp the state would offer exactly nothing here.
func TestTheFaceDownCastIsOfferedBesideThePrintedOne(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withMorphCast(t, "{1}{U}")
	withCastTargetMode(t, morphMovesOracle, "creature")
	card := handCard(active, morphCreature())
	advanceTo(t, g, game.StepPrecombatMain)
	for i := 0; i < 4; i++ {
		battlefieldCard(g, active, basic("Island", "Island"))
	}

	moves := legal.EnumerateFor(g, active.ID)
	casts := movesOfKindFor(moves, legal.KindCast, card)
	faceDown, faceUp := 0, 0
	for _, m := range casts {
		if strings.Contains(m.Label, "Morph") {
			faceDown++
			continue
		}
		faceUp++
	}
	if faceDown != 1 {
		t.Errorf("face-down casts offered = %d, want 1 — got %v", faceDown, labels(casts))
	}
	// The printed cast is the one the S13.1 target mode makes
	// unenumerable: the engine would demand a target and nothing says
	// which are legal, so the walk declines to offer it. Asserted so
	// this test says what it is really holding — the face-down cast
	// survives a clause the face-up cast does not.
	if faceUp != 0 {
		t.Errorf("printed casts offered = %d, want 0 (an S13.1 target mode with no spec) — got %v", faceUp, labels(casts))
	}
	// The soundness half: the engine accepts every move offered.
	dispatchAll(t, g, active.ID, moves)
}

// TestTurnFaceUpIsEnumeratedFromTheBattlefield: the special-action
// walk reaches the seat's own face-down permanents, and nothing else
// on the board.
func TestTurnFaceUpIsEnumeratedFromTheBattlefield(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withMorphCast(t, "{1}{U}")
	advanceTo(t, g, game.StepPrecombatMain)

	down := battlefieldCard(g, active, morphCreature())
	up := battlefieldCard(g, active, creature("Grizzly Bears", "{1}{G}", 2, 2))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == down {
				g.Battlefield.Cards[i].SetFaceDown(game.FaceDownMorphed)
			}
		}
	})

	// No mana: the row is priced through the same gate every other
	// special action is, so an unaffordable morph cost is not offered.
	if acts := specialActionsOf(legal.EnumerateFor(g, active.ID), down); len(acts) != 0 {
		t.Fatalf("a seat with no mana: want no turn-face-up move, got %v", labels(acts))
	}

	for i := 0; i < 2; i++ {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	moves := legal.EnumerateFor(g, active.ID)
	acts := specialActionsOf(moves, down)
	if len(acts) != 1 {
		t.Fatalf("want exactly one turn-face-up move, got %v", labels(acts))
	}
	if acts[0].Type != legal.TypeSpecialAction {
		t.Errorf("move type = %q, want %q", acts[0].Type, legal.TypeSpecialAction)
	}
	if got := specialActionsOf(moves, up); len(got) != 0 {
		t.Errorf("a face-UP permanent was offered %v", labels(got))
	}
	dispatchAll(t, g, active.ID, moves)
}

// TestAnOpponentsMorphIsNotEnumerated is CR 708.6: the CONTROLLER
// turns a face-down permanent face up, and the battlefield is the one
// zone that holds everybody's cards.
func TestAnOpponentsMorphIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	clearHand(other)
	withMorphCast(t, "{1}{U}")
	advanceTo(t, g, game.StepPrecombatMain)

	theirs := battlefieldCard(g, other, morphCreature())
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].SetFaceDown(game.FaceDownMorphed)
			}
		}
	})
	for i := 0; i < 4; i++ {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	if acts := specialActionsOf(legal.EnumerateFor(g, active.ID), theirs); len(acts) != 0 {
		t.Errorf("a seat was offered %v on a morph it does not control", labels(acts))
	}
}
