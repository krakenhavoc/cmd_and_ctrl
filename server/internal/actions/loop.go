package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loop.go — the action half of ADR 0193 (CR 732.2): the five verbs of
// the demonstrated-loop conversation, and the recorder that turns every
// dispatched action into a step while a loop is being recorded.
//
// Recording happens here, in Dispatch, because Dispatch is the one
// place every seat's action passes through: a browser's and an MCP
// seat's (ws/hub.go), a bot's (aiseat), an improvisation's bundle. A
// step is recorded only once its action succeeded, with every ID of an
// object created inside the iteration rewritten to a reference
// (game.LoopRefPrefix), so a replay can find the same object in a later
// iteration. ws.Room's replay reverses the rewrite with ResolveLoopStep.

// The loop verbs. None mints an undo entry (MintsNoUndo): they are the
// shortcut conversation, not plays, and the replay they lead to is one
// undo entry of its own.
const (
	// TypeLoopRecordStart starts recording a loop for `player`, who
	// must hold priority with no prompt open. No params.
	TypeLoopRecordStart Type = "loop_record_start"
	// TypeLoopRecordStop ends the demonstration; the closure check
	// decides whether the loop is ready. No params.
	TypeLoopRecordStop Type = "loop_record_stop"
	// TypeLoopRecordCancel discards the recording and any proposal.
	TypeLoopRecordCancel Type = "loop_record_cancel"
	// TypeLoopPropose proposes `{count}` more iterations (1..1000).
	TypeLoopPropose Type = "loop_propose"
	// TypeResolveLoopProposal answers a loop_proposal prompt:
	// `{choice_id, accept: true}` or `{choice_id, shorten: {iteration,
	// step}}` (CR 732.2b).
	TypeResolveLoopProposal Type = "resolve_loop_proposal"
)

// isLoopVerb reports whether t is one of the loop verbs above.
func isLoopVerb(t Type) bool {
	switch t {
	case TypeLoopRecordStart, TypeLoopRecordStop, TypeLoopRecordCancel, TypeLoopPropose, TypeResolveLoopProposal:
		return true
	}
	return false
}

// recordsAsLoopStep reports whether an action of type t is a step of a
// recording. The loop verbs are the conversation about the recording,
// the settings and the table roll are not plays, and a concession ends
// the recording rather than being part of it.
func recordsAsLoopStep(t Type) bool {
	if isLoopVerb(t) {
		return false
	}
	switch t {
	case TypeRollOpening, TypeHostRollRemaining, TypeChooseStartingPlayer,
		TypeRollTableDie, TypeSetTableSettings, TypeSetUndoLimit,
		TypeSetTriggerOrderPreference, TypeSetAutoAnswers, TypeConcede:
		return false
	}
	return true
}

// loopDraft is a step being recorded, with the public name of the card
// it is about as it was before the action (a sacrificed permanent is
// named before it leaves; a spell cast from a hand after it is cast).
type loopDraft struct {
	step   *game.LoopStepDraft
	source string
}

// beginLoopStep drafts the step `a` is about to be, or nil when no loop
// is being recorded or `a` is not a step.
func beginLoopStep(g *game.Game, a Action) *loopDraft {
	if !recordsAsLoopStep(a.Type) {
		return nil
	}
	d := g.BeginLoopStep()
	if d == nil {
		return nil
	}
	return &loopDraft{step: d, source: loopSourceName(g, a)}
}

// commitLoopStep records `a`, which succeeded, as the drafted step.
func commitLoopStep(g *game.Game, ld *loopDraft, a Action) {
	if ld == nil {
		return
	}
	d := ld.step
	params, err := rewriteParamIDs(a.Params, func(s string) (string, bool) {
		id, err := uuid.Parse(s)
		if err != nil {
			return "", false
		}
		return d.RefFor(id)
	})
	if err != nil {
		// Params that decoded for the dispatch always re-encode; keep
		// them verbatim if they somehow do not.
		params = a.Params
	}
	label, source := describeLoopStep(g, a, ld.source)
	g.CommitLoopStep(d, game.LoopStep{
		Actor:  a.Caller,
		Admin:  a.Admin,
		Player: a.Player,
		Type:   string(a.Type),
		Params: string(params),
		Label:  label,
		Source: source,
	})
}

// ResolveLoopStep rebuilds the action a recorded step was, for one
// replayed iteration: every reference in its params resolved against
// that iteration's births. ok is false when a reference names nothing
// in this iteration — a divergence (ADR 0193 §5).
func ResolveLoopStep(step game.LoopStep, births *game.LoopBirths) (Action, error) {
	var missing string
	params, err := rewriteParamIDs(json.RawMessage(step.Params), func(s string) (string, bool) {
		id, isRef, ok := births.Resolve(s)
		if !isRef {
			return "", false
		}
		if !ok {
			missing = s
			return s, true
		}
		return id.String(), true
	})
	if err != nil {
		return Action{}, fmt.Errorf("%w: %v", ErrLoopStepUnreadable, err)
	}
	if missing != "" {
		return Action{}, ErrLoopReferenceGone
	}
	return Action{
		Type:   Type(step.Type),
		Player: step.Player,
		Caller: step.Actor,
		Admin:  step.Admin,
		Params: params,
	}, nil
}

// ErrLoopReferenceGone is a replayed step naming an object the
// iteration did not create.
var ErrLoopReferenceGone = errors.New("actions: a step names something this iteration did not create")

// ErrLoopStepUnreadable is a recorded step whose params are not JSON.
var ErrLoopStepUnreadable = errors.New("actions: a recorded step could not be read")

// rewriteParamIDs walks a params object and replaces every string `fn`
// claims. Numbers are kept as written.
func rewriteParamIDs(raw json.RawMessage, fn func(string) (string, bool)) (json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return raw, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	changed := false
	var walk func(any) any
	walk = func(x any) any {
		switch t := x.(type) {
		case string:
			if s, ok := fn(t); ok {
				changed = true
				return s
			}
			return t
		case []any:
			for i := range t {
				t[i] = walk(t[i])
			}
			return t
		case map[string]any:
			for k := range t {
				t[k] = walk(t[k])
			}
			return t
		}
		return x
	}
	v = walk(v)
	if !changed {
		return raw, nil
	}
	return json.Marshal(v)
}

// loopSourceName is the public name of the card an action is about, if
// it names one the table can see right now.
func loopSourceName(g *game.Game, a Action) string {
	var p struct {
		CardID       string `json:"card_id"`
		InstanceID   string `json:"instance_id"`
		SourceCardID string `json:"source_card_id"`
	}
	_ = json.Unmarshal(a.Params, &p)
	for _, s := range []string{p.CardID, p.InstanceID, p.SourceCardID} {
		if id, err := uuid.Parse(s); err == nil {
			if name := g.LoopPublicName(id); name != "" {
				return name
			}
		}
	}
	return ""
}

// describeLoopStep is a step's label and the public name of the card it
// is about: as it is now, or else `before`, as it was before the action.
func describeLoopStep(g *game.Game, a Action, before string) (label, source string) {
	who := g.LoopPublicName(a.Player)
	if who == "" {
		who = g.LoopPublicName(a.Caller)
	}
	if who == "" {
		who = "The admin"
	}
	source = loopSourceName(g, a)
	if source == "" {
		source = before
	}
	card := source
	if card == "" {
		card = "a card"
	}
	switch a.Type {
	case TypePassPriority:
		return who + " passes", ""
	case TypeActivateAbility:
		return who + " activates " + card, source
	case TypeActivateManaAbility:
		return who + " activates " + card + " for mana", source
	case TypeCastSpell:
		return who + " casts " + card, source
	case TypeResolveChoice:
		return who + " answers a prompt", source
	case TypeSacrificePermanent:
		return who + " sacrifices " + card, source
	}
	label = who + ": " + strings.ReplaceAll(string(a.Type), "_", " ")
	if source != "" {
		label += " " + source
	}
	return label, source
}

// dispatchLoopVerb runs one of the loop verbs.
func dispatchLoopVerb(g *game.Game, a Action) error {
	if a.Player == uuid.Nil {
		return ErrInvalidPlayer
	}
	switch a.Type {
	case TypeLoopRecordStart:
		return g.StartLoopRecording(a.Player)
	case TypeLoopRecordStop:
		return g.StopLoopRecording(a.Player)
	case TypeLoopRecordCancel:
		return g.CancelLoopRecording(a.Player)
	case TypeLoopPropose:
		var p struct {
			Count int `json:"count"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		return g.ProposeLoop(a.Player, p.Count)
	case TypeResolveLoopProposal:
		var p struct {
			ChoiceID string          `json:"choice_id"`
			Accept   bool            `json:"accept,omitempty"`
			Shorten  *game.LoopPoint `json:"shorten,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		choiceID, err := uuid.Parse(p.ChoiceID)
		if err != nil {
			return fmt.Errorf("resolve_loop_proposal choice_id: %w", err)
		}
		if p.Accept == (p.Shorten != nil) {
			return fmt.Errorf("%w: resolve_loop_proposal takes accept or shorten, not both", ErrMissingParams)
		}
		return g.ResolveLoopProposal(choiceID, a.Player, p.Shorten)
	}
	return ErrUnknownType
}
