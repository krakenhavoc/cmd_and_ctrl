package tiers_test

import (
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
)

// fakeOwnKeys records the seats it was asked for and hands each a
// client that also has a notice to give.
type fakeOwnKeys struct {
	mu    sync.Mutex
	seats []aiseat.SeatSpec
}

type noticingClient struct {
	model.FakeClient
	notice string
}

func (c *noticingClient) TakeNotice() string {
	n := c.notice
	c.notice = ""
	return n
}

func (f *fakeOwnKeys) SeatClient(seat aiseat.SeatSpec) model.Client {
	f.mu.Lock()
	f.seats = append(f.seats, seat)
	f.mu.Unlock()
	return &noticingClient{notice: "my key stopped working"}
}

// An own-key seat is built on its owner's client even on a server with
// no model of its own, keeps every optional extension of the policy
// underneath, and exposes the client's notices to the runner.
func TestAnOwnKeySeatRunsOnItsOwnersClient(t *testing.T) {
	keys := &fakeOwnKeys{}
	f := tiers.NewFactory(tiers.FactoryOptions{OwnKeys: keys})
	owner := uuid.New()
	for _, tier := range []string{"assisted", "strong"} {
		seat := aiseat.SeatSpec{PlayerID: uuid.New(), Tier: tier, KeyOwner: owner, KeyModel: "their-model", GameID: uuid.New()}
		p, err := f.NewPolicy(seat)
		if err != nil {
			t.Fatalf("%s: %v", tier, err)
		}
		if p.Name() != tier {
			t.Errorf("policy name %q, want %q", p.Name(), tier)
		}
		n, ok := aiseat.Capability[aiseat.Noticer](p)
		if !ok {
			t.Fatalf("%s: the runner cannot find the seat client's notices", tier)
		}
		if got := n.TakeNotice(); got != "my key stopped working" {
			t.Errorf("notice = %q", got)
		}
		if _, ok := aiseat.Capability[aiseat.Tracer](p); !ok {
			t.Errorf("%s: the wrapper hid the model policy's Tracer", tier)
		}
	}
	keys.mu.Lock()
	defer keys.mu.Unlock()
	if len(keys.seats) != 2 || keys.seats[0].KeyOwner != owner || keys.seats[0].KeyModel != "their-model" {
		t.Errorf("seats asked for: %+v", keys.seats)
	}
}

// A seat that names a key owner on a server where the feature is off
// (the sealing key was removed, say) is still a complete policy: it
// plays the heuristic under its tier's name rather than sitting idle
// and stalling the table.
func TestAnOwnKeySeatWithTheFeatureOffStillPlays(t *testing.T) {
	f := tiers.NewFactory(tiers.FactoryOptions{})
	p, err := f.NewPolicy(aiseat.SeatSpec{PlayerID: uuid.New(), Tier: "assisted", KeyOwner: uuid.New()})
	if err != nil {
		t.Fatalf("NewPolicy: %v — an idle seat stalls its table", err)
	}
	if _, ok := aiseat.Capability[aiseat.Noticer](p); ok {
		t.Error("a seat with no client has a notice source")
	}
	if p.Name() != "assisted" {
		t.Errorf("name %q", p.Name())
	}
}

// The server's own spend gate is for the server's own bill: an own-key
// seat is built while it is closed. A plain model seat on the same
// server is still built too (buildable), and the PICKER is what the
// gate closes, which the lobby does not consult for an own-key seat.
func TestTheServerGateDoesNotApplyToAnOwnKeySeat(t *testing.T) {
	f := tiers.NewFactory(tiers.FactoryOptions{OwnKeys: &fakeOwnKeys{}, Gate: closedGate{}})
	if _, err := f.NewPolicy(aiseat.SeatSpec{PlayerID: uuid.New(), Tier: "strong", KeyOwner: uuid.New()}); err != nil {
		t.Fatalf("NewPolicy: %v", err)
	}
	if st := f.TierStatus(aiseat.TierStrong); st.Available {
		t.Error("the server tier is offered with no server client")
	}
}

type closedGate struct{}

func (closedGate) Blocked() string { return "switched off" }
