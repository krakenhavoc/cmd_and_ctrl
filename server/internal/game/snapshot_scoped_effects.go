package game

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
)

// snapshot_scoped_effects.go is the snapshot half of ADR 0041 phase 3's
// data records (#1497): the deep copy at the snapshot boundary, and the
// restore-time refusal of an effect key this binary cannot interpret
// (ErrUnknownEffectKey, ADR 0041 P4). The record itself, its vocabulary
// and its interpreter are in scoped_effects.go.

// ---------------------------------------------------------------
// Restore validation
// ---------------------------------------------------------------

// checkEffectKeys refuses a snapshot naming an effect key this binary
// cannot interpret (ADR 0041 P4): a mod kind, or a delayed-trigger or
// stack-item body or condition key. Every key a binary can write is
// one it registered, so the only way to meet an unknown one is a
// ROLLBACK — the file was written by a newer build — and the answer is
// the one Decision 5 gives a too-new file: refuse it, keep it.
//
// Effect-body and condition keys are ADR 0041 P2's (tier 2): a key
// this binary registered restores; any other is refused.
func (s *GameSnapshot) checkEffectKeys() error {
	unknown := append([]string(nil), s.unknownEffectFields...)
	// ADR 0188 Decision 3b: the effects an extra turn is queued with are
	// records too, registered as the turn begins, so they are held to the
	// same vocabulary as the live ones.
	records := s.ScopedEffects
	for _, et := range s.ExtraTurns {
		if len(et.Effects) > 0 {
			records = append(append([]ScopedEffect(nil), records...), et.Effects...)
		}
	}
	unknown = append(unknown, unknownScopedEffectRecordKeys(records)...)
	// ADR 0203: a venture marker on a dungeon or room this binary does
	// not register.
	unknown = append(unknown, unknownDungeonKeys(s.Seats)...)
	// ADR 0109 Shared machinery 2: every stored duration, wherever it
	// is — a scoped effect's, a delayed trigger's, a cast permission's,
	// a player static's, an untap hold's — is one this binary can read,
	// or the file is refused (snapshot_durations.go).
	unknown = append(unknown, s.unknownDurations()...)
	// #3192: a cast permission's condition is a key too. One this binary
	// never registered would restore as an unconditional grant, which is
	// stronger than the file's writer meant, so the file is refused.
	for _, seat := range s.Seats {
		for _, perm := range seat.CastPermissions {
			if perm.Condition != nil && !perm.Condition.Key.Known() {
				unknown = append(unknown, "cast-permission condition "+string(perm.Condition.Key))
			}
			// #4361 S13: and its follow-ups, which an unknown key would
			// run as nothing (runCastFollowUpLocked).
			unknown = append(unknown, unknownCastFollowUps(&perm)...)
		}
	}
	// ADR 0109 §11: a mana spend rider's kind, and a counted rider's
	// count key, are keys too — wherever a token is carried
	// (snapshot_mana_riders.go).
	unknown = append(unknown, s.unknownManaRiders()...)
	for _, d := range s.DelayedTriggers {
		if d.Body != "" && !KnownEffectBody(d.Body) {
			unknown = append(unknown, "delayed-trigger body "+d.Body)
		}
		if d.Condition != "" && !KnownEffectCondition(d.Condition) {
			unknown = append(unknown, "delayed-trigger condition "+d.Condition)
		}
		// A spell filter is a closed vocabulary too (#1568 review).
		for _, p := range []*EffectParams{d.Params, d.CondParams} {
			if p != nil && !p.Filter.Valid() {
				unknown = append(unknown, fmt.Sprintf("delayed-trigger spell filter %v", p.Filter.Types))
			}
			// An ability ref is an ability's stack item's, never a
			// delayed trigger's: no binary writes one here.
			if p != nil && p.Ability != nil {
				unknown = append(unknown, "delayed-trigger params with an ability ref")
			}
		}
	}
	// ADR 0107 §6: an owed follow-up names its body as a key.
	for _, f := range s.PreventionFollowUps {
		// #2047: a CR 615.13 announcement names no body; it is emitted
		// as an event.
		if f.Announce && f.Body == "" {
			continue
		}
		if !KnownEffectBody(f.Body) {
			unknown = append(unknown, "prevention follow-up body "+f.Body)
		}
	}
	// ADR 0115 §8: a pending-choice kind is a key too. A newer binary's
	// prompt would restore as a question this one cannot answer, which
	// wedges the table, so the file is refused and kept for the
	// roll-forward.
	for _, c := range s.PendingChoices {
		if !KnownChoiceKind(c.Kind) {
			unknown = append(unknown, "pending-choice kind "+string(c.Kind))
		}
		// #2115: a revealed-hand pick's continuation is a key, and its
		// destination a closed vocabulary. Restored without either, the
		// pick would move the card somewhere else or skip the rest of
		// the card.
		// #2854: on an option_pick it is a keyed option pick's
		// continuation, from the other registry.
		if c.ReplacementResume != nil && (c.PickThen == "" || !keyedReplacementPromptKind(c.Kind)) {
			// #4253: a frame on a prompt kind that cannot carry one, or
			// with no tail key, is a newer binary's shape.
			unknown = append(unknown, fmt.Sprintf("replacement resume frame on a %s prompt with tail %q", c.Kind, c.PickThen))
		}
		if c.SearchPrompt != nil && c.Kind != PendingChoiceSearchLibrary {
			// #4253 slice 5: a search frame on another kind is a newer
			// binary's shape.
			unknown = append(unknown, fmt.Sprintf("search prompt frame on a %s prompt", c.Kind))
		}
		if c.TriggerFrame != nil && (!carriesTriggerFrame(c.Kind) || c.PickThen != TriggerFrameKey) {
			// #4361 S14a: a trigger frame on another kind, or without its
			// marker, is a newer binary's shape.
			unknown = append(unknown, fmt.Sprintf("trigger frame on a %s prompt with tail %q", c.Kind, c.PickThen))
		}
		// #4361 S14b: a target walk on another kind, or without its
		// prompt's marker, is a newer binary's shape, and so is a
		// pick_target marker with no walk.
		unknown = append(unknown, unknownTargetWalkKeys(&c)...)
		if c.Parked != nil {
			// #4361 S13: a parked announcement on a kind that cannot carry
			// one, under another marker, or without its params.
			unknown = append(unknown, unknownParkedAnnouncementKeys(&c)...)
		}
		// #4361 S15: a copy walk likewise, and the keys its item names.
		unknown = append(unknown, unknownCopyWalkKeys(&c)...)
		switch {
		case c.Parked != nil && c.Kind != PendingChoiceOptionPick:
			// #4361 S13: the marker is the parked frame's, checked above;
			// an option pick's is its own continuation, checked below.
		case carriesTriggerFrame(c.Kind) && c.PickThen != "":
			// #4361 S14a: a written trigger frame names its marker and
			// its catalog row.
			unknown = append(unknown, unknownTriggerFrameKeys(&c)...)
		case c.Kind == PendingChoicePickTarget && c.PickThen != "":
			// #4361 S14b and S15: checked above (unknownTargetWalkKeys,
			// unknownCopyWalkKeys).
		case c.Kind == PendingChoiceSearchLibrary && (c.PickThen != "" || c.SearchPrompt != nil):
			// #4253 slice 5: a written search prompt names the marker, its
			// set rule and the card's tail.
			unknown = append(unknown, unknownSearchPromptKeys(&c)...)
		case c.Kind == PendingChoiceDivideShield:
			// #4361 S16: a written divide_shield prompt names its landing
			// step's marker, and a walk's tails.
			unknown = append(unknown, unknownDivideShieldKeys(&c)...)
		case c.PickThen == "":
		case keyedReplacementPromptKind(c.Kind):
			// #4253: a keyed CR 616 frame's tail, and the bodies and stack
			// item its event names.
			unknown = append(unknown, unknownReplacementResumeKeys(&c)...)
		case c.Kind == PendingChoiceOptionPick:
			if !KnownOptionPickThen(c.PickThen) {
				unknown = append(unknown, "option pick continuation "+c.PickThen)
			}
			if c.DiceRoll != nil {
				// #4307: a paused dice roll's keyed "then".
				unknown = append(unknown, unknownKeyedTail(c.DiceRoll.Then)...)
			}
			// #4253 slice 10b: and the tail it hands on.
			unknown = append(unknown, unknownKeyedTail(c.PickTail)...)
		case c.Kind == PendingChoiceUntapChoice:
			// #4361 S7a: a paused untap step names its continuation and
			// carries its plan as data.
			if !KnownChooseCardsThen(c.PickThen) {
				unknown = append(unknown, "choose-cards continuation "+c.PickThen)
			}
			if untapChoicePlanOf(c.OptionCarry, c.PickNums) == nil {
				unknown = append(unknown, "untap choice plan")
			}
		case c.Kind == PendingChoiceChooseCards || c.Kind == PendingChoiceResolutionCast || c.Kind == PendingChoiceRevealPick || c.Kind == PendingChoiceChooseSource ||
			// #4253 slice 10b: a keyed proliferate or Ring-bearer choice.
			c.Kind == PendingChoiceProliferate || c.Kind == PendingChoiceRingBearer:
			// #2040: a keyed choose_cards pick's continuation. #4250: a
			// resolution_cast's "after the casts" is one too. #4264: and a
			// keyed reveal pick's.
			if !KnownChooseCardsThen(c.PickThen) {
				unknown = append(unknown, "choose-cards continuation "+c.PickThen)
			}
			// #4253 slice 8: and the tail it hands on.
			unknown = append(unknown, unknownKeyedTail(c.PickTail)...)
			// #4253 slice 10c: a choose_source's shield names its follow-up.
			unknown = append(unknown, unknownSourceShieldKeys(c.SourceShield)...)
		case c.Kind == PendingChoiceOutsideCollection:
			// #4256: a keyed collection pick's filter and continuation,
			// registered under one key.
			if !KnownOutsidePick(c.PickThen) {
				unknown = append(unknown, "outside pick "+c.PickThen)
			}
		case c.Kind == PendingChoicePayUnless && c.AnyPlayerOffer != nil:
			// #4264: a keyed any-player offer's consequences, which are
			// registered apart from a plain pay-unless's branches.
			if !KnownAnyPlayerPays(c.PickThen) {
				unknown = append(unknown, "any-player consequences "+c.PickThen)
			}
		case c.Kind == PendingChoicePayAmount && c.MayPayX != nil:
			// #4289: a keyed MayPayX's "choose X" names the branches the
			// payment that follows runs.
			if !KnownPayUnlessElse(c.PickThen) {
				unknown = append(unknown, "pay-unless branches "+c.PickThen)
			}
		case c.Kind == PendingChoicePayAmount:
			// #4294: a keyed number prompt's branch (an assist helper's is
			// the engine's own).
			if !KnownNumberThen(c.PickThen) {
				unknown = append(unknown, "number branch "+c.PickThen)
			}
		case c.Kind == PendingChoicePayUnless:
			// #4256: a keyed pay-unless prompt's decline (and, since
			// #4264, its payment's branch).
			if !KnownPayUnlessElse(c.PickThen) {
				unknown = append(unknown, "pay-unless branches "+c.PickThen)
			}
		case carriesValueThen(c.Kind):
			// #4253 slice 10c: a card-name prompt's keyed "then"; since the
			// sweep's second pass, a colour or creature-type prompt's too.
			if c.NameThen == nil || c.NameThen.Key != c.PickThen {
				unknown = append(unknown, string(c.Kind)+" continuation "+c.PickThen)
			} else {
				unknown = append(unknown, unknownKeyedTail(c.NameThen)...)
			}
		case carriesLookAtThen(c.Kind):
			// #4253 slice 3: a look-at's keyed "then".
			if c.ScryThen == nil || c.ScryThen.Key != c.PickThen || !KnownTail(c.PickThen) {
				unknown = append(unknown, "look-at continuation "+c.PickThen)
			}
		case c.Kind == PendingChoiceSacrifice || isPermanentPickKind(c.Kind):
			// #4290: a keyed run's sacrifice or permanent-pick leg names
			// its run's key, so a binary that cannot finish the run
			// refuses the file rather than answering into nobody.
			if !KnownPromptRunThen(c.PickThen) {
				unknown = append(unknown, "prompted-run continuation "+c.PickThen)
			}
		case c.Kind == PendingChoicePutInLibrary:
			// #4301: an ordering prompt's keyed "then".
			if c.LibraryOrder == nil || c.LibraryOrder.Then == nil || c.LibraryOrder.Then.Key != c.PickThen {
				unknown = append(unknown, "put-in-library continuation "+c.PickThen)
			} else {
				unknown = append(unknown, unknownKeyedTail(c.LibraryOrder.Then)...)
			}
		case c.Kind == PendingChoiceCoinCall:
			// #4253 slice 10b: a keyed coin call's continuation.
			if !KnownCoinFlipThen(c.PickThen) {
				unknown = append(unknown, "coin-flip continuation "+c.PickThen)
			}
		case c.Kind == PendingChoiceConfirm:
			// #4264: a keyed confirm's answer.
			if !KnownConfirmThen(c.PickThen) {
				unknown = append(unknown, "confirm answer "+c.PickThen)
			}
		case c.Kind == PendingChoiceMayCast:
			// #4361: a keyed may_cast's answer.
			if !KnownMayCastThen(c.PickThen) {
				unknown = append(unknown, "may-cast answer "+c.PickThen)
			}
		case !KnownRevealedPickThen(c.PickThen):
			unknown = append(unknown, "revealed-hand pick continuation "+c.PickThen)
		}
		// #4290: a keyed ordered collection names its continuation.
		unknown = append(unknown, unknownOrderedPicksKeys(c.Ordered)...)
		// #4294: a tokens-attacking split's tail is a registered key too.
		if c.AttackSplit != nil {
			unknown = append(unknown, unknownKeyedTail(c.AttackSplit.Tail)...)
		}
		// #4264: a keyed pay-unless's waterbend clause is built by a
		// registered kind; restored without it the payment would be plain
		// mana, harder to pay than printed.
		if c.PayTap != nil && !KnownTapCost(c.PayTap.Kind) {
			unknown = append(unknown, "tap cost "+c.PayTap.Kind)
		}
		// ADR 0146: a ballot's vote continuation is a key too.
		if c.CouncilVote != nil && !KnownVoteThen(c.CouncilVote.Then) {
			unknown = append(unknown, "vote continuation "+c.CouncilVote.Then)
		}
		if !knownPickDestination(c.PickDestination) {
			unknown = append(unknown, "revealed-hand pick destination "+string(c.PickDestination))
		}
	}
	// #4162: a parked continuation names its steps by key. Restored
	// without them, the rest of the continuation would never run.
	unknown = append(unknown, unknownThenSteps(s.PausedStepWalk)...)
	// #4290: a keyed prompted run names its continuation by key.
	// Restored without it, the rest of the card would never run.
	unknown = append(unknown, unknownPromptRunKeys(s.PromptRuns)...)
	// #4253 slice 10c: a parked draw's "then" body is a key too.
	unknown = append(unknown, unknownParkedDrawKeys(s.ParkedDraws)...)
	// #4253: a keyed damage park's tail is a key too.
	unknown = append(unknown, unknownFollowUpParkKeys(s.FollowUpParks)...)
	// #4361 S17: a frozen entry's event names its bodies and tails too.
	for i := range s.FrozenEntries {
		unknown = append(unknown, unknownReplacementEventKeys(&s.FrozenEntries[i].Event)...)
	}
	lastKnown := make([]stackItemSnapshot, 0, len(s.LastKnownStack))
	for _, e := range s.LastKnownStack {
		lastKnown = append(lastKnown, e.Item)
	}
	for _, list := range [][]stackItemSnapshot{s.StackMeta, s.PendingTriggers, lastKnown} {
		for _, it := range list {
			if it.Body != "" && !KnownEffectBody(it.Body) {
				unknown = append(unknown, "stack-item body "+it.Body)
			}
			if it.Params != nil && !it.Params.Filter.Valid() {
				unknown = append(unknown, fmt.Sprintf("stack-item spell filter %v", it.Params.Filter.Types))
			}
			unknown = append(unknown, unknownAbilityRef(it)...)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUnknownEffectKey, strings.Join(unknown, ", "))
}

// unknownAbilityRef is the ability-ref half of checkEffectKeys (ADR
// 0041 P9, #1497). A catalog/* body is read through its ref, so an item
// naming one must carry a ref this binary can read: a slot it knows, a
// key, and a ref in ADR 0093's grammar. Whether the catalog still HAS
// that row is not a refusal — that is the owner's Q3, answered at
// restore. And a ref on an item whose body does not read one is a
// newer binary's shape, so it is refused rather than ignored.
func unknownAbilityRef(it stackItemSnapshot) []string {
	var ref *AbilityRef
	if it.Params != nil {
		ref = it.Params.Ability
	}
	slot, catalog := catalogBodySlot(it.Body)
	if !catalog {
		if ref != nil {
			return []string{fmt.Sprintf("ability ref on a stack item whose body %q does not read one", it.Body)}
		}
		return nil
	}
	switch {
	case ref == nil:
		return []string{it.Body + " stack item with no ability ref"}
	case ref.Slot != slot:
		return []string{fmt.Sprintf("ability-ref slot %q under %s", ref.Slot, it.Body)}
	case !wellFormedAbilityRef(*ref):
		return []string{fmt.Sprintf("ability ref %q of %q", ref.Ref, ref.Key)}
	}
	return nil
}

// unknownGrantBundles is the grantAbilities half of checkEffectKeys
// (ADR 0093 PR 4, #1584): every bundle a grant names must be one this
// binary's catalog registers, because the bundle IS the ability the
// effect gives. A newer build's bundle restored here would put a grant
// on the recipient that composes into a key nothing answers — an
// ability that silently does nothing — so it is refused like an
// unknown mod kind. A mod naming no bundle at all is refused too:
// registration never writes one.
func unknownGrantBundles(keys []string) []string {
	if len(keys) == 0 {
		return []string{"grantAbilities mod with no bundle"}
	}
	var out []string
	for _, k := range keys {
		if k == "" || catalogDef(GrantKey(k)) == nil {
			out = append(out, "ability-grant bundle "+fmt.Sprintf("%q", k))
		}
	}
	return out
}

// unknownScopedEffectRecordKeys is checkEffectKeys' line for each key or
// shape in records this binary does not register: the live registry's,
// an extra turn's queued effects (ADR 0188 Decision 3b) and a keyed
// tail's (#4361 S12, KeyedTail.Records) alike.
func unknownScopedEffectRecordKeys(records []ScopedEffect) []string {
	var unknown []string
	for _, e := range records {
		for _, m := range e.Mods {
			if !KnownModKind(m.Kind) {
				unknown = append(unknown, "mod kind "+string(m.Kind))
			}
			if m.Kind == ModGrantAbilities {
				unknown = append(unknown, unknownGrantBundles(m.Grants)...)
			}
			// ADR 0155 Decision 9: a lent row is a catalog row, and one
			// this binary does not have would be an ability that
			// silently does nothing.
			unknown = append(unknown, unknownLentRows(m)...)
			// #1593: a copy mod carries exactly one set of values, and
			// the ability bundles its except clause granted are keys
			// like a grant mod's — a bundle this binary does not
			// register would be an ability that silently does nothing.
			if problem := copyModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0160: a text-box mod carries exactly one text box.
			if problem := textBoxModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0160's amendment: a text-word mod carries one sound
			// change.
			if problem := textWordsModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			for _, v := range m.Copy {
				for _, k := range v.GrantedAbilities {
					if k == "" || catalogDef(GrantKey(k)) == nil {
						unknown = append(unknown, fmt.Sprintf("copied ability-grant bundle %q", k))
					}
				}
			}
			// #1715: a capacity mod's count is part of what it means.
			if problem := blockCapacityModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #1597: the requirement a block-requirement mod names is as
			// much a key as the kind itself.
			if m.Kind == ModAddBlockRequirement && !KnownBlockRequirementKind(BlockRequirementKind(m.Text)) {
				unknown = append(unknown, "block requirement "+m.Text)
			}
			// Tier 3b: a delayed-trigger body a replacement mod names
			// (Cosmic Intervention's per-card return) is as much a key
			// as a delayed trigger's own body.
			if m.Then != "" && !KnownEffectBody(m.Then) {
				unknown = append(unknown, "scoped-effect then body "+m.Then)
			}
			// ADR 0107 §6: the next-damage shield's fields are part of
			// what it means, and a source-zone field on another kind is
			// a newer binary's shape.
			if problem := nextFromSourceModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0108 §3: a multiplier's vocabulary (Sources,
			// Recipients) is as much a key as its kind, and its fields
			// on another kind are a newer binary's shape.
			if problem := multiplyDamageModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #1879: a granted attack restriction that asks for nothing.
			if problem := defenderControlsModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0109 §1: a land-type set naming no basic land type.
			if problem := landTypesModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0108 PR 8: a follow-up's recipient on a mod that has
			// no follow-up to deal it is a newer binary's shape.
			if problem := followUpModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0108 §9: a redirection's destination is part of what it
			// means, and toSourceController on another kind is a newer
			// binary's shape.
			if problem := redirectDamageModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #1859: a row removal naming a slot this binary cannot
			// honour is a newer binary's shape.
			if problem := loseOwnAbilityModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #2562: a name or type line with nothing in it, or
			// supertypes on another kind, is a newer binary's shape.
			if problem := becomeNamedModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #2144: a creature-type set naming no creature type.
			if problem := setCreatureTypesModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #4072: a subtype loss naming no subtype.
			if problem := loseSubtypesModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #3398: an attack tax with no player or an unreadable cost.
			if problem := attackTaxModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0188: a skip debt owed by no one, or owing nothing.
			if problem := skipTurnsModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// ADR 0188 Decision 3a: a skipped step that names no turn or
			// no real step, or a step field on another kind.
			if problem := skipExtraTurnStepModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #3995: a player-wide untap skip naming no player or no
			// permanent type, or anyController on another kind.
			if problem := skipNextUntapStepModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
			// #2567: an attack-this-permanent requirement naming no
			// permanent, or a turn stamp on another kind.
			if problem := attackPermanentModProblem(m); problem != "" {
				unknown = append(unknown, problem)
			}
		}
		// ADR 0107 §3: a stack pin may carry only what this binary's
		// stack step applies (layer 2 control, layer-6 keywords). A
		// newer binary's spell effect would otherwise restore as a
		// record that silently does nothing.
		if problem := stackPinProblem(e.Affected, e.Mods); problem != "" {
			unknown = append(unknown, "scoped effect "+problem)
		}
		// #2045: a recipient set is read only on an unpinned record.
		if problem := recipientFilterScopeProblem(e.Scope, e.Mods); problem != "" {
			unknown = append(unknown, problem)
		}
		// #1946: a dealt-damage-by record names one dealer per mod, and
		// its fields on any other record are a newer binary's shape.
		if problem := dealtDamageByScopeProblem(e.Scope, e.Mods); problem != "" {
			unknown = append(unknown, problem)
		}
		if !KnownAffectedScope(e.Scope) {
			unknown = append(unknown, "scoped-effect scope "+string(e.Scope))
		}
	}
	return unknown
}

// deepCopyScopedEffects copies records and their inner slices, for the
// snapshot boundary, where nothing may alias live game state.
func deepCopyScopedEffects(in []ScopedEffect) []ScopedEffect {
	if len(in) == 0 {
		return nil
	}
	out := make([]ScopedEffect, len(in))
	for i, e := range in {
		e.Affected = append([]AffectedObject(nil), e.Affected...)
		e.Mods = cloneMods(e.Mods)
		out[i] = e
	}
	return out
}

// unknownScopedEffectFields re-reads the raw snapshot and lists every
// key this binary's effect types do not declare. encoding/json drops an
// unknown key without a word, and on these types an unknown key is
// exactly what a newer build's vocabulary looks like: a field a future
// mod kind reads, a duration field a future condition needs, a param a
// future body reads. So it is refused (ErrUnknownEffectKey), never
// dropped. ADR 0041 P4.
//
// Covered: a scopedEffects record, its affected members and its mods
// (tier 1), and since #4361 S18 every other record the file stores (a
// keyed tail's records, an extra turn's effects) — its duration, like every other stored duration, is
// unknownDurationFields' (ADR 0109 Shared machinery 2); and — #1568 review — every EffectParams a
// delayed trigger or a stack item carries (`params`, `condParams`),
// down through its `filter`, its `object` and — tier 4 — its `ability`.
//
// Tier 4 (ADR 0041 P10, #1497) extends it to the WHOLE stack-item
// record — on the stack, queued, or remembered in `lastKnownStack` —
// for a file of the current schema. That does not protect a binary from
// before this change (the catalog/* body is its refusal token); it
// means every stack-item field added after it is refused by every
// binary from here on. It is asked only of the current schema because
// in an OLDER file an unknown key can only be one a bump removed,
// which is that bump's migration to handle, never a refusal.
//
// ADR 0108 PR 0 extends it the same way to an owed prevention follow-up
// (`preventionFollowUps[]`), for a file of the current schema. The binary
// before this change does not check that list, so it would restore a
// follow-up's `instance` by dropping it and grouping the entry by its
// batch, which it also carries; every binary from here on refuses a
// follow-up field it does not know.
func unknownScopedEffectFields(data []byte, schema int) ([]string, error) {
	var envelope struct {
		ScopedEffects       []map[string]json.RawMessage `json:"scopedEffects"`
		DelayedTriggers     []map[string]json.RawMessage `json:"delayedTriggers"`
		StackMeta           []map[string]json.RawMessage `json:"stackMeta"`
		PendingTriggers     []map[string]json.RawMessage `json:"pendingTriggers"`
		LastKnownStack      []map[string]json.RawMessage `json:"lastKnownStack"`
		Resolving           map[string]json.RawMessage   `json:"resolving"`
		PreventionFollowUps []map[string]json.RawMessage `json:"preventionFollowUps"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	var out []string
	check := func(where string, obj map[string]json.RawMessage, known map[string]bool) {
		for k := range obj {
			if !known[k] {
				out = append(out, fmt.Sprintf("unknown field %q on %s", k, where))
			}
		}
	}
	object := func(raw json.RawMessage) (map[string]json.RawMessage, error) {
		if len(raw) == 0 || string(raw) == "null" {
			return nil, nil
		}
		var obj map[string]json.RawMessage
		err := json.Unmarshal(raw, &obj)
		return obj, err
	}
	nested := func(where string, raw json.RawMessage, known map[string]bool, many bool) error {
		if len(raw) == 0 || string(raw) == "null" {
			return nil
		}
		if !many {
			obj, err := object(raw)
			if err != nil {
				return err
			}
			check(where, obj, known)
			return nil
		}
		var objs []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &objs); err != nil {
			return err
		}
		for _, obj := range objs {
			check(where, obj, known)
		}
		return nil
	}
	params := func(where string, raw json.RawMessage) error {
		obj, err := object(raw)
		if err != nil || obj == nil {
			return err
		}
		check(where, obj, effectParamsJSONKeys)
		if err := nested(where+" filter", obj["filter"], castFilterJSONKeys, false); err != nil {
			return err
		}
		if err := nested(where+" ability", obj["ability"], abilityRefJSONKeys, false); err != nil {
			return err
		}
		if err := nested(where+" objects", obj["objects"], objectRefJSONKeys, true); err != nil {
			return err
		}
		return nested(where+" object", obj["object"], objectRefJSONKeys, false)
	}
	record := func(where string, rec map[string]json.RawMessage) error {
		check(where, rec, scopedEffectJSONKeys)
		if err := nested("an affected object", rec["affected"], affectedObjectJSONKeys, true); err != nil {
			return err
		}
		if err := nested("a mod", rec["mods"], modJSONKeys, true); err != nil {
			return err
		}
		// #2026: a shield's source filter is a mod's own vocabulary, so
		// a key a newer binary added to it is refused here too, down to
		// its queries and its pinned objects.
		if err := sourceFilterFields(rec["mods"], nested); err != nil {
			return err
		}
		// #2045: so is a shield's recipient filter, down to its queries.
		return recipientFilterFields(rec["mods"], nested)
	}
	for _, rec := range envelope.ScopedEffects {
		if err := record("a scoped effect", rec); err != nil {
			return nil, err
		}
	}
	// #4361 S18: a record stored anywhere else in the file is held to the
	// same keys: a keyed tail's records, wherever a tail is written
	// (KeyedTail.Records, #4361 S12), and an extra turn's rules effects
	// (ExtraTurn.Effects, ADR 0188). The routes come from the snapshot
	// types (reachesScopedEffect), so a field added later that stores
	// records is covered the day it lands. The top-level list is checked
	// above, by name.
	var walkErr error
	err := walkRawTo(data, reflect.TypeOf(GameSnapshot{}), "", scopedEffectType, reachesScopedEffect,
		func(where string, rec map[string]json.RawMessage) {
			if where == "scopedEffects[]" || walkErr != nil {
				return
			}
			walkErr = record("a scoped effect at "+where, rec)
		})
	if err == nil {
		err = walkErr
	}
	if err != nil {
		return nil, err
	}
	for _, dt := range envelope.DelayedTriggers {
		if err := params("a delayed trigger's params", dt["params"]); err != nil {
			return nil, err
		}
		if err := params("a delayed trigger's condParams", dt["condParams"]); err != nil {
			return nil, err
		}
	}
	current := schema >= SnapshotSchemaVersion
	stackItem := func(where string, it map[string]json.RawMessage) error {
		if current {
			check(where, it, stackItemJSONKeys)
		}
		return params(where+"'s params", it["params"])
	}
	for _, list := range [][]map[string]json.RawMessage{envelope.StackMeta, envelope.PendingTriggers} {
		for _, it := range list {
			if err := stackItem("a stack item", it); err != nil {
				return nil, err
			}
		}
	}
	if current {
		for _, f := range envelope.PreventionFollowUps {
			check("a prevention follow-up", f, preventionFollowUpJSONKeys)
		}
	}
	for _, e := range envelope.LastKnownStack {
		if current {
			check("a lastKnownStack entry", e, lastKnownSpellJSONKeys)
		}
		it, err := object(e["item"])
		if err != nil {
			return nil, err
		}
		if it != nil {
			if err := stackItem("a lastKnownStack item", it); err != nil {
				return nil, err
			}
		}
	}
	// #4270: the resolving slot is a stack item and a card of its own.
	if envelope.Resolving != nil {
		if current {
			check("the resolving slot", envelope.Resolving, resolvingJSONKeys)
		}
		it, err := object(envelope.Resolving["item"])
		if err != nil {
			return nil, err
		}
		if it != nil {
			if err := stackItem("the resolving slot's item", it); err != nil {
				return nil, err
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// sourceFilterFields walks each mod's sourceFilter entries with
// `nested`, the unknown-key check: the filter's own keys, its Except
// queries' and its ExceptObjects' (#2026).
func sourceFilterFields(mods json.RawMessage, nested func(string, json.RawMessage, map[string]bool, bool) error) error {
	if len(mods) == 0 || string(mods) == "null" {
		return nil
	}
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(mods, &list); err != nil {
		return err
	}
	for _, m := range list {
		raw := m["sourceFilter"]
		if err := nested("a mod's source filter", raw, damageSourceFilterJSONKeys, true); err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		var filters []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &filters); err != nil {
			return err
		}
		for _, f := range filters {
			if err := nested("a source filter's except query", f["except"], permanentQueryJSONKeys, true); err != nil {
				return err
			}
			if err := nested("a source filter's excepted object", f["exceptObjects"], objectRefJSONKeys, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// recipientFilterFields walks each mod's recipientFilter entries with
// `nested`, the unknown-key check: the filter's own keys and its Match
// queries' (#2045).
func recipientFilterFields(mods json.RawMessage, nested func(string, json.RawMessage, map[string]bool, bool) error) error {
	if len(mods) == 0 || string(mods) == "null" {
		return nil
	}
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(mods, &list); err != nil {
		return err
	}
	for _, m := range list {
		raw := m["recipientFilter"]
		if err := nested("a mod's recipient filter", raw, damageRecipientFilterJSONKeys, true); err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		var filters []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &filters); err != nil {
			return err
		}
		for _, f := range filters {
			if err := nested("a recipient filter's query", f["match"], permanentQueryJSONKeys, true); err != nil {
				return err
			}
		}
	}
	return nil
}

var (
	damageRecipientFilterJSONKeys = jsonKeysOf(reflect.TypeOf(DamageRecipientFilter{}))
	damageSourceFilterJSONKeys    = jsonKeysOf(reflect.TypeOf(DamageSourceFilter{}))
	permanentQueryJSONKeys        = jsonKeysOf(reflect.TypeOf(PermanentQuery{}))
	scopedEffectJSONKeys          = jsonKeysOf(reflect.TypeOf(ScopedEffect{}))
	affectedObjectJSONKeys        = jsonKeysOf(reflect.TypeOf(AffectedObject{}))
	modJSONKeys                   = jsonKeysOf(reflect.TypeOf(Mod{}))
	durationJSONKeys              = jsonKeysOf(reflect.TypeOf(Duration{}))
	effectParamsJSONKeys          = jsonKeysOf(reflect.TypeOf(EffectParams{}))
	castFilterJSONKeys            = jsonKeysOf(reflect.TypeOf(CastFilter{}))
	objectRefJSONKeys             = jsonKeysOf(reflect.TypeOf(ObjectRef{}))
	abilityRefJSONKeys            = jsonKeysOf(reflect.TypeOf(AbilityRef{}))
	stackItemJSONKeys             = jsonKeysOf(reflect.TypeOf(stackItemSnapshot{}))
	lastKnownSpellJSONKeys        = jsonKeysOf(reflect.TypeOf(lastKnownSpellSnapshot{}))
	resolvingJSONKeys             = jsonKeysOf(reflect.TypeOf(resolvingSnapshot{}))
	// ADR 0108 PR 0: an owed follow-up's fields are checked too.
	preventionFollowUpJSONKeys = jsonKeysOf(reflect.TypeOf(PreventionFollowUp{}))
)

var (
	scopedEffectType = reflect.TypeOf(ScopedEffect{})

	scopedEffectReachOnce sync.Once
	scopedEffectReach     map[reflect.Type]bool
)

// reachesScopedEffect reports whether a value of type t can hold a
// ScopedEffect record in the JSON encoding/json writes for it (#4361
// S18), as reachesDuration does for a Duration.
func reachesScopedEffect(t reflect.Type) bool {
	scopedEffectReachOnce.Do(func() { scopedEffectReach = snapshotTypeReach(scopedEffectType) })
	return scopedEffectReach[derefType(t)]
}

// scopedEffectRoutes lists every JSON path in a GameSnapshot at which a
// ScopedEffect record is stored, for tests: array elements are "[]" and
// map values "{}"; a recursive type is followed one level.
func scopedEffectRoutes() []string {
	set := map[string]bool{}
	var walk func(t reflect.Type, path string, stack []reflect.Type)
	walk = func(t reflect.Type, path string, stack []reflect.Type) {
		t = derefType(t)
		if t == scopedEffectType {
			set[path] = true
			return
		}
		if !reachesScopedEffect(t) {
			return
		}
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			walk(t.Elem(), path+"[]", stack)
		case reflect.Map:
			walk(t.Elem(), path+"{}", stack)
		case reflect.Struct:
			if slices.Contains(stack, t) {
				return
			}
			for _, f := range jsonFields(t) {
				child := f.name
				if path != "" {
					child = path + "." + f.name
				}
				walk(f.typ, child, append(stack, t))
			}
		}
	}
	walk(reflect.TypeOf(GameSnapshot{}), "", nil)
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// jsonKeysOf is the set of keys encoding/json writes for a struct's
// exported fields: the tag name, or the field name when untagged.
func jsonKeysOf(t reflect.Type) map[string]bool {
	out := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name := f.Name
		if tag, ok := f.Tag.Lookup("json"); ok {
			if n := strings.Split(tag, ",")[0]; n == "-" {
				continue
			} else if n != "" {
				name = n
			}
		}
		out[name] = true
	}
	return out
}
