package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// adventure_moves_test.go — the enumerator half of CR 715 (#719).
//
// The contract is the one every other enumerator test has and it is
// the only reason these exist: the enumerator and the engine read the
// same functions (Card.CastableFaces from hand, CastPermission from
// everywhere else), so every move offered is a move CastSpell
// accepts. dispatchAll proves it by applying each one.

// adventureInHand seats a two-faced adventure card in a player's hand
// and returns its instance ID. {0} on both halves so the enumerator's
// affordability filter is not the thing under test.
func adventureInHand(p *game.Player) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   "adventure-enum-oracle",
		Layout:     game.LayoutAdventure,
		Owner:      p.ID,
		Controller: p.ID,
		Faces: []game.Face{
			{Name: "Enum Knight", TypeLine: "Creature — Knight", ManaCost: "{0}", Power: 1, Toughness: 1},
			{Name: "Enum Insight", TypeLine: "Instant — Adventure", ManaCost: "{0}"},
		},
	}
	c.SetFace(0)
	p.Hand.PushTop(c)
	return c.InstanceID
}

// faceOf pulls the announced face out of a cast move's params. Absent
// means face 0 — the wire omits the default (ADR 0034).
func faceOf(t *testing.T, m legal.Move) int {
	t.Helper()
	var p struct {
		Face int `json:"face"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode cast params %s: %v", string(m.Params), err)
	}
	return p.Face
}

// CR 715.3: from HAND the bot is offered both halves, because the
// caster chooses which spell to cast.
func TestEnumeratorOffersBothAdventureHalvesFromHand(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	id := adventureInHand(seat)
	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, id)
	if len(got) != 2 {
		t.Fatalf("adventure card offered %d casts, want both halves: %v", len(got), labels(moves))
	}
	seen := map[int]bool{}
	for _, m := range got {
		seen[faceOf(t, m)] = true
	}
	if !seen[0] || !seen[1] {
		t.Errorf("faces offered = %v, want the creature (0) and the Adventure (1)", seen)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// CR 715.4: from EXILE under the Adventure grant the bot is offered
// exactly one cast, and it is the creature. The Adventure half was
// cast once and is spent, and a bot offered it would be offered a
// move the engine settles to something else.
func TestEnumeratorOffersOnlyTheCreatureFromTheAdventureExile(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	// Cast the Adventure half and let it resolve, so the grant under
	// test is the one the engine actually writes rather than one the
	// test invented.
	id := adventureInHand(seat)
	if err := g.CastSpell(seat.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	for i := 0; i < 16 && g.Stack.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the resolved Adventure spell is not in exile")
	}
	advanceTo(t, g, game.StepPostcombatMain)

	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, id)
	if len(got) != 1 {
		t.Fatalf("exiled adventure card offered %d casts, want just the creature: %v",
			len(got), labels(moves))
	}
	if face := faceOf(t, got[0]); face != 0 {
		t.Errorf("offered face %d from exile, want the creature face 0", face)
	}
	dispatchAll(t, g, seat.ID, moves)
}
