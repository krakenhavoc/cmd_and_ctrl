package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// goaders_view_test.go — #1598, ADR 0045 Decision 53: the wire half of
// "goad remembers every goader". `goaders` lists every goader oldest
// first, `goaded_by` keeps its one-ID shape as the latest of them, and
// an ungoaded creature carries neither key.
func TestGoadersViewListsEveryGoader(t *testing.T) {
	g := threeSeatsInDeclareAttackers(t)
	s := g.Turn.ActiveSeat
	active, first, second := g.Seats[s], g.Seats[(s+1)%3], g.Seats[(s+2)%3]
	bear, plain := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{bear, plain} {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
		})
	}
	for _, p := range []*game.Player{first, second} {
		if err := g.SetGoaded(bear, p.ID); err != nil {
			t.Fatal(err)
		}
	}
	views := map[string]CardView{}
	for _, c := range ViewOfGame(g).Battlefield.Cards {
		views[c.InstanceID] = c
	}
	v := views[bear.String()]
	if len(v.Goaders) != 2 || v.Goaders[0] != first.ID.String() || v.Goaders[1] != second.ID.String() {
		t.Errorf("goaders = %v, want [%s %s]", v.Goaders, first.ID, second.ID)
	}
	if v.GoadedBy != second.ID.String() {
		t.Errorf("goaded_by = %q, want the latest goader %s", v.GoadedBy, second.ID)
	}
	raw, err := json.Marshal(views[plain.String()])
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]any
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"goaders", "goaded_by"} {
		if _, ok := keys[k]; ok {
			t.Errorf("an ungoaded creature carries %q", k)
		}
	}
}
