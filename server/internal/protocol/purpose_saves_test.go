package protocol

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2939 gap 3: a survive or protect row that covers more than its own
// source rides the wire as purpose.saves, and is not an amount.
func TestPurposeViewCarriesSaves(t *testing.T) {
	v := viewOfPurpose(game.Purpose{
		Answers: game.AnswerSurvive,
		Saves:   &game.Saves{Matches: game.SweepCreatures, Partial: true},
	})
	if v == nil || v.Saves == nil || v.Saves.Matches != "creatures" || !v.Saves.Partial {
		t.Fatalf("saves projected as %+v", v)
	}
	if v.Priced() {
		t.Error("a purpose that declares only answers and what they cover is Priced")
	}
	raw, err := json.Marshal(viewOfPurpose(game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures}}))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"answers":["survive"],"saves":{"matches":"creatures"}}` {
		t.Errorf("saves marshals as %s", raw)
	}
	if v := viewOfPurpose(game.Purpose{Answers: game.AnswerSurvive}); v.Saves != nil {
		t.Errorf("a row that covers only its source carries saves %+v", v.Saves)
	}
}
