package game

import "github.com/google/uuid"

// triggers.go is the S19 auto-fire dispatcher: a single per-game
// listener that watches the event log and queues triggered abilities
// onto PendingTriggers without players having to click "trigger" by
// hand. Replaces the manual AnnounceTrigger path from S13.1 for any
// catalog card that declares a TriggeredAbility — manual announce
// stays for non-catalog and "hidden info" cases (cards in hand with
// cast-replacement triggers).
//
// Architecture:
//
//   - One process-lifetime triggerHarvester{} listener registered at
//     NewGame, sitting alongside layerVersionBump from S16. The
//     harvester carries no state — it walks per-event.
//   - Each card declares its triggers in effects.Spec.Triggered.
//     Cards/effects/wire.go bridges the slice into the game package
//     via the CatalogTriggers hook (effect_hooks.go).
//   - On every Event the harvester walks the battlefield, looks up
//     each card's TriggeredAbilities, and for matching event kinds
//     calls Predicate then Build. A non-nil StackItem from Build is
//     appended to g.PendingTriggers — same queue AnnounceTrigger
//     uses, drained APNAP-style by the existing S13.1 pipeline.
//   - LTB-style triggers (dies-triggers) need pre-move
//     characteristics (CR 603.10 last-known-information). Game keeps
//     a per-game snapshot map updated by the few LTB-emitting sites
//     in effect_api.go — see snapshotLKILocked + lastKnownBattlefield.
//
// Sub-PR 1 ships the dispatcher framework with zero card migrations.
// Sub-PRs 3-7 fill in the per-card Triggered declarations.

// TriggeredAbility declares one auto-fire trigger on a catalog card.
// Cards declare a list (effects.Spec.Triggered) — typical creatures
// have zero entries, ETB-trigger creatures have one or two.
//
// The shape mirrors StaticAbility / ReplacementEffect: a Watches
// pre-filter for cheap event-kind dispatch, an AppliesTo predicate
// evaluated per event when the kind matches, and a Build callback
// that constructs the StackItem to enqueue.
type TriggeredAbility struct {
	// Watches is the set of event kinds this ability cares about. The
	// harvester skips the AppliesTo / Build pair when the live event's
	// Kind isn't in this list — a cheap pre-filter that keeps the
	// per-event walk linear in matching cards rather than total cards.
	// Empty / nil means "never fires" — declaring a trigger with no
	// Watches is a programming error and the harvester skips it.
	Watches []EventKind

	// AppliesTo decides whether this specific event triggers this
	// specific source card. Receives a live pointer to the source on
	// the battlefield (or wherever the harvester found it) and the
	// last-known-information snapshot for LTB-style cases. For an ETB
	// trigger the AppliesTo typically reads ev.CardID == source.InstanceID;
	// for a "whenever an opponent draws" trigger it checks ev.Actor !=
	// source.Controller. Returning false silently skips the trigger.
	//
	// Runs under g.mu held in write mode (the harvester runs from
	// notifyListenersLocked). MUST NOT call public locking mutators.
	AppliesTo func(ev Event, source *Card, sourceLKI Characteristic, g *Game) bool

	// Build constructs the StackItem to append to PendingTriggers.
	// Returning nil suppresses the trigger — used for optional
	// triggers when the controller declines (sub-PR 2 routes the
	// declination through OptionalPrompt; mandatory Build callbacks
	// rarely return nil).
	//
	// Runs under g.mu held in write mode. MUST NOT call public
	// locking mutators.
	Build func(ev Event, source *Card, sourceLKI Characteristic, g *Game) *StackItem

	// OptionalPrompt is the declarative "you may" gate for CR 603.4
	// optional triggers. When non-nil, the harvester does NOT call
	// Build immediately on a matching event — it queues a
	// PendingChoiceTriggerPrompt to the source's controller (or the
	// override-chooser specified in OptionalPrompt). On
	// resolve_choice with `apply: true`, Build runs against the
	// captured event + LKI and queues the StackItem. On `apply:
	// false`, the trigger drops without effect.
	//
	// Nil means mandatory — Build runs unconditionally. Modal-choice
	// triggers ("draw a card OR gain 3 life") aren't represented
	// here; they need a separate ModePrompt slot when the first
	// modal catalog card ships. Added in S19 sub-PR 2.
	OptionalPrompt *TriggerOptionalPrompt
}

// TriggerOptionalPrompt is the declarative payload for the "ask
// before firing" gate. Question is rendered in the client prompt
// dialog; Chooser optionally overrides the default chooser
// (source.Controller) for opponent-prompted triggers.
type TriggerOptionalPrompt struct {
	// Question is the dialog header text. Empty falls back to the
	// source card's name on the client.
	Question string

	// Chooser overrides the default chooser (source.Controller) for
	// triggers that prompt someone other than the source's
	// controller. Nil means "use source.Controller". Runs under
	// g.mu — MUST NOT take public locks.
	Chooser func(ev Event, source *Card, g *Game) uuid.UUID
}

// triggerHarvester is the process-lifetime listener that
// auto-announces triggered abilities. Stateless; reads g state and
// the catalog hook on every event. Registered at NewGame next to
// layerVersionBump.
type triggerHarvester struct{}

// OnEvent walks the battlefield (and the LKI map for LTB events),
// collects matching catalog-declared TriggeredAbility entries, and
// appends the Build output onto g.PendingTriggers. Caller (the
// notifyListenersLocked path) holds g.mu in write mode.
//
// The harvester does NOT emit EventTrigger here — that event is the
// "added to PendingTriggers" breadcrumb and is emitted once per
// queued item below. EventTrigger lets downstream listeners observe
// trigger creation without polling the queue.
func (triggerHarvester) OnEvent(g *Game, ev Event) {
	if CatalogTriggers == nil {
		return
	}
	// Skip the harvester's own follow-on emits — EventTrigger is the
	// "queued onto PendingTriggers" breadcrumb. A trigger that fires
	// further triggers happens at resolution time, not at announce
	// time, so re-entry through EventTrigger here would just spam.
	if ev.Kind == EventTrigger {
		return
	}
	g.harvestFromZone(ev, g.Battlefield)
	// LTB-style events: the source has already moved off the
	// battlefield, so harvestFromZone(BF) won't find it. The card
	// lives in its destination zone now (graveyard / exile / hand);
	// the harvester finds it there and pulls battlefield
	// characteristics from the LKI snapshot.
	if ev.Kind == EventLTB && ev.CardID != uuid.Nil {
		g.harvestLTB(ev)
	}
}

// harvestFromZone is the per-zone scan used for "live" triggers (ETB,
// cast, draw, combat damage, upkeep). For every battlefield card with
// a TriggeredAbility whose Watches contains ev.Kind, run AppliesTo
// then Build and queue the result.
func (g *Game) harvestFromZone(ev Event, z *Zone) {
	if z == nil || CatalogTriggers == nil {
		return
	}
	for i := range z.Cards {
		card := &z.Cards[i]
		oracle := card.OracleID
		if oracle == "" {
			continue
		}
		triggers := CatalogTriggers(oracle)
		if len(triggers) == 0 {
			continue
		}
		lki := card.Effective()
		for _, t := range triggers {
			if !triggerWatches(t.Watches, ev.Kind) {
				continue
			}
			if t.AppliesTo != nil && !t.AppliesTo(ev, card, lki, g) {
				continue
			}
			if t.OptionalPrompt != nil {
				g.queueTriggerPromptLocked(ev, *card, lki, t)
				continue
			}
			if t.Build == nil {
				continue
			}
			item := t.Build(ev, card, lki, g)
			if item == nil {
				continue
			}
			g.queueHarvestedTriggerLocked(item)
		}
	}
}

// harvestLTB walks LTB triggers for a card whose battlefield exit
// just emitted EventLTB. The card now lives in its destination zone;
// look it up by ID and pull its battlefield characteristics from the
// LKI map. After the harvest, clear the LKI entry — keeping it would
// leak across future events and re-fire stale triggers.
func (g *Game) harvestLTB(ev Event) {
	defer delete(g.lastKnownBattlefield, ev.CardID)
	card := g.findCardByIDLocked(ev.CardID)
	if card == nil {
		return
	}
	oracle := card.OracleID
	if oracle == "" {
		return
	}
	triggers := CatalogTriggers(oracle)
	if len(triggers) == 0 {
		return
	}
	lki, ok := g.lastKnownBattlefield[ev.CardID]
	if !ok {
		// Missing LKI is a programming error — every battlefield exit
		// is supposed to stamp the snapshot first. Fall back to the
		// post-move characteristics so the trigger still fires;
		// log-shaped EventEffectError surfaces the gap during dev.
		lki = card.Effective()
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   ev.CardID,
			ErrorMsg: "S19: LTB trigger fired without LKI snapshot",
		})
	}
	for _, t := range triggers {
		if !triggerWatches(t.Watches, ev.Kind) {
			continue
		}
		if t.AppliesTo != nil && !t.AppliesTo(ev, card, lki, g) {
			continue
		}
		if t.OptionalPrompt != nil {
			g.queueTriggerPromptLocked(ev, *card, lki, t)
			continue
		}
		if t.Build == nil {
			continue
		}
		item := t.Build(ev, card, lki, g)
		if item == nil {
			continue
		}
		g.queueHarvestedTriggerLocked(item)
	}
}

// queueHarvestedTriggerLocked appends a freshly-built StackItem to
// PendingTriggers and emits the EventTrigger breadcrumb. Mirrors the
// tail of AnnounceTrigger but does not take the lock (caller already
// holds g.mu) and does not validate playerID / sourceCardID — Build
// is trusted to populate those.
//
// Stamps a fresh ID if Build left one as uuid.Nil — Build callbacks
// typically don't bother minting an ID.
func (g *Game) queueHarvestedTriggerLocked(item *StackItem) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.Kind == "" {
		item.Kind = StackItemTriggered
	}
	g.PendingTriggers = append(g.PendingTriggers, item)
	g.EmitEvent(Event{
		Kind:   EventTrigger,
		Actor:  item.Controller,
		Source: item.SourceCardID,
		Label:  item.Label,
	})
}

// snapshotLKILocked records the source card's effective
// characteristics into lastKnownBattlefield just before a
// battlefield-exit move. Called from each LTB-emitting site in
// effect_api.go before the MoveCard call. No-op if cardID isn't on
// the battlefield (the source might already have left for an
// unrelated reason).
//
// Caller must hold g.mu in write mode. The map is initialised lazily.
func (g *Game) snapshotLKILocked(cardID uuid.UUID) {
	if cardID == uuid.Nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != cardID {
			continue
		}
		if g.lastKnownBattlefield == nil {
			g.lastKnownBattlefield = make(map[uuid.UUID]Characteristic)
		}
		g.lastKnownBattlefield[cardID] = c.Effective()
		return
	}
}

// findCardByIDLocked scans every zone for a card with the given
// instance ID and returns a pointer into the slice (or nil if not
// found). Used by harvestLTB to locate a card whose zone is unknown
// to the harvester — the LTB destination depends on the cause
// (graveyard for destroy, exile for Path, hand for Unsummon).
//
// Caller must hold g.mu. Returns a pointer into the live slice;
// callers MUST NOT mutate the card through it (the harvester only
// reads).
func (g *Game) findCardByIDLocked(cardID uuid.UUID) *Card {
	scan := func(z *Zone) *Card {
		if z == nil {
			return nil
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == cardID {
				return &z.Cards[i]
			}
		}
		return nil
	}
	if c := scan(g.Battlefield); c != nil {
		return c
	}
	if c := scan(g.Stack); c != nil {
		return c
	}
	if c := scan(g.Exile); c != nil {
		return c
	}
	for _, p := range g.Seats {
		if c := scan(p.Hand); c != nil {
			return c
		}
		if c := scan(p.Library); c != nil {
			return c
		}
		if c := scan(p.Graveyard); c != nil {
			return c
		}
		if c := scan(p.Command); c != nil {
			return c
		}
	}
	return nil
}

// triggerWatches reports whether kinds contains kind. Linear scan;
// the typical TriggeredAbility watches one or two kinds, so a
// bitmask isn't worth the API friction.
func triggerWatches(kinds []EventKind, kind EventKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}
