package game

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
)

// auto_answer.go — answering repeated prompts for a player (ADR 0127,
// #1961).
//
// Some cards ask the same yes/no question many times a game: an
// opponent's Rhystic Study asks "pay {1}?" on every spell, a
// Consecrated Sphinx asks "draw two cards?" on every opposing draw. A
// player may answer such a prompt once and for all: Always or Never,
// per card and prompt. This file is the engine half:
//
//   - which prompts are COVERED, decided once when a prompt is queued
//     from the prompt's own facts (§1);
//   - the KEY a standing answer is filed under, derived by the server
//     from the catalog row that asked (§2);
//   - the seat's rules, Player.AutoAnswers, set by set_auto_answers
//     (§3);
//   - NextAutoAnswer and AutoAnswer, which the room calls after every
//     commit to answer a prompt as a commit of its own (§4);
//   - canPayCostLocked, the read-only "could Always pay pay this?"
//     (§5).
//
// An automatic answer is not a player decision: it does not restart
// the CR 732 loop run (§7, loop_breaker.go), and while a loop notice
// is up nothing is answered automatically.

// AutoAnswer is a seat's standing answer to one prompt (ADR 0127 §3).
// Ask is the absence of a rule, so it has no value here.
type AutoAnswer string

const (
	// AutoAnswerAlways answers Yes, or Pay when auto-pay can cover the
	// cost.
	AutoAnswerAlways AutoAnswer = "always"
	// AutoAnswerNever answers No, or Don't pay.
	AutoAnswerNever AutoAnswer = "never"
)

// Valid reports whether a is one of the two standing answers.
func (a AutoAnswer) Valid() bool { return a == AutoAnswerAlways || a == AutoAnswerNever }

// MaxAutoAnswerRules bounds a seat's rules (ADR 0127 §3): 100 rules of
// about 200 bytes each fit inside the account's 32 KiB of settings.
const MaxAutoAnswerRules = 100

// MaxAutoAnswerKeyLen bounds one rule's key, in bytes.
const MaxAutoAnswerKeyLen = 256

// ErrTooManyAutoAnswers refuses a set_auto_answers above
// MaxAutoAnswerRules.
var ErrTooManyAutoAnswers = errors.New("game: too many automatic answers")

// ErrNotAutoAnswerable refuses an AutoAnswer for a prompt that does not
// qualify now (NextAutoAnswer did not name it).
var ErrNotAutoAnswerable = errors.New("game: prompt cannot be answered automatically")

// AskedByHand says why a prompt that has a standing answer is asked by
// hand instead (ADR 0127 §4). The zero value means it is not.
type AskedByHand string

const (
	// AskedByHandNoMana: "Always pay", and auto-pay cannot cover the
	// cost (CR 118.3).
	AskedByHandNoMana AskedByHand = "no_mana"
	// AskedByHandEmptyLibrary: "Always", and the chooser's library is
	// empty (owner decision 9: the engine cannot tell which Yes draws).
	AskedByHandEmptyLibrary AskedByHand = "empty_library"
	// AskedByHandLoop: the CR 732 loop notice is up (§7).
	AskedByHandLoop AskedByHand = "loop"
	// AskedByHandUndone: the chooser undid the automatic answer (§6).
	AskedByHandUndone AskedByHand = "undone"
)

// SetAutoAnswers replaces playerID's standing answers with rules (ADR
// 0127 §3). It is a seat setting, not a play: it emits no event and
// mints no undo entry (actions.MintsNoUndo), and RestoreFrom carries it
// across an undo. A bot seat has none; the room would ignore them
// anyway. Takes the write lock.
func (g *Game) SetAutoAnswers(playerID uuid.UUID, rules map[string]AutoAnswer) error {
	if len(rules) > MaxAutoAnswerRules {
		return ErrTooManyAutoAnswers
	}
	for key, answer := range rules {
		if key == "" || len(key) > MaxAutoAnswerKeyLen || !answer.Valid() {
			return ErrInvalidParam
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.IsBot {
		return ErrInvalidParam
	}
	p.AutoAnswers = copyAutoAnswers(rules)
	return nil
}

// copyAutoAnswers deep-copies a seat's rules; nil for none.
func copyAutoAnswers(in map[string]AutoAnswer) map[string]AutoAnswer {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]AutoAnswer, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// --- §1 and §2: coverage and keys -------------------------------------

// promptKeyState is the transient bookkeeping behind the keys of
// prompts a resolution or a branch queues (ADR 0127 §2). It lives only
// inside one locked call — a resolution function, or an answered
// prompt's branch — and is reset where each of those begins, so it is
// neither cloned nor persisted.
type promptKeyState struct {
	// item is the stack item whose resolution counts belongs to.
	item uuid.UUID
	// counts is how many prompts of each kind that resolution queued.
	counts map[PendingChoiceKind]int
	// branch is set while an answered prompt's branch runs.
	branch *promptBranch
}

// promptBranch is the branch of an answered prompt that is running:
// the answered prompt's key (empty when it had none, which leaves every
// prompt the branch queues without one) and the prompts it has queued.
type promptBranch struct {
	origin string
	counts map[PendingChoiceKind]int
}

// beginPromptKeysForItemLocked resets the resolution's ordinals as an
// item begins to resolve. Caller must hold g.mu.
func (g *Game) beginPromptKeysForItemLocked(item *StackItem) {
	g.promptKeys.item = uuid.Nil
	if item != nil {
		g.promptKeys.item = item.ID
	}
	g.promptKeys.counts = nil
}

// runPromptBranchLocked runs fn as the branch of the answered prompt
// whose key is origin: a prompt fn queues is keyed `origin>kind#n`.
// Caller must hold g.mu.
func (g *Game) runPromptBranchLocked(origin string, fn func()) {
	saved := g.promptKeys.branch
	g.promptKeys.branch = &promptBranch{origin: origin}
	defer func() { g.promptKeys.branch = saved }()
	fn()
}

// payUnlessCovered reports whether a pay_unless can take a standing
// answer (ADR 0127 §1): a plain mana payment with no X and no Phyrexian
// symbol, guarding no spell, owed in no step, and paid with nothing
// but mana.
func payUnlessCovered(c *PendingChoice) bool {
	if c.GuardsStackItem != uuid.Nil || c.OwedInStep != (TurnStep{}) {
		return false
	}
	f := c.payUnlessResume
	if f == nil || f.action != nil || !f.tap.Empty() {
		return false
	}
	return f.cost.XSlots == 0 && !f.cost.HasPhyrexian
}

// confirmCovered reports whether a confirm can take a standing answer:
// a plain "you may", with the default Yes / No labels and no life to
// pay.
func confirmCovered(c *PendingChoice) bool {
	return c.confirmResume != nil && c.LifeCost == 0 && c.AcceptLabel == "" && c.DeclineLabel == ""
}

// triggerPromptCovered reports whether an optional trigger can take a
// standing answer: no target clause, no modes, no trade, and not the
// miracle reveal.
func triggerPromptCovered(t TriggeredAbility) bool {
	if t.Targets != nil || t.TargetsFrom != nil || t.Modes != nil {
		return false
	}
	if t.OptionalPrompt != nil && t.OptionalPrompt.Trade {
		return false
	}
	return t.Keyword != AltCostKeyMiracle
}

// triggerPromptKey is the key of an optional trigger's prompt:
// `<catalog key>|triggered|<row key>`, or "" when the row is not a
// catalog row, the prompt is not covered, or the source is face down.
func triggerPromptKey(source Card, t TriggeredAbility) string {
	if t.row.key == "" || t.Key == "" || source.FaceDown || !triggerPromptCovered(t) {
		return ""
	}
	return t.row.key + "|" + AbilitySlotTriggered + "|" + t.Key
}

// keyResolutionPromptLocked files a pay_unless or confirm that is being
// queued under its key (ADR 0127 §2), when it is covered: the branch's
// origin when an answered prompt's branch is running, else the
// resolving item's row when a resolution function is. Anything else —
// a prompt the engine cannot trace to a catalog row, or one whose
// source is face down — gets no key. Caller must hold g.mu.
func (g *Game) keyResolutionPromptLocked(c *PendingChoice) {
	switch c.Kind {
	case PendingChoicePayUnless:
		if !payUnlessCovered(c) {
			return
		}
	case PendingChoiceConfirm:
		if !confirmCovered(c) {
			return
		}
	default:
		return
	}
	var prefix string
	var counts *map[PendingChoiceKind]int
	switch {
	case g.promptKeys.branch != nil:
		if g.promptKeys.branch.origin == "" {
			return
		}
		prefix = g.promptKeys.branch.origin + ">"
		counts = &g.promptKeys.branch.counts
	case g.resolutionDepth > 0 && g.resolving != nil && g.resolving.item != nil:
		prefix = g.resolvingPromptPrefixLocked()
		if prefix == "" {
			return
		}
		if g.promptKeys.item != g.resolving.item.ID {
			g.beginPromptKeysForItemLocked(g.resolving.item)
		}
		counts = &g.promptKeys.counts
	default:
		return
	}
	name := ""
	if c.Source != uuid.Nil {
		card := g.findCardByIDLocked(c.Source)
		if card == nil {
			if r := g.resolving; r != nil && r.hasCard && r.card.InstanceID == c.Source {
				card = &r.card
			}
		}
		if card != nil {
			if card.FaceDown {
				return
			}
			name = card.Name
		}
	}
	if *counts == nil {
		*counts = map[PendingChoiceKind]int{}
	}
	(*counts)[c.Kind]++
	c.AutoAnswerKey = prefix + string(c.Kind) + "#" + strconv.Itoa((*counts)[c.Kind])
	c.AutoAnswerCard = name
	c.AutoAnswerPrompt = autoAnswerPromptText(c)
}

// resolvingPromptPrefixLocked is the key prefix of the item resolving
// now: its catalog row (`<key>|<slot>|<row label>|`), or for a spell its
// card's catalog key (`<key>|spell|`). Empty for an item with neither,
// and for a face-down spell. Caller must hold g.mu.
func (g *Game) resolvingPromptPrefixLocked() string {
	r := g.resolving
	if r == nil || r.item == nil {
		return ""
	}
	if ab := r.item.Params.Ability; ab != nil && ab.Key != "" {
		return ab.Key + "|" + ab.Slot + "|" + ab.Name + "|"
	}
	if r.hasCard && !r.card.FaceDown {
		if key := CatalogKey(r.card); key != "" {
			return key + "|spell|"
		}
	}
	return ""
}

// autoAnswerPromptText is the question the client shows beside a rule:
// the prompt's own words, or a plain fallback.
func autoAnswerPromptText(c *PendingChoice) string {
	if c.Reason != "" {
		return c.Reason
	}
	if c.Kind == PendingChoicePayUnless {
		return "Pay " + c.PayCost + "?"
	}
	return ""
}

// --- §4: answering ----------------------------------------------------

// NextAutoAnswer is the first prompt in the queue the server may answer
// for its chooser now (ADR 0127 §4), and that chooser. A prompt with a
// rule that fails one of the hand checks — the loop notice, an empty
// library under Always, an Always pay auto-pay cannot cover — is marked
// asked by hand on the way and never looked at again.
//
// "An answer the enumerator would list now" is legal.choiceMoves's
// condition: the game is active, the opening roll and the mulligan are
// over, and the chooser is still in the game. Every prompt it lists for
// a seat is answerable in queue order, which is the order this takes.
//
// Takes the write lock (the mark is a write).
func (g *Game) NextAutoAnswer() (choiceID, chooser uuid.UUID, ok bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive || g.OpeningRoll != nil || g.MulligansOpen {
		return uuid.Nil, uuid.Nil, false
	}
	for _, c := range g.PendingChoices {
		answer, live := g.autoAnswerRuleLocked(c)
		if !live {
			continue
		}
		if reason := g.autoAnswerHandReasonLocked(c, answer); reason != "" {
			c.AskedByHand = reason
			continue
		}
		return c.ID, c.Chooser, true
	}
	return uuid.Nil, uuid.Nil, false
}

// autoAnswerRuleLocked is the chooser's standing answer to c, and
// whether one applies at all: c has a key, its chooser is a human seat
// still in the game with a rule for that key, and it is not asked by
// hand. Caller must hold g.mu.
func (g *Game) autoAnswerRuleLocked(c *PendingChoice) (AutoAnswer, bool) {
	if c == nil || c.AutoAnswerKey == "" || c.AskedByHand != "" {
		return "", false
	}
	p := g.playerByIDLocked(c.Chooser)
	if p == nil || p.Eliminated || p.IsBot {
		return "", false
	}
	answer, ok := p.AutoAnswers[c.AutoAnswerKey]
	if !ok || !answer.Valid() {
		return "", false
	}
	return answer, true
}

// autoAnswerHandReasonLocked is why c must be asked by hand despite its
// rule, or "" when the rule may answer it. Caller must hold g.mu.
func (g *Game) autoAnswerHandReasonLocked(c *PendingChoice, answer AutoAnswer) AskedByHand {
	if g.LoopNotice != nil {
		return AskedByHandLoop
	}
	if answer != AutoAnswerAlways {
		return ""
	}
	p := g.playerByIDLocked(c.Chooser)
	if p == nil || p.Library == nil || p.Library.Size() == 0 {
		return AskedByHandEmptyLibrary
	}
	if c.Kind == PendingChoicePayUnless {
		if c.payUnlessResume == nil || !g.canPayCostLocked(p, c.payUnlessResume.cost, c.Source) {
			return AskedByHandNoMana
		}
	}
	return ""
}

// MarkAskedByHand marks an open prompt to be asked by hand, so the room
// does not answer it automatically again — the room calls it when the
// chooser undoes an automatic answer (ADR 0127 §6). A prompt no longer
// open is ignored. Takes the write lock.
func (g *Game) MarkAskedByHand(choiceID uuid.UUID, reason AskedByHand) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, c := g.findChoiceLocked(choiceID); c != nil && c.AutoAnswerKey != "" {
		c.AskedByHand = reason
	}
}

// AutoAnswer answers the prompt NextAutoAnswer named, with its
// chooser's standing answer (ADR 0127 §4). It checks the prompt again,
// writes the EventAutoAnswer line, and answers through the kind's own
// resolver body — except that the prompt leaves the queue without
// counting as a player decision, so the CR 732 loop run is not
// restarted (§7). Takes the write lock.
func (g *Game) AutoAnswer(choiceID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, c := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	answer, live := g.autoAnswerRuleLocked(c)
	if !live || g.autoAnswerHandReasonLocked(c, answer) != "" {
		return ErrNotAutoAnswerable
	}
	yes := answer == AutoAnswerAlways
	ev := Event{Kind: EventAutoAnswer, Actor: c.Chooser, Source: c.Source}
	switch c.Kind {
	case PendingChoicePayUnless:
		ev.Call = AutoAnswerCallDontPay
		if yes {
			ev.Call = AutoAnswerCallPay
			ev.Label = c.PayCost
		}
	case PendingChoiceTriggerPrompt:
		ev.Call = AutoAnswerCallNo
		if yes {
			ev.Call = AutoAnswerCallYes
		}
		if c.triggerResume != nil {
			ev.Label = c.triggerResume.ability.Key
		}
	case PendingChoiceConfirm:
		ev.Call = AutoAnswerCallNo
		if yes {
			ev.Call = AutoAnswerCallYes
		}
	default:
		return ErrNotAutoAnswerable
	}
	g.EmitEvent(ev)
	switch c.Kind {
	case PendingChoicePayUnless:
		g.answerPayUnlessLocked(idx, c, c.Chooser, yes, nil, nil, false, true)
	case PendingChoiceTriggerPrompt:
		g.answerTriggerPromptLocked(idx, c, yes, true)
	case PendingChoiceConfirm:
		g.answerConfirmLocked(idx, c, c.Chooser, yes, true)
	}
	return nil
}

// The four answers an EventAutoAnswer records, in Event.Call.
const (
	AutoAnswerCallPay     = "pay"
	AutoAnswerCallDontPay = "dont_pay"
	AutoAnswerCallYes     = "yes"
	AutoAnswerCallNo      = "no"
)

// answerChoiceLocked takes an answered prompt out of the queue. A
// player's answer is a decision and restarts the CR 732 loop run
// (dequeueChoiceLocked); an automatic one is not (ADR 0127 §7). Caller
// must hold g.mu.
func (g *Game) answerChoiceLocked(idx int, auto bool) {
	if auto {
		g.removeChoiceAtLocked(idx)
		return
	}
	g.dequeueChoiceLocked(idx)
}

// --- §5: can Always pay pay it? ---------------------------------------

// canPayCostLocked reports whether payCostLocked would pay cost for p
// now: the same spend context, the same cost-as-paid rules and the same
// planner with its pool top-up. It asks by paying on a throwaway clone,
// so the answer is payCostLocked's own and the live game is untouched.
// Caller must hold g.mu.
func (g *Game) canPayCostLocked(p *Player, cost ParsedCost, source uuid.UUID) bool {
	if p == nil {
		return false
	}
	trial := g.cloneLocked()
	tp := trial.playerByIDLocked(p.ID)
	if tp == nil {
		return false
	}
	return trial.payCostLocked(tp, cost, source, nil)
}

// EventAutoAnswer records that the server answered a prompt with its
// chooser's standing answer (ADR 0127 §6). Actor is the chooser, Source
// the card that asked, Call the answer (AutoAnswerCall*), and Label the
// cost paid ("{1}") on a pay, or the trigger's stack label on an
// optional trigger. Projected to the public log as LogAutoAnswer,
// marked "(automatic)": CR 732.1a asks that the table understand each
// player's shortcut.
const EventAutoAnswer EventKind = "auto_answer"
