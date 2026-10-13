package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The connection cap (ADR 0148, #2977): a newcomer with no seat is
// turned away with an error frame carrying the stable code and a 1013
// close; a seated player and an admin always connect, so a table that
// is mid-game never loses a reconnect to the cap.

func TestConnectionCapTurnsAwayANewcomerWithAFriendlyFrame(t *testing.T) {
	at := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	hub, base := liveSocketsHub(t, at)
	hub.SetMaxConnections(2)

	// Two newcomers fill the cap.
	for i := 0; i < 2; i++ {
		c := dial(t, base+"?ro=1")
		t.Cleanup(func() { _ = c.Close() })
	}
	waitForHubCount(t, hub, 2)

	rejectedBefore := metricstest.Value(t, "cmdctrl_ws_upgrade_rejections_total", metricstest.L{"reason": metrics.RejectCapacity})
	refusedBefore := metricstest.Value(t, "cmdctrl_capacity_refusals_total", metricstest.L{"reason": metrics.RefuseWSConnections})

	c := dial(t, base+"?ro=1")
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("a refused newcomer should be told why, not dropped: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	if f.Kind != protocol.KindError {
		t.Fatalf("frame kind %q, want error", f.Kind)
	}
	var p protocol.ErrorPayload
	if err := json.Unmarshal(f.Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Code != "server_full" || p.Message != capacity.Message {
		t.Errorf("error payload %+v, want the stable server_full code and the friendly message", p)
	}
	// Then the close: 1013, try again later.
	_, _, err = c.ReadMessage()
	var ce *websocket.CloseError
	if !asCloseError(err, &ce) || ce.Code != websocket.CloseTryAgainLater {
		t.Fatalf("close = %v, want 1013", err)
	}

	if hub.Count() != 2 {
		t.Errorf("the refused socket was admitted: count %d", hub.Count())
	}
	if got := metricstest.Value(t, "cmdctrl_ws_upgrade_rejections_total", metricstest.L{"reason": metrics.RejectCapacity}); got != rejectedBefore+1 {
		t.Errorf("rejections = %v, want %v", got, rejectedBefore+1)
	}
	if got := metricstest.Value(t, "cmdctrl_capacity_refusals_total", metricstest.L{"reason": metrics.RefuseWSConnections}); got != refusedBefore+1 {
		t.Errorf("refusals = %v, want %v", got, refusedBefore+1)
	}
}

func asCloseError(err error, out **websocket.CloseError) bool {
	ce, ok := err.(*websocket.CloseError)
	if ok {
		*out = ce
	}
	return ok
}

func TestConnectionCapNeverRefusesASeatedPlayerOrAnAdmin(t *testing.T) {
	hub, base := liveSocketsHub(t, time.Now())
	hub.SetMaxConnections(1)
	c := dial(t, base+"?ro=1")
	t.Cleanup(func() { _ = c.Close() })
	waitForHubCount(t, hub, 1)

	// The cap is full. A player with a seat (reconnecting mid-game) and
	// an admin both get in.
	seat := dial(t, base+"?player="+uuid.NewString())
	t.Cleanup(func() { _ = seat.Close() })
	admin := dial(t, base+"?admin=1")
	t.Cleanup(func() { _ = admin.Close() })
	waitForHubCount(t, hub, 3)

	// And with the cap off nobody is turned away.
	hub.SetMaxConnections(0)
	more := dial(t, base+"?ro=1")
	t.Cleanup(func() { _ = more.Close() })
	waitForHubCount(t, hub, 4)
}
