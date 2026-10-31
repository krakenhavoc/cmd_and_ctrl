package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// up_to_cast_x_test.go — #4617: the enumerator's half of "choose up to
// X" (The Ruinous Wrecking Crew). The mode_pick prompt carries the
// bound, so a bot is offered every selection of up to X bullets, none
// of more, and every one dispatches; at X = 0 it owes no prompt at all.

const oracleTheRuinousWreckingCrew = "0ff9c0cc-4fb3-4117-9a8e-5b51f62306c4"

// crewEnters casts the Crew for `x` from seat 0's hand and passes
// priority until the table stops on a prompt or the stack is empty.
func crewEnters(t *testing.T, g *game.Game, x int) *game.Player {
	t.Helper()
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCard(me, game.Card{
		Name: "The Ruinous Wrecking Crew", TypeLine: "Legendary Creature — Human Villain",
		OracleID: oracleTheRuinousWreckingCrew, ManaCost: "{X}{B}{R}", Power: 2, Toughness: 2,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast for X=%d: %v", x, err)
	}
	for i := 0; i < 8 && (g.Stack.Size() > 0 || len(g.PendingTriggers) > 0); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return me
}

func modePickOwed(g *game.Game, seat uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceModePick && c.Chooser == seat {
			return c
		}
	}
	return nil
}

func TestUpToCastXModePickOffersUpToXBullets(t *testing.T) {
	g := newTable(t)
	me := crewEnters(t, g, 2)
	c := modePickOwed(g, me.ID)
	if c == nil {
		t.Fatal("X = 2: the seat owes the mode choice")
	}
	moves := legal.EnumerateFor(g, me.ID)
	sizes := map[int]bool{}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("an owed prompt crowds out everything else: %+v", m)
		}
		var p struct {
			Modes []int `json:"modes"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		if len(p.Modes) > 2 {
			t.Errorf("X = 2: a move takes %d bullets: %v", len(p.Modes), p.Modes)
		}
		sizes[len(p.Modes)] = true
	}
	for _, n := range []int{0, 1, 2} {
		if !sizes[n] {
			t.Errorf("X = 2: no move takes %d bullets (have %v)", n, sizes)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}

func TestUpToCastXAtZeroOwesNoModePick(t *testing.T) {
	g := newTable(t)
	me := crewEnters(t, g, 0)
	if c := modePickOwed(g, me.ID); c != nil {
		t.Fatalf("X = 0: no mode can be chosen, so nothing is asked: %+v", c.ModeOptionIndex)
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeResolveChoice {
			t.Errorf("X = 0: no choice to answer, got %q", m.Label)
		}
	}
}
