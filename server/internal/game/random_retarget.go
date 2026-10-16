package game

import (
	"strings"

	"github.com/google/uuid"
)

// random_retarget.go — "reselect the targets at random" (#1815), and
// the engine half of "target … chosen at random" on a trigger.
//
// Chef's Kiss: "Gain control of target spell that targets only a
// single permanent or player. Copy it, then reselect the targets at
// random for the spell and the copy. The new targets can't be you or a
// permanent you control."
//
// It is CR 115.7 run by the engine instead of a player. Nothing about
// which targets are legal changes — the same clause, the same
// TargetSource (the ITEM's controller, retargetSourceLocked), the same
// gate (retargetCheckLocked) a "choose new targets" answer passes. What
// changes is who picks: nobody is asked, and each target is drawn
// uniformly from the legal set through the keyed random stream (ADR
// 0054), so an undo replays the same draw.
//
// Three facts from the 2021-06-18 rulings decide the shape:
//
//   - The draw is over ALL the legal targets, the current one
//     included. "Reselect" is not "change to another": Grip of Chaos'
//     reminder text says "(Select from among all legal targets.)" and
//     its ruling "If there are multiple legal targets, it may still
//     choose the original one."
//   - "If there are no legal targets that could be selected, the
//     targets remain unchanged" — per slot, and even if the original
//     target is illegal by now (CR 115.7a's outcome).
//   - "If a spell has multiple instances of the word 'target,' but all
//     of them target the same permanent or player … a new target will
//     be reselected at random for each instance of the word target."
//     So every slot is drawn on its own, in announce order, each draw
//     judged against the list as the earlier draws left it (CR 115.7e:
//     only the final set must be legal, and the last draw's check is
//     the whole list's).

// RandomRetarget is one "reselect the targets at random" instruction.
type RandomRetarget struct {
	// ItemID is the stack item whose targets are reselected. A spell's
	// item id is its card's instance id.
	ItemID uuid.UUID

	// Drawer is the player whose effect draws: the stream is keyed to
	// them and the source (RandomDraw), and the log names them.
	Drawer uuid.UUID

	// Source is the object whose effect this is (Chef's Kiss).
	Source uuid.UUID

	// Exclude is "the new targets can't be you or a permanent you
	// control": this player, and every permanent they control, is left
	// out of every draw. uuid.Nil leaves nothing out.
	Exclude uuid.UUID
}

// ReselectTargetsAtRandomForEffect applies a RandomRetarget.
//
// Outcomes, none of them a prompt:
//
//   - the item is not on the stack: ErrCardNotFound, which a card treats
//     as "nothing happens" (it was countered or resolved in response);
//   - it has no targets: nothing, nil;
//   - a slot with no target that could be selected keeps its target;
//   - otherwise each slot gets a target drawn uniformly from the ones
//     that could be selected, the item is rewritten through
//     RetargetStackItemForEffect (so the new targets become targets, CR
//     115.7, and a divided spell keeps its division, CR 115.7f), and
//     the draw is logged.
//
// Caller must hold g.mu.
func (g *Game) ReselectTargetsAtRandomForEffect(r RandomRetarget) error {
	item := g.StackMeta[r.ItemID]
	if item == nil {
		return ErrCardNotFound
	}
	if countRealTargets(item.Targets) == 0 {
		return nil
	}
	steps := g.itemAnnouncedClauses(item)
	src := g.retargetSourceLocked(item)
	old := item.Targets
	next := append([]TargetRef(nil), old...)
	draw := RandomDraw{Player: r.Drawer, Source: r.Source}
	drawn := 0
	for i, ref := range old {
		if ref.Kind != TargetCard && ref.Kind != TargetPlayer {
			continue
		}
		clause := retargetClauseFor(steps, ref)
		if clause == nil {
			// A free-form S13.1 announcement: nothing says what a
			// legal target for it would be, so it is left alone, as
			// the player-driven retarget leaves it.
			continue
		}
		cands := g.randomRetargetCandidatesLocked(src, steps, clause, old, next, i, r.Exclude)
		if len(cands) == 0 {
			continue
		}
		ids := make([]uuid.UUID, len(cands))
		byID := make(map[uuid.UUID]TargetRef, len(cands))
		for k, c := range cands {
			ids[k] = c.ID
			byID[c.ID] = c
		}
		picked := g.ChooseAtRandomForEffect(draw, ids, 1)
		if len(picked) != 1 {
			continue
		}
		next[i] = byID[picked[0]]
		drawn++
	}
	if len(changedRefs(old, next)) > 0 {
		if err := g.RetargetStackItemForEffect(item.ID, r.Drawer, RetargetChooseNew, next); err != nil {
			return err
		}
	}
	// A draw that landed on the target already there changed nothing
	// and is still a draw: the log says what the spell targets now.
	g.emitRandomRetargetLocked(r, item, next, drawn > 0)
	return nil
}

// randomRetargetCandidatesLocked is what slot `i` could be reselected
// to: every legal target of the slot's clause (the current one
// included), less the excluded player and their permanents, kept only
// when the list with slot `i` replaced by it passes CR 115.7's gate —
// the same-clause duplicate rule, Distinct, the set rules and the
// relations all fall out of retargetCheckLocked rather than being
// restated here. Each candidate carries the slot's (Mode, Slot).
//
// Caller must hold g.mu.
func (g *Game) randomRetargetCandidatesLocked(src TargetSource, steps []AnnouncedClause, clause *TargetClause, old, next []TargetRef, i int, exclude uuid.UUID) []TargetRef {
	lt := g.legalTargetsLocked(src, clause)
	var out []TargetRef
	try := func(ref TargetRef) {
		trial := append([]TargetRef(nil), next...)
		trial[i] = ref
		if g.retargetCheckLocked(src, steps, old, trial, RetargetChooseNew) == nil {
			out = append(out, ref)
		}
	}
	for _, p := range lt.Players {
		if exclude != uuid.Nil && p == exclude {
			continue
		}
		try(TargetRef{Kind: TargetPlayer, ID: p, Mode: old[i].Mode, Slot: old[i].Slot})
	}
	for _, id := range lt.Cards {
		if exclude != uuid.Nil && g.permanentControlledByLocked(id, exclude) {
			continue
		}
		try(TargetRef{Kind: TargetCard, ID: id, Mode: old[i].Mode, Slot: old[i].Slot})
	}
	return out
}

// permanentControlledByLocked reports whether `id` is a permanent on
// the battlefield that `player` controls. Caller must hold g.mu.
func (g *Game) permanentControlledByLocked(id, player uuid.UUID) bool {
	if g.Battlefield == nil || !g.Battlefield.Contains(id) {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Controller == player
}

// emitRandomRetargetLocked logs the outcome of a random reselection:
// "<spell> now targets <names> (at random)", or, when no slot had a
// target that could be selected, that it kept its targets. A
// choose-option line, like a mode drawn at random (#3399).
func (g *Game) emitRandomRetargetLocked(r RandomRetarget, item *StackItem, targets []TargetRef, drawn bool) {
	var names []string
	for _, t := range targets {
		if t.Kind != TargetCard && t.Kind != TargetPlayer {
			continue
		}
		names = append(names, g.targetRefNameLocked(t))
	}
	what := "the spell"
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Name != "" {
		what = c.Name
	}
	label := what + " keeps its targets (no other target could be selected)"
	if drawn && len(names) > 0 {
		label = what + " now targets " + strings.Join(names, ", ") + " (at random)"
	}
	g.EmitEvent(Event{
		Kind:   EventOptionChosen,
		Actor:  r.Drawer,
		CardID: r.Source,
		Label:  label,
	})
}

// targetRefNameLocked is a target's name for a log line: the player's
// name, or the card's. Caller must hold g.mu.
func (g *Game) targetRefNameLocked(t TargetRef) string {
	if t.Kind == TargetPlayer {
		if p := g.playerByIDLocked(t.ID); p != nil && p.Name != "" {
			return p.Name
		}
		return "a player"
	}
	if c, ok := g.LookupCardForEffect(t.ID); ok && c.Name != "" {
		return c.Name
	}
	return "a card"
}

// StackItemSoleTargetForEffect is the CR 115.9c reading of "a spell
// that targets only [something]": the one object or player the item
// targets, and true, when exactly one DIFFERENT object or player was
// chosen as its targets (as modified by effects that changed them) —
// so a spell that chose the same creature for two instances of the
// word "target" still targets only that creature. The caller checks
// that target's current state (CR 115.9b): an object no longer in the
// zone it is expected to be in, or a player who has left the game, is
// ignored, and then the item targets nothing of that kind.
//
// False for an item that is not on the stack, has no target, or
// targets two or more different objects or players.
//
// Caller must hold g.mu.
func (g *Game) StackItemSoleTargetForEffect(itemID uuid.UUID) (TargetRef, bool) {
	item := g.StackMeta[itemID]
	if item == nil {
		return TargetRef{}, false
	}
	var sole TargetRef
	found := false
	for _, t := range item.Targets {
		if t.Kind != TargetCard && t.Kind != TargetPlayer {
			continue
		}
		if !found {
			sole, found = t, true
			continue
		}
		if t.Kind != sole.Kind || t.ID != sole.ID {
			return TargetRef{}, false
		}
	}
	return sole, found
}

// CopySpellKeepingTargetsForEffect is CopySpellForEffect with no "you
// may choose new targets for the copy" (CR 707.10: the copy has the
// original's targets), returning the copy's id so the caller can act
// on it next — Chef's Kiss's "copy it, then reselect the targets at
// random for the spell and the copy". The copy is on the stack when it
// returns: nothing is asked between the copy and its creation.
//
// Errors: ErrCardNotFound when `spellID` is no longer a spell on the
// stack (countered in response — "the copy effect did nothing").
//
// Caller must hold g.mu.
func (g *Game) CopySpellKeepingTargetsForEffect(spellID, controller uuid.UUID) (uuid.UUID, error) {
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok || item.Kind != StackItemSpell {
		return uuid.Nil, ErrCardNotFound
	}
	g.createSpellCopyLocked(src, item, controller, item.Targets)
	top, err := g.Stack.Top()
	if err != nil {
		return uuid.Nil, ErrCardNotFound
	}
	return top.InstanceID, nil
}

// drawTriggerTargetLocked is TargetSpec.AtRandom in the CR 603.3d
// target walk: the frame's current step, whose legal set `lt` is not
// empty, gets one target drawn uniformly from it on the "pick" stream
// keyed to the trigger's controller and source, and the draw is logged
// ("<clause>: <name> (at random)"). Players come first in the draw's
// list and cards after, each in the legal set's own stable order, so
// the same board and the same stream give the same target.
//
// Caller must hold g.mu.
func (g *Game) drawTriggerTargetLocked(f *pickTargetFrame, lt LegalTargets) TargetRef {
	cur := f.steps[f.step]
	ids := make([]uuid.UUID, 0, len(lt.Players)+len(lt.Cards))
	isPlayer := make(map[uuid.UUID]bool, len(lt.Players))
	for _, p := range lt.Players {
		ids = append(ids, p)
		isPlayer[p] = true
	}
	ids = append(ids, lt.Cards...)
	picked := g.ChooseAtRandomForEffect(RandomDraw{Player: f.source.Controller, Source: f.source.InstanceID}, ids, 1)
	ref := TargetRef{Kind: TargetCard, ID: picked[0], Mode: cur.Mode, Slot: cur.Slot}
	if isPlayer[ref.ID] {
		ref.Kind = TargetPlayer
	}
	label := cur.Clause.Label
	if label == "" {
		label = "target"
	}
	g.EmitEvent(Event{
		Kind:   EventOptionChosen,
		Actor:  f.source.Controller,
		CardID: f.source.InstanceID,
		Label:  label + ": " + g.targetRefNameLocked(ref) + " (at random)",
	})
	return ref
}
