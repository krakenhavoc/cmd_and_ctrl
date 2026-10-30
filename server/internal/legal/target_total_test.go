package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// target_total_test.go — #4616, the enumerator's half of the total set
// rule: The Super Hero Civil War's chapter I asks for "up to two target
// creatures with total mana value 6 or less". Every answer the bot is
// offered stays within the total and the engine accepts it (#544), a
// prompt none of whose candidates fits still has its "choose nothing"
// answer, and the view ships the values the picker greys by.

const oracleSuperHeroCivilWar = "11d74dc2-8c11-4d4b-b727-8216d2eb6b4e"

// civilWarPrompt casts The Super Hero Civil War for the active seat and
// passes until chapter I's pick_target prompt is open.
func civilWarPrompt(t *testing.T, g *game.Game, active *game.Player) *game.PendingChoice {
	t.Helper()
	clearHand(active)
	saga := handCard(active, game.Card{
		Name: "The Super Hero Civil War", TypeLine: "Enchantment — Saga",
		OracleID: oracleSuperHeroCivilWar, ManaCost: "{3}{R}{W}",
	})
	lands(g, active, "Mountain", "Mountain", 4)
	lands(g, active, "Plains", "Plains", 1)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, saga, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast The Super Hero Civil War: %v", err)
	}
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.PendingChoices) == 0 || g.PendingChoices[0].Kind != game.PendingChoicePickTarget {
		t.Fatal("expected chapter I's pick_target prompt")
	}
	return g.PendingChoices[0]
}

// answerTargets decodes a pick_target answer's targets, whichever field
// carries them.
func answerTargets(t *testing.T, m legal.Move) []uuid.UUID {
	t.Helper()
	var p struct {
		Targets []struct {
			ID string `json:"id"`
		} `json:"targets"`
		Target *struct {
			ID string `json:"id"`
		} `json:"target"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %q: %v", m.Label, err)
	}
	var out []uuid.UUID
	if p.Target != nil {
		p.Targets = append(p.Targets, *p.Target)
	}
	for _, tw := range p.Targets {
		id, err := uuid.Parse(tw.ID)
		if err != nil {
			t.Fatalf("%q: target id %q: %v", m.Label, tw.ID, err)
		}
		out = append(out, id)
	}
	return out
}

func TestSuperHeroCivilWarEnumeratesOnlySetsWithinTheTotal(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mv := map[uuid.UUID]int{}
	add := func(name, cost string, v int) uuid.UUID {
		id := battlefieldCard(g, opp, creature(name, cost, 2, 2))
		mv[id] = v
		return id
	}
	ox := add("Ox", "{3}{G}", 4)
	add("Elk", "{3}{G}", 4)
	bear := add("Bear", "{1}{G}", 2)
	wurm := add("Wurm", "{6}{G}", 7)
	odd := battlefieldCard(g, opp, creature("Odd", "{1}{R} // {1}{U}", 2, 2))
	civilWarPrompt(t, g, active)

	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) == 0 {
		t.Fatal("no answers to chapter I's prompt — the seat is stuck")
	}
	sawEmpty, sawOxBear := false, false
	for _, m := range moves {
		ids := answerTargets(t, m)
		if len(ids) > 2 {
			t.Errorf("%q: %d targets, the clause takes up to two", m.Label, len(ids))
		}
		sum := 0
		for _, id := range ids {
			if id == wurm || id == odd {
				t.Errorf("%q picks a creature that fits no total", m.Label)
			}
			sum += mv[id]
		}
		if sum > 6 {
			t.Errorf("%q totals mana value %d, over 6", m.Label, sum)
		}
		switch {
		case len(ids) == 0:
			sawEmpty = true
		case len(ids) == 2 && (ids[0] == ox && ids[1] == bear || ids[0] == bear && ids[1] == ox):
			sawOxBear = true
		}
	}
	if !sawEmpty {
		t.Error("\"up to two\": choosing nothing is an answer")
	}
	if !sawOxBear {
		t.Errorf("the 4 and the 2 total 6 and should be offered: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)

	// The view ships the rule and each candidate's value; the unreadable
	// cost has none.
	v := protocol.ViewOfGameFor(g, active.ID.String())
	var pt *protocol.LegalTargetsView
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoicePickTarget) {
			pt = v.PendingChoices[i].PickTarget
		}
	}
	if pt == nil || pt.Total == nil {
		t.Fatalf("the prompt ships its total rule: %+v", pt)
	}
	if pt.Total.Label != "have total mana value 6 or less" || pt.Total.Max != 6 {
		t.Errorf("total = %+v", pt.Total)
	}
	if pt.Total.Values[ox.String()] != 4 || pt.Total.Values[wurm.String()] != 7 {
		t.Errorf("values = %v", pt.Total.Values)
	}
	if _, ok := pt.Total.Values[odd.String()]; ok {
		t.Error("an unreadable cost ships no value")
	}
}

// A lone candidate over the total: the enumerator offers "choose
// nothing" and nothing else, rather than a single-slot answer the
// engine refuses.
func TestSuperHeroCivilWarLoneCandidateOverTheTotalStillHasAnAnswer(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	battlefieldCard(g, opp, creature("Wurm", "{6}{G}", 7, 7))
	civilWarPrompt(t, g, active)

	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) != 1 {
		t.Fatalf("answers = %v, want just \"choose nothing\"", labels(moves))
	}
	if ids := answerTargets(t, moves[0]); len(ids) != 0 {
		t.Errorf("the one answer picks %v, want nothing", ids)
	}
	dispatchAll(t, g, active.ID, moves)
}
