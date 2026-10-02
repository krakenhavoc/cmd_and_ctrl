package game

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// state_triggers.go — CR 603.8 state triggers (ADR 0107 §1, #1858).
//
// "Some triggered abilities trigger when a game state (such as a player
// controlling no permanents of a particular card type) is true, rather
// than triggering when an event occurs. These abilities trigger as soon
// as the game state matches the condition. They'll go onto the stack at
// the next available opportunity." (CR 603.8)
//
// A row that sets TriggeredAbility.State is one. It watches no event:
// the engine asks it.
//
// # When it is asked (owner decision 1)
//
//   - After every event the engine emits: the trigger harvester's last
//     step (triggerHarvester.OnEvent). That is what makes CR 603.8's own
//     example come out right — a hand that is empty only for a moment,
//     in the middle of "discard your hand, then draw that many cards",
//     still triggers "whenever you have no cards in hand".
//   - In each pass of the CR 704.3 loop (runStateChecksLocked), after the
//     state-based actions and before the drain. That is the
//     authoritative check: it catches a state no event announced (a
//     continuous effect that started or ended in a layer pass).
//
// Not inside an event the rules see as ONE event that the engine emits
// in pieces. A board wipe or a state-based-action sweep moves its
// permanents one at a time under an open simultaneous exit
// (simultaneous.go), and combat damage is dealt attacker by attacker; a
// board that is half-moved, or a life total that has taken half the
// damage, is not a game state, and a state trigger must not see it
// (CR 603.10: objects are checked "immediately after an event"). Those
// sections hold the per-event check (stateTriggerHold); the next event,
// or the CR 704.3 loop, reads the settled board.
//
// # The latch (CR 603.8)
//
// "A state-triggered ability doesn't trigger again until the ability has
// resolved, has been countered, or has otherwise left the stack. Then, if
// the object with the ability is still in the same zone and the game
// state still matches its trigger condition, the ability will trigger
// again." The latch is DERIVED, never stored, so a snapshot needs no
// field for it: an ability is latched while any of these holds an item
// of the same row from the same object (ObjectRef, CR 400.7):
//
//   - the trigger queue (PendingTriggers), waiting to go on the stack;
//   - an open announcement prompt (its "you may", its mode pick or its
//     target pick — CR 603.3c-d happen as it is put on the stack);
//   - the stack (StackMeta);
//   - the resolving slot, while that resolution is still open: the
//     ability has not "resolved" until its resolution finishes.
//
// A copy of the ability (Strionic Resonator, CR 707.10) is not the
// ability, and does not latch it.
//
// # A state about each of several objects (#1858)
//
// Bomb Squad's "Whenever a creature has four or more fuse counters on
// it, remove all fuse counters from it and destroy it" is a state of
// ONE creature, and the battlefield can hold several creatures in it at
// once. The game state that matches the trigger condition is "this
// creature has four or more fuse counters", so the ability triggers once
// for each creature in that state, and CR 603.8's "doesn't trigger again
// until the ability has resolved" is about the instance that creature's
// state triggered: an instance about one creature waiting on the stack
// does not stop a second creature reaching four counters from
// triggering it. Bomb Squad's ruling says the same from the other side:
// two Bomb Squads both trigger for one creature, because each has the
// ability.
//
// Such a row registers a StateEachCondition (RegisterStateEachCondition)
// instead of a StateCondition, in the same key namespace. The walk asks
// it once per battlefield permanent, and each instance it queues carries
// the permanent it is about as its trigger context's Object
// (TriggerContext.Object: the ID and the epoch, CR 400.7), with no event
// kind, since nothing happened. The latch reads that Object too, so it
// stays derived: the object an instance is about is already snapshot
// data on every queued and stacked item (its "trigger" key, recorded by
// the shape guard), and a restored table's waiting instance still
// latches its creature. The effect reads the creature through
// effects.Context.TriggeringPermanent, which answers with last-known
// information once it has gone (CR 608.2h).
//
// # What it costs
//
// The registry tells the engine which catalog keys carry a state trigger
// as it files them (IdentifyCatalogRows), so the per-event walk is one
// map lookup per battlefield permanent, and the layer cache is only
// caught up when some permanent might have one.

// stateTriggerKeys is the set of catalog keys whose definition has at
// least one state-trigger row: a card, a face, a token template, an
// emblem or a granted bundle. Written at registration (IdentifyCatalogRows),
// read on every event; the RWMutex keeps a test that registers a card
// while another test's game runs race-clean.
var stateTriggerKeys struct {
	sync.RWMutex
	keys map[string]struct{}
}

// StateCondition is a state trigger's condition (TriggeredAbility.State
// names one): is the game in the printed state, for this permanent and
// its controller? A pure read of the board, called under g.mu in write
// mode; it MUST NOT call public locking mutators.
type StateCondition func(g *Game, source *Card, controller uuid.UUID) bool

// stateConditions is the registry TriggeredAbility.State keys into. It
// is package-level and no Game field reaches it, which is the point:
// the condition is code, and the row carries only its name.
var stateConditions struct {
	sync.RWMutex
	m map[string]stateConditionEntry
}

// StateEachCondition is the condition of a state trigger about each of
// several objects (#1858, Bomb Squad's "Whenever a creature has four or
// more fuse counters on it"): is `obj`, a permanent on the battlefield,
// in the printed state? `source` and `controller` are the permanent with
// the ability and its controller. The same contract as StateCondition: a
// pure read under g.mu in write mode that MUST NOT call public locking
// mutators.
type StateEachCondition func(g *Game, source *Card, controller uuid.UUID, obj *Card) bool

// stateConditionEntry is one registered condition. Exactly one of the
// two forms is set.
type stateConditionEntry struct {
	whole StateCondition
	each  StateEachCondition
}

// RegisterStateCondition files `cond` under `key`, for a catalog row to
// name in TriggeredAbility.State. Call it at init, from the catalog's
// constructors (effects.WhenState). Panics on an empty key, a nil
// condition or a key already taken, the way the catalog's Register
// does: two rows that meant different conditions under one name would
// otherwise silently share one.
//
// The key never reaches a snapshot (see TriggeredAbility.State), so an
// unknown key is not a restore question. A row naming a key this binary
// has not registered never triggers; the catalog's Register refuses such
// a row at boot, so that can only be a hand-built test row.
func RegisterStateCondition(key string, cond StateCondition) {
	if cond == nil {
		panic(fmt.Sprintf("game: state condition %q has no function", key))
	}
	registerStateConditionEntry(key, stateConditionEntry{whole: cond})
}

// RegisterStateEachCondition files a condition about each of several
// objects under `key` (the file comment's "A state about each of several
// objects"): the row that names it triggers once for each battlefield
// permanent `cond` accepts. It shares RegisterStateCondition's key
// namespace and panics.
func RegisterStateEachCondition(key string, cond StateEachCondition) {
	if cond == nil {
		panic(fmt.Sprintf("game: state condition %q has no function", key))
	}
	registerStateConditionEntry(key, stateConditionEntry{each: cond})
}

func registerStateConditionEntry(key string, e stateConditionEntry) {
	if key == "" {
		panic("game: a state condition needs a key")
	}
	stateConditions.Lock()
	defer stateConditions.Unlock()
	if stateConditions.m == nil {
		stateConditions.m = make(map[string]stateConditionEntry)
	}
	if _, dup := stateConditions.m[key]; dup {
		panic(fmt.Sprintf("game: state condition %q is already registered", key))
	}
	stateConditions.m[key] = e
}

// StateConditionFor is the condition registered under key, when it is a
// condition over the board as a whole.
func StateConditionFor(key string) (StateCondition, bool) {
	e, _ := stateConditionEntryFor(key)
	return e.whole, e.whole != nil
}

// StateEachConditionFor is the condition registered under key, when it
// is a condition about each of several objects.
func StateEachConditionFor(key string) (StateEachCondition, bool) {
	e, _ := stateConditionEntryFor(key)
	return e.each, e.each != nil
}

// StateConditionKnown reports whether a condition of either form is
// registered under key. The catalog's Register refuses a row naming a
// key that is not.
func StateConditionKnown(key string) bool {
	_, ok := stateConditionEntryFor(key)
	return ok
}

func stateConditionEntryFor(key string) (stateConditionEntry, bool) {
	stateConditions.RLock()
	defer stateConditions.RUnlock()
	e, ok := stateConditions.m[key]
	return e, ok
}

// noteStateTriggerKey records that key's definition carries a state
// trigger. Called by IdentifyCatalogRows.
func noteStateTriggerKey(key string, d *CardDef) {
	has := false
	for i := range d.Triggered {
		if d.Triggered[i].State != "" {
			has = true
			break
		}
	}
	if !has {
		return
	}
	stateTriggerKeys.Lock()
	defer stateTriggerKeys.Unlock()
	if stateTriggerKeys.keys == nil {
		stateTriggerKeys.keys = make(map[string]struct{})
	}
	stateTriggerKeys.keys[key] = struct{}{}
}

// mayHaveStateTrigger is the cheap pre-filter: could this permanent's
// ability key name a definition with a state trigger? It reads the
// object's identity and its grants, not its layered abilities, so a yes
// is only a candidate — TriggersForCard, on a caught-up layer cache, is
// the answer (a permanent under an ability-removing effect has none).
func mayHaveStateTrigger(c *Card, keys map[string]struct{}) bool {
	has := func(k string) bool {
		_, ok := keys[k]
		return ok && k != ""
	}
	// The object's own key — its face, its token key, the abilities a
	// copy effect granted it — and then each layer-6 grant. The layer
	// cache may be behind the event that brought us here, which is why
	// this is only a pre-filter: an own key the cache would have
	// emptied (a new ability-removing effect) is still a candidate, and
	// TriggersForCard on the caught-up cache says no.
	key := CatalogKey(*c)
	if has(key) {
		return true
	}
	if strings.Contains(key, grantKeySeparator) {
		for _, part := range strings.Split(key, grantKeySeparator) {
			if has(part) {
				return true
			}
		}
	}
	if c.effective != nil {
		for _, gr := range c.effective.GrantedAbilities {
			if has(gr.Key) {
				return true
			}
		}
	}
	return false
}

// holdStateTriggersLocked opens a section the per-event check skips
// (see the file comment). The returned func closes it; defer it.
// Caller must hold g.mu.
func (g *Game) holdStateTriggersLocked() func() {
	g.stateTriggerHold++
	return func() {
		if g.stateTriggerHold > 0 {
			g.stateTriggerHold--
		}
	}
}

// stateTriggersAfterEventLocked is the per-event check, run by the
// trigger harvester after its zone walks. Caller must hold g.mu.
func (g *Game) stateTriggersAfterEventLocked() {
	if g.stateTriggerHold > 0 || len(g.simultaneousExit) > 0 {
		return
	}
	g.stateTriggersLocked()
}

// stateTriggersLocked asks every battlefield permanent's state triggers
// whether the game state matches them, and queues each one that does
// and is not latched, as an ordinary triggered ability with the
// permanent as its source: its "you may", mode pick and target pick run
// through the same dispatch an event trigger's do (CR 603.3c-d), and it
// goes on the stack at the next drain.
//
// Re-entrant calls (an event emitted while this runs — a layer
// recompute's control-change events, a prompt's own events) return at
// once: the walk in progress reads the board after them anyway.
//
// Caller must hold g.mu in write mode.
func (g *Game) stateTriggersLocked() {
	if g.stateTriggerChecking || g.State != StateActive || g.Battlefield == nil || CatalogTriggers == nil {
		return
	}
	stateTriggerKeys.RLock()
	keys := stateTriggerKeys.keys
	stateTriggerKeys.RUnlock()
	if len(keys) == 0 {
		return
	}
	candidate := false
	for i := range g.Battlefield.Cards {
		if mayHaveStateTrigger(&g.Battlefield.Cards[i], keys) {
			candidate = true
			break
		}
	}
	if !candidate {
		return
	}
	g.stateTriggerChecking = true
	defer func() { g.stateTriggerChecking = false }()
	// The condition reads types, controllers and abilities, and the
	// event that brought us here may have left the layer cache behind
	// (an Island-making Aura just entered, a creature just changed
	// control). Catch it up first; a recompute's own events re-enter
	// and return above.
	g.RecomputeLayersIfStaleLocked()
	type match struct {
		source Card
		t      TriggeredAbility
		// about is the permanent a per-object instance is about
		// (StateEachCondition); uuid.Nil for a state of the board.
		about uuid.UUID
	}
	var matches []match
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !mayHaveStateTrigger(c, keys) {
			continue
		}
		for _, t := range TriggersForCard(*c) {
			if t.State == "" || !TriggerWatchesFromZone(t, ZoneBattlefield) {
				continue
			}
			e, _ := stateConditionEntryFor(t.State)
			switch {
			case e.each != nil:
				// One instance per permanent in the state, each latched
				// on its own (the file comment's "A state about each of
				// several objects").
				for j := range g.Battlefield.Cards {
					obj := &g.Battlefield.Cards[j]
					about := ObjectRef{ID: obj.InstanceID, Epoch: obj.ObjectEpoch}
					if g.stateTriggerLatchedLocked(c, t, &about) || !e.each(g, c, c.Controller, obj) {
						continue
					}
					matches = append(matches, match{source: *c, t: t, about: obj.InstanceID})
				}
			case e.whole != nil:
				if g.stateTriggerLatchedLocked(c, t, nil) || !e.whole(g, c, c.Controller) {
					continue
				}
				matches = append(matches, match{source: *c, t: t})
			}
		}
	}
	// Dispatched after the walk: a dispatch can open a prompt, and
	// nothing it does may move the slice the walk was reading.
	for _, m := range matches {
		// A per-object instance names its permanent the way an event
		// trigger names the object its event was about, so the context
		// the dispatch builds carries that object's ID and epoch
		// (triggerContextLocked) and stampTriggerContext puts it on the
		// item. The event has no Kind, because nothing happened:
		// TriggerContext.Fired stays false.
		g.dispatchTriggerInstanceLocked(Event{CardID: m.about}, m.source, m.source.Effective(), m.t, doublerRef{})
	}
}

// stateTriggerLatchedLocked is CR 603.8's "doesn't trigger again until
// the ability has resolved, has been countered, or has otherwise left
// the stack", read off the places an item of this row from this object
// can be (see the file comment). `about`, when set, is the permanent a
// per-object instance is about, and only an instance about that same
// object (CR 400.7) latches it. Caller must hold g.mu.
func (g *Game) stateTriggerLatchedLocked(source *Card, t TriggeredAbility, about *ObjectRef) bool {
	obj := ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
	ref := TriggeredAbilityRef(t)
	isAbout := func(tc *TriggerContext) bool {
		if about == nil {
			return true
		}
		return tc != nil && tc.Object != nil && tc.Object.Ref() == *about
	}
	isThis := func(it *StackItem) bool {
		if it == nil || it.IsCopy || it.SourceObject != obj || !isAbout(it.Trigger) {
			return false
		}
		if ref != nil && it.Params.Ability != nil {
			return it.Params.Ability.Key == ref.Key && it.Params.Ability.Ref == ref.Ref
		}
		return it.Label == t.Key
	}
	for _, it := range g.PendingTriggers {
		if isThis(it) {
			return true
		}
	}
	for _, it := range g.StackMeta {
		if isThis(it) {
			return true
		}
	}
	if g.resolutionOpen && g.resolving != nil && isThis(g.resolving.item) {
		return true
	}
	sameRow := func(s Card, a TriggeredAbility) bool {
		if s.InstanceID != obj.ID || s.ObjectEpoch != obj.Epoch {
			return false
		}
		if t.row.key != "" {
			return a.row == t.row
		}
		return a.Key == t.Key
	}
	for _, c := range g.PendingChoices {
		if c == nil {
			continue
		}
		switch {
		case c.triggerResume != nil && sameRow(c.triggerResume.source, c.triggerResume.ability) && isAbout(&c.triggerResume.tc):
			return true
		case c.pickTargetResume != nil && sameRow(c.pickTargetResume.source, c.pickTargetResume.ability) && isAbout(&c.pickTargetResume.tc):
			return true
		case c.modePickResume != nil && sameRow(c.modePickResume.source, c.modePickResume.ability) && isAbout(&c.modePickResume.tc):
			return true
		}
	}
	return false
}

// ControlsMatchingForEffect reports whether `player` controls a
// permanent other than `except` that matches any of qs — the read behind
// "When you control no Islands" and "no OTHER creatures" (ADR 0107 §1,
// effects.WhenYouControlNo). uuid.Nil for `except` excludes nothing. It
// reads current characteristics, through the matcher ADR 0107 §2's
// attack restriction uses. Caller must hold g.mu.
func (g *Game) ControlsMatchingForEffect(player, except uuid.UUID, qs ...PermanentQuery) bool {
	if g.Battlefield == nil || player == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player || (except != uuid.Nil && c.InstanceID == except) {
			continue
		}
		for _, q := range qs {
			if q.matchesLocked(g, c) {
				return true
			}
		}
	}
	return false
}

// CountControlledMatchingForEffect is how many permanents `player`
// controls that match any of qs — "When you control seven or more
// Thrulls". Caller must hold g.mu.
func (g *Game) CountControlledMatchingForEffect(player uuid.UUID, qs ...PermanentQuery) int {
	if g.Battlefield == nil || player == uuid.Nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player {
			continue
		}
		for _, q := range qs {
			if q.matchesLocked(g, c) {
				n++
				break
			}
		}
	}
	return n
}

// AnyPermanentMatchingForEffect reports whether any permanent on the
// battlefield, whoever controls it, matches any of qs — "When there are
// no creatures on the battlefield". Caller must hold g.mu.
func (g *Game) AnyPermanentMatchingForEffect(qs ...PermanentQuery) bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		for _, q := range qs {
			if q.matchesLocked(g, &g.Battlefield.Cards[i]) {
				return true
			}
		}
	}
	return false
}
