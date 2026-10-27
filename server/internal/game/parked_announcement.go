package game

import (
	"fmt"

	"github.com/google/uuid"
)

// parked_announcement.go — #4361 S13, owner decision D7: a cast or an
// activation parked before anything is paid, written into the snapshot.
//
// Two prompts park an announcement and make it again when answered:
//
//   - the CR 903.9b question a cost asks a commander's owner before the
//     payment returns the commander to a hand or a library
//     (cost_commander_choice.go, costCommanderFrame);
//   - "target … of an opponent's choice" (#2055, opponent_target.go,
//     opponentTargetFrame): the controller's "which opponent" option pick
//     and that opponent's pick_target.
//
// Both frames used to differ only in how they re-ran it: the commander
// question held a closure over the caller's params, the opponent's pick
// the params themselves. They now share one value, parkedAnnouncement:
// who announces, what (a spell, or which activated ability of which
// card) and the params the announcement was made with. A re-run is
// castSpellLocked or activateCatalogAbilityLocked on those params, with
// the new answer attached, against whichever game it is handed — so a
// restored table re-runs it exactly as the live one would have.
//
// The snapshot writes it as pendingChoices[].parked: the exported params
// as themselves, and the unexported answers a re-run also reads (the
// commanders' owners' answers, the opponents' picks, the spliced targets,
// an activation's random draws) beside them. The prompt's PickThen names
// the frame (CostCommanderAnnounceKey, OpponentTargetAnnounceKey), so a
// binary from before this refuses the file; the "which opponent" option
// pick keeps its own key (option-pick/opponent-target-chooser).

// CostCommanderAnnounceKey is the PickThen marker of an optional_replacement
// that parks an announcement on a CR 903.9b question (#4361 S13). An
// on-disk key: never renamed, never reused.
const CostCommanderAnnounceKey = "announce/cost-commander"

// OpponentTargetAnnounceKey is the PickThen marker of a pick_target an
// opponent answers for a parked announcement (#4361 S13). An on-disk key:
// never renamed, never reused.
const OpponentTargetAnnounceKey = "announce/opponent-target"

// parkedAnnouncement is a cast or an activation waiting on an answer:
// nothing has been paid or moved. Plain data, and immutable once parked:
// an answer builds a new value (withCommanderAnswers, withOpponentPicks)
// rather than writing this one, so the undo snapshot can share it.
type parkedAnnouncement struct {
	// controller is the player making the announcement.
	controller uuid.UUID
	// cardID is the spell's card, or the source of the ability.
	cardID uuid.UUID
	// name is the spell's or the source's name, for the log.
	name string
	// activation says which: an activated ability (index, activate) or
	// a spell (cast).
	activation bool
	index      int
	activate   ActivateAbilityParams
	cast       CastSpellParams
}

// castAnnouncement is a spell's announcement.
func castAnnouncement(controller, cardID uuid.UUID, name string, p CastSpellParams) parkedAnnouncement {
	return parkedAnnouncement{controller: controller, cardID: cardID, name: name, cast: p}
}

// activationAnnouncement is an activated ability's announcement.
func activationAnnouncement(controller, cardID uuid.UUID, name string, index int, p ActivateAbilityParams) parkedAnnouncement {
	return parkedAnnouncement{controller: controller, cardID: cardID, name: name, activation: true, index: index, activate: p}
}

// opponentPicks is what the announcement already carries.
func (a parkedAnnouncement) opponentPicks() []opponentPick {
	if a.activation {
		return a.activate.opponentPicks
	}
	return a.cast.opponentPicks
}

// withOpponentPicks is the announcement with `picks` attached.
func (a parkedAnnouncement) withOpponentPicks(picks []opponentPick) parkedAnnouncement {
	if a.activation {
		a.activate.opponentPicks = picks
	} else {
		a.cast.opponentPicks = picks
	}
	return a
}

// withCommanderAnswers is the announcement with the CR 903.9 answers
// `answers` attached.
func (a parkedAnnouncement) withCommanderAnswers(answers map[uuid.UUID]bool) parkedAnnouncement {
	if a.activation {
		a.activate.commanderAnswers = answers
	} else {
		a.cast.commanderAnswers = answers
	}
	return a
}

// run makes the announcement again, from scratch.
//
// Caller must hold g.mu.
func (a parkedAnnouncement) run(g *Game) error {
	if a.activation {
		return g.activateCatalogAbilityLocked(a.controller, a.cardID, a.index, a.activate)
	}
	return g.castSpellLocked(a.controller, a.cardID, a.cast)
}

// sameAs reports whether two announcements are of the same thing: one
// source, one ability (or the spell itself).
func (a parkedAnnouncement) sameAs(o parkedAnnouncement) bool {
	return a.cardID == o.cardID && a.activation == o.activation && a.index == o.index
}

// parkedAnnouncementSnapshot is a parked frame as data: the announcement,
// and the frame's own fields. Which frame it is follows from the prompt's
// kind: an optional_replacement's is a CR 903.9b question
// (costCommanderFrame), an option pick's or a pick_target's an opponent's
// target (opponentTargetFrame).
type parkedAnnouncementSnapshot struct {
	Controller uuid.UUID `json:"controller"`
	Card       uuid.UUID `json:"card"`
	Name       string    `json:"name,omitempty"`
	Activation bool      `json:"activation,omitempty"`
	Index      int       `json:"index,omitempty"`
	// Cast is a spell's params, the same data a parked assist prompt
	// writes (AssistAnnounce). Activate is an ability's. Exactly one is
	// set, by Activation.
	Cast     *AssistAnnounce   `json:"cast,omitempty"`
	Activate *ActivateAnnounce `json:"activate,omitempty"`

	// The CR 903.9b question's frame: the commander it asks about, where
	// a decline sends it, and the answers already given for other cards
	// of the same payment.
	Commander uuid.UUID          `json:"commander,omitempty"`
	HeadedFor ZoneKind           `json:"headedFor,omitempty"`
	Answers   map[uuid.UUID]bool `json:"answers,omitempty"`

	// The opponent's target's frame: the clause being asked.
	Mode int `json:"mode,omitempty"`
	Slot int `json:"slot,omitempty"`
}

// ActivateAnnounce is an activation's params as data (#4361 S13): the
// exported fields of ActivateAbilityParams as themselves, and the
// unexported answers a re-run also reads beside them.
type ActivateAnnounce struct {
	Params ActivateAbilityParams `json:"params"`
	// Commander is ActivateAbilityParams.commanderAnswers.
	Commander map[uuid.UUID]bool `json:"commander,omitempty"`
	// Opponent is ActivateAbilityParams.opponentPicks.
	Opponent []AssistOpponentPick `json:"opponent,omitempty"`
	// RandomDiscards is ActivateAbilityParams.randomDiscards, the cards a
	// "discard a card at random" cost drew before the park.
	RandomDiscards []uuid.UUID `json:"randomDiscards,omitempty"`
	// RandomTargets is ActivateAbilityParams.randomTargets, the targets
	// the engine drew at random before the park.
	RandomTargets []TargetRef `json:"randomTargets,omitempty"`
}

func copyCommanderAnswers(in map[uuid.UUID]bool) map[uuid.UUID]bool {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func snapshotOpponentPicks(in []opponentPick) []AssistOpponentPick {
	var out []AssistOpponentPick
	for _, o := range in {
		out = append(out, AssistOpponentPick{
			Mode: o.mode, Slot: o.slot, Chooser: o.chooser, Answered: o.answered,
			Refs: append([]TargetRef(nil), o.refs...),
		})
	}
	return out
}

func restoreOpponentPicks(in []AssistOpponentPick) []opponentPick {
	var out []opponentPick
	for _, o := range in {
		out = append(out, opponentPick{
			mode: o.Mode, slot: o.Slot, chooser: o.Chooser, answered: o.Answered,
			refs: append([]TargetRef(nil), o.Refs...),
		})
	}
	return out
}

func newActivateAnnounce(p ActivateAbilityParams) *ActivateAnnounce {
	out := &ActivateAnnounce{
		Params:         p,
		Commander:      copyCommanderAnswers(p.commanderAnswers),
		Opponent:       snapshotOpponentPicks(p.opponentPicks),
		RandomDiscards: copyUUIDs(p.randomDiscards),
		RandomTargets:  append([]TargetRef(nil), p.randomTargets...),
	}
	if len(out.RandomTargets) == 0 {
		out.RandomTargets = nil
	}
	out.Params.commanderAnswers, out.Params.opponentPicks = nil, nil
	out.Params.randomDiscards, out.Params.randomTargets = nil, nil
	return out
}

func (a *ActivateAnnounce) params() ActivateAbilityParams {
	p := a.Params
	p.commanderAnswers = copyCommanderAnswers(a.Commander)
	p.opponentPicks = restoreOpponentPicks(a.Opponent)
	p.randomDiscards = copyUUIDs(a.RandomDiscards)
	if len(a.RandomTargets) > 0 {
		p.randomTargets = append([]TargetRef(nil), a.RandomTargets...)
	}
	return p
}

// snapshot writes the announcement's half of a frame. ok is false for an
// announcement that cannot be written: a cast already carrying a helper's
// assist answer, which no park reaches before (assist parks after both),
// is left to the census rather than written without it.
func (a parkedAnnouncement) snapshot() (out parkedAnnouncementSnapshot, ok bool) {
	out = parkedAnnouncementSnapshot{
		Controller: a.controller, Card: a.cardID, Name: a.name,
		Activation: a.activation, Index: a.index,
	}
	if a.activation {
		out.Activate = newActivateAnnounce(a.activate)
		return out, true
	}
	if a.cast.assistPaid != nil {
		return parkedAnnouncementSnapshot{}, false
	}
	out.Cast = newAssistAnnounce(a.cast)
	return out, true
}

// announcement reads the announcement back.
func (s *parkedAnnouncementSnapshot) announcement() parkedAnnouncement {
	a := parkedAnnouncement{
		controller: s.Controller, cardID: s.Card, name: s.Name,
		activation: s.Activation, index: s.Index,
	}
	switch {
	case s.Activation && s.Activate != nil:
		a.activate = s.Activate.params()
	case !s.Activation && s.Cast != nil:
		a.cast = s.Cast.params()
	}
	return a
}

// snapshotParkedAnnouncement writes a prompt's parked frame and names it
// on PickThen, or returns nil when the prompt holds none (or holds one
// that cannot be written, which the census then counts).
func snapshotParkedAnnouncement(c *PendingChoice) (*parkedAnnouncementSnapshot, string) {
	switch {
	case c.Kind == PendingChoiceOptionalReplacement && c.costCommanderResume != nil:
		f := c.costCommanderResume
		out, ok := f.announcement.snapshot()
		if !ok {
			return nil, ""
		}
		out.Commander, out.HeadedFor, out.Answers = f.card, f.headedFor, copyCommanderAnswers(f.answers)
		return &out, CostCommanderAnnounceKey
	case c.Kind == PendingChoicePickTarget && c.opponentTargetResume != nil && c.pickTargetResume == nil && c.copyResume == nil:
		f := c.opponentTargetResume
		out, ok := f.snapshot()
		if !ok {
			return nil, ""
		}
		out.Mode, out.Slot = f.mode, f.slot
		return &out, OpponentTargetAnnounceKey
	case c.Kind == PendingChoiceOptionPick && c.opponentTargetResume != nil && c.pickTargetResume == nil && c.PickThen == opponentTargetChooserThen.key:
		f := c.opponentTargetResume
		out, ok := f.snapshot()
		if !ok {
			return nil, ""
		}
		out.Mode, out.Slot = f.mode, f.slot
		// The option pick keeps its own continuation key.
		return &out, c.PickThen
	}
	return nil, ""
}

// restoreParkedAnnouncement puts a written frame back on its prompt.
func restoreParkedAnnouncement(c *PendingChoice, s *parkedAnnouncementSnapshot) {
	switch c.Kind {
	case PendingChoiceOptionalReplacement:
		c.costCommanderResume = &costCommanderFrame{
			card: s.Commander, headedFor: s.HeadedFor, answers: copyCommanderAnswers(s.Answers),
			announcement: s.announcement(),
		}
		c.PickThen = ""
	case PendingChoicePickTarget:
		c.opponentTargetResume = &opponentTargetFrame{parkedAnnouncement: s.announcement(), mode: s.Mode, slot: s.Slot}
		// The pick is for the announcement's controller (CR 601.2c: they
		// still target), which the wire shows.
		c.PickTargetFor = s.Controller
		c.PickThen = ""
	case PendingChoiceOptionPick:
		c.opponentTargetResume = &opponentTargetFrame{parkedAnnouncement: s.announcement(), mode: s.Mode, slot: s.Slot}
	}
}

// unknownParkedAnnouncementKeys is checkEffectKeys' line for a parked
// frame this binary cannot restore: on a kind that does not carry one,
// under the wrong marker, or with the params of the other announcement.
func unknownParkedAnnouncementKeys(c *pendingChoiceSnapshot) []string {
	s := c.Parked
	var want string
	switch c.Kind {
	case PendingChoiceOptionalReplacement:
		want = CostCommanderAnnounceKey
	case PendingChoicePickTarget:
		want = OpponentTargetAnnounceKey
	case PendingChoiceOptionPick:
		want = opponentTargetChooserThen.key
	}
	switch {
	case want == "" || c.PickThen != want:
		return []string{fmt.Sprintf("parked announcement on a %s prompt with tail %q", c.Kind, c.PickThen)}
	case s.Activation != (s.Activate != nil) || s.Activation == (s.Cast != nil):
		return []string{fmt.Sprintf("parked announcement on a %s prompt with no params", c.Kind)}
	}
	return nil
}
