package game

import (
	"slices"

	"github.com/google/uuid"
)

// opponent_target.go — #2055, ADR 0019's 2026-10-16 amendment: "target
// … of an opponent's choice" (Nova Pentacle, Evangelize, Volcanic
// Offering, Arena, Cuombajj Witches, Mausoleum Turnkey).
//
// The rulings say the same thing for every such card: as the spell or
// ability is put on the stack, its controller chooses an opponent, and
// that opponent chooses the target (Nova Pentacle, Evangelize, Preacher;
// CR 601.2c via CR 602.2b for an activation, CR 603.3d for a trigger).
// The controller still controls the spell or ability and is still the
// player targeting it, so legality is judged for the controller: an
// opponent cannot name their own hexproof creature (Cuombajj Witches).
// With several such clauses the controller may name the same opponent
// or different ones (Volcanic Offering), and the choice of opponent is
// part of the announcement: a retarget keeps it (Magus of the Arena).
//
// # A spell or an activation is PARKED, the shape #1397 built
//
// castSpellLocked and activateCatalogAbilityLocked validate the whole
// announcement with the opponent's clauses left out (the controller's
// payload names targets for their own clauses only), and then, at the
// point where the CR 903.9 commander question parks an announcement and
// before anything is paid, park it on the first unanswered clause:
//
//  1. If more than one opponent could fill the clause, the controller is
//     asked which opponent (an ordinary choose-a-player option pick).
//     Only opponents who could fill it are offered: an opponent with no
//     creature cannot choose "target creature … they control", and an
//     announcement that names them could not be completed (CR 601.2c).
//     With exactly one such opponent the engine names them.
//  2. That opponent is asked for the clause's targets: an ordinary
//     pick_target prompt, its legal set judged for the controller.
//
// Each answer makes the announcement again from scratch with what has
// been chosen so far attached (CastSpellParams.opponentPicks), so the
// re-run re-validates everything against the game as it is now, and the
// final run pays and puts the item on the stack like any other. Nothing
// is paid or moved while the question is open. A re-run that no longer
// validates is the controller's log line, not the answering opponent's
// error.
//
// The frames are plain data: the closure ratchet (closure_fields.txt)
// gains no route. The choose-an-opponent prompt rides the existing
// option-pick continuation; the pick rides opponentTargetFrame.
//
// # A trigger walks it
//
// A triggered ability's targets are already a server-side walk, one
// pick_target prompt per clause (pickTargetFrame). An opponent's clause
// adds the choose-an-opponent question in front of its step and sends
// the step's prompt to that opponent (queuePickTargetStepLocked).

// TargetChooser records who chose the targets of one clause of an
// announcement: the opponent named for an "of an opponent's choice"
// clause, by the clause's (Mode, Slot) coordinates.
type TargetChooser struct {
	Mode   int       `json:"mode,omitempty"`
	Slot   int       `json:"slot,omitempty"`
	Player uuid.UUID `json:"player"`
}

// copyTargetChoosers is a fresh copy, nil for none.
func copyTargetChoosers(in []TargetChooser) []TargetChooser {
	if len(in) == 0 {
		return nil
	}
	return append([]TargetChooser(nil), in...)
}

// bindStepChoosers writes each recorded chooser onto its clause copy,
// so ControlledByChooser reads "controlled by THAT opponent". The steps
// are the per-announcement value copies AnnouncedClauses hands out.
func bindStepChoosers(steps []AnnouncedClause, choosers []TargetChooser) {
	for _, ch := range choosers {
		for i := range steps {
			if steps[i].Mode == ch.Mode && steps[i].Slot == ch.Slot && steps[i].Clause.ChosenByOpponent {
				steps[i].Clause.chooser = ch.Player
			}
		}
	}
}

// StepChosenByOpponent reports whether a step is an "of an opponent's
// choice" clause, which the controller's own announcement leaves out
// (the client's walk and the bot's enumerator ask the same question).
func StepChosenByOpponent(s AnnouncedClause) bool {
	return s.Clause.ChosenByOpponent
}

// StepsChosenByController is the announcement's steps without the
// clauses an opponent chooses: the ones the controller's payload names
// targets for. The same slice when there are none.
func StepsChosenByController(steps []AnnouncedClause) []AnnouncedClause {
	if !slices.ContainsFunc(steps, StepChosenByOpponent) {
		return steps
	}
	out := make([]AnnouncedClause, 0, len(steps))
	for _, s := range steps {
		if !s.Clause.ChosenByOpponent {
			out = append(out, s)
		}
	}
	return out
}

// opponentPick is one "of an opponent's choice" clause of a parked
// announcement: the opponent named for it, and, once they have answered,
// their picks. answered is false between the controller naming the
// opponent and the opponent choosing.
type opponentPick struct {
	mode, slot int
	chooser    uuid.UUID
	answered   bool
	refs       []TargetRef
}

// withOpponentPick is `picks` with the entry for (mode, slot) replaced
// by `p`, in a fresh slice: a parked frame's params are never written.
func withOpponentPick(picks []opponentPick, p opponentPick) []opponentPick {
	out := make([]opponentPick, 0, len(picks)+1)
	for _, q := range picks {
		if q.mode == p.mode && q.slot == p.slot {
			continue
		}
		out = append(out, q)
	}
	p.refs = append([]TargetRef(nil), p.refs...)
	return append(out, p)
}

func opponentPickFor(picks []opponentPick, mode, slot int) (opponentPick, bool) {
	for _, p := range picks {
		if p.mode == mode && p.slot == slot {
			return p, true
		}
	}
	return opponentPick{}, false
}

func stepIndexOf(steps []AnnouncedClause, mode, slot int) int {
	for i := range steps {
		if steps[i].Mode == mode && steps[i].Slot == slot {
			return i
		}
	}
	return -1
}

// opponentAnnouncement is an announcement's targets with its opponents'
// clauses accounted for.
type opponentAnnouncement struct {
	// bound is every step, each answered opponent clause's chooser
	// bound onto its copy.
	bound []AnnouncedClause
	// check is the steps the announce gate validates now: every step but
	// the opponents' clauses still unanswered.
	check []AnnouncedClause
	// targets is the controller's picks and the answered opponents'
	// picks, in step order: the announcement's Targets.
	targets []TargetRef
	// pending indexes (into bound) the opponents' clauses still to ask.
	pending []int
	// choosers records who chose each answered opponent clause, for the
	// item (StackItem.TargetChoosers).
	choosers []TargetChooser
}

// splitOpponentTargets merges the controller's own picks with the
// opponents' answers. `own` has been slotted against
// StepsChosenByController. A ref of the controller's that names an
// opponent's clause, or an answer for a clause that is not one, is
// refused: the controller never chooses an opponent's target, and
// nothing but a resume carries answers. An announcement with no such
// clause comes back exactly as it went in.
func splitOpponentTargets(steps []AnnouncedClause, own []TargetRef, picks []opponentPick) (opponentAnnouncement, error) {
	if !slices.ContainsFunc(steps, StepChosenByOpponent) {
		if len(picks) > 0 {
			return opponentAnnouncement{}, ErrInvalidParam
		}
		return opponentAnnouncement{bound: steps, check: steps, targets: own}, nil
	}
	var out opponentAnnouncement
	for _, p := range picks {
		i := stepIndexOf(steps, p.mode, p.slot)
		if i < 0 || !steps[i].Clause.ChosenByOpponent || p.chooser == uuid.Nil {
			return opponentAnnouncement{}, ErrInvalidParam
		}
		if p.answered {
			out.choosers = append(out.choosers, TargetChooser{Mode: p.mode, Slot: p.slot, Player: p.chooser})
		}
	}
	out.bound = append([]AnnouncedClause(nil), steps...)
	bindStepChoosers(out.bound, out.choosers)
	var unslotted []TargetRef
	for _, t := range own {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			unslotted = append(unslotted, t)
			continue
		}
		i := stepIndexOf(steps, t.Mode, t.Slot)
		if i < 0 || steps[i].Clause.ChosenByOpponent {
			return opponentAnnouncement{}, ErrInvalidParam
		}
	}
	for i, s := range out.bound {
		if s.Clause.ChosenByOpponent {
			p, ok := opponentPickFor(picks, s.Mode, s.Slot)
			if !ok || !p.answered {
				out.pending = append(out.pending, i)
				continue
			}
			out.check = append(out.check, s)
			for _, r := range p.refs {
				r.Mode, r.Slot = s.Mode, s.Slot
				out.targets = append(out.targets, r)
			}
			continue
		}
		out.check = append(out.check, s)
		for _, t := range own {
			if (t.Kind == TargetPlayer || t.Kind == TargetCard) && t.Mode == s.Mode && t.Slot == s.Slot {
				out.targets = append(out.targets, t)
			}
		}
	}
	out.targets = append(out.targets, unslotted...)
	return out, nil
}

// opponentChoosersForStepLocked lists the opponents of the controller
// who could fill step `idx` given the picks made so far: every opponent
// still in the game for whom the clause, bound to them, has enough legal
// targets. In the order a choose-a-player prompt offers them.
//
// Caller must hold g.mu.
func (g *Game) opponentChoosersForStepLocked(src TargetSource, steps []AnnouncedClause, idx int, picked []TargetRef) []uuid.UUID {
	if idx < 0 || idx >= len(steps) {
		return nil
	}
	var out []uuid.UUID
	for _, opp := range g.GiftOpponentsLocked(src.Controller) {
		if _, ok := g.opponentStepLegalLocked(src, steps, idx, picked, opp); ok {
			out = append(out, opp)
		}
	}
	return g.eligibleChosenPlayersLocked(out)
}

// opponentStepLegalLocked is step `idx`'s legal set with `opp` as its
// chooser, and whether it can be filled.
//
// Caller must hold g.mu.
func (g *Game) opponentStepLegalLocked(src TargetSource, steps []AnnouncedClause, idx int, picked []TargetRef, opp uuid.UUID) (LegalTargets, bool) {
	bound := append([]AnnouncedClause(nil), steps...)
	bound[idx].Clause.chooser = opp
	clause := &bound[idx].Clause
	lt := g.LegalTargetsForAnnouncedClauseForEffect(src, bound, idx, append([]TargetRef(nil), picked...))
	lt = withoutPicked(lt, picked, clause.Distinct)
	need := max(clause.Min, 1)
	return lt, g.fillableCountLocked(clause, lt) >= need
}

// opponentTargetFrame is a parked announcement waiting on an
// opponent's target: what to make again, and which clause is being
// asked. Plain data and immutable once queued; the answer builds new
// params rather than writing these.
type opponentTargetFrame struct {
	controller uuid.UUID
	cardID     uuid.UUID
	name       string
	// activation says which announcement: an activated ability (index,
	// activate) or a spell (cast).
	activation bool
	index      int
	activate   ActivateAbilityParams
	cast       CastSpellParams
	// mode / slot are the clause this prompt asks.
	mode, slot int
}

// picks is what the parked announcement already carries.
func (f *opponentTargetFrame) picks() []opponentPick {
	if f.activation {
		return f.activate.opponentPicks
	}
	return f.cast.opponentPicks
}

// rerun makes the announcement again with `picks` attached.
//
// Caller must hold g.mu.
func (f *opponentTargetFrame) rerun(g *Game, picks []opponentPick) error {
	if f.activation {
		p := f.activate
		p.opponentPicks = picks
		return g.activateCatalogAbilityLocked(f.controller, f.cardID, f.index, p)
	}
	p := f.cast
	p.opponentPicks = picks
	return g.castSpellLocked(f.controller, f.cardID, p)
}

// sameAnnouncement reports whether two frames park the same
// announcement: one source, one ability (or the spell itself).
func (f *opponentTargetFrame) sameAnnouncement(o *opponentTargetFrame) bool {
	return f != nil && o != nil && f.cardID == o.cardID && f.activation == o.activation && f.index == o.index
}

// parkForOpponentTargetLocked parks a validated announcement on its
// first unanswered opponent's clause (opp.pending[0]): it asks the
// controller which opponent, or, once one is named (or only one could
// fill it), asks that opponent for the targets. Nothing has been paid.
// An announcement no opponent could complete is refused with
// ErrIllegalTarget (CR 601.2c: there is no legal choice), and a second
// park of the same announcement while one is open with ErrChoicePending.
//
// Caller must hold g.mu.
func (g *Game) parkForOpponentTargetLocked(f *opponentTargetFrame, src TargetSource, opp opponentAnnouncement) error {
	for _, c := range g.PendingChoices {
		if c != nil && f.sameAnnouncement(c.opponentTargetResume) {
			return ErrChoicePending
		}
	}
	idx := opp.pending[0]
	step := opp.bound[idx]
	f.mode, f.slot = step.Mode, step.Slot
	eligible := g.opponentChoosersForStepLocked(src, opp.bound, idx, opp.targets)
	if named, ok := opponentPickFor(f.picks(), step.Mode, step.Slot); ok && named.chooser != uuid.Nil {
		if !slices.Contains(eligible, named.chooser) {
			return ErrIllegalTarget
		}
		return g.queueOpponentTargetPickLocked(f, src, opp, idx, named.chooser)
	}
	switch len(eligible) {
	case 0:
		return ErrIllegalTarget
	case 1:
		return g.queueOpponentTargetPickLocked(f, src, opp, idx, eligible[0])
	}
	frame := f
	id := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  f.controller,
		Source:   f.cardID,
		Question: f.name + ": choose an opponent to choose " + clauseLabel(&step.Clause),
		Options:  g.seatChoiceOptionsLocked(eligible),
		// A marker: a second announcement of the same thing is refused
		// while this is open (OptionPickPrompt.parked).
		parked: frame,
		ThenSeat: func(g *Game, chosen uuid.UUID) error {
			if chosen == uuid.Nil {
				// Dropped unanswered: the controller has left, and the
				// announcement they were making with them (CR 800.4a).
				return nil
			}
			picks := withOpponentPick(frame.picks(), opponentPick{mode: frame.mode, slot: frame.slot, chooser: chosen})
			g.rerunOpponentTargetLocked(frame, picks)
			return nil
		},
	})
	if id == uuid.Nil {
		return ErrIllegalTarget
	}
	return nil
}

// queueOpponentTargetPickLocked asks `chooser` for step `idx`'s targets.
//
// Caller must hold g.mu.
func (g *Game) queueOpponentTargetPickLocked(f *opponentTargetFrame, src TargetSource, opp opponentAnnouncement, idx int, chooser uuid.UUID) error {
	lt, ok := g.opponentStepLegalLocked(src, opp.bound, idx, opp.targets, chooser)
	if !ok {
		return ErrIllegalTarget
	}
	clause := opp.bound[idx].Clause
	who := g.seatLabelLocked(f.controller)
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:                 PendingChoicePickTarget,
		Chooser:              chooser,
		Count:                1,
		Source:               f.cardID,
		Reason:               who + "'s " + f.name + ": choose " + clauseLabel(&clause),
		PickTargetPlayers:    lt.Players,
		PickTargetCards:      lt.Cards,
		PickTargetMin:        clause.Min,
		PickTargetMax:        clause.Max,
		PickTargetFor:        f.controller,
		opponentTargetResume: f,
	})
	if id == uuid.Nil {
		return ErrIllegalTarget
	}
	return nil
}

// clauseLabel is a clause's printed words for a prompt.
func clauseLabel(c *TargetClause) string {
	if c == nil || c.Label == "" {
		return "a target"
	}
	return c.Label
}

// resolveOpponentTargetsLocked answers a parked announcement's
// opponent's pick: the refs are checked against the prompt's frozen
// legal set and count (a bad answer leaves the prompt open), then the
// announcement is made again with them attached. The answering player is
// the prompt's chooser now, which a departure may have changed (CR
// 800.4g).
//
// Caller must hold g.mu.
func (g *Game) resolveOpponentTargetsLocked(idx int, choice *PendingChoice, f *opponentTargetFrame, chooserID uuid.UUID, targets []TargetRef) error {
	legal := make(map[uuid.UUID]TargetRefKind, len(choice.PickTargetPlayers)+len(choice.PickTargetCards))
	for _, id := range choice.PickTargetPlayers {
		legal[id] = TargetPlayer
	}
	for _, id := range choice.PickTargetCards {
		legal[id] = TargetCard
	}
	seen := make(map[uuid.UUID]bool, len(targets))
	refs := make([]TargetRef, 0, len(targets))
	for _, t := range targets {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
		if kind, ok := legal[t.ID]; !ok || kind != t.Kind {
			return ErrIllegalTarget
		}
		if seen[t.ID] {
			return ErrInvalidParam
		}
		seen[t.ID] = true
		refs = append(refs, TargetRef{Kind: t.Kind, ID: t.ID, Mode: f.mode, Slot: f.slot})
	}
	if len(refs) < choice.PickTargetMin || (choice.PickTargetMax > 0 && len(refs) > choice.PickTargetMax) {
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)
	picks := withOpponentPick(f.picks(), opponentPick{mode: f.mode, slot: f.slot, chooser: chooserID, answered: true, refs: refs})
	g.rerunOpponentTargetLocked(f, picks)
	return nil
}

// rerunOpponentTargetLocked makes a parked announcement again with
// `picks` and reports a refusal to its controller. A controller who has
// left has no announcement to make (CR 800.4a).
//
// Caller must hold g.mu.
func (g *Game) rerunOpponentTargetLocked(f *opponentTargetFrame, picks []opponentPick) {
	if !g.playerInGameLocked(f.controller) {
		return
	}
	if err := f.rerun(g, picks); err != nil {
		what := "cast"
		if f.activation {
			what = "activated"
		}
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    f.controller,
			Source:   f.cardID,
			ErrorMsg: f.name + " couldn't be " + what + ": " + err.Error(),
		})
		return
	}
	g.noteActionTakenLocked(f.controller)
}

// withStepChooser is `choosers` with step s's chooser set to `who`, in
// a fresh slice.
func withStepChooser(choosers []TargetChooser, s AnnouncedClause, who uuid.UUID) []TargetChooser {
	out := make([]TargetChooser, 0, len(choosers)+1)
	for _, ch := range choosers {
		if ch.Mode == s.Mode && ch.Slot == s.Slot {
			continue
		}
		out = append(out, ch)
	}
	return append(out, TargetChooser{Mode: s.Mode, Slot: s.Slot, Player: who})
}

// OpponentStepFillableForEffect reports whether some opponent of the
// walk's controller could choose step idx's targets, given the picks
// made so far: the question the bot's enumerator asks of a clause the
// controller does not pick (#2055), the same gate the announcement
// parks on.
//
// Caller must hold g.mu.
func (g *Game) OpponentStepFillableForEffect(src TargetSource, steps []AnnouncedClause, idx int, picked []TargetRef) bool {
	return len(g.opponentChoosersForStepLocked(src, steps, idx, picked)) > 0
}
