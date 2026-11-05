package game

import (
	"errors"

	"github.com/google/uuid"
)

// entry_attack.go — "If a creature you control would enter while you
// control an attacking creature, instead it enters tapped and
// attacking" (Kira Nerys, Firebrand). Issue #4973.
//
//	CR 614.1d  Continuous effects that read "[This permanent] enters . . ."
//	           or "[Objects] enter [the battlefield] . . ." are
//	           replacement effects.
//	CR 614.12a If a replacement effect that modifies how a permanent
//	           enters the battlefield requires a choice, that choice is
//	           made before the permanent enters the battlefield.
//	CR 508.4   If a creature is put onto the battlefield attacking, its
//	           controller chooses which defending player, planeswalker a
//	           defending player controls, or battle a defending player
//	           protects it's attacking as it enters the battlefield
//	           (unless the effect that put it onto the battlefield
//	           specifies what it's attacking). [...] Such creatures are
//	           "attacking" but, for the purposes of trigger events and
//	           effects, they never "attacked."
//	CR 508.4a  If a creature would be put onto the battlefield attacking
//	           a certain player, and that player is no longer in the
//	           game, the creature is put onto the battlefield but is
//	           never considered an attacking creature. [...]
//
// So the replacement does two things to the entry event: it sets
// EntersTapped, and it fills ReplacementEvent.EntersAttacking with the
// defender the entering creature's controller chooses. The choice is a
// question asked INSIDE the CR 614 window (CR 614.12a), the way
// entry_controller.go's and read_ahead.go's are, and it pauses the
// entry through the same replacementResume frame:
//
//	applyReplacementsLocked sees EntryAttack != nil
//	  → offerEntryAttackLocked sets EntersTapped and either settles the
//	    defender inline (none or one to choose from, or an entry that
//	    cannot pause) or queues PendingChoiceEntryAttackTarget and bails
//	  → ResolveEntryAttackTarget writes ev.EntersAttacking and re-enters
//	    the apply-loop
//	  → the shared finisher lands the creature, and
//	    stampEntryAttackerLocked makes it an attacker that was never
//	    declared (CR 506.3c, 508.4)
//
// The defenders offered are AttackTargetsForEffect of the chooser, with
// the one the chooser's attacking creatures are already attacking first:
// it is the default an entry that cannot pause takes, and the bot's
// answer. With nothing to attack the creature still enters tapped, and
// not attacking (CR 508.4a's outcome).
//
// The effect is mandatory, so a window that cannot ask settles it on
// the default rather than skipping it (entry_controller's rule).

// PendingChoiceEntryAttackTarget is CR 508.4's question for a creature a
// replacement effect puts onto the battlefield attacking: "which player,
// planeswalker or battle does it attack?" Answered with
// {option_index: N}, the payload option_pick and entry_controller take,
// routed by kind for their reason (the meaningful value is zero). The
// options are the defenders, carried in PickOptions: a player's option
// has Player set, a planeswalker's or battle's has the permanent as its
// one card.
//
// A kind of its own rather than an option_pick for entry_controller's
// reason: its continuation is a paused entry (replacementResume), and
// it resumes through the replacement pipeline, not an option chain.
//
// The chooser is the entering creature's would-be controller.
const PendingChoiceEntryAttackTarget PendingChoiceKind = "entry_attack_target"

// EntryAttackChoice is the declaration on ReplacementEffect: "instead it
// enters tapped and attacking" (CR 614.1d), with CR 508.4's choice of
// what it attacks. Data: the card's AppliesTo says WHEN it applies, and
// the engine does the rest. Built by effects.EntersTappedAndAttacking.
//
// The effect's own Replace is never called.
type EntryAttackChoice struct {
	// Question heads the prompt. Falls back to the effect's
	// PromptQuestion and then its Label.
	Question string
}

// entryAttackCandidatesLocked is what the entering creature may attack:
// every legal attack target for `chooser` (CR 508.4: a defending player,
// a planeswalker a defending player controls, a battle a defending
// player protects), with the target the chooser's attacking creatures
// are attacking first.
//
// Caller must hold g.mu.
func (g *Game) entryAttackCandidatesLocked(chooser uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, ref := range g.AttackTargetsForEffect(chooser) {
		out = append(out, ref.ID)
	}
	prefer := uuid.Nil
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == chooser && c.AttackingTarget != uuid.Nil && c.AttackingTarget != AttackingNothing {
			prefer = c.AttackingTarget
			break
		}
	}
	for i, id := range out {
		if id == prefer && i > 0 {
			rest := append([]uuid.UUID{id}, out[:i]...)
			return append(rest, out[i+1:]...)
		}
	}
	return out
}

// entryAttackOptionsLocked is the prompt's option list, one per
// defender in candidate order.
//
// Caller must hold g.mu.
func (g *Game) entryAttackOptionsLocked(targets []uuid.UUID) []ChoiceOption {
	out := make([]ChoiceOption, 0, len(targets))
	for _, id := range targets {
		if g.classifyAttackTargetLocked(id) == AttackTargetPlayer {
			out = append(out, ChoiceOption{Label: "Attack " + g.seatLabelLocked(id), Player: id})
			continue
		}
		name := "that permanent"
		if c := findBattlefieldCard(g, id); c != nil && c.Name != "" {
			name = c.Name
		}
		out = append(out, ChoiceOption{Label: "Attack " + name, Cards: []uuid.UUID{id}})
	}
	return out
}

// entryAttackOptionTarget is the defender an option names.
func entryAttackOptionTarget(opt ChoiceOption) uuid.UUID {
	if opt.Player != uuid.Nil {
		return opt.Player
	}
	if len(opt.Cards) == 1 {
		return opt.Cards[0]
	}
	return uuid.Nil
}

// offerEntryAttackLocked applies an "enters tapped and attacking"
// replacement: the entry is made tapped, and CR 508.4's defender is
// settled inline or asked. Returns true when the prompt was queued and
// the caller must bail.
//
// It settles inline, with no prompt, when there is nothing to choose
// (no defender, or one) or nobody to ask (an entry that cannot pause, a
// chooser who has left or cannot be named): the first candidate, which
// is what the chooser is already attacking when there is such a thing.
//
// Caller must hold g.mu.
func (g *Game) offerEntryAttackLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	g.markReplacementAppliedLocked(ev, chosen.id)
	if ev.Canceled {
		return false
	}
	ev.EntersTapped = true
	if ev.EntersAttacking != uuid.Nil {
		// The effect putting it onto the battlefield already says what
		// it attacks (ninjutsu, a token created attacking), so there is
		// nothing to choose (CR 508.4's "unless the effect that put it
		// onto the battlefield specifies what it's attacking"). A second
		// such replacement lands here too.
		return false
	}
	chooser := g.entryChoicePlayerLocked(ev, chosen)
	candidates := g.entryAttackCandidatesLocked(chooser)
	if len(candidates) == 0 {
		// Nothing it could attack: it enters tapped, never attacking
		// (CR 508.4a).
		return false
	}
	if len(candidates) == 1 || ev.mustSettleNow || !g.optionalReplacementResumableLocked(ev) ||
		chooser == uuid.Nil || g.chooserGoneLocked(chooser) {
		g.setEntryAttackTargetLocked(ev, chooser, candidates[0])
		return false
	}
	question := ""
	if spec := chosen.effect.EntryAttack; spec != nil {
		question = spec.Question
	}
	if question == "" {
		question = chosen.effect.PromptQuestion
	}
	if question == "" {
		question = chosen.effect.Label
	}
	var source uuid.UUID
	if chosen.source != nil {
		source = chosen.source.InstanceID
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:                 PendingChoiceEntryAttackTarget,
		Chooser:              chooser,
		FromPlayer:           chooser,
		Count:                1,
		Source:               source,
		Reason:               question,
		PickOptions:          g.entryAttackOptionsLocked(candidates),
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	})
	if id == uuid.Nil {
		// QueueChoiceForEffect refused the chooser (#864): the default
		// stands rather than a paused entry nobody can resume.
		g.setEntryAttackTargetLocked(ev, chooser, candidates[0])
		return false
	}
	return true
}

// settleEntryAttackByDefaultLocked applies an "enters tapped and
// attacking" effect with no question asked: the path a window that
// cannot ask anybody takes (applyFirstGatheredLocked,
// skipOwnQuestionLocked). The effect is mandatory, so it is applied on
// its default rather than skipped. The caller has already marked it
// applied.
//
// Caller must hold g.mu.
func (g *Game) settleEntryAttackByDefaultLocked(ev *ReplacementEvent, chosen activeReplacement) {
	if ev == nil || ev.Canceled {
		return
	}
	ev.EntersTapped = true
	if ev.EntersAttacking != uuid.Nil {
		return
	}
	chooser := g.entryChoicePlayerLocked(ev, chosen)
	if candidates := g.entryAttackCandidatesLocked(chooser); len(candidates) > 0 {
		g.setEntryAttackTargetLocked(ev, chooser, candidates[0])
	}
}

// setEntryAttackTargetLocked records the defender on the entry event and
// says so: "Ian chose attacking Bob for Grizzly Bears".
//
// Caller must hold g.mu.
func (g *Game) setEntryAttackTargetLocked(ev *ReplacementEvent, chooser, target uuid.UUID) {
	ev.EntersAttacking = target
	if target == uuid.Nil {
		return
	}
	label := "attacking that permanent"
	if g.classifyAttackTargetLocked(target) == AttackTargetPlayer {
		label = "attacking " + g.seatLabelLocked(target)
	} else if c := findBattlefieldCard(g, target); c != nil && c.Name != "" {
		label = "attacking " + c.Name
	}
	g.EmitEvent(Event{Kind: EventOptionChosen, Actor: chooser, CardID: ev.CardID, Target: target, Label: label})
}

// ResolveEntryAttackTarget answers a PendingChoiceEntryAttackTarget:
// `index` is the offset into the prompt's defenders. The defender is
// read off the option picked (#994's rule). A defender that can no
// longer be attacked by the time it is answered — a player who has left,
// a planeswalker that is gone — is still a legal answer, and the
// creature enters tapped but not attacking (CR 508.4a).
//
// An out-of-range index is refused with the prompt still open. Once
// answered, the apply-loop is re-entered so CR 616.1f gathers anything
// left, and the settled entry is landed.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryAttackTarget(choiceID, chooserID uuid.UUID, index int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceEntryAttackTarget {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if index < 0 || index >= len(choice.PickOptions) {
		return ErrInvalidParam
	}
	target := entryAttackOptionTarget(choice.PickOptions[index])
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil {
		return ErrInvalidParam
	}
	if g.dropStaleReplacementResumeLocked(frame) {
		return nil
	}
	ev := frame.ev
	if !ev.Canceled {
		g.setEntryAttackTargetLocked(ev, chooserID, target)
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	return g.finishReplacementResumeLocked(ev, out)
}

// EntryWouldBeCreatureForEffect reports whether the permanent an entry
// event is putting onto the battlefield would be a creature there
// (CR 614.12: its characteristics as it would exist on the battlefield,
// taking into account the replacements already applied to its entry, its
// own static abilities and the continuous effects that already exist).
// A face-down entry is a creature (CR 708.2a). For a replacement's
// AppliesTo that asks "if a creature would enter" (Kira Nerys,
// Firebrand).
//
// Caller must hold g.mu (write), which every replacement gather does:
// the look-ahead's dry pass assigns layer caches on a copy of the
// battlefield.
func (g *Game) EntryWouldBeCreatureForEffect(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.Canceled {
		return false
	}
	if ev.FaceDown != FaceDownNone {
		return true
	}
	perm, ok := g.entryPermanentForTypeClausesLocked(ev)
	return ok && perm.IsCreature()
}

// EntryWouldBeControllerForEffect is the player an entering permanent
// would enter under: the entry's actor, which every battlefield door
// stamps and an entry_controller answer rewrites, or the card's own
// controller and then its owner for an entry that stamped none.
//
// Caller must hold g.mu.
func (g *Game) EntryWouldBeControllerForEffect(ev *ReplacementEvent) uuid.UUID {
	if ev == nil {
		return uuid.Nil
	}
	if ev.Actor != uuid.Nil {
		return ev.Actor
	}
	if c := g.inZoneControllerLocked(ev); c != uuid.Nil {
		return c
	}
	if c, ok := g.LookupCardForEffect(ev.CardID); ok {
		return c.Owner
	}
	return uuid.Nil
}
