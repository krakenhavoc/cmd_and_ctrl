package legal_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// value_only_test.go — #2877, ADR 0142's amendment of 2026-10-30 (owner,
// option (a)). A cast whose announcement declares exactly `value` (Opt,
// Brainstorm, Fact or Fiction, Mystic Confluence drawing three) carries
// value_only, which smart autopass reads as "no answer to an opponent's
// stack item" by default. A cast that declares another answer (Altar's
// Reap: sac_outlet), declares nothing (Fog), targets the stack (Mystic
// Confluence's counter) or targets a creature with an undeclared mode
// does not.

const (
	oracleOptVO              = "713332c1-5bd8-400f-bfff-c1ca0697a043"
	oracleBrainstormVO       = "36cd2364-d113-47d1-b2c4-b088d9eb88dd"
	oracleFactOrFictionVO    = "437b2dab-15e0-4b9a-a204-58622d37a3b3"
	oracleMysticConfluenceVO = "cd11c27b-9368-4622-bc5a-2e0b993cc42b"
	oracleAltarsReapVO       = "6a125750-2b8c-4f9d-8173-ac8d14c91ddb"
	oracleFogVO              = "27e9db49-7af7-4bef-ad4c-bf5dfb92030d"
)

type valueBoard struct {
	g                                      *game.Game
	seat, opp                              *game.Player
	opt, brainstorm, fof, confluence, reap uuid.UUID
	fog, bears, oppBears                   uuid.UUID
	oppSpell                               uuid.UUID
}

// newValueBoard is the seat at its precombat main with six Islands, two
// Swamps and a Forest, a Grizzly Bears, and a hand of instants. With
// `oppSpell` an opponent's creature spell waits on the stack.
func newValueBoard(t *testing.T, oppSpell bool) *valueBoard {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	b := &valueBoard{g: g, seat: seat, opp: opponentOf(g, seat)}
	for range 6 {
		battlefieldCard(g, seat, basic("Island", "Island"))
	}
	for range 2 {
		battlefieldCard(g, seat, basic("Swamp", "Swamp"))
	}
	battlefieldCard(g, seat, basic("Forest", "Forest"))
	b.bears = battlefieldCard(g, seat, creature("Grizzly Bears", "{1}{G}", 2, 2))
	b.oppBears = battlefieldCard(g, b.opp, creature("Runeclaw Bear", "{1}{G}", 2, 2))
	b.opt = handCard(seat, game.Card{Name: "Opt", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleOptVO})
	b.brainstorm = handCard(seat, game.Card{Name: "Brainstorm", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleBrainstormVO})
	b.fof = handCard(seat, game.Card{Name: "Fact or Fiction", TypeLine: "Instant", ManaCost: "{3}{U}", OracleID: oracleFactOrFictionVO})
	b.confluence = handCard(seat, game.Card{Name: "Mystic Confluence", TypeLine: "Instant", ManaCost: "{3}{U}{U}{U}", OracleID: oracleMysticConfluenceVO})
	b.reap = handCard(seat, game.Card{Name: "Altar's Reap", TypeLine: "Instant", ManaCost: "{1}{B}", OracleID: oracleAltarsReapVO})
	b.fog = handCard(seat, game.Card{Name: "Fog", TypeLine: "Instant", ManaCost: "{G}", OracleID: oracleFogVO})
	if oppSpell {
		b.oppSpell = uuid.New()
		g.WithWriteLock(func() {
			g.Stack.PushTop(game.Card{
				InstanceID: b.oppSpell, Name: "Runeclaw Bear", TypeLine: "Creature — Bear",
				Owner: b.opp.ID, Controller: b.opp.ID,
			})
			if g.StackMeta == nil {
				g.StackMeta = make(map[uuid.UUID]*game.StackItem)
			}
			g.StackMeta[b.oppSpell] = &game.StackItem{
				ID: b.oppSpell, Kind: game.StackItemSpell, Controller: b.opp.ID, Owner: b.opp.ID, SourceCardID: b.oppSpell,
			}
		})
	}
	return b
}

// casts is every cast move of source, failing when there is none.
func (b *valueBoard) casts(t *testing.T, moves []legal.Move, name string, source uuid.UUID) []legal.Move {
	t.Helper()
	got := movesFrom(moves, source, legal.KindCast)
	if len(got) == 0 {
		t.Fatalf("%s offered no cast: %v", name, labels(moves))
	}
	return got
}

// wantValueOnly asserts value_only on every cast of source.
func (b *valueBoard) wantValueOnly(t *testing.T, moves []legal.Move, name string, source uuid.UUID, want bool) {
	t.Helper()
	for _, m := range b.casts(t, moves, name, source) {
		if m.ValueOnly != want {
			t.Errorf("%s: value_only=%v, want %v (%s)", name, m.ValueOnly, want, m.Label)
		}
		if m.Interacts || m.CombatInteracts {
			t.Errorf("%s: a cast carries interacts or combat_interacts (%s)", name, m.Label)
		}
	}
}

func TestValueOnlyMarksADeclaredValueCast(t *testing.T) {
	for _, c := range []struct {
		name     string
		oppSpell bool
	}{
		{"an opponent's spell on the stack", true},
		{"the seat's own main phase, an empty stack", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newValueBoard(t, c.oppSpell)
			moves := legal.EnumerateFor(b.g, b.seat.ID)
			b.wantValueOnly(t, moves, "Opt", b.opt, true)
			b.wantValueOnly(t, moves, "Brainstorm", b.brainstorm, true)
			b.wantValueOnly(t, moves, "Fact or Fiction", b.fof, true)
			b.wantValueOnly(t, moves, "Altar's Reap (sac_outlet)", b.reap, false)
			b.wantValueOnly(t, moves, "Fog (undeclared)", b.fog, false)
		})
	}
}

// Mystic Confluence declares `value` on its draw mode only. Three draws
// answer nothing; any announcement with the counter (it targets the
// stack) or the bounce (an undeclared mode, which the card does not
// cover) is an instant like any other.
func TestValueOnlyReadsTheChosenModes(t *testing.T) {
	b := newValueBoard(t, true)
	moves := legal.EnumerateFor(b.g, b.seat.ID)
	var sawValue, sawOther bool
	for _, m := range b.casts(t, moves, "Mystic Confluence", b.confluence) {
		var p struct {
			Modes []int `json:"modes"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		drawsOnly := len(p.Modes) > 0 && !slices.ContainsFunc(p.Modes, func(i int) bool { return i != 2 })
		if m.ValueOnly != drawsOnly {
			t.Errorf("modes %v: value_only=%v, want %v (%s)", p.Modes, m.ValueOnly, drawsOnly, m.Label)
		}
		if drawsOnly {
			sawValue = true
		} else {
			sawOther = true
		}
		if slices.Contains(p.Modes, 0) && !m.TargetsStack {
			t.Errorf("modes %v: the counter mode does not target the stack (%s)", p.Modes, m.Label)
		}
	}
	if !sawValue || !sawOther {
		t.Fatalf("want both a draw-three cast and another one, saw value=%v other=%v", sawValue, sawOther)
	}
}
