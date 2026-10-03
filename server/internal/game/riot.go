package game

import (
	"errors"

	"github.com/google/uuid"
)

// riot.go — riot (CR 702.136) and unleash (CR 702.98), the two
// keyword-shaped entry replacements (ADR 0109 §10, #1556).
//
//	CR 702.136a  Riot is a static ability. "Riot" means "You may have
//	             this permanent enter with an additional +1/+1 counter
//	             on it. If you don't, it gains haste."
//	CR 702.136b  If a permanent has multiple instances of riot, each
//	             works separately.
//	CR 702.98a   Unleash is a keyword that represents two static
//	             abilities. "Unleash" means "You may have this permanent
//	             enter with an additional +1/+1 counter on it" and "This
//	             permanent can't block as long as it has a +1/+1 counter
//	             on it."
//
// Both are entry replacements (CR 614.1c), and which ones apply is
// decided by the permanent "as it would exist on the battlefield"
// (CR 614.12): the entry look-ahead (entry_lookahead.go) counts the
// instances, and the gather adds one replacement per instance. So a
// printed riot under Rhythm of the Wild asks twice, a creature entering
// under Dress Down is not asked at all, and a riot granted to the spell
// on the stack (CR 400.7a) is asked like a printed one. The choice is
// made before the permanent enters (CR 614.12a).
//
// # Riot's question
//
// PendingChoiceEntryRiot, a mandatory two-answer prompt. The counter is
// the effect's Replace; haste sets ReplacementEvent.EntersWithHaste,
// which the landing turns into an indefinite haste record pinned to the
// permanent (grantRiotHasteLocked) — data, the earthbend shape. An entry
// that cannot pause takes the counter (ADR 0109 §10 decision 3), the
// default an entry_controller question takes in the same place.
//
// # Unleash
//
// Its first half is an ordinary Optional entry counter, asked through
// the existing "may" prompt. Its second half is a rule: foldUnleashLocked
// adds CantBlock to a permanent that has unleash and a +1/+1 counter,
// once the layer pass is over, so the block gate, the enumerator and
// the view all read it through Restricted.

const (
	// KeywordRiot is CR 702.136's canonical token.
	KeywordRiot = "riot"
	// KeywordUnleash is CR 702.98's canonical token.
	KeywordUnleash = "unleash"
)

// PendingChoiceEntryRiot is riot's question: "a +1/+1 counter, or
// haste?" Answered with the `{apply: bool}` payload the yes/no kinds
// share — `true` takes the counter (the "you may"), `false` takes haste
// (the "if you don't") — and routed by kind in the actions dispatcher.
// The chooser is the entering permanent's would-be controller.
//
// A kind of its own rather than an optional_replacement because "no" is
// not "nothing": riot is mandatory, and an Optional decline runs no
// replacement at all.
const PendingChoiceEntryRiot PendingChoiceKind = "entry_riot"

// RiotHasteLabel is the label of the haste record riot's "if you don't"
// writes. The view reads it to name the haste chip "Riot"
// (RiotHasteForEffect).
const RiotHasteLabel = "riot — haste"

// The keyword entry replacements' IDs sit in the self-replacement range,
// above the entering card's own slots (one stride) and the copied card's
// (a second): riot's instances at the third stride, unleash's at the
// fourth. One ID per instance, so CR 614.5's once-per-event tracking
// asks each instance once.
const (
	riotReplacementIDBase    = selfReplacementIDBase + 2*MaxCatalogReplacementSlots
	unleashReplacementIDBase = selfReplacementIDBase + 3*MaxCatalogReplacementSlots
)

// entryKeywordReplacementID is the ID of instance `i` of an entry
// keyword.
func entryKeywordReplacementID(keyword string, i int) ReplacementEffectID {
	if keyword == KeywordUnleash {
		return unleashReplacementIDBase + ReplacementEffectID(i)
	}
	return riotReplacementIDBase + ReplacementEffectID(i)
}

// EntryKeywordOfReplacement reports which entry keyword a replacement
// ID stands for — KeywordRiot, KeywordUnleash, or "" for any other
// effect. The view reads it to tell a policy and the table which
// question an optional_replacement or entry_riot prompt is.
func EntryKeywordOfReplacement(id ReplacementEffectID) string {
	switch {
	case id >= riotReplacementIDBase && id < riotReplacementIDBase+MaxCatalogReplacementSlots:
		return KeywordRiot
	case id >= unleashReplacementIDBase && id < unleashReplacementIDBase+MaxCatalogReplacementSlots:
		return KeywordUnleash
	}
	return ""
}

// entryKeywordLabel is the name an entry keyword's replacement carries
// into a CR 616 ordering prompt.
func entryKeywordLabel(keyword string) string {
	switch keyword {
	case KeywordRiot:
		return "Riot"
	case KeywordUnleash:
		return "Unleash"
	}
	return ""
}

// sameEntryKeyword reports whether every gathered replacement is an
// instance of one entry keyword. Two riots work separately (CR
// 702.136b) and ask the same question, so ordering them is a prompt
// with one answer.
func sameEntryKeyword(applicable []activeReplacement) bool {
	if len(applicable) < 2 {
		return false
	}
	kw := applicable[0].effect.entryKeyword
	if kw == "" {
		return false
	}
	for _, a := range applicable[1:] {
		if a.effect.entryKeyword != kw {
			return false
		}
	}
	return true
}

// gatherEntryKeywordReplacementsLocked appends one replacement per
// riot and unleash instance the look-ahead finds on the entering
// permanent, skipping those already applied to this event.
//
// Caller must hold g.mu (write).
func (g *Game) gatherEntryKeywordReplacementsLocked(ev *ReplacementEvent, applied map[ReplacementEffectID]bool, out []activeReplacement) []activeReplacement {
	la := g.entryLookAheadLocked(ev)
	if la.riot == 0 && la.unleash == 0 {
		return out
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return out
	}
	src := entering
	for i := 0; i < la.riot && i < MaxCatalogReplacementSlots; i++ {
		if id := entryKeywordReplacementID(KeywordRiot, i); !applied[id] {
			out = append(out, activeReplacement{effect: riotReplacement(src.Name), source: &src, id: id})
		}
	}
	for i := 0; i < la.unleash && i < MaxCatalogReplacementSlots; i++ {
		if id := entryKeywordReplacementID(KeywordUnleash, i); !applied[id] {
			out = append(out, activeReplacement{effect: unleashReplacement(src.Name), source: &src, id: id})
		}
	}
	return out
}

// riotReplacement is one riot instance's entry replacement. Its Replace
// is the counter; haste is the other answer, which the prompt's resolve
// writes (ResolveEntryRiot).
func riotReplacement(name string) ReplacementEffect {
	return ReplacementEffect{
		Watches:         []EventKind{EventZoneMove},
		Replace:         addEntryPlusOneCounter,
		Controller:      entryKeywordController,
		SelfReplacement: true,
		PromptQuestion:  "Riot — " + name + " enters with your choice of a +1/+1 counter or haste",
		Label:           "Riot",
		entryKeyword:    KeywordRiot,
	}
}

// unleashReplacement is unleash's first half: "you may have this
// permanent enter with an additional +1/+1 counter on it".
func unleashReplacement(name string) ReplacementEffect {
	return ReplacementEffect{
		Watches:         []EventKind{EventZoneMove},
		Replace:         addEntryPlusOneCounter,
		Controller:      entryKeywordController,
		SelfReplacement: true,
		Optional:        true,
		PromptQuestion:  "Unleash — have " + name + " enter with a +1/+1 counter? It can't block while it has one.",
		Label:           "Unleash",
		entryKeyword:    KeywordUnleash,
	}
}

// addEntryPlusOneCounter is the "additional +1/+1 counter" both
// keywords give.
func addEntryPlusOneCounter(ev *ReplacementEvent, _ *Game, _ *Card) error {
	ev.AddCounterAtETB(CounterPlusOne, 1)
	return nil
}

// entryKeywordController is who answers: the player the permanent would
// enter under (ev.Actor, which an entry_controller answer rewrites),
// then the entering card's controller, then its owner.
func entryKeywordController(ev *ReplacementEvent, _ *Game, src *Card) uuid.UUID {
	if ev != nil && ev.Actor != uuid.Nil {
		return ev.Actor
	}
	if src == nil {
		return uuid.Nil
	}
	if src.Controller != uuid.Nil {
		return src.Controller
	}
	return src.Owner
}

// offerEntryRiotLocked asks riot's question, or settles it on the
// counter when it cannot be asked: the entry cannot pause (mustSettleNow,
// or nothing can resume it), or its chooser has left or cannot be named.
// Returns true when the prompt was queued and the caller must bail.
//
// Caller must hold g.mu.
func (g *Game) offerEntryRiotLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	chooser := g.entryChoicePlayerLocked(ev, chosen)
	if ev.mustSettleNow || !g.optionalReplacementResumableLocked(ev) || chooser == uuid.Nil || g.chooserGoneLocked(chooser) {
		g.settleRiotOnTheCounterLocked(ev, chosen)
		return false
	}
	var source uuid.UUID
	if chosen.source != nil {
		source = chosen.source.InstanceID
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:                 PendingChoiceEntryRiot,
		Chooser:              chooser,
		FromPlayer:           chooser,
		Count:                1,
		Source:               source,
		Reason:               chosen.effect.PromptQuestion,
		AcceptLabel:          "+1/+1 counter",
		DeclineLabel:         "Haste",
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	})
	if id == uuid.Nil {
		// QueueChoiceForEffect refused the chooser (#864): the default
		// stands rather than a paused entry nobody can resume.
		g.settleRiotOnTheCounterLocked(ev, chosen)
		return false
	}
	return true
}

// settleRiotOnTheCounterLocked applies riot's default: the counter.
//
// Caller must hold g.mu.
func (g *Game) settleRiotOnTheCounterLocked(ev *ReplacementEvent, chosen activeReplacement) {
	g.markReplacementAppliedLocked(ev, chosen.id)
	if chosen.effect.Replace != nil && !ev.Canceled {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}
}

// ResolveEntryRiot answers a PendingChoiceEntryRiot. `counter` true
// takes the +1/+1 counter; false takes haste. Either way the instance
// is marked applied (CR 614.5), the apply-loop is re-entered so CR
// 616.1f gathers anything left — a second riot asks next — and the
// settled entry is landed.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryRiot(choiceID, chooserID uuid.UUID, counter bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceEntryRiot {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil || len(frame.applicable) == 0 {
		return ErrInvalidParam
	}
	if g.dropStaleReplacementResumeLocked(frame) {
		return nil
	}
	ev := frame.ev
	chosen := frame.applicable[0]
	if counter {
		g.settleRiotOnTheCounterLocked(ev, chosen)
	} else {
		g.markReplacementAppliedLocked(ev, chosen.id)
		ev.EntersWithHaste = true
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

// grantRiotHasteLocked writes riot's "it gains haste" for a permanent
// that has just entered with EntersWithHaste: one addKeywords record,
// pinned to the permanent by its entry stamp and lasting for as long as
// it is that object (CR 400.7). It is data, so a table holding one is a
// restore point, and it sorts in layer 6 at its own timestamp, so an
// ability-removing effect that arrives later takes the haste away with
// everything else (CR 613.7).
//
// Called from announceEntryLocked once the zone move has stamped the
// permanent, and before EventETB.
//
// Caller must hold g.mu (write).
func (g *Game) grantRiotHasteLocked(ev *ReplacementEvent, entered uuid.UUID) {
	if ev == nil || !ev.EntersWithHaste {
		return
	}
	c, ok := g.battlefieldCardLocked(entered)
	if !ok {
		return
	}
	g.registerScopedEffectLocked(entered,
		[]AffectedObject{PinObject(entered, c.EnteredBattlefieldAt)},
		[]Mod{AddKeywordsMod("haste")},
		g.PinnedTo(IndefiniteDuration(), entered), RiotHasteLabel, timeNowUnixNano())
}

// RiotHasteForEffect reports whether the permanent `id` has haste from
// its own riot — a live riot haste record names it. The view reads it to
// label the haste chip "Riot".
//
// Caller must hold g.mu.
func (g *Game) RiotHasteForEffect(id uuid.UUID) bool {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Label == RiotHasteLabel && affectedPredicate(e.Affected)(c, g, nil) {
			return true
		}
	}
	return false
}

// foldUnleashLocked is unleash's second half (CR 702.98a): "This
// permanent can't block as long as it has a +1/+1 counter on it." A
// rules effect of the permanent's own ability, applied once the layer
// pass is over (CR 613.11), so it reads the finished ability list — a
// permanent that has lost unleash may block — and the counters the pass
// counted. A counter placed or removed bumps the layer version, so the
// next read sees the change.
//
// Caller must hold g.mu (write).
func (g *Game) foldUnleashLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil || c.Counters[CounterPlusOne] <= 0 || !containsKeyword(c.effective.Abilities, KeywordUnleash) {
			continue
		}
		c.effective.Restrictions |= CantBlock
	}
}
