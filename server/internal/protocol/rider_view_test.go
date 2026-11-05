package protocol

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rider_view_test.go — #4978: the riders a purpose declares beside its
// main effect ride the wire under their own names.
func TestPurposeViewCarriesRiders(t *testing.T) {
	v := viewOfPurpose(game.Purpose{Draws: 1, Scry: 2, Learn: 1, Discover: 4, Amass: 2, RingTempts: 1})
	want := PurposeView{Draws: 1, Scry: 2, Learn: 1, Discover: 4, Amass: 2, RingTempts: 1}
	if v == nil || *v != want {
		t.Fatalf("purpose = %+v, want %+v", v, want)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"draws":1,"scry":2,"learn":1,"discover":4,"amass":2,"ring_tempts":1}`; got != want {
		t.Errorf("wire = %s, want %s", got, want)
	}
	if viewOfPurpose(game.Purpose{}) != nil {
		t.Error("the zero purpose has a view")
	}
}
