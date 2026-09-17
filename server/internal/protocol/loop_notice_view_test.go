package protocol

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loop_notice_view_test.go — the wire half of #628 (CR 726).
//
// The client's autopass toggle is the thing being suspended, and the
// only way it learns about the suspension is this field. It is public
// by design: a loop is something the whole table can watch running,
// and nothing in the notice names a card in a hidden zone.

func TestLoopNoticeIsAbsentByDefault(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)
	if v.LoopNotice != nil {
		t.Fatalf("LoopNotice = %+v on an ordinary game, want nil", v.LoopNotice)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if bytes.Contains(raw, []byte(`"loop_notice"`)) {
		t.Errorf("quiet game still ships loop_notice: %s", raw)
	}
}

func TestLoopNoticeReachesEverySeat(t *testing.T) {
	g := buildActiveGame(t)
	source := uuid.New()
	controller := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.LoopNotice = &game.LoopNotice{
			Source:     source,
			Label:      "Mirror Engine — create a Spark",
			Controller: controller,
			Count:      game.DefaultLoopThreshold,
		}
	})

	v := ViewOfGame(g)
	if v.LoopNotice == nil {
		t.Fatal("LoopNotice missing from the unfiltered view")
	}
	if v.LoopNotice.Label != "Mirror Engine — create a Spark" {
		t.Errorf("label = %q", v.LoopNotice.Label)
	}
	if v.LoopNotice.Source != source.String() {
		t.Errorf("source = %q, want %q", v.LoopNotice.Source, source)
	}
	if v.LoopNotice.Controller != controller.String() {
		t.Errorf("controller = %q, want %q", v.LoopNotice.Controller, controller)
	}
	if v.LoopNotice.Count != game.DefaultLoopThreshold {
		t.Errorf("count = %d, want %d", v.LoopNotice.Count, game.DefaultLoopThreshold)
	}

	// Every seat, and the spectator: the suspension is table-wide, so
	// a seat that lost it would keep autopassing into the loop.
	for _, viewerID := range []string{g.Seats[0].ID.String(), g.Seats[1].ID.String(), ""} {
		filtered := FilterViewFor(v, viewerID)
		if filtered.LoopNotice == nil {
			t.Errorf("viewer %q lost the loop notice", viewerID)
			continue
		}
		if *filtered.LoopNotice != *v.LoopNotice {
			t.Errorf("viewer %q notice = %+v, want %+v", viewerID, filtered.LoopNotice, v.LoopNotice)
		}
	}

	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded GameView
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.LoopNotice == nil || *decoded.LoopNotice != *v.LoopNotice {
		t.Errorf("round-tripped notice = %+v, want %+v", decoded.LoopNotice, v.LoopNotice)
	}
}
