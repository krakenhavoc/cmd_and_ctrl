package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assist_test.go — ADR 0150 §4: cast_spell carries `assist_player`,
// and `cancel_assist` withdraws the caster's own parked request.
func TestCastSpellAssistPlayerAndCancelAssist(t *testing.T) {
	g := newGame(t)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	helper := g.Seats[1-g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Fan Thing", TypeLine: "Creature — Human Rogue", ManaCost: "{3}{B}",
			Owner: me.ID, Controller: me.ID, Keywords: []string{game.KeywordAssist},
		})
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: helper.ID, Controller: helper.ID})
	})

	bad, err := Decode(string(TypeCastSpell), me.ID.String(), params(t, map[string]any{"instance_id": id.String(), "assist_player": "nope"}))
	if err != nil {
		t.Fatal(err)
	}
	bad.Caller = me.ID
	if err := Dispatch(g, bad); err == nil {
		t.Error("a malformed assist_player was accepted")
	}

	cast, err := Decode(string(TypeCastSpell), me.ID.String(), params(t, map[string]any{
		"instance_id": id.String(), "strict": true, "auto_tap": true, "assist_player": helper.ID.String(),
	}))
	if err != nil {
		t.Fatal(err)
	}
	cast.Caller = me.ID
	if err := Dispatch(g, cast); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Assist == nil || g.PendingChoices[0].Chooser != helper.ID {
		t.Fatalf("the cast is not parked on the helper: %d prompts", len(g.PendingChoices))
	}

	// The helper can't withdraw the caster's request, nor act as them.
	notMine, err := Decode(string(TypeCancelAssist), me.ID.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	notMine.Caller = helper.ID
	if err := Dispatch(g, notMine); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("helper withdrawing for the caster: %v", err)
	}
	theirs, err := Decode(string(TypeCancelAssist), helper.ID.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	theirs.Caller = helper.ID
	if err := Dispatch(g, theirs); !errors.Is(err, game.ErrPendingChoiceNotFound) {
		t.Errorf("helper withdrawing a request they did not make: %v", err)
	}

	cancel, err := Decode(string(TypeCancelAssist), me.ID.String(), params(t, map[string]any{"choice_id": g.PendingChoices[0].ID.String()}))
	if err != nil {
		t.Fatal(err)
	}
	cancel.Caller = me.ID
	if err := Dispatch(g, cancel); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left after the cancel", len(g.PendingChoices))
	}
	inHand := false
	for _, c := range me.Hand.Cards {
		inHand = inHand || c.InstanceID == id
	}
	if !inHand {
		t.Error("the withdrawn cast moved the card")
	}
}
