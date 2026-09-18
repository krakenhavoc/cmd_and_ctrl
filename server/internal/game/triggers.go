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
//
// Stack timing: a harvested item goes PendingTriggers → StackMeta
// (APNAP drain at the next priority boundary) → resolveTopAbilityLocked
// when every player passes in succession, which runs the item's
// Effect callback. Sub-PRs 3-5 originally applied effects inline
// from Build and returned nil (no stack item, no response window);
// that shortcut was retired — Build now returns a real item via
// NewTriggeredItem and the effect waits for resolution.

// NewTriggeredItem builds the StackItem for a triggered ability
// whose source is `source`: Kind StackItemTriggered, controller and
// owner = source.Controller, Label for the stack overlay, and an
// Effect callback that runs at resolution. Targets / Modes / X are
// left empty — a targeted trigger sets item.Targets on the returned
// value before handing it back from Build so the CR 608.2b re-check
// applies at resolve time.
//
// The ID is left Nil; queueHarvestedTriggerLocked mints one. The
// effect must read the controller / source / targets off the item
// it receives (not off `source`, which is a pointer into a zone
// slice that may have been reallocated or moved by resolve time).
func NewTriggeredItem(source *Card, label string, effect func(g *Game, item *StackItem) error) *StackItem {
	return &StackItem{
		Kind:         StackItemTriggered,
		Controller:   source.Controller,
		Owner:        source.Controller,
		SourceCardID: source.InstanceID,
		Label:        label,
		Effect:       effect,
	}
}

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
	// The item carries the trigger's announce-time choices (targets
	// picked now, per CR 603.3d) and an Effect callback that runs
	// when the item resolves off the stack — see NewTriggeredItem
	// for the one-liner most cards want. Build itself must NOT
	// mutate game state: the ability hasn't resolved yet, and
	// applying the effect here would deny every player their
	// response window (Stifle, counter-the-trigger, sacrifice in
	// response). Returning nil suppresses the trigger — e.g. a
	// targeted trigger with no legal target (CR 603.3d: it's removed
	// from the stack / never put there).
	//
	// Runs under g.mu held in write mode. MUST NOT call public
	// locking mutators.
	Build func(ev Event, source *Card, sourceLKI Characteristic, g *Game) *StackItem

	// OptionalPrompt is the declarative "you may" gate for CR 603.5
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

	// Targets is the S20 structured target clause for a targeted
	// trigger ("destroy target artifact or enchantment"). When set,
	// the harvester computes the legal set at trigger time (CR
	// 603.3d — targets are chosen as the ability is put on the
	// stack): an empty set drops the trigger without any prompt; a
	// non-empty one queues a pick_target prompt for the controller
	// (after the OptionalPrompt's "yes", if there is one). The chosen
	// TargetRef is stamped onto the built item's Targets, the item
	// remembers the spec, and resolution re-checks it (CR 608.2b).
	// Build / Effect read item.Targets[0]. Added in S20 sub-PR 2.
	Targets *TargetSpec

	// HasLegalTarget reports whether the trigger has a legal
	// target / will actually do something at fire time. Optional —
	// nil means "assume the effect always has an effect" (e.g.
	// Mulldrifter's unconditional draw). When set, the harvester
	// evaluates it at OptionalPrompt-queue time and stamps the
	// result onto the PendingChoice (NoLegalTarget). The client uses
	// this to warn the chooser that answering "Yes" will pass without
	// effect — the S19 sandbox auto-targeter (pickFirstOpponent*)
	// only ever picks opponent-controlled permanents and there is no
	// target-selection UI until S20, so a "Yes" with no legal target
	// silently no-ops, which reads as a bug. Surfacing the warning
	// removes that confusion until the S20 picker lands.
	//
	// Runs under g.mu held in write mode. MUST NOT call public
	// locking mutators. Added in S19 follow-up.
	HasLegalTarget func(ev Event, source *Card, sourceLKI Characteristic, g *Game) bool

	// FromStack marks a trigger that fires while its source is a
	// SPELL ON THE STACK rather than a permanent on the battlefield
	// — "When you cast this spell, ..." (cascade, CR 702.85a; storm;
	// the ripple / replicate family). Nothing else in the catalog
	// triggers from there.
	//
	// The harvester scans the battlefield, because that is where
	// abilities live (CR 113.6). A trigger like cascade's is not on
	// a permanent: it is on an object that exists only between
	// announce and resolution, and by the time it would reach the
	// battlefield the event has long passed.
	//
	// An opt-in FLAG rather than a blanket stack scan, and the
	// distinction matters. A creature spell sitting on the stack
	// carries its permanent's declared triggers with it — "whenever
	// another creature enters, ..." — and a blanket scan would fire
	// those from the stack, a turn early and from a zone where the
	// ability does not exist. The flagged scan is narrow in both
	// directions: only EventCast, and only the card that event names.
	//
	// Added in S28.
	FromStack bool

	// Key names this ability for the OncePerBatch check: the stack
	// label it announces with. The effects constructors stamp it from
	// the label they are given; a hand-written ability may set it, or
	// leave it empty to mean "any trigger from this source". Not
	// persisted — it is catalog data.
	Key string

	// OncePerBatch is "whenever ONE OR MORE …": the engine emits one
	// event per creature that attacks, enters or deals damage, and a
	// printed once-per-batch ability must not fire once per event.
	// With this set the harvester fires the ability (matched by Key,
	// or by source when Key is empty) for the FIRST matching event of
	// an event batch and declines every later event of that SAME
	// batch — and fires again for the next batch, whatever is still
	// on the stack from the last one (CR 603.2c). Before #587 every
	// card that needed it scanned by hand, four slightly different
	// ways.
	//
	// A batch is every event emitted between two points where play
	// moves on: a stack item beginning to resolve, and the turn cursor
	// entering a new step. One resolution is one batch, one turn-based
	// action is one batch. #594 shipped this flag with an "is an item
	// of this ability in flight" check standing in for the batch;
	// #829 replaced that with the real batch identity on Event.Batch,
	// because an item on the stack outlives the batch that put it
	// there and was swallowing the next one. See event_batch.go.
	OncePerBatch bool

	// BatchKey is the SECOND dimension of the OncePerBatch key: the
	// distinct object the printed clause quantifies over, read off
	// the event (#784).
	//
	// CR 603.2c fires one trigger per occurrence, and a clause that
	// names an object counts one occurrence per object — "whenever
	// one or more creatures you control deal combat damage to A
	// PLAYER" is once per player dealt damage, not once per damage
	// step (Keeper of Fables, ruling 2019-10-04), and "whenever you
	// attack A PLAYER" is once per player attacked (Horizon
	// Explorer, Neyali). Key alone cannot say that: it is static
	// catalog data and the player is only known from the event.
	//
	// With this set the guard becomes "(source, key, this player)
	// once per batch": three creatures hitting three opponents in
	// one damage step are three triggers, and two creatures hitting
	// one opponent are one. Nil — the usual case — leaves the guard
	// keyed on (source, key) alone.
	//
	// Runs under g.mu in write mode, from the harvest path. MUST NOT
	// call public locking mutators. Not persisted; catalog data, like
	// Key. See effects.OncePerBatchPerPlayer for the one reading the
	// catalog uses.
	BatchKey func(ev Event, source *Card, g *Game) string

	// Chapter is the Saga chapter number this ability is printed
	// against — 1 for "I —", 3 for "III —" (CR 714.2b). Zero for
	// every ability that is not a chapter, which is every ability on
	// every card that is not a Saga.
	//
	// It is declarative data, not a second trigger condition: the
	// ability still watches EventSagaChapter and still predicates on
	// the chapter number, exactly as effects.ChapterTrigger writes
	// it. What this field adds is a way for the ENGINE to read the
	// card's FINAL chapter off the declarations, which is what the
	// CR 704.5s sacrifice needs and what nothing else on the card can
	// tell it. See game.SagaFinalChapter.
	//
	// Added in S27.
	Chapter int
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
	pass := g.newHarvestPassLocked(ev)
	g.harvestFromZone(&pass, g.Battlefield)
	// #623 / CR 114.3: an emblem's triggered abilities function in the
	// command zone. One more zone into the same walk — see emblem.go.
	g.harvestFromEmblemsLocked(&pass)
	// S28: "When you cast this spell, ..." — cascade. The source is
	// the spell that was just announced, which is on the stack and
	// invisible to the battlefield scan above. Narrow on purpose:
	// only EventCast, only the one card the event names, and only
	// abilities that declared FromStack.
	if ev.Kind == EventCast && ev.CardID != uuid.Nil {
		g.harvestCastFromStack(&pass)
	}
	// LTB-style events: the source has already moved off the
	// battlefield, so harvestFromZone(BF) won't find it. The card
	// lives in its destination zone now (graveyard / exile / hand);
	// the harvester finds it there and pulls battlefield
	// characteristics from the LKI snapshot.
	if ev.Kind == EventLTB && ev.CardID != uuid.Nil {
		g.harvestLTB(&pass)
	}
	// S23: watchers that left the battlefield EARLIER IN THIS SAME
	// event — the Blood Artist that was wiped alongside the creatures
	// it is supposed to see die. harvestFromZone cannot find them
	// (they are off the battlefield) and harvestLTB is about the one
	// card whose death is being reported, so a wipe needs its own
	// pass. No-op unless a simultaneous batch is open.
	g.harvestSimultaneousExitLocked(&pass)
}

// harvestFromZone is the per-zone scan used for "live" triggers (ETB,
// cast, draw, combat damage, upkeep). For every battlefield card with
// a TriggeredAbility whose Watches contains ev.Kind, run AppliesTo
// then Build and queue the result.
func (g *Game) harvestFromZone(pass *harvestPass, z *Zone) {
	ev := pass.ev
	if z == nil || CatalogTriggers == nil {
		return
	}
	for i := range z.Cards {
		card := &z.Cards[i]
		// During a simultaneous exit the live card may already have lost a
		// continuous effect because another batch member moved. Triggers see
		// the shared pre-exit snapshot instead (the card can still be in the
		// battlefield while this particular event is harvested).
		source := card
		if isBattlefieldExitEvent(ev) {
			if batch, ok := g.simultaneousExitCardLocked(card.InstanceID); ok {
				source = &batch
			}
		}
		// CatalogAbilityKey: a permanent under a CR 613.1f
		// ability-removing effect has no triggered abilities to
		// harvest. Off the battlefield this is CatalogKey exactly.
		oracle := CatalogAbilityKey(*source)
		if oracle == "" {
			continue
		}
		triggers := CatalogTriggers(oracle)
		if len(triggers) == 0 {
			continue
		}
		lki := source.Effective()
		for _, t := range triggers {
			if !triggerWatches(t.Watches, ev.Kind) {
				continue
			}
			if t.AppliesTo != nil && !t.AppliesTo(ev, source, lki, g) {
				continue
			}
			g.harvestMatchLocked(pass, *source, lki, t, false)
		}
	}
}

// harvestCastFromStack fires the FromStack triggers of the spell
// named by an EventCast — cascade's "when you cast this spell".
//
// Finds the one card the event names in the stack zone rather than
// walking it, because "this spell" means exactly that object: a
// second cascading spell already on the stack under this one does
// not re-trigger when something is cast above it.
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestCastFromStack(pass *harvestPass) {
	ev := pass.ev
	if g.Stack == nil || CatalogTriggers == nil {
		return
	}
	for i := range g.Stack.Cards {
		card := &g.Stack.Cards[i]
		if card.InstanceID != ev.CardID {
			continue
		}
		oracle := CatalogAbilityKey(*card)
		if oracle == "" {
			return
		}
		lki := card.Effective()
		for _, t := range CatalogTriggers(oracle) {
			if !t.FromStack || !triggerWatches(t.Watches, ev.Kind) {
				continue
			}
			if t.AppliesTo != nil && !t.AppliesTo(ev, card, lki, g) {
				continue
			}
			g.harvestMatchLocked(pass, *card, lki, t, true)
		}
		return
	}
}

// dispatchTriggerLocked takes a matched TriggeredAbility from
// "AppliesTo said yes" to either a queued prompt or an item on
// PendingTriggers:
//
//  1. Targeted (Targets != nil) with no legal target right now →
//     the trigger is removed without any prompt (CR 603.3d).
//  2. OptionalPrompt → the yes/no prompt; on "yes" the flow re-enters
//     here at step 3 via ResolveTriggerPrompt.
//  3. Targeted → pick_target prompt; on the pick, Build runs and the
//     chosen ref is stamped onto the item.
//  4. Otherwise Build → queue.
//
// Caller must hold g.mu. Added in S20 sub-PR 2 (steps 1 and 3).
func (g *Game) dispatchTriggerLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility) {
	if t.OncePerBatch && !g.oncePerBatchAllowsLocked(ev.Batch, source.InstanceID, oncePerBatchKeyLocked(t, ev, &source, g)) {
		return
	}
	g.dispatchTriggerInstanceLocked(ev, source, lki, t, doublerRef{})
}

// harvestMatchLocked fixes the additional-instance count at the moment this
// ability triggered, before prompts or items are queued.
func (g *Game) harvestMatchLocked(pass *harvestPass, source Card, lki Characteristic, t TriggeredAbility, fromSpell bool) {
	if t.OncePerBatch && !g.oncePerBatchAllowsLocked(pass.ev.Batch, source.InstanceID, oncePerBatchKeyLocked(t, pass.ev, &source, g)) {
		return
	}
	extra := g.triggerDoublersLocked(pass, source, lki, t, fromSpell)
	g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, doublerRef{})
	for _, d := range extra {
		g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, d)
	}
}

func (g *Game) dispatchTriggerInstanceLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef) {
	if t.Targets != nil {
		lt := g.legalTargetsLocked(source.Controller, t.Targets)
		if len(lt.Players) == 0 && len(lt.Cards) == 0 {
			return
		}
	}
	if t.OptionalPrompt != nil {
		g.queueTriggerPromptLocked(ev, source, lki, t, doubledBy)
		return
	}
	g.buildOrPickTriggerLocked(ev, source, lki, t, doubledBy)
}

// buildOrPickTriggerLocked is the post-"yes" half of the dispatch:
// queue the target picker for a targeted trigger, or Build and queue
// the item directly. Caller must hold g.mu.
func (g *Game) buildOrPickTriggerLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef) {
	if t.Build == nil {
		return
	}
	if t.Targets != nil {
		g.queuePickTargetLocked(ev, source, lki, t, doubledBy)
		return
	}
	item := t.Build(ev, &source, lki, g)
	if item == nil {
		return
	}
	item.DoubledBy, item.DoubledByName = doubledBy.id, doubledBy.name
	g.queueHarvestedTriggerLocked(item)
}

// harvestLTB walks LTB triggers for a card whose battlefield exit
// just emitted EventLTB. The card now lives in its destination zone;
// look it up by ID and pull its battlefield characteristics from the
// LKI map. After the harvest, clear the LKI entry — keeping it would
// leak across future events and re-fire stale triggers.
func (g *Game) harvestLTB(pass *harvestPass) {
	ev := pass.ev
	defer delete(g.lastKnownBattlefield, ev.CardID)
	defer delete(g.lastKnownTriggerIdentity, ev.CardID)
	card := g.findCardByIDLocked(ev.CardID)
	if card == nil {
		return
	}
	source := g.withLastKnownTriggerIdentityLocked(*card)
	if batch, ok := g.simultaneousExitCardLocked(ev.CardID); ok {
		source = batch
	}
	oracle := CatalogKey(source)
	if oracle == "" {
		return
	}
	triggers := CatalogTriggers(oracle)
	if len(triggers) == 0 {
		return
	}
	lki, ok := g.lastKnownBattlefield[ev.CardID]
	if batch, inBatch := g.simultaneousExitCardLocked(ev.CardID); inBatch {
		lki, ok = batch.Effective(), true
	}
	// S24 layer 6: read the removal off the LKI SNAPSHOT, not off the
	// card. CatalogAbilityKey cannot answer here — the permanent has
	// already left the battlefield and clearEffectiveCacheLocked has
	// dropped its layer cache — and CR 603.10 says an LTB trigger is
	// judged on what the permanent looked like while it was still
	// there. A creature that died under a Kenrith's Transformation
	// has no dies-trigger, and that stays true for the beat between
	// the death and the Aura falling off.
	if ok && lki.AbilitiesRemoved {
		return
	}
	if !ok {
		// Missing LKI is a programming error — every battlefield exit
		// is supposed to stamp the snapshot first. Fall back to the
		// post-move characteristics so the trigger still fires;
		// log-shaped EventEffectError surfaces the gap during dev.
		lki = source.Effective()
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
		if t.AppliesTo != nil && !t.AppliesTo(ev, &source, lki, g) {
			continue
		}
		g.harvestMatchLocked(pass, source, lki, t, false)
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
		if g.lastKnownTriggerIdentity == nil {
			g.lastKnownTriggerIdentity = make(map[uuid.UUID]triggerIdentityLKI)
		}
		g.lastKnownTriggerIdentity[cardID] = triggerIdentityLKI{OracleID: c.OracleID, ActiveFace: c.ActiveFace, AttachedTo: c.AttachedTo}
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
