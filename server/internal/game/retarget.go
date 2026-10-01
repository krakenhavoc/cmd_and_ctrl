package game

import "github.com/google/uuid"

// retarget.go — CR 115.7, changing the targets of a spell or ability
// that is already on the stack.
//
// "Change the target of target spell or ability with a single target"
// (Bolt Bend, Misdirection, Ricochet Trap, Imp's Mischief) and "you
// may choose new targets for target spell or ability" (Deflecting
// Swat) are the two printed shapes, and CR 707.10c — a copy's "you
// may choose new targets for the copy" — is the third by reference.
//
// The whole design is one observation: RETARGETING IS THE ANNOUNCE
// GATE RUN A SECOND TIME, FOR A DIFFERENT CHOOSER. Nothing about
// which targets are legal changes when a Swat redirects a Bolt — the
// same clause, the same TargetSource, the same protection and
// hexproof check (CanBeTargetedBy, through targetLegalLocked). What
// changes is who answers the question. So this file is an entry point
// over clauses.go's walk, not a second walk.
//
// Three things ARE the retarget's own, and they are the only three:
//
//  1. ONE-FOR-ONE. The new list is the same length as the item's and
//     each ref answers the same (Mode, Slot) step. CR 115.7 changes
//     WHICH objects are targeted, never HOW MANY, and never which
//     clause a slot belongs to.
//
//  2. A CHANGE COUNT. RetargetChangeOne permits one slot to differ,
//     RetargetChooseNew permits every slot to.
//
//  3. AN UNCHANGED SLOT IS WAIVED. CR 115.7c lets a player leave a
//     target alone EVEN IF IT IS NOW ILLEGAL, so the predicate runs
//     on the slots that changed and is waived on the ones that did
//     not. That is the `waive` parameter on
//     validateAnnouncedTargetsWithLocked and the only change this
//     seam made to the announce path.
//
// A fourth printed shape rides the same gate: "change a target of
// target spell or ability TO THIS CREATURE" (Spellskite, Mizzium
// Meddler, Hydroelectric Specimen, #1743). Its destination is fixed,
// so it asks no "where" — RetargetOffer.To pins it, and a slot is
// changed only if the pinned object passes exactly the check a free
// choice of it would (pinnedRetargetSlotsLocked).
//
// See docs/decisions/0019-structured-targeting.md, the 2026-09-22
// and 2026-10-01 amendments.

// PendingChoiceRetarget is the CR 115.7 prompt: "choose the new
// target for …", addressed to the RETARGETING effect's controller
// over one slot of an item that is already on the stack.
//
// It is not a PendingChoicePickTarget, though it asks the same
// question shape and rides the same wire projection and the same
// board picker. The difference that earns it a kind is the DEPARTURE
// table: pick_target is `{reassign: true}` — a trigger's CR 603.3d
// pick is about the whole board and a survivor can answer it —
// whereas nobody else gets to aim another player's Deflecting Swat,
// and CR 800.4 leaves the targets unchanged when the player who would
// have changed them is gone. See leave_game.go.
//
// It carries DATA, not a continuation: the item id, the policy,
// whether declining is allowed and which slot is being asked.
// Answering it rewrites an object already on the stack, so there is
// no closure to resume and nothing for ContinuationCensus to count —
// a game paused on a retarget prompt is a restorable snapshot, which
// is not true of any other prompt in this family.
const PendingChoiceRetarget PendingChoiceKind = "retarget"

// RetargetPolicy is which of CR 115.7's two printed sentences an
// effect is applying.
type RetargetPolicy uint8

const (
	// RetargetChangeOne is "change the target of target spell or
	// ability" (CR 115.7b): at most ONE of the item's target slots
	// may end up different.
	//
	// Every card that prints it with a free choice also prints "with
	// a single target", so OfferRetargetForEffect refuses the prompt
	// for an item with more than one — a multi-slot "change a target"
	// would have to ask WHICH slot first and then WHERE, and there is
	// no card to design that prompt against. A PINNED offer
	// (RetargetOffer.To, #1743) is exempt: Spellskite's "change a
	// target … to this creature" fixes the where, so which slot is
	// the only question left, and it is asked.
	RetargetChangeOne RetargetPolicy = iota

	// RetargetChooseNew is "choose new targets for" (CR 115.7c):
	// every slot may be changed, any number may be left unchanged,
	// and one left unchanged may stay illegal. Deflecting Swat prints
	// it; so does CR 707.10c for a copy.
	RetargetChooseNew
)

// RetargetOffer is one CR 115.7 offer: whose targets are being
// changed, who is choosing, and under which sentence.
type RetargetOffer struct {
	// ItemID is the stack item whose targets change. A SPELL's item
	// id is the spell card's instance id.
	ItemID uuid.UUID

	// Chooser is the player answering — the RETARGETING effect's
	// controller, which is deliberately not the item's own
	// controller. Legality is still judged for the item (see
	// retargetSourceLocked).
	Chooser uuid.UUID

	// Policy is which sentence is printed.
	Policy RetargetPolicy

	// Optional marks a "you MAY change the target". A mandatory
	// change with an alternative available must be taken; an
	// optional one may be declined, and so may any slot under
	// RetargetChooseNew (CR 115.7c).
	Optional bool

	// Source is the card doing the retargeting, for the prompt's
	// banner. Cosmetic.
	Source uuid.UUID

	// Reason is the prompt header ("Deflecting Swat — choose a new
	// target"). Empty falls back to the clause label.
	Reason string

	// To pins the DESTINATION (#1743): "change a target of target
	// spell or ability to this creature" (Spellskite, Mizzium
	// Meddler, Hydroelectric Specimen). The zero value (Kind "") is
	// the free choice every other retarget card prints.
	//
	// A pinned offer asks no "where": the only target a slot can be
	// changed to is To, and only when To is a legal new target for
	// that slot under the very gate a free choice runs
	// (retargetCheckLocked) — the slot's own clause and zones, the
	// item's controller's view for hexproof / shroud / protection,
	// the same-clause duplicate rule, Distinct and the set rules, so
	// a change that would make another target illegal is refused
	// (the 2020-08-07 Spellskite ruling). A slot To already holds is
	// not a change. With no slot eligible the offer does nothing
	// (CR 115.7a). See pinnedRetargetSlotsLocked.
	//
	// The one question a pinned offer can still ask is WHICH slot:
	// "change A target" over an item with several (CR 115.7b), which
	// the chooser decides as the ability resolves. So a pinned offer
	// takes RetargetChangeOne over an item with any number of
	// targets — the restriction a free RetargetChangeOne carries
	// exists because it would have to ask both questions at once.
	To TargetRef
}

// pinnedRetarget reports whether `to` names a destination — the
// RetargetOffer.To / PendingChoice.RetargetTo "is this pinned" test.
func pinnedRetarget(to TargetRef) bool {
	return to.Kind == TargetCard || to.Kind == TargetPlayer
}

// RetargetableForEffect reports whether an item on the stack can be
// retargeted at all: it exists and it has at least one real target.
// A spell or ability with no targets cannot be retargeted (there is
// nothing to change), which is what keeps Wrath of God off Bolt
// Bend's picker.
//
// Caller must hold g.mu.
func (g *Game) RetargetableForEffect(itemID uuid.UUID) bool {
	item := g.StackMeta[itemID]
	return item != nil && countRealTargets(item.Targets) > 0
}

// StackItemTargetCountForEffect is how many real targets an item on
// the stack has — the "with a single target" half of the printed
// clause, which every "change the target" card carries. 0 for an
// item that is not on the stack.
//
// Counts SLOTS, not distinct objects: a spell that chose the same
// creature for two different instances of the word "target" (CR
// 115.3's AllowSame) has two targets and is not a single-target
// spell.
//
// Caller must hold g.mu.
func (g *Game) StackItemTargetCountForEffect(itemID uuid.UUID) int {
	item := g.StackMeta[itemID]
	if item == nil {
		return 0
	}
	return countRealTargets(item.Targets)
}

// retargetSourceLocked is the TargetSource a retarget's legality is
// judged under, and the one decision the rest of this file falls out
// of: it is the ITEM's, not the chooser's.
//
// A Deflecting Swat pointed at an opponent's "target creature you
// control" can only move it to a creature THAT OPPONENT controls, and
// protection is tested against the redirected spell's own colour
// rather than against the Swat. Keeping the two players in different
// variables is the whole of it.
//
// Caller must hold g.mu.
func (g *Game) retargetSourceLocked(item *StackItem) TargetSource {
	return g.stackItemSourceLocked(item)
}

// RetargetStackItemForEffect is THE entry point: rewrite `itemID`'s
// targets to `next`, applying CR 115.7.
//
// `next` is the WHOLE new list, one-for-one with the item's current
// one — a slot the chooser left alone appears in it unchanged. That
// shape is what lets one call express "changed slot 1, left slot 0",
// and it is what the copy path's answer is too.
//
// Everything but the targets survives: PaidCost, Modes, XValue,
// AltCost, Foretold, CastFromZone and IsCopy are untouched, because
// CR 115.7 changes no other announce-time choice. Distribution is
// REMAPPED rather than kept, since it is keyed by target id — see
// remapDistributionLocked.
//
// Errors: ErrCardNotFound (no such item on the stack),
// ErrInvalidParam (no targets to change, a list that is not
// one-for-one, or more slots changed than the policy allows),
// ErrIllegalTarget (a changed slot's new pick is not legal).
//
// Caller must hold g.mu.
func (g *Game) RetargetStackItemForEffect(itemID, chooser uuid.UUID, policy RetargetPolicy, next []TargetRef) error {
	item := g.StackMeta[itemID]
	if item == nil {
		return ErrCardNotFound
	}
	if countRealTargets(item.Targets) == 0 {
		// CR 115.7: there is no target to change. Refused rather
		// than silently treated as a no-op, so a card whose clause
		// forgot to exclude untargeted spells fails loudly.
		return ErrInvalidParam
	}
	old := item.Targets
	steps := g.itemAnnouncedClauses(item)
	if err := g.retargetCheckLocked(g.retargetSourceLocked(item), steps, old, next, policy); err != nil {
		return err
	}
	changed := changedRefs(old, next)
	if len(changed) == 0 {
		// A declined "you may", or a change to the same objects.
		// Nothing moved, so nothing became a target (CR 115.7) and
		// there is nothing to emit.
		return nil
	}
	item.Distribution = remapDistributionLocked(item.Distribution, old, next)
	item.Targets = append([]TargetRef(nil), next...)
	// CR 115.7: the new picks have become the target of the spell or
	// ability — of the ITEM, so the actor is its controller, which is
	// what a ward trigger or Monk Gyatso reads ("a spell or ability
	// an opponent controls"). The objects that were already targeted
	// and stayed do not become targets a second time.
	g.emitBecameTargetLocked(item.Controller, item.SourceCardID, item.ID, changed)
	return nil
}

// retargetCheckLocked is CR 115.7's gate, shared by the stack-item
// entry point above and by the CR 707.10c copy re-target
// (resolveCopySpellTargetsLocked). It is the announce walk with two
// extra rules and one waiver; see the file header.
//
// `old` is the list being replaced — the item's for a retarget, the
// COPIED item's for a copy, which is why this takes the list rather
// than the item.
//
// Caller must hold g.mu.
func (g *Game) retargetCheckLocked(src TargetSource, steps []AnnouncedClause, old, next []TargetRef, policy RetargetPolicy) error {
	pairs, err := pairRetargetSlots(old, next)
	if err != nil {
		return err
	}
	changed := 0
	unchanged := make(map[int]bool, len(pairs))
	for _, p := range pairs {
		if p.changed {
			changed++
		} else {
			unchanged[p.nextIdx] = true
		}
	}
	if policy == RetargetChangeOne && changed > 1 {
		return ErrInvalidParam
	}
	// The announce gate, with the unchanged slots' predicate waived:
	// CR 115.7c lets a player leave a target alone even if it would
	// now be illegal. Counts, duplicates and Distinct are still
	// checked across the WHOLE list, which is CR 115.7c's "must not
	// cause any unchanged targets to become illegal".
	return g.validateAnnouncedTargetsWithLocked(src, steps, next, func(i int, _ TargetRef) bool {
		return unchanged[i]
	})
}

// retargetPair is one slot of the announcement seen twice: where it
// was, where it is going, and whether that is a move.
type retargetPair struct {
	oldIdx, nextIdx int
	old, next       TargetRef
	changed         bool
}

// pairRetargetSlots lines the new list up against the old one, REAL
// SLOT BY REAL SLOT, and is where CR 115.7's "one-for-one" lives.
//
// TargetSelf / TargetNone placeholders are not targets and are
// skipped on both sides rather than required to line up positionally:
// a caller that hands back only the real picks (the CR 707.10c copy
// prompt does) is answering the same question as one that hands back
// the whole list with its placeholders intact.
//
// Errors with ErrInvalidParam when the counts differ (CR 115.7
// changes WHICH objects are targeted, never HOW MANY), when a ref
// answers a different (Mode, Slot) step than the one it replaces (the
// clause a slot answers is part of the announcement, not part of the
// choice), or when a new ref is neither a player nor a card.
func pairRetargetSlots(old, next []TargetRef) ([]retargetPair, error) {
	oldIdx := realTargetIndexes(old)
	nextIdx := realTargetIndexes(next)
	if len(oldIdx) != len(nextIdx) {
		return nil, ErrInvalidParam
	}
	out := make([]retargetPair, 0, len(oldIdx))
	for i := range oldIdx {
		o, n := old[oldIdx[i]], next[nextIdx[i]]
		if n.Kind != TargetPlayer && n.Kind != TargetCard {
			return nil, ErrInvalidParam
		}
		if n.Mode != o.Mode || n.Slot != o.Slot {
			return nil, ErrInvalidParam
		}
		out = append(out, retargetPair{
			oldIdx:  oldIdx[i],
			nextIdx: nextIdx[i],
			old:     o,
			next:    n,
			changed: n.Kind != o.Kind || n.ID != o.ID,
		})
	}
	return out, nil
}

// realTargetIndexes is the positions of the refs that are actually
// targets — the placeholders validateAnnouncedTargetsLocked skips.
func realTargetIndexes(refs []TargetRef) []int {
	var out []int
	for i, t := range refs {
		if t.Kind == TargetSelf || t.Kind == TargetNone {
			continue
		}
		out = append(out, i)
	}
	return out
}

// changedRefs is the slots whose object actually moved.
func changedRefs(old, next []TargetRef) []TargetRef {
	pairs, err := pairRetargetSlots(old, next)
	if err != nil {
		return nil
	}
	var out []TargetRef
	for _, p := range pairs {
		if p.changed {
			out = append(out, p.next)
		}
	}
	return out
}

// remapDistributionLocked carries a divided-damage / divided-counters
// record across a retarget.
//
// StackItem.Distribution is keyed by TARGET ID, so a slot that moves
// would otherwise leave its portion stranded on the old key and the
// new target would get nothing. CR 115.7c says the division can't be
// changed, so the portion travels with the SLOT: positionally, old[i]
// → next[i].
//
// Two targets that collapse onto one object (legal only under
// AllowSame) collapse their portions too, last slot winning; no
// printed card can reach that, and the alternative — a second keying
// scheme for one impossible board — is worse.
func remapDistributionLocked(dist map[uuid.UUID]int, old, next []TargetRef) map[uuid.UUID]int {
	if len(dist) == 0 {
		return dist
	}
	pairs, err := pairRetargetSlots(old, next)
	if err != nil {
		return dist
	}
	out := make(map[uuid.UUID]int, len(dist))
	moved := make(map[uuid.UUID]bool, len(pairs))
	for _, p := range pairs {
		moved[p.old.ID] = true
		if v, ok := dist[p.old.ID]; ok {
			out[p.next.ID] = v
		}
	}
	// Anything the item divided among something that is not one of
	// its target slots is left exactly where it was.
	for k, v := range dist {
		if !moved[k] {
			out[k] = v
		}
	}
	return out
}

// OfferRetargetForEffect is the card-facing half: open the CR 115.7
// prompt for `offer`, or do nothing when there is nothing to offer.
//
// "Nothing to offer" is an OUTCOME, not a failure. CR 115.7a: if a
// target can't be changed to another legal target, the original
// target is unchanged, even if that target is illegal. A Bolt Bend on
// a Lightning Bolt that is already pointed at the only creature on
// the board resolves and does nothing, which is what the card does in
// paper.
//
// Errors: ErrCardNotFound (the item is no longer on the stack — its
// controller may have had it countered in response, a normal outcome
// callers should treat as "the effect did nothing"), ErrInvalidParam
// (the item has no targets, a free RetargetChangeOne was offered for
// an item with more than one, or a pinned offer — RetargetOffer.To —
// came with any policy but RetargetChangeOne).
//
// Caller must hold g.mu.
func (g *Game) OfferRetargetForEffect(offer RetargetOffer) error {
	item := g.StackMeta[offer.ItemID]
	if item == nil {
		return ErrCardNotFound
	}
	n := countRealTargets(item.Targets)
	if n == 0 {
		return ErrInvalidParam
	}
	if pinnedRetarget(offer.To) {
		if offer.Policy != RetargetChangeOne {
			// "Choose new targets … to X" is not a printed sentence:
			// every pinned card says "change the target" or "change
			// a target".
			return ErrInvalidParam
		}
		return g.offerPinnedRetargetLocked(offer, item)
	}
	if offer.Policy == RetargetChangeOne && n > 1 {
		// See RetargetChangeOne's doc comment: no printed card gets
		// here, and the prompt for it is undesigned.
		return ErrInvalidParam
	}
	g.queueRetargetStepLocked(offer, 0)
	return nil
}

// queueRetargetStepLocked opens the prompt for the first slot at or
// after `from` that has somewhere else to go, and returns having
// queued nothing when no slot does.
//
// Skipping a slot with no alternative is CR 115.7a, and skipping it
// SILENTLY is the point: a prompt with an empty option list is a
// prompt nobody can answer, which since #791 is a table that cannot
// move.
//
// Caller must hold g.mu.
func (g *Game) queueRetargetStepLocked(offer RetargetOffer, from int) {
	item := g.StackMeta[offer.ItemID]
	if item == nil {
		// Countered or resolved while the walk was open. Nothing
		// left to retarget, and nothing to report — the rest of the
		// retargeting card has already run.
		return
	}
	steps := g.itemAnnouncedClauses(item)
	src := g.retargetSourceLocked(item)
	for i := from; i < len(item.Targets); i++ {
		ref := item.Targets[i]
		if ref.Kind != TargetCard && ref.Kind != TargetPlayer {
			continue
		}
		clause := retargetClauseFor(steps, ref)
		if clause == nil {
			// A free-form S13.1 announcement: no structured clause,
			// so nothing can say what a legal new target would be.
			continue
		}
		alt := g.legalTargetsLocked(src, clause)
		alt = retargetAlternatives(alt, item.Targets, i, clause)
		alt = g.withoutSetRuleConflictsLocked(clause, alt, item.Targets, i)
		if len(alt.Players) == 0 && len(alt.Cards) == 0 {
			continue
		}
		min := 1
		if offer.Policy == RetargetChooseNew || offer.Optional {
			// CR 115.7c: any number of the targets may be left
			// unchanged. A mandatory "change the target" with an
			// alternative on the board must be taken.
			min = 0
		}
		reason := offer.Reason
		if reason == "" {
			reason = "Choose a new target"
			if clause.Label != "" {
				reason = "Choose the new target for " + clause.Label
			}
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:              PendingChoiceRetarget,
			Chooser:           offer.Chooser,
			Count:             1,
			Source:            offer.Source,
			Reason:            reason,
			PickTargetPlayers: alt.Players,
			PickTargetCards:   alt.Cards,
			PickTargetMin:     min,
			PickTargetMax:     1,
			RetargetItem:      offer.ItemID,
			RetargetPolicy:    offer.Policy,
			RetargetOptional:  offer.Optional,
			RetargetSlot:      i,
			RetargetReason:    offer.Reason,
		})
		return
	}
}

// retargetClauseFor is the announced clause a ref belongs to, by
// (Mode, Slot) — clauseForRefLocked's pure half, over a step list the
// caller already has.
func retargetClauseFor(steps []AnnouncedClause, ref TargetRef) *TargetClause {
	for i := range steps {
		if steps[i].Mode == ref.Mode && steps[i].Slot == ref.Slot {
			return &steps[i].Clause
		}
	}
	return nil
}

// retargetAlternatives narrows a slot's legal set to the objects it
// could actually be changed TO: everything legal, minus the target
// already sitting in this slot ("changed only to ANOTHER legal
// target", CR 115.7a), minus whatever the rest of the announcement
// forbids — another slot of the same clause unless AllowSame, or any
// earlier slot's pick when the clause is Distinct.
func retargetAlternatives(lt LegalTargets, targets []TargetRef, slot int, clause *TargetClause) LegalTargets {
	taken := map[uuid.UUID]bool{targets[slot].ID: true}
	for i, t := range targets {
		if i == slot || (t.Kind != TargetCard && t.Kind != TargetPlayer) {
			continue
		}
		sameClause := t.Mode == targets[slot].Mode && t.Slot == targets[slot].Slot
		if sameClause && !clause.AllowSame {
			taken[t.ID] = true
		}
		if clause.Distinct && !sameClause && i < slot {
			taken[t.ID] = true
		}
	}
	out := LegalTargets{}
	for _, id := range lt.Players {
		if !taken[id] {
			out.Players = append(out.Players, id)
		}
	}
	for _, id := range lt.Cards {
		if !taken[id] {
			out.Cards = append(out.Cards, id)
		}
	}
	return out
}

// pinnedRetargetSlotsLocked is the slots of `item` whose target can be
// changed to `to` — the whole of the pinned variant's legality, and
// deliberately no new rule: each candidate is the item's list with
// that ONE slot replaced by `to`, run through retargetCheckLocked
// under RetargetChangeOne, the exact check a free "change the target"
// answer passes. So everything Spellskite's rulings ask for falls out
// of the gate rather than being restated:
//
//   - `to` must satisfy the slot's own clause — its predicate, its
//     zones (a Spellskite that left the battlefield is in no
//     battlefield clause's zones), and CanBeTargetedBy judged for the
//     ITEM's controller, so an opponent's spell cannot be pulled onto
//     a hexproof Spellskite and nobody's onto a shrouded one;
//   - the whole new list must stand: `to` may not join a clause that
//     already names it unless the clause is AllowSame, and may not
//     land where a Distinct clause or a set rule forbids it — "if
//     changing one target … would make other targets … illegal, that
//     target can't be changed";
//   - a slot `to` already holds is not a change (CR 115.7a's
//     "ANOTHER legal target"), so it is never eligible.
//
// Slots with no structured clause (a free-form S13.1 announcement)
// are skipped, as queueRetargetStepLocked skips them: nothing can say
// what a legal new target for them would be, and the announce walk
// accepts anything for a list with no clauses.
//
// Caller must hold g.mu.
func (g *Game) pinnedRetargetSlotsLocked(item *StackItem, to TargetRef) []int {
	if item == nil || !pinnedRetarget(to) {
		return nil
	}
	steps := g.itemAnnouncedClauses(item)
	src := g.retargetSourceLocked(item)
	var out []int
	for i, ref := range item.Targets {
		if ref.Kind != TargetCard && ref.Kind != TargetPlayer {
			continue
		}
		if ref.Kind == to.Kind && ref.ID == to.ID {
			continue
		}
		if retargetClauseFor(steps, ref) == nil {
			continue
		}
		next := pinnedRetargetList(item.Targets, i, to)
		if g.retargetCheckLocked(src, steps, item.Targets, next, RetargetChangeOne) == nil {
			out = append(out, i)
		}
	}
	return out
}

// pinnedRetargetList is `targets` with slot `i` changed to `to`. The
// replacement keeps the slot's (Mode, Slot) — a retarget never moves a
// ref between clauses.
func pinnedRetargetList(targets []TargetRef, i int, to TargetRef) []TargetRef {
	next := append([]TargetRef(nil), targets...)
	next[i] = TargetRef{Kind: to.Kind, ID: to.ID, Mode: targets[i].Mode, Slot: targets[i].Slot}
	return next
}

// offerPinnedRetargetLocked is OfferRetargetForEffect for a pinned
// offer (#1743). Three outcomes, by what could be changed to the
// destination:
//
//   - NOTHING: nothing happens, and nobody is asked. CR 115.7a — the
//     original target is unchanged — and Spellskite's ruling that its
//     ability may be activated at a spell it cannot steal, or at one
//     with no targets at all, "in which case no targets are changed".
//   - ONE object, and the change is mandatory: it is made, with no
//     prompt. There is no choice to ask about.
//   - otherwise: a PendingChoiceRetarget asking WHICH target to
//     change — "you choose which one target you're changing … as
//     Spellskite's ability resolves" — or, for a "you may" (Mizzium
//     Meddler), whether to change it at all. The options are the
//     objects CURRENTLY in the eligible slots, so the board picker
//     answers it unchanged; `RetargetTo` is what tells the answer
//     apart from a free retarget's.
//
// Two eligible slots that hold the SAME object (two instances of the
// word "target" that chose one creature, CR 601.2c) are one option on
// the board, and the answer changes the earlier of them —
// resolvePinnedRetargetLocked. The board cannot point at a slot, only
// at an object.
//
// Caller must hold g.mu.
func (g *Game) offerPinnedRetargetLocked(offer RetargetOffer, item *StackItem) error {
	eligible := g.pinnedRetargetSlotsLocked(item, offer.To)
	if len(eligible) == 0 {
		return nil
	}
	var players, cards []uuid.UUID
	seen := make(map[uuid.UUID]bool, len(eligible))
	for _, i := range eligible {
		ref := item.Targets[i]
		if seen[ref.ID] {
			continue
		}
		seen[ref.ID] = true
		if ref.Kind == TargetPlayer {
			players = append(players, ref.ID)
		} else {
			cards = append(cards, ref.ID)
		}
	}
	if len(seen) == 1 && !offer.Optional {
		// One thing the board could offer, and no "you may": the
		// answer is forced, so it is made rather than asked. It is the
		// earliest eligible slot — the one the prompt's answer would
		// have changed. RetargetStackItemForEffect re-runs the check
		// that slot just passed.
		return g.RetargetStackItemForEffect(item.ID, offer.Chooser, RetargetChangeOne,
			pinnedRetargetList(item.Targets, eligible[0], offer.To))
	}
	min := 1
	if offer.Optional {
		min = 0
	}
	reason := offer.Reason
	if reason == "" {
		reason = "Choose the target to change"
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoiceRetarget,
		Chooser:           offer.Chooser,
		Count:             1,
		Source:            offer.Source,
		Reason:            reason,
		PickTargetPlayers: players,
		PickTargetCards:   cards,
		PickTargetMin:     min,
		PickTargetMax:     1,
		RetargetItem:      offer.ItemID,
		RetargetPolicy:    RetargetChangeOne,
		RetargetOptional:  offer.Optional,
		RetargetSlot:      eligible[0],
		RetargetReason:    offer.Reason,
		RetargetTo:        offer.To,
	})
	return nil
}

// resolvePinnedRetargetLocked answers a pinned retarget prompt. The
// one ref, if any, names the object CURRENTLY in the slot to change;
// the destination is the prompt's RetargetTo, never the answer.
//
// Eligibility is recomputed rather than read off the prompt, so the
// change is judged against the board as it is answered. If no slot is
// eligible any more the prompt is dropped and nothing moves (CR
// 115.7a); an answer naming an object that is in no eligible slot is
// refused and the prompt stays open.
//
// Caller must hold g.mu.
func (g *Game) resolvePinnedRetargetLocked(idx int, choice *PendingChoice, item *StackItem, targets []TargetRef) error {
	switch len(targets) {
	case 0:
		if choice.PickTargetMin > 0 {
			return ErrInvalidParam
		}
		g.dequeueChoiceLocked(idx)
	case 1:
		t := targets[0]
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
		eligible := g.pinnedRetargetSlotsLocked(item, choice.RetargetTo)
		if len(eligible) == 0 {
			g.dequeueChoiceLocked(idx)
			break
		}
		slot := -1
		for _, i := range eligible {
			if item.Targets[i].Kind == t.Kind && item.Targets[i].ID == t.ID {
				slot = i
				break
			}
		}
		if slot < 0 {
			return ErrIllegalTarget
		}
		if err := g.RetargetStackItemForEffect(item.ID, choice.Chooser, RetargetChangeOne,
			pinnedRetargetList(item.Targets, slot, choice.RetargetTo)); err != nil {
			return err
		}
		g.dequeueChoiceLocked(idx)
	default:
		return ErrInvalidParam
	}
	g.runStateChecksLocked()
	return nil
}

// ResolveRetarget is the submit half of the CR 115.7 prompt: the
// chooser's answer for one slot, which is either one ref or — when
// declining is allowed — none.
//
// Each slot is applied AS IT IS ANSWERED rather than accumulated:
// the object being rewritten is already on the stack, so there is no
// continuation to hold, and a walk cut short (the chooser leaves the
// game) leaves the slots already changed changed — each of which was
// legal on its own when it was made.
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolveRetarget(choiceID, chooserID uuid.UUID, targets []TargetRef) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceRetarget {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	offer := RetargetOffer{
		ItemID:   choice.RetargetItem,
		Chooser:  choice.Chooser,
		Policy:   choice.RetargetPolicy,
		Optional: choice.RetargetOptional,
		Source:   choice.Source,
		Reason:   choice.RetargetReason,
	}
	item := g.StackMeta[offer.ItemID]
	if item == nil {
		// The item left the stack under an open prompt. Nothing to
		// change; drop the question rather than wedge the table.
		g.dequeueChoiceLocked(idx)
		g.runStateChecksLocked()
		return nil
	}
	if pinnedRetarget(choice.RetargetTo) {
		// #1743: the answer names WHICH slot, not where it goes.
		return g.resolvePinnedRetargetLocked(idx, choice, item, targets)
	}
	slot := choice.RetargetSlot
	if slot < 0 || slot >= len(item.Targets) {
		return ErrInvalidParam
	}
	switch len(targets) {
	case 0:
		if choice.PickTargetMin > 0 {
			// A mandatory "change the target" with somewhere to go.
			return ErrInvalidParam
		}
		g.dequeueChoiceLocked(idx)
	case 1:
		t := targets[0]
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
		next := append([]TargetRef(nil), item.Targets...)
		t.Mode, t.Slot = next[slot].Mode, next[slot].Slot
		next[slot] = t
		if err := g.RetargetStackItemForEffect(offer.ItemID, chooserID, offer.Policy, next); err != nil {
			return err
		}
		g.dequeueChoiceLocked(idx)
	default:
		// One slot is asked at a time.
		return ErrInvalidParam
	}
	if offer.Policy == RetargetChooseNew {
		// CR 115.7c: every target may be changed, so the walk goes
		// on to the next slot that has somewhere to go.
		g.queueRetargetStepLocked(offer, slot+1)
	}
	g.runStateChecksLocked()
	return nil
}
