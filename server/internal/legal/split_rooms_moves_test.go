package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// split_rooms_moves_test.go — the enumerator half of ADR 0103: every
// move offered for a split card or a Room is one the engine accepts
// (dispatchAll), and nothing the rules forbid is offered.

func splitEnumCard(owner uuid.UUID, oracle string, left, right game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutSplit,
		Owner:      owner,
		Controller: owner,
		Faces:      []game.Face{left, right},
	}
	c.SettleImported()
	return c
}

func splitCastParams(t *testing.T, m legal.Move) (face int, fuse bool) {
	t.Helper()
	var p struct {
		Face int  `json:"face"`
		Fuse bool `json:"fuse"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %s: %v", string(m.Params), err)
	}
	return p.Face, p.Fuse
}

// CR 709.3 / 702.102a: a fuse card in hand offers its left half, its
// right half and the fused cast.
func TestEnumeratorOffersBothHalvesAndTheFusedCast(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	c := splitEnumCard(seat.ID, "fuse-enum-oracle",
		game.Face{Name: "Left", TypeLine: "Sorcery", ManaCost: "{0}", OracleText: "Left.\nFuse (You may cast one or both halves of this card from your hand.)"},
		game.Face{Name: "Right", TypeLine: "Sorcery", ManaCost: "{0}", OracleText: "Right.\nFuse (You may cast one or both halves of this card from your hand.)"},
	)
	seat.Hand.PushTop(c)
	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, c.InstanceID)
	seen := map[string]bool{}
	for _, m := range got {
		face, fuse := splitCastParams(t, m)
		switch {
		case fuse:
			seen["fused"] = true
		case face == 0:
			seen["left"] = true
		case face == 1:
			seen["right"] = true
		}
	}
	if !seen["left"] || !seen["right"] || !seen["fused"] {
		t.Fatalf("offered %v, want left, right and fused: %v", seen, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// CR 702.127a: the aftermath half is never offered from hand; from the
// graveyard only it is, and the first half is not.
func TestEnumeratorOffersTheAftermathHalfOnlyFromTheGraveyard(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mk := func() game.Card {
		return splitEnumCard(seat.ID, "aftermath-enum-oracle",
			game.Face{Name: "First", TypeLine: "Sorcery", ManaCost: "{0}", OracleText: "First."},
			game.Face{Name: "Later", TypeLine: "Sorcery", ManaCost: "{0}", OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nLater."},
		)
	}
	inHand, inYard := mk(), mk()
	seat.Hand.PushTop(inHand)
	seat.Graveyard.PushTop(inYard)
	moves := legal.EnumerateFor(g, seat.ID)
	for _, m := range castMovesFor(moves, inHand.InstanceID) {
		if face, _ := splitCastParams(t, m); face == 1 {
			t.Errorf("the aftermath half was offered from hand: %s", m.Label)
		}
	}
	yard := castMovesFor(moves, inYard.InstanceID)
	if len(yard) != 1 {
		t.Fatalf("graveyard offered %d casts, want exactly the aftermath half: %v", len(yard), labels(moves))
	}
	if face, _ := splitCastParams(t, yard[0]); face != 1 {
		t.Errorf("graveyard cast is face %d, want the aftermath half 1", face)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// CR 709.5e: a Room with a locked door offers an unlock per locked
// door, naming the door, at sorcery speed only.
func TestEnumeratorOffersUnlockPerLockedDoor(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	room := splitEnumCard(seat.ID, "room-enum-oracle",
		game.Face{Name: "Hall", TypeLine: "Enchantment — Room", ManaCost: "{0}"},
		game.Face{Name: "Cellar", TypeLine: "Enchantment — Room", ManaCost: "{0}"},
	)
	g.WithWriteLock(func() {
		seat.Hand.PushTop(room)
		if _, err := game.MoveCard(seat.Hand, g.Battlefield, room.InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	unlocks := func() []legal.Move {
		var out []legal.Move
		for _, m := range legal.EnumerateFor(g, seat.ID) {
			if m.Kind == legal.KindSpecialAction && m.Source == room.InstanceID {
				out = append(out, m)
			}
		}
		return out
	}
	if got := unlocks(); len(got) != 0 {
		t.Errorf("unlocks offered outside a main phase: %v", labels(got))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	got := unlocks()
	if len(got) != 2 {
		t.Fatalf("unlocks offered = %v, want one per locked door", labels(got))
	}
	doors := map[string]bool{}
	for _, m := range got {
		var p struct {
			Kind string `json:"kind"`
			Door string `json:"door"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.Kind != "unlock" {
			t.Errorf("kind %q, want unlock", p.Kind)
		}
		doors[p.Door] = true
	}
	if !doors["left"] || !doors["right"] {
		t.Errorf("doors offered = %v, want left and right", doors)
	}
	dispatchAll(t, g, seat.ID, got)
}
