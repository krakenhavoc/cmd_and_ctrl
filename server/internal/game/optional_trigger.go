package game

import "github.com/google/uuid"

// optional_trigger.go — an optional triggered ability's "you may" is
// asked as it resolves (CR 603.5, #2494, ADR 0197).
//
// # The rule
//
// CR 603.5: an optional triggered ability goes on the stack whether or
// not its controller means to take the option, and "the choice is made
// when the ability resolves". Its modes and targets are still chosen as
// it is put on the stack (CR 603.3c, 603.3d), so a "may" whose target
// clause has nothing to point at is removed then, with no question.
//
// Before #2494 the engine asked the question as the ability triggered
// and a "No" dropped it before the stack. The player decided with less
// than the rules give them, and nobody could respond to the trigger
// before the decision. The harvest now builds the item like any other
// and stamps it Optional (stampOptionalTriggerLocked); the
// resolution asks (askOptionalTriggerLocked) and a "Yes" runs the rest
// (answerOptionalTriggerLocked).
//
// # The one "may" that stays early
//
// Miracle's "you may reveal this card as you draw it" (CR 702.94a) is a
// choice the rules make before anything triggers: the reveal is what
// makes the trigger. TriggerOptionalPrompt.BeforeStack keeps it on the
// harvest-time prompt (queueTriggerPromptLocked).
//
// # The prompt is part of the resolution
//
// The question is an ordinary trigger_prompt queued while the item
// resolves, so QueueChoiceForEffect stamps it midResolution: SBAs and
// the trigger drain wait for the answer (resolution_pause.go), and the
// item's log line and stack entry wait with it (resolution_tail.go,
// which parks the item). The prompt names the item by ID
// (PendingChoice.OptionalItem) and finds it on the parked tail, so the
// prompt is plain data: Clone and the snapshot carry it, and a restore
// into the open question still finishes the resolution.

// OptionalTrigger is the CR 603.5 "you may" an optional triggered
// ability carries on the stack: who is asked, and how. Plain data,
// stamped as the item is built and read when it resolves.
type OptionalTrigger struct {
	// Question is the prompt's header (TriggerOptionalPrompt.Question).
	Question string `json:"question,omitempty"`
	// Chooser is the player asked when it is not the item's controller
	// (TriggerOptionalPrompt.Chooser: Edric's "that player may draw").
	// Nil means the controller, read as the item resolves.
	Chooser uuid.UUID `json:"chooser,omitempty"`
	// AnswerKey, AnswerCard and AnswerPrompt are the ADR 0127 standing
	// answer's key and its two labels (triggerPromptKey). Empty when
	// the prompt takes no standing answer.
	AnswerKey    string `json:"answerKey,omitempty"`
	AnswerCard   string `json:"answerCard,omitempty"`
	AnswerPrompt string `json:"answerPrompt,omitempty"`
	// TradeFor is the object a "yes" would trade the source for
	// (TriggerOptionalPrompt.Trade, ADR 0104: Perplexing Chimera's
	// spell), or nil.
	TradeFor uuid.UUID `json:"tradeFor,omitempty"`
}

// cloneOptionalTrigger copies o for Clone and the snapshot.
func cloneOptionalTrigger(o *OptionalTrigger) *OptionalTrigger {
	if o == nil {
		return nil
	}
	out := *o
	return &out
}

// stampOptionalTriggerLocked marks a freshly built item as an optional
// trigger when its row declares a CR 603.5 "you may" asked at
// resolution. Caller must hold g.mu.
func (g *Game) stampOptionalTriggerLocked(item *StackItem, t TriggeredAbility, ev Event, source Card) {
	p := t.OptionalPrompt
	if item == nil || p == nil || p.BeforeStack {
		return
	}
	o := &OptionalTrigger{Question: p.Question}
	if p.Chooser != nil {
		if who := p.Chooser(ev, &source, g); who != uuid.Nil && who != item.Controller {
			o.Chooser = who
		}
	}
	if key := triggerPromptKey(source, t); key != "" {
		o.AnswerKey, o.AnswerCard, o.AnswerPrompt = key, source.Name, p.Question
		if o.AnswerPrompt == "" {
			o.AnswerPrompt = t.Key
		}
	}
	if p.Trade {
		o.TradeFor = ev.CardID
	}
	item.Optional = o
}

// optionalChooser is who answers item's "you may".
func optionalChooser(item *StackItem) uuid.UUID {
	if item.Optional != nil && item.Optional.Chooser != uuid.Nil {
		return item.Optional.Chooser
	}
	return item.Controller
}

// askOptionalTriggerLocked asks an optional item's "you may" as it
// resolves, after the CR 608.2b target check. A chooser who has left
// the game cannot take the option: QueueChoiceForEffect drops the
// prompt, nothing waits, and the item resolves doing nothing.
//
// Caller must hold g.mu in write mode, inside the resolution function.
func (g *Game) askOptionalTriggerLocked(item *StackItem) {
	o := item.Optional
	g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceTriggerPrompt,
		Chooser:          optionalChooser(item),
		Count:            1,
		Source:           item.SourceCardID,
		Reason:           o.Question,
		NoLegalTarget:    g.optionalTriggerHasNoTargetLocked(item),
		AutoAnswerKey:    o.AnswerKey,
		AutoAnswerCard:   o.AnswerCard,
		AutoAnswerPrompt: o.AnswerPrompt,
		OptionalItem:     item.ID,
		OptionalTradeFor: o.TradeFor,
	})
}

// optionalTriggerHasNoTargetLocked is the warning a resolution-time
// prompt carries when its row's HasLegalTarget says "Yes" would do
// nothing. Rows that declare no predicate, and items whose row or
// source cannot be found, carry no warning. Caller must hold g.mu.
func (g *Game) optionalTriggerHasNoTargetLocked(item *StackItem) bool {
	if item.Params.Ability == nil || item.Trigger == nil {
		return false
	}
	row, _, outcome := resolveTriggeredAbilityRef(*item.Params.Ability)
	if outcome == abilityRefLost || row.HasLegalTarget == nil {
		return false
	}
	card := g.findCardByIDLocked(item.SourceCardID)
	if card == nil {
		return false
	}
	return !row.HasLegalTarget(item.Trigger.Event, card, card.Effective(), g)
}

// answerOptionalTriggerLocked applies the answer to a resolution-time
// "you may": "Yes" runs the item's effect and its chosen modes as the
// rest of its resolution; "No" runs nothing. Either way the resolution
// then finishes at the boundary (its log line, CR 704.3).
//
// Caller must hold g.mu.
func (g *Game) answerOptionalTriggerLocked(idx int, choice *PendingChoice, apply, auto bool) {
	itemID := choice.OptionalItem
	g.answerChoiceLocked(idx, auto)
	if apply {
		if item := g.optionalResolvingItemLocked(itemID); item != nil {
			g.runOptionalTriggerEffectLocked(item)
		}
	}
	g.runStateChecksLocked()
}

// optionalResolvingItemLocked is the parked resolution's item when it
// is the one with id, or nil. Caller must hold g.mu.
func (g *Game) optionalResolvingItemLocked(id uuid.UUID) *StackItem {
	t := g.pausedResolutionTail
	if id == uuid.Nil || t == nil || t.item == nil || t.item.ID != id {
		return nil
	}
	return t.item
}

// runOptionalTriggerEffectLocked is the part of resolveTopAbilityLocked
// an optional item skipped while it asked: the effect, then the chosen
// bullets (CR 608.2c), back inside the resolution so a prompt either of
// them asks is part of it. Caller must hold g.mu.
func (g *Game) runOptionalTriggerEffectLocked(item *StackItem) {
	if g.State != StateActive {
		return
	}
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
	if item.Effect != nil {
		if err := item.Effect(g, item); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    item.Controller,
				Source:   item.SourceCardID,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runChosenModeEffectsLocked(item, item.modeSpec, "")
}

// optionalStandingDeclineLocked reports whether item is an optional
// trigger its chooser has a standing "No" for (ADR 0127), so it will
// resolve doing nothing: the CR 603.3b order check reads it as touching
// nothing. The same conditions as the automatic answer itself, except
// that a loop notice, which only asks by hand, cannot make "No" do
// anything. Caller must hold g.mu.
func (g *Game) optionalStandingDeclineLocked(item *StackItem) bool {
	if item == nil || item.Optional == nil || item.Optional.AnswerKey == "" {
		return false
	}
	chooser := optionalChooser(item)
	p := g.playerByIDLocked(chooser)
	if p == nil || p.Eliminated || p.IsBot || g.ControllerOfPlayerLocked(chooser) != chooser {
		return false
	}
	return p.AutoAnswers[item.Optional.AnswerKey] == AutoAnswerNever
}
