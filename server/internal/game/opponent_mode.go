package game

import (
	"slices"

	"github.com/google/uuid"
)

// opponent_mode.go — #5771, ADR 0065's 2026-11-19 amendment: "An
// opponent chooses one —" (Fatal Lore, Library of Lat-Nam, Misfortune).
//
// CR 700.2e: "Some spells and abilities specify that a player other than
// their controller chooses a mode for it. In that case, the other player
// does so when the spell or ability's controller normally would do so.
// If there is more than one other player who could make such a choice,
// the spell or ability's controller decides which of those players will
// make the choice." For a spell that moment is CR 601.2b, as it is cast,
// and the targets come after (CR 601.2c): Fatal Lore's ruling is "The
// opponent chooses which option they want (this is a mode choice). If
// they choose the 'destroy' option, you choose the zero, one, or two
// targets after that." The opponent who chose is "that player" in the
// bullets.
//
// # A cast is PARKED, the shape "of an opponent's choice" built
//
// opponent_target.go parks a cast before anything is paid while an
// opponent picks a target, and each answer makes the announcement again
// from scratch with what has been chosen so far attached. A mode is the
// same question one step earlier:
//
//  1. The caster's payload names no mode and no target (the caster
//     cannot know the bullet). castSpellLocked validates the rest of the
//     announcement and, where an opponent's target would park it, parks
//     it on the mode instead.
//  2. If more than one opponent could choose, the caster is asked which
//     (an ordinary choose-a-player option pick); with one, the engine
//     names them. An opponent "could choose" when at least one bullet is
//     choosable with them as "that player" (CR 700.2a: a bullet with no
//     legal target can't be chosen).
//  3. That opponent answers an ordinary mode_pick whose options are the
//     bullets choosable for them. The prompt carries the parked cast as
//     its opponentTargetResume and names the caster (ModePickFor), so the
//     client can say whose spell it is and the bot prices the bullets
//     the other way round.
//  4. The answer casts it again with the modes attached
//     (CastSpellParams.opponentMode). If the chosen bullet targets, its
//     clauses are parked steps the caster is asked for, one pick_target
//     each (parkForModeTargetsLocked), through opponent_target.go's own
//     pick machinery; a clause with nothing to pick and "up to" in its
//     count is answered with nothing without asking.
//  5. The last re-run pays and puts the spell on the stack, with the
//     opponent recorded as StackItem.ModeChooser.
//
// The frames are plain data, so no closure route is added
// (closure_fields.txt), and the snapshot writes them: the mode_pick
// under OpponentModeAnnounceKey, the "which opponent" option pick under
// its own key (option-pick/opponent-mode-chooser), the caster's
// pick_target under OpponentTargetAnnounceKey like any parked pick. A
// table waiting on any of them is a restore point.
//
// # Leaving
//
// The opponent's mode_pick is a choice an object (the spell being cast)
// requires of a player, so if that player leaves it passes to another
// opponent, as CR 800.4g says, by the departure table's ordinary
// inheritor (reassignDepartedChoiceLocked), and the offer is judged
// again for them. A caster who leaves takes the parked cast with them:
// the prompts their card asks are withdrawn with it (CR 800.4a), and an
// answer that arrives after does nothing (rerunParkedAnnouncementLocked).

// OpponentModeAnnounceKey is the PickThen marker of a mode_pick an
// opponent answers for a parked cast (#5771). An on-disk key: never
// renamed, never reused.
const OpponentModeAnnounceKey = "announce/opponent-mode"

// opponentModeAnswer is the mode half of a parked cast whose modes an
// opponent chooses: the opponent the caster named (or the engine, when
// only one could choose), and, once they have answered, their modes.
// Immutable once attached; an answer builds a new one.
type opponentModeAnswer struct {
	chooser  uuid.UUID
	answered bool
	modes    []int
}

// chosenBy is the opponent who chose the modes, or uuid.Nil while they
// have not (or for a cast whose modes are its caster's).
func (a *opponentModeAnswer) chosenBy() uuid.UUID {
	if a == nil || !a.answered {
		return uuid.Nil
	}
	return a.chooser
}

// settleOpponentModes is the cast's CR 601.2b mode step for a spell whose
// modes an opponent chooses. It reports whether the choice is still to
// be made, in which case the cast parks before anything is paid. Once the
// opponent has answered, their modes become the announcement's.
//
// A payload naming modes for such a spell is refused: the caster never
// chooses them. So is an opponent's answer on any other spell, which no
// payload can carry but a parked frame restored onto a changed catalog
// row could.
func settleOpponentModes(ms *ModeSpec, params *CastSpellParams) (bool, error) {
	if ms == nil || !ms.ChosenByOpponent {
		if params.opponentMode != nil {
			return false, ErrInvalidParam
		}
		return false, nil
	}
	m := params.opponentMode
	if m == nil || !m.answered {
		if len(params.Modes) > 0 {
			return false, ErrInvalidParam
		}
		return true, nil
	}
	params.Modes = append([]int(nil), m.modes...)
	return false, nil
}

// markAfterOpponentMode marks the announcement's steps that belong to a
// bullet an opponent chose: the caster's payload could not name them, so
// they are parked steps the caster is asked for (parkForModeTargetsLocked).
// The steps are AnnouncedClauses' per-announcement copies.
func markAfterOpponentMode(ms *ModeSpec, steps []AnnouncedClause) {
	if ms == nil || !ms.ChosenByOpponent {
		return
	}
	for i := range steps {
		if steps[i].Option >= 0 {
			steps[i].Clause.afterOpponentMode = true
		}
	}
}

// withOpponentMode is the announcement with the mode answer `m` attached.
// Only a spell's modes are an opponent's.
func (a parkedAnnouncement) withOpponentMode(m *opponentModeAnswer) parkedAnnouncement {
	if !a.activation {
		a.cast.opponentMode = m
	}
	return a
}

// opponentModeSourceFor is the cast's source with `opp` as the opponent
// choosing the mode: the walk the bullets' clauses are judged on.
func opponentModeSourceFor(src TargetSource, opp uuid.UUID) TargetSource {
	src.modeChooser = opp
	return src
}

// opponentModeOptionsLocked is the bullets `opp` could choose: those
// with a legal target with them as "that player", or with no target at
// all (CR 700.2a).
//
// Caller must hold g.mu.
func (g *Game) opponentModeOptionsLocked(src TargetSource, ms *ModeSpec, opp uuid.UUID) []int {
	return g.choosableModeOptionsLocked(opponentModeSourceFor(src, opp), ms, ModeAbility{})
}

// opponentModeChoosersLocked lists the opponents of the caster who could
// choose the mode, in the order a choose-a-player prompt offers them.
//
// Caller must hold g.mu.
func (g *Game) opponentModeChoosersLocked(src TargetSource, ms *ModeSpec) []uuid.UUID {
	var out []uuid.UUID
	for _, opp := range g.GiftOpponentsLocked(src.Controller) {
		if EnoughChoosableModes(len(g.opponentModeOptionsLocked(src, ms, opp)), ms) {
			out = append(out, opp)
		}
	}
	return g.eligibleChosenPlayersLocked(out)
}

// OpponentCouldChooseModeForEffect reports whether some opponent of the
// cast's controller could choose a mode of `ms`: the question the bot's
// enumerator asks before offering the cast, the same gate the park asks
// (CR 700.2a).
//
// Caller must hold g.mu.
func (g *Game) OpponentCouldChooseModeForEffect(src TargetSource, ms *ModeSpec) bool {
	return len(g.opponentModeChoosersLocked(src, ms)) > 0
}

// parkForOpponentModeLocked parks a validated cast on its opponent's
// mode choice: it asks the caster which opponent, or, once one is named
// (or only one could choose), asks that opponent for the mode. Nothing
// has been paid. A cast no opponent could choose a mode for is refused
// with ErrIllegalTarget (CR 700.2a: no bullet could be chosen), and a
// second park of the same cast while one is open with ErrChoicePending.
//
// Caller must hold g.mu.
func (g *Game) parkForOpponentModeLocked(f *opponentTargetFrame, src TargetSource, ms *ModeSpec) error {
	for _, c := range g.PendingChoices {
		if c != nil && f.sameAnnouncement(c.opponentTargetResume) {
			return ErrChoicePending
		}
	}
	eligible := g.opponentModeChoosersLocked(src, ms)
	if named := f.cast.opponentMode; named != nil && named.chooser != uuid.Nil {
		if !slices.Contains(eligible, named.chooser) {
			return ErrIllegalTarget
		}
		return g.queueOpponentModePickLocked(f, src, ms, named.chooser)
	}
	switch len(eligible) {
	case 0:
		return ErrIllegalTarget
	case 1:
		return g.queueOpponentModePickLocked(f, src, ms, eligible[0])
	}
	id := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  f.controller,
		Source:   f.cardID,
		Question: f.name + ": choose an opponent to choose the mode",
		Options:  g.seatChoiceOptionsLocked(eligible),
		// The parked cast: a second cast of the same card is refused
		// while this is open, and the keyed answer reads it back.
		parked:  f,
		ThenKey: opponentModeChooserThen,
	})
	if id == uuid.Nil {
		return ErrIllegalTarget
	}
	return nil
}

// opponentModeChooserThen is the caster's answer to "which opponent
// chooses the mode" (#5771): the cast is made again with that opponent
// named. The parked cast is the prompt's own opponentTargetResume, so the
// key captures nothing. A dropped prompt (the caster has left, and the
// cast with them, CR 800.4a) does nothing. An on-disk key: never renamed,
// never reused.
var opponentModeChooserThen OptionPickThen

// Registered from init: the continuation can reach code that names the
// key, which a package-level initializer cannot.
func init() {
	opponentModeChooserThen = RegisterOptionPickThen("option-pick/opponent-mode-chooser", opponentModeChooserAnswered)
}

func opponentModeChooserAnswered(g *Game, r OptionPicked) error {
	chosen, frame := chosenSeat(r), r.parked
	if chosen == uuid.Nil || frame == nil {
		return nil
	}
	g.rerunParkedAnnouncementLocked(frame.withOpponentMode(&opponentModeAnswer{chooser: chosen}))
	return nil
}

// parkedCardKeyLocked is the catalog key of the card a parked cast is
// casting, "" when it is in no zone any more.
//
// Caller must hold g.mu.
func (g *Game) parkedCardKeyLocked(f *opponentTargetFrame) string {
	card := g.findCardByIDLocked(f.cardID)
	if card == nil {
		return ""
	}
	return CatalogKey(*card)
}

// queueOpponentModePickLocked asks `chooser` for the cast's mode: the
// bullets choosable with them as "that player", bounded as the caster's
// own choice would be.
//
// Caller must hold g.mu.
func (g *Game) queueOpponentModePickLocked(f *opponentTargetFrame, src TargetSource, ms *ModeSpec, chooser uuid.UUID) error {
	options := g.opponentModeOptionsLocked(src, ms, chooser)
	if !EnoughChoosableModes(len(options), ms) {
		return ErrIllegalTarget
	}
	modeMin, modeMax := g.modeBoundsLocked(ms, ModeCountQuery{
		Chooser:       f.controller,
		OracleID:      g.parkedCardKeyLocked(f),
		OptionalCosts: f.cast.OptionalCosts,
	})
	labels := make([]string, 0, len(options))
	for _, i := range options {
		labels = append(labels, ms.Options[i].Label)
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoiceModePick,
		Chooser: chooser,
		Count:   1,
		Source:  f.cardID,
		// The caster's spell, and the card's own "choose one" (no
		// printed card asks an opponent for more than one).
		Reason:               g.seatLabelLocked(f.controller) + "'s " + f.name + ": choose one",
		ModeOptionIndex:      options,
		ModeOptionLabel:      labels,
		ModeMin:              modeMin,
		ModeMax:              modeMax,
		ModeRepeatable:       ms.Repeatable,
		ModePickFor:          f.controller,
		opponentTargetResume: f,
	})
	if id == uuid.Nil {
		return ErrIllegalTarget
	}
	return nil
}

// resolveOpponentModeLocked answers a parked cast's opponent's mode
// pick: the modes have passed the prompt's own gate (validateModePick),
// and the cast is made again with them attached. The answering player is
// the prompt's chooser now, which a departure may have changed (CR
// 800.4g).
//
// Caller must hold g.mu.
func (g *Game) resolveOpponentModeLocked(idx int, f *opponentTargetFrame, chooserID uuid.UUID, modes []int) {
	g.dequeueChoiceLocked(idx)
	a := f.withOpponentMode(&opponentModeAnswer{chooser: chooserID, answered: true, modes: append([]int(nil), modes...)})
	if !g.rerunParkedAnnouncementLocked(a) {
		return
	}
	ms := ModeSpecFor(g.parkedCardKeyLocked(f))
	if ms == nil {
		return
	}
	// The log says what was chosen and by whom: the caster did not choose
	// it, so their cast line alone would misreport it.
	for _, m := range modes {
		if m >= 0 && m < len(ms.Options) {
			g.EmitEvent(Event{
				Kind:   EventOptionChosen,
				Actor:  chooserID,
				CardID: f.cardID,
				Label:  ms.Options[m].Label,
			})
		}
	}
}

// parkForModeTargetsLocked asks the caster for the targets of step `idx`,
// a clause of the bullet an opponent chose (#5771). The caster is the
// player targeting (CR 601.2c), so the pick is theirs, and its legal set
// is judged with the opponent as "that player". A clause that has no
// legal target and may be left empty ("up to two") is answered with
// nothing at once; one that needs a target and has none refuses the cast
// (CR 700.2a), which only a board changed since the mode was offered can
// reach.
//
// Caller must hold g.mu.
func (g *Game) parkForModeTargetsLocked(f *opponentTargetFrame, src TargetSource, opp opponentAnnouncement, idx int) error {
	step := opp.bound[idx]
	if _, ok := g.opponentStepLegalLocked(src, opp.bound, idx, opp.targets, f.controller); !ok {
		if step.Clause.Min > 0 {
			return ErrIllegalTarget
		}
		picks := withOpponentPick(f.picks(), opponentPick{mode: step.Mode, slot: step.Slot, chooser: f.controller, answered: true})
		return f.rerun(g, picks)
	}
	return g.queueOpponentTargetPickLocked(f, src, opp, idx, f.controller)
}

// isOpponentModePick reports whether a prompt is an opponent's mode
// choice for a parked cast.
func (c *PendingChoice) isOpponentModePick() bool {
	return c != nil && c.Kind == PendingChoiceModePick && c.opponentTargetResume != nil
}

// OpponentModeSpellName is, on an opponent's mode choice for a parked
// cast, the name of the spell being cast, for the wire: the card is still
// in its caster's hand while the cast is parked, where the chooser cannot
// see it, but the spell is public from the moment it is cast (CR 601.2a).
// "" on every other prompt.
func (c *PendingChoice) OpponentModeSpellName() string {
	if !c.isOpponentModePick() {
		return ""
	}
	return c.opponentTargetResume.name
}

// opponentModePickSpecLocked is ModePickSpecLocked for an opponent's mode
// choice: the spell's ModeSpec and its cast's source with the prompt's
// chooser as "that player", so the wire projects each bullet's legal
// targets as the chooser would face them.
//
// Caller must hold g.mu.
func (g *Game) opponentModePickSpecLocked(c *PendingChoice) (TargetSource, *ModeSpec, bool) {
	f := c.opponentTargetResume
	card := g.findCardByIDLocked(f.cardID)
	if card == nil {
		return TargetSource{}, nil, false
	}
	ms := ModeSpecFor(CatalogKey(*card))
	if ms == nil || !ms.ChosenByOpponent {
		return TargetSource{}, nil, false
	}
	return opponentModeSourceFor(SourceCast(f.controller, card), c.Chooser), ms, true
}

// refreshOpponentModeOfferLocked judges an opponent's mode pick again for
// the player it has just passed to (CR 800.4g): the bullets choosable
// with them as "that player". It reports false when they can choose
// none, and the prompt is then not handed to them.
//
// Caller must hold g.mu.
func (g *Game) refreshOpponentModeOfferLocked(c *PendingChoice, newChooser uuid.UUID) bool {
	f := c.opponentTargetResume
	card := g.findCardByIDLocked(f.cardID)
	if card == nil {
		return false
	}
	ms := ModeSpecFor(CatalogKey(*card))
	if ms == nil {
		return false
	}
	src := SourceCast(f.controller, card)
	src.X = f.cast.XValue
	options := g.opponentModeOptionsLocked(src, ms, newChooser)
	if !EnoughChoosableModes(len(options), ms) {
		return false
	}
	labels := make([]string, 0, len(options))
	for _, i := range options {
		labels = append(labels, ms.Options[i].Label)
	}
	c.ModeOptionIndex, c.ModeOptionLabel = options, labels
	return true
}
