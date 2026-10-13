package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// assist_test.go — ADR 0150 §6: a caster's moves never ask for help
// (bots never ask), and the helper's pay_amount is offered to the
// helper only, as "don't help" and "help pay {N}".
func TestAssistIsNeverAskedAndIsOfferedToTheHelperOnly(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	helper := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Fan Thing", TypeLine: "Creature — Human Rogue", ManaCost: "{3}{B}",
			Owner: me.ID, Controller: me.ID, Keywords: []string{game.KeywordAssist},
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
		for _, sub := range []string{"Swamp", "Swamp", "Swamp", "Swamp"} {
			g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: sub, TypeLine: "Basic Land — " + sub, Owner: me.ID, Controller: me.ID})
		}
		for range 2 {
			g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: helper.ID, Controller: helper.ID})
		}
	})
	casts := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if strings.Contains(string(m.Params), "assist_player") {
			t.Errorf("the caster was offered a cast that asks for help: %s", m.Label)
		}
		if m.Type == legal.TypeCastSpell && strings.Contains(string(m.Params), id.String()) {
			casts++
		}
	}
	if casts == 0 {
		t.Fatal("the spell with assist is not castable alone")
	}

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, AssistPlayer: helper.ID}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, s := range g.Seats {
		moves := legal.EnumerateFor(g, s.ID)
		if s.ID != helper.ID {
			if len(moves) != 0 {
				t.Errorf("seat %s was offered %v while the helper decides", s.Name, labels(moves))
			}
			continue
		}
		var none, most bool
		for _, m := range moves {
			switch {
			case strings.HasSuffix(m.Label, ": don't help"):
				none = true
			case strings.HasSuffix(m.Label, ": help pay {2}"):
				most = true
			default:
				t.Errorf("unexpected move %q", m.Label)
			}
		}
		if !none || !most {
			t.Errorf("the helper was offered %v; want don't help and help pay {2}", labels(moves))
		}
	}
}
