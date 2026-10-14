package tiers

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// OwnKeyClients builds the model client for a seat that runs on its
// owner's own key (ADR 0182). *modelkeys.Service implements it; main.go
// hands it to the factory. An interface here, not an import, for the
// same reason as everything else in this package: the direction of the
// dependency.
type OwnKeyClients interface {
	// SeatClient returns the client for one seat. It must be cheap and
	// must not block: NewPolicy runs under the Manager's lock, so the
	// key is read on the seat's first call, not here.
	SeatClient(seat aiseat.SeatSpec) model.Client
}

// ownKeyPolicy builds a model tier on the seat owner's key.
//
// It is always buildable, and that is deliberate. The lobby only seats
// one when the owner has a key, but a seat outlives that check: a
// restart brings it back, the owner may remove the key, the server may
// lose its sealing key. A seat whose NewPolicy failed would sit idle and
// stall the table (ADR 0033 §10), so whatever is missing the seat is a
// complete policy, and it plays the heuristic under its tier's name
// until a model call can be made. The server's own spend gate does not
// apply: these calls are on the owner's account, under the owner's own
// limits.
func (f *Factory) ownKeyPolicy(tier Tier, seat aiseat.SeatSpec) (aiseat.Policy, error) {
	opt := Options{
		Meter:       f.opt.Meter,
		MaxThink:    f.opt.MaxThink,
		NoImprovise: f.opt.NoImprovise,
		MaxTokens:   f.opt.MaxTokens,
	}
	var client model.Client
	if f.opt.OwnKeys != nil {
		client = f.opt.OwnKeys.SeatClient(seat)
		opt.Client = client
		// The seat client sets the model on every request; naming it
		// here as well keeps the decision log and the trace honest.
		opt.Models = Models{Routine: seat.KeyModel}
	}
	if f.opt.DeckProfile != nil && seat.Deck != "" {
		if profile, ok := f.opt.DeckProfile(seat.Deck); ok {
			opt.Deck = profile
		}
	}
	p, err := New(tier, opt)
	if err != nil {
		return nil, err
	}
	if n, ok := client.(aiseat.Noticer); ok {
		return &noticePolicy{Policy: p, notices: n}, nil
	}
	return p, nil
}

// noticePolicy lets the runner find the seat client's notices
// (aiseat.Noticer) through the policy it holds. Everything else passes
// through Unwrap, so every optional extension of the policy underneath
// still works (aiseat/capability.go).
type noticePolicy struct {
	aiseat.Policy
	notices aiseat.Noticer
}

func (p *noticePolicy) Unwrap() aiseat.Policy { return p.Policy }

func (p *noticePolicy) TakeNotice() string { return p.notices.TakeNotice() }
