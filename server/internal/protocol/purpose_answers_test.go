package protocol

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0142 §7: a declared Answers rides the wire as purpose.answers, in
// the vocabulary's order, and ["value"] is sent so a reader can tell a
// declared "answers nothing" from an undeclared row.
func TestPurposeViewCarriesAnswers(t *testing.T) {
	v := viewOfPurpose(game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect})
	if v == nil || v.Answers == nil {
		t.Fatalf("answers-only purpose projected as %+v", v)
	}
	if got := *v.Answers; !slices.Equal(got, []string{"protect", "sac_outlet"}) {
		t.Fatalf("answers = %v, want [protect sac_outlet]", got)
	}
	raw, err := json.Marshal(viewOfPurpose(game.Purpose{Answers: game.AnswerValue}))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"answers":["value"]}` {
		t.Fatalf("declared value marshals as %s", raw)
	}
	if v := viewOfPurpose(game.Purpose{Draws: 1}); v.Answers != nil {
		t.Fatalf("an undeclared row carries answers %v", *v.Answers)
	}
}

// The Priced() trap fix: a purpose that declares only what it answers
// is not a purpose the bot prices, so a reader keeps its proxy price.
func TestPurposeViewPriced(t *testing.T) {
	var none *PurposeView
	for _, c := range []struct {
		name string
		v    *PurposeView
		want bool
	}{
		{"nil", none, false},
		{"answers only", viewOfPurpose(game.Purpose{Answers: game.AnswerPump}), false},
		{"draws", viewOfPurpose(game.Purpose{Draws: 1}), true},
		{"draws and answers", viewOfPurpose(game.Purpose{Draws: 1, Answers: game.AnswerValue}), true},
		{"sweep", viewOfPurpose(game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}), true},
		{"pump", viewOfPurpose(game.Purpose{Pump: &game.Pump{Power: 1}}), true},
		{"targets", viewOfPurpose(game.Purpose{Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Damage: 3})}), true},
		{"death payoff", viewOfPurpose(game.Purpose{DeathPayoff: true}), true},
	} {
		if got := c.v.Priced(); got != c.want {
			t.Errorf("%s: Priced() = %v, want %v", c.name, got, c.want)
		}
	}
}
