package game

import (
	"strconv"

	"github.com/google/uuid"
)

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
// a Repeatable spec means none at all, or a "choose up to X" whose X
// is 0 (#4617) — in which case the caller drops the trigger (CR
// 603.3d, CR 700.2b).
//
// Caller must hold g.mu.
func (g *Game) queueModePickLocked(tc TriggerContext, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef) bool {
	ms := t.Modes
	// ADR 0097: the modes this object's ability has already chosen are
	// not on offer. They ride the prompt as the USED list, so the
	// picker can show them greyed out rather than silently missing.
	ab := ModeAbilityOf(source, t.Key)
	options := g.choosableModeOptionsLocked(SourceObject(source.Controller, &source).WithTrigger(tc), ms, ab)
	if !EnoughChoosableModes(len(options), ms) {
		return false
	}
	// #1590 / #1655: the bounds as they stand NOW, as the trigger goes
	// on the stack (CR 603.3c) — the prompt carries them, so the
	// answer gate and the enumerator read the same numbers. A forced
	// count ("choose both instead" — Prophetic Titan, Depth Defiler)
	// raises the minimum too, but CR 603.3c says a mode with no legal
	// target "can't be chosen" rather than that the ability is
	// removed, so a raised minimum asks for no more than the bullets
	// actually on offer. The printed Min keeps its old meaning, which
	// EnoughChoosableModes has already enforced above.
	q := g.modeQueryForSourceLocked(source, source.Controller)
	// "Choose one at random. If <condition>, you choose one instead":
	// while the condition holds the controller picks, and the prompt
	// below is the ordinary "Choose one".
	chosenInstead := ms.AtRandom && ms.NotRandomIf.holds(g, q)
	if ms.AtRandom && !chosenInstead {
		g.pickModeAtRandomLocked(tc, source, lki, t, doubledBy, options)
		return true
	}
	if ms.UpToCastX {
		// #4617, CR 107.3m: "choose up to X" in an enters trigger reads
		// the X paid for the spell that became the source. `source` is
		// the dispatch's copy of the permanent as it entered, and the
		// entry finisher stamped its CastProvenance before EventETB, so
		// the X is already there; a permanent that was not cast has none.
		q.X = source.CastX()
	}
	modeMin, modeMax := g.modeBoundsLocked(ms, q)
	if !ms.Repeatable && modeMin > len(options) {
		modeMin = max(ms.Min, len(options))
	}
	if ms.UpToCastX {
		if modeMax == 0 {
			// X = 0: "choose up to 0" chooses no mode, and a modal
			// trigger with no mode chosen is removed from the stack
			// (CR 700.2b, CR 603.3c). Nobody is asked.
			return false
		}
		// The bullets on offer are the most that can be chosen (CR
		// 700.2d), so the prompt's count says what is reachable.
		modeMax = min(modeMax, len(options))
	}
	labels := make([]string, 0, len(options))
	for _, i := range options {
		labels = append(labels, ms.Options[i].Label)
	}
	prompt := ms.Prompt
	if prompt == "" || chosenInstead {
		prompt = "Choose one"
	}
	if ms.UpToCastX {
		// The printed "choose up to X" with the X it was cast for.
		prompt += " (X is " + strconv.Itoa(q.X) + ")"
	}
	usedIdx, usedLabel := usedModeOptions(ms, g.modesChosenLocked(ms, ab))
	g.QueueChoiceForEffect(PendingChoice{
		Kind:            PendingChoiceModePick,
		Chooser:         source.Controller,
		Count:           1,
		Source:          source.InstanceID,
		Reason:          prompt,
		ModeOptionIndex: options,
		ModeOptionLabel: labels,
		ModeMin:         modeMin,
		ModeMax:         modeMax,
		ModeRepeatable:  ms.Repeatable,
		ModeUsedIndex:   usedIdx,
		ModeUsedLabel:   usedLabel,
		ModeNotChosen:   ms.NotChosen,
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

// pickModeAtRandomLocked is CR 700.2's "choose one at random" for a
// triggered ability: one of the bullets on offer, drawn from the
// "pick" stream keyed to the controller and the source (so an undo
// rewinds the counter and replays the same draw), and then straight
// into the CR 603.3d target walk, as if the controller had answered a
// mode_pick. Nothing is asked of anyone for the mode itself.
//
// `options` is never empty: EnoughChoosableModes has already
// refused an ability with fewer bullets than Min, and an at-random
// spec has Min 1.
//
// Caller must hold g.mu.
func (g *Game) pickModeAtRandomLocked(tc TriggerContext, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef, options []int) {
	ms := t.Modes
	// The shuffle works on ids; an option's id is its index in the
	// first bytes of an otherwise empty uuid, which is stable and
	// distinct per bullet.
	ids := make([]uuid.UUID, len(options))
	byID := make(map[uuid.UUID]int, len(options))
	for k, i := range options {
		var id uuid.UUID
		id[0], id[1] = byte(i>>8), byte(i)
		id[15] = 1
		ids[k] = id
		byID[id] = i
	}
	picked := g.ChooseAtRandomForEffect(RandomDraw{Player: source.Controller, Source: source.InstanceID}, ids, 1)
	if len(picked) != 1 {
		return
	}
	mode := byID[picked[0]]
	g.recordModesChosenLocked(ms, ModeAbilityOf(source, t.Key), []int{mode})
	g.EmitEvent(Event{
		Kind:   EventOptionChosen,
		Actor:  source.Controller,
		CardID: source.InstanceID,
		Label:  ms.Options[mode].Label + " (at random)",
	})
	g.buildOrPickTriggerLocked(tc, source, lki, t, doubledBy, []int{mode})
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
	// ADR 0097: the answer gate re-checks the LIVE memory, not only the
	// prompt's offer list, so an answer can never name a mode another
	// instance of the same ability took first.
	if frame != nil && g.anyModeChosenLocked(frame.ability.Modes, ModeAbilityOf(frame.source, frame.ability.Key), modes) {
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)
	if frame == nil {
		g.runStateChecksLocked()
		return nil
	}
	// ADR 0097 Decision 4: the mode is used the moment it is chosen,
	// not when the ability resolves — a trigger countered on the stack
	// still spent it. Recorded after the prompt has left the queue, so
	// the re-narrowing of the ability's other open prompts never
	// touches this one.
	g.recordModesChosenLocked(frame.ability.Modes, ModeAbilityOf(frame.source, frame.ability.Key), modes)
	if len(modes) == 0 {
		// A "choose up to" answered with no mode (#4617): "If no mode
		// is chosen, the ability is removed from the stack" (CR 603.3c,
		// CR 700.2b). It is not built and never reaches the stack.
		// Passing the empty answer on would also be read as "not asked
		// yet" (a nil `modes`) and ask the same question again.
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
	// Then the rest: the distinct combinations, or for a repeatable
	// prompt every multiset (CR 700.2d, #2681), mixed ones such as
	// [bounce, bounce, draw] included.
	var rec func(start int, cur []int) bool
	rec = func(start int, cur []int) bool {
		if len(cur) >= max(lo, 1) && len(cur) <= hi {
			if choice.ModeRepeatable && SameModeThroughout(cur) {
				// Already emitted by the all-one-option pass.
			} else if !add(cur) {
				return false
			}
		}
		if len(cur) == hi {
			return true
		}
		for i := start; i < len(choice.ModeOptionIndex); i++ {
			next := i + 1
			if choice.ModeRepeatable {
				next = i
			}
			if !rec(next, append(cur, choice.ModeOptionIndex[i])) {
				return false
			}
		}
		return true
	}
	rec(0, nil)
	return out
}

// SameModeThroughout reports whether a non-empty selection repeats one
// mode: an all-one-option selection.
func SameModeThroughout(sel []int) bool {
	if len(sel) == 0 {
		return false
	}
	for _, m := range sel[1:] {
		if m != sel[0] {
			return false
		}
	}
	return true
}
