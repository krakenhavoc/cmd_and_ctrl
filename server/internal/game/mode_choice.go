package game

import "github.com/google/uuid"

// mode_choice.go — #764 / ADR 0065 §4: the mode choice of a
// TRIGGERED ability, made as the ability is put on the stack
// (CR 603.3c), after the CR 603.5 "you may" prompt and before the
// CR 603.3d target choice.
//
// A spell and an activated ability need no prompt: their modes are
// announced by the player who casts or activates them, in the same
// message (CR 601.2b, CR 602.2b). A trigger has nobody to ask,
// because the engine is what puts it on the stack — which is the
// whole reason this kind exists. Approximating it with a chain of
// confirm prompts at RESOLUTION would ask a priority round too late
// and give nobody a window to respond to the announced mode.

// PendingChoiceModePick is "choose one —" for a triggered ability,
// answered with the list of chosen OPTION INDEXES in the order they
// were chosen. Repeats are legal when the ability's ModeSpec is
// Repeatable (CR 700.2d); the count must fall within Min..Max.
//
// The name was reserved for this at the top of pending_choice.go in
// S20, when Spec.Modes shipped for spells only.
//
// Only the options that can actually be taken are offered: an option
// whose target clause has no legal target is dropped before the
// prompt is queued (CR 603.3d), and if that leaves too few to fill
// Min the trigger is removed from the stack without any prompt at
// all.
const PendingChoiceModePick PendingChoiceKind = "mode_pick"

// modePickFrame is the server-only continuation for a mode_pick: the
// captured trigger, so the answer can continue into the CR 603.3d
// target walk and then Build.
type modePickFrame struct {
	// tc is the triggering event (#1223) — see pickTargetFrame.tc.
	tc        TriggerContext
	source    Card
	lki       Characteristic
	ability   TriggeredAbility
	doubledBy doublerRef
}

// queueModePickLocked queues the CR 603.3c mode choice for a modal
// triggered ability. Returns false when the ability cannot be put on
// the stack at all — too few choosable options to fill Min, which for
// a Repeatable spec means none at all — in which case the caller
// drops the trigger (CR 603.3d).
//
// Caller must hold g.mu.
func (g *Game) queueModePickLocked(tc TriggerContext, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef) bool {
	ms := t.Modes
	options := g.choosableModeOptionsLocked(SourceObject(source.Controller, &source), ms)
	if !EnoughChoosableModes(len(options), ms) {
		return false
	}
	labels := make([]string, 0, len(options))
	for _, i := range options {
		labels = append(labels, ms.Options[i].Label)
	}
	prompt := ms.Prompt
	if prompt == "" {
		prompt = "Choose one"
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:            PendingChoiceModePick,
		Chooser:         source.Controller,
		Count:           1,
		Source:          source.InstanceID,
		Reason:          prompt,
		ModeOptionIndex: options,
		ModeOptionLabel: labels,
		ModeMin:         ms.Min,
		ModeMax:         ms.Max,
		ModeRepeatable:  ms.Repeatable,
		modePickResume: &modePickFrame{
			tc:        tc,
			source:    source,
			lki:       lki,
			ability:   t,
			doubledBy: doubledBy,
		},
	})
	return true
}

// ResolveModePick processes the controller's mode choice for a
// PendingChoiceModePick. `modes` are OPTION indexes in the ModeSpec,
// in the order chosen; every one must be on the prompt's offer list,
// the count must fall within Min..Max, and a repeat is legal only
// when the ability is Repeatable (CR 700.2d).
//
// On success the flow continues where a non-modal trigger already
// is: the CR 603.3d target walk for a targeted ability, or Build for
// an untargeted one.
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolveModePick(choiceID, chooserID uuid.UUID, modes []int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceModePick {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if err := validateModePick(choice, modes); err != nil {
		return err
	}
	frame := choice.modePickResume
	g.dequeueChoiceLocked(idx)
	if frame == nil {
		g.runStateChecksLocked()
		return nil
	}
	g.buildOrPickTriggerLocked(frame.tc, frame.source, frame.lki, frame.ability, frame.doubledBy, append([]int(nil), modes...))
	// Answering is the moment the ability is put on the stack (CR
	// 603.3), so drain and run SBAs here rather than waiting for the
	// next pass around the table — the same reason
	// ResolveTriggerPrompt does.
	g.runStateChecksLocked()
	return nil
}

// validateModePick is the answer gate, kept separate so the bot's
// enumerator and the tests can reason about it directly.
func validateModePick(choice *PendingChoice, modes []int) error {
	offered := make(map[int]bool, len(choice.ModeOptionIndex))
	for _, i := range choice.ModeOptionIndex {
		offered[i] = true
	}
	seen := make(map[int]bool, len(modes))
	for _, m := range modes {
		if !offered[m] {
			return ErrInvalidParam
		}
		if seen[m] && !choice.ModeRepeatable {
			return ErrInvalidParam
		}
		seen[m] = true
	}
	if len(modes) < choice.ModeMin || (choice.ModeMax > 0 && len(modes) > choice.ModeMax) {
		return ErrInvalidParam
	}
	return nil
}

// ModePickSelections lists every legal answer to a mode_pick prompt,
// in a deterministic order, capped at `budget`. The bot's move
// enumerator uses it, and so does the client's "choose for me".
//
// The order is the enumeration policy of ADR 0065 §6: the
// all-one-option selections first (so a repeatable prompt with one
// legal option is never crowded out by mixed multisets), then the
// mixed ones in index order. Only options the prompt offered appear,
// which is already "prefer the modes that have legal targets" — an
// option with no legal target never reached the prompt.
func ModePickSelections(choice *PendingChoice, budget int) [][]int {
	if choice == nil || len(choice.ModeOptionIndex) == 0 || budget <= 0 {
		return nil
	}
	lo, hi := choice.ModeMin, choice.ModeMax
	if hi <= 0 || (!choice.ModeRepeatable && hi > len(choice.ModeOptionIndex)) {
		hi = len(choice.ModeOptionIndex)
	}
	if lo < 0 {
		lo = 0
	}
	var out [][]int
	add := func(sel []int) bool {
		out = append(out, append([]int(nil), sel...))
		return len(out) < budget
	}
	if lo == 0 {
		if !add(nil) {
			return out
		}
	}
	// All-one-option first.
	if choice.ModeRepeatable {
		for _, opt := range choice.ModeOptionIndex {
			for n := max(lo, 1); n <= hi; n++ {
				sel := make([]int, n)
				for i := range sel {
					sel[i] = opt
				}
				if !add(sel) {
					return out
				}
			}
		}
	}
	// Then the distinct combinations, shortest first.
	var rec func(start int, cur []int) bool
	rec = func(start int, cur []int) bool {
		if len(cur) >= max(lo, 1) && len(cur) <= hi {
			if choice.ModeRepeatable && len(cur) == 1 {
				// Already emitted by the all-one-option pass.
			} else if !add(cur) {
				return false
			}
		}
		if len(cur) == hi {
			return true
		}
		for i := start; i < len(choice.ModeOptionIndex); i++ {
			if !rec(i+1, append(cur, choice.ModeOptionIndex[i])) {
				return false
			}
		}
		return true
	}
	rec(0, nil)
	return out
}
