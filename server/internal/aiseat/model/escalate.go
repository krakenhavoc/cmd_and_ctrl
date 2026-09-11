package model

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// escalate.go decides which model answers a window.
//
// ADR 0033 §5 lists five triggers and they are all the same shape:
// "this is a window where being wrong is expensive". A routine window
// gets the cheap model; a window on this list gets the frontier one.
// The cost model in the ADR assumes roughly a fifth of surviving
// windows escalate, so these predicates are deliberately narrow —
// each one names a concrete thing on the board, not a mood.

// Escalation reasons. Stable strings: they are the label a per-
// decision record is aggregated by.
const (
	// ReasonStackTargetsMe — something on the stack is pointed at
	// this seat or one of its permanents. Responding correctly is
	// the highest-leverage decision in Magic and the window closes
	// when it resolves.
	ReasonStackTargetsMe = "stack-targets-me"
	// ReasonCombat — attacks or blocks are on offer. Combat is where
	// games are decided and where the heuristic is weakest.
	ReasonCombat = "combat"
	// ReasonRemovalVsThreat — the bot may point something at an
	// opponent's permanent or at their spell, and that opponent's
	// board has crossed the threat threshold. Spending removal on
	// the wrong permanent is the classic rule-based-AI mistake.
	ReasonRemovalVsThreat = "removal-vs-threat"
	// ReasonCloseCall — the heuristic's top two candidates are within
	// Epsilon. The scorer is admitting it cannot separate them.
	ReasonCloseCall = "close-call"
	// ReasonComplexChoice — a modal spell, an X cost, a move with
	// more than one target, or a pick_target choice with a wide
	// candidate list. These are the shapes the scorer reads least
	// well, because their meaning is in oracle text it cannot see.
	ReasonComplexChoice = "complex-choice"
	// ReasonTier — the `strong` tier escalates every window that
	// Layer A did not absorb (ADR 0033 §6: "A + C, wider candidates").
	ReasonTier = "tier"
)

// escalationParams is the subset of the wire cast/activate payloads
// the triggers read. Declared here rather than imported for the same
// reason aiseat/heuristic declares its own: legal's param structs are
// unexported, and what a policy actually depends on is the WIRE,
// which is stable.
type escalationParams struct {
	Targets []struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	} `json:"targets"`
	Modes  []int `json:"modes"`
	XValue int   `json:"x_value"`
}

// board is the once-per-decision index the triggers share.
type board struct {
	me        string
	byID      map[string]*protocol.CardView
	stackByID map[string]*protocol.CardView
}

func newBoard(in aiseat.Input) *board {
	b := &board{
		me:        in.Seat.String(),
		byID:      make(map[string]*protocol.CardView, len(in.View.Battlefield.Cards)),
		stackByID: make(map[string]*protocol.CardView, len(in.View.Stack.Cards)),
	}
	for i := range in.View.Battlefield.Cards {
		c := &in.View.Battlefield.Cards[i]
		b.byID[c.InstanceID] = c
	}
	for i := range in.View.Stack.Cards {
		c := &in.View.Stack.Cards[i]
		b.stackByID[c.InstanceID] = c
	}
	return b
}

// mine reports whether an instance ID is a permanent this seat
// controls.
func (b *board) mine(id string) bool {
	c := b.byID[id]
	return c != nil && c.Controller == b.me
}

// theirs reports whether an instance ID is a permanent or a stack
// item some OTHER seat controls.
func (b *board) theirs(id string) bool {
	if c := b.byID[id]; c != nil {
		return c.Controller != b.me
	}
	if c := b.stackByID[id]; c != nil {
		return c.Controller != b.me
	}
	return false
}

// escalationReasons returns every trigger that fired, sorted, so the
// per-decision record is stable and aggregatable. An empty result
// means the cheap model answers.
func (p *Policy) escalationReasons(in aiseat.Input, cands []heuristic.Candidate) []string {
	fired := map[string]bool{}
	if p.cfg.AlwaysEscalate {
		fired[ReasonTier] = true
	}
	b := newBoard(in)

	// 1. Something on the stack is aimed at this seat.
	for i := range in.View.StackItems {
		it := &in.View.StackItems[i]
		if it.Controller == b.me {
			continue
		}
		for _, t := range it.Targets {
			if (t.Kind == "player" && t.ID == b.me) || b.mine(t.ID) {
				fired[ReasonStackTargetsMe] = true
			}
		}
	}

	// 2. Combat declarations.
	// 3. Removal or a counter, against a board that has grown teeth.
	// 5. Modal / X / multi-target.
	threatening := p.boardIsThreatening(in, b)
	for i := range in.Moves {
		m := &in.Moves[i]
		switch m.Kind {
		case legal.KindAttack, legal.KindBlock:
			fired[ReasonCombat] = true
			continue
		case legal.KindCast, legal.KindActivate:
		default:
			continue
		}
		ep := decodeParams(m.Params)
		if len(ep.Modes) > 0 || ep.XValue > 0 || len(ep.Targets) > 1 {
			fired[ReasonComplexChoice] = true
		}
		if threatening {
			for _, t := range ep.Targets {
				if t.Kind != "player" && b.theirs(t.ID) {
					fired[ReasonRemovalVsThreat] = true
				}
			}
		}
	}

	// 5b. A pick_target choice with a wide candidate list — ADR 0033
	// says ">4 candidates" and this is that, read off the choice
	// rather than guessed from the move count.
	for i := range in.View.PendingChoices {
		ch := &in.View.PendingChoices[i]
		if ch.Chooser != b.me {
			continue
		}
		if ch.Kind == "pick_target" && ch.PickTarget != nil &&
			len(ch.PickTarget.Players)+len(ch.PickTarget.Cards) > p.cfg.WideTargetCount {
			fired[ReasonComplexChoice] = true
		}
	}

	// 4. The heuristic cannot separate its top two.
	if len(cands) >= 2 && cands[0].Value > 0 &&
		cands[0].Value-cands[1].Value <= p.cfg.Epsilon {
		fired[ReasonCloseCall] = true
	}

	out := make([]string, 0, len(fired))
	for k := range fired {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// boardIsThreatening is the "high-threat board" half of trigger 3.
// It is a board fact, not a judgement: some opponent has assembled
// enough power to matter, or this seat is low enough that anything
// matters.
func (p *Policy) boardIsThreatening(in aiseat.Input, b *board) bool {
	var myLife int
	power := map[string]int{}
	biggest := 0
	for i := range in.View.Seats {
		s := &in.View.Seats[i]
		if s.ID == b.me {
			myLife = s.Life
		}
	}
	for i := range in.View.Battlefield.Cards {
		c := &in.View.Battlefield.Cards[i]
		if c.Controller == b.me || !strings.Contains(strings.ToLower(c.TypeLine), "creature") {
			continue
		}
		power[c.Controller] += c.Power
		if c.Power > biggest {
			biggest = c.Power
		}
	}
	for _, tot := range power {
		if tot >= p.cfg.ThreatPower {
			return true
		}
	}
	return biggest >= p.cfg.ThreatCreature || (myLife > 0 && myLife <= p.cfg.DangerLife)
}

func decodeParams(raw json.RawMessage) escalationParams {
	var out escalationParams
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out
}
