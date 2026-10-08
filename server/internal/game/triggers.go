package game

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

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
// owner = source.Controller, and Label for the stack overlay.
// Targets / Modes / X are left empty — a targeted trigger sets
// item.Targets on the returned value before handing it back from
// Build so the CR 608.2b re-check applies at resolve time.
//
// It takes NO effect (ADR 0041 P9, #1497, tier 4-final): what the
// item does is always data. A catalog row declares its Effect and
// the engine installs it and names the row on the item
// (buildTriggerItemLocked); an engine trigger with no row names a
// registered body (NewKeyedTriggeredItem). A catalog Build that calls
// this is a fill-in: it sets a label, a controller or Params and
// leaves item.Effect nil.
//
// The ID is left Nil; queueHarvestedTriggerLocked mints one.
func NewTriggeredItem(source *Card, label string) *StackItem {
	return &StackItem{
		Kind:         StackItemTriggered,
		Controller:   source.Controller,
		Owner:        source.Controller,
		SourceCardID: source.InstanceID,
		Label:        label,
	}
}

// NewKeyedTriggeredItem is NewTriggeredItem for an ENGINE trigger with
// no catalog row (ADR 0041 P9, #1497, tier 4): prowess, suspend,
// madness, the monarch, evoke's sacrifice, face-down ward. Effect is
// derived from the registered body rather than captured directly, and
// Body/Params are stamped alongside it, so a table with one of these on
// the stack is still a restore point — the same shape
// dt.stackItem() stamps for a fired delayed trigger.
func NewKeyedTriggeredItem(source *Card, label string, body BodyRef, params EffectParams) *StackItem {
	item := NewTriggeredItem(source, label)
	item.Effect = bodyEffect(body.key, params)
	item.Body, item.Params = body.key, params
	return item
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
	// Watches is a programming error and the harvester skips it —
	// unless State is set, which makes the row a state trigger that
	// watches no event at all.
	Watches []EventKind

	// State makes this row a CR 603.8 STATE TRIGGER: "When you control
	// no Islands, sacrifice this creature", "When there are five or
	// more plot counters on this enchantment, …". It is a condition
	// over the board, the source permanent and its controller, and the
	// ability triggers as soon as the game state matches it — not when
	// an event happens (CR 603.2's "or game state").
	//
	// The engine asks it after every event it emits and in each pass
	// of the CR 704.3 loop (ADR 0107 §1, owner decision 1), so a state
	// that holds only for a moment inside a resolution still triggers,
	// as CR 603.8's own example says. It asks only battlefield
	// permanents, through TriggersForCard, so a permanent that has lost
	// its abilities (CR 613.1f) has no state triggers.
	//
	// The CR 603.8 latch — "doesn't trigger again until the ability has
	// resolved, has been countered, or has otherwise left the stack" —
	// is derived, never stored: the ability is latched while the stack,
	// the trigger queue, the resolving slot or an open announcement
	// prompt holds an item of this row from this OBJECT (CR 400.7). See
	// state_triggers.go.
	//
	// A state trigger watches nothing (Register refuses Watches beside
	// State) and must declare its Effect, so its stack item is keyed
	// (ADR 0041 P9) and a table with one waiting is a restore point.
	//
	// The field is a KEY, not the condition: the condition is a
	// StateCondition registered under it (RegisterStateCondition), in a
	// package-level registry no Game reaches. So the declaration a
	// trigger's resume frame carries holds no new func, and the ADR 0041
	// closure ratchet gains no route (AGENTS.md §5). The key lives only
	// on catalog rows, which the running binary rebuilds on restore; it
	// is never written to a snapshot, because a queued or stacked state
	// trigger names its ROW (AbilityRef), not its condition. Empty means
	// "not a state trigger".
	State string

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
	//
	// ADR 0041 P9 (#1497, tier 4-2): a declaration that sets Effect
	// below needs no Build at all — the engine builds the item. A Build
	// kept beside an Effect only FILLS IN data the engine cannot know
	// (a controller override, a computed label, Params) and must leave
	// item.Effect nil; the engine then installs Effect itself. A Build
	// that sets item.Effect while Effect is declared is an
	// effectKeyFault. A Build with no Effect beside it is the legacy
	// shape the tail of tier 4 retires (testdata/legacy_trigger_builds.txt).
	Build func(ev Event, source *Card, sourceLKI Characteristic, g *Game) *StackItem

	// Effect is what the ability does when its item resolves, declared
	// on the row itself (ADR 0041 P9, #1497, tier 4-2). When it is set
	// the engine builds the item — NewTriggeredItem(source, Key) with
	// Effect installed, or Build's item with Effect installed — and, for a row
	// the catalog registered, stamps it with the row's name
	// (Body "catalog/triggered" plus Params.Ability), so a table with
	// the trigger waiting on the stack is still a restore point:
	// restore looks the row up again and takes its Effect, its target
	// clause and its mode clause from the running binary.
	//
	// The contract StackItem.Effect has always had, and now it is load
	// bearing: read everything off the item (its Trigger, Targets,
	// Params, Payload) and the game handed in, and capture nothing
	// that was only true when the ability triggered. A value that was
	// is either on item.Trigger already (#1223) or goes into Params
	// from a fill-in Build.
	Effect func(g *Game, item *StackItem) error

	// TargetsFromReadsBoard declares that TargetsFrom reads more than
	// its trigger context and the source's identity — the board, or the
	// source's CURRENT controller (ADR 0041 P9's TargetsFrom audit,
	// tier 4-2). Restore re-derives a stamped trigger's clause by
	// calling TargetsFrom again on the RESTORED board, which is exact
	// only for a clause that reads nothing the board can change, so a
	// row that sets this is never stamped: its item stays a census
	// closure, and the row is on the legacy list. The seat list is not
	// the board (a seat is permanent once the game starts), and the
	// controller as the ability triggered is on the trigger context
	// (tc.Object), which is how Molten Primordial's one clause per
	// opponent stopped needing the flag. No catalog row sets it today.
	TargetsFromReadsBoard bool

	// row is the catalog identity the registry stamped on this row
	// (IdentifyCatalogRows): the key whose Triggered list holds it and
	// its index there. Zero for a row the catalog did not register — an
	// engine trigger, a test stub, a reflexive trigger — which is never
	// stamped. See ability_ref.go.
	row catalogRowID

	// OptionalPrompt is the declarative "you may" gate for CR 603.5
	// optional triggers. When non-nil, the harvester does NOT call
	// Build immediately on a matching event — it queues a
	// PendingChoiceTriggerPrompt to the source's controller (or the
	// override-chooser specified in OptionalPrompt). On
	// resolve_choice with `apply: true`, Build runs against the
	// captured event + LKI and queues the StackItem. On `apply:
	// false`, the trigger drops without effect.
	//
	// Nil means mandatory — Build runs unconditionally. Added in S19
	// sub-PR 2.
	OptionalPrompt *TriggerOptionalPrompt

	// Modes is the CR 700.2 mode clause of a modal triggered ability
	// ("choose one — put a +1/+1 counter on this creature; create a
	// Treasure token; you gain 2 life"). The same game.ModeSpec a
	// modal spell and a modal activated ability declare — one struct,
	// three owners (#764, ADR 0065 §3).
	//
	// The choice is made as the ability is put on the stack (CR
	// 603.3c): after the OptionalPrompt's "yes", before the CR
	// 603.3d target pick, through a mode_pick prompt to the source's
	// controller. An option whose target clause has no legal target
	// is not offered, and if that leaves fewer than Min the trigger
	// is removed without any prompt at all.
	//
	// The chosen occurrences land on the built item's Modes and each
	// bullet's ModeOption.Effect runs at resolution in announce
	// order; a card that would rather branch by hand reads
	// effects.Context.HasMode in the Build closure's effect instead.
	// Added by #764.
	Modes *ModeSpec

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

	// TargetsFrom is `Targets` for a clause that is a fact about WHAT
	// HAPPENED: "return to your hand target artifact card in your
	// graveyard with lesser mana value" (Scrap Trawler), where
	// "lesser" is lesser than the mana value of the artifact that was
	// just put into a graveyard.
	//
	// Non-nil wins over `Targets`, which such a card leaves nil —
	// declaring both would be two answers to one question. Called
	// with the trigger's own CR 603.10 context (trigger_event.go) and
	// a value copy of the source, at every point the dispatch needs
	// the clause: the CR 603.3d "can this be filled at all" check,
	// the target walk, and the spec stamped onto the item for the
	// CR 608.2b re-check. Returning nil is "this trigger targets
	// nothing after all", which is a legal answer and drops the
	// clause rather than the trigger.
	//
	// A FUNCTION rather than a predicate that receives the context,
	// because the event changes the clause's COUNT and LABEL as
	// readily as its predicate — "up to X target creatures", where X
	// is the damage that was dealt — and a spec is the only thing
	// that can say all three.
	//
	// Runs under g.mu in write mode. MUST NOT call public locking
	// mutators, and must be a pure read of its arguments plus the
	// board: it is called more than once for one trigger (once per
	// prompt step) and every call must agree. Not persisted — catalog
	// data, like Targets. Added in #1223.
	TargetsFrom func(tc TriggerContext, source *Card, g *Game) *TargetSpec

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

	// Keyword is the machine-readable name of the KEYWORD ability this
	// trigger is — "cascade" (CR 702.85), "storm" (CR 702.40),
	// "prowess" (CR 702.108) — and empty for every hand-written
	// trigger, which is almost all of them (#1258).
	//
	// Stamped by the keyword's constructor (effects.Cascade,
	// effects.Storm) or by the engine's own keyword-trigger table
	// (prowess.go), never by a card file, so it cannot disagree with
	// what the ability does. It exists to be READ: the caveat drift
	// guard (cards/coverage) asks "does this card have cascade" by
	// walking the card's triggers for the name, the way it already
	// asks about flashback through an alternative cost's key.
	//
	// Catalog data, not persisted, and not a behaviour switch —
	// nothing in the harvest path branches on it.
	Keyword string

	// Zones is WHERE this ability watches from (CR 113.6, #925). Nil
	// — the answer for all but a handful of cards — means the
	// battlefield, which is where abilities live. {ZoneGraveyard} is
	// "when you cycle this card" (CR 702.29c) and Bloodghast's
	// landfall; {ZoneExile} is suspend's upkeep countdown (CR
	// 702.62b).
	//
	// A declared zone list IS the list: an ability that names the
	// graveyard does not also fire from the battlefield. The same
	// default and the same rule as ActivatedAbilityShape.Zones (ADR
	// 0062 Decision 1) and CastableZones, read through TriggerZones /
	// TriggerWatchesFromZone in trigger_zones.go, which is also where
	// the per-zone index and the one extra harvest live.
	Zones []ZoneKind

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

	// PerCounter is "whenever a [kind] counter is put on ~": the
	// counter kind whose placement this ability triggers on ONCE PER
	// COUNTER (#1841, CR 603.2c and the Fathom Mage rulings). The
	// engine emits one EventCounterPlaced per placement, and a
	// replacement (Doubling Season, Hardened Scales) has already
	// settled the count by then, so the number of counters the event
	// put is the number of occurrences: two counters are two triggers,
	// each its own stack object and each its own "you may". A removal
	// or another kind is zero occurrences. Counters a permanent enters
	// with are placed after it arrives, so they count too (CR 122.6).
	// Empty — every other ability — is one occurrence per event. Plain
	// data, not a func. See counterPlacedDeltaLocked.
	PerCounter string

	// PerCounterRemoved is the removal twin of PerCounter (#2466):
	// "whenever a [kind] counter is removed from ~" triggers ONCE PER
	// COUNTER removed (CR 603.2c, Protean Hydra's ruling). The same
	// EventCounterPlaced carries a removal as a lower post-change
	// total, so the number of counters it took off is the number of
	// occurrences. Removal by any cause counts — a cost, an effect, the
	// CR 704.5q cancel, prevented damage — and a placement or another
	// kind is zero occurrences. Counters that go away because the
	// permanent LEFT the battlefield are not removed at all (CR 122.2,
	// the permanent no longer exists): the engine clears them with the
	// card and emits no event, so nothing fires. Plain data, not a
	// func. See counterRemovedDeltaLocked.
	PerCounterRemoved string

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

	// AtBatchEnd defers the ability until the event batch it triggered
	// in has finished (#2183, life_batch.go). The ability still
	// watches events and AppliesTo still runs per event, but only as a
	// loose pre-filter ("an opponent lost life"): a match is staged,
	// and when play moves on (the next priority grant, or the next
	// batch opening) the condition is asked ONCE with the batch's
	// final totals and the ability dispatches if it holds. That is
	// what "exactly 1 life" needs: the total is not known until every
	// simultaneous loss has landed.
	//
	// Plain data, not a func, so an ability carried in a prompt's
	// resume frame adds nothing to the closure census. Needs a
	// non-empty Key (the staged match is found again by it), and fires
	// from a permanent only. Dispatch then runs the same suppression,
	// OncePerBatch and doubling steps as any trigger, so set
	// OncePerBatch as well for a "one or more" clause. Not persisted;
	// catalog data.
	AtBatchEnd *BatchEndCondition

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

	// ActiveWhen is the CR 716 / 719 / 721 / 709.5 designation gate:
	// this trigger exists only while its source permanent has the
	// designation named. A Case's "Solved — whenever …" is
	// CaseSolved(); a Class's level-3 trigger is ClassLevel(3). The
	// zero value is "no gate".
	//
	// Evaluated in TriggersForCard / TriggersForKey and nowhere else,
	// so a gated-off trigger is never matched, never prompted and
	// never queued. Distinct from Chapter, which is declarative data
	// about a Saga rather than a condition. See designations.go and
	// ADR 0071.
	ActiveWhen Designation

	// Purpose is what the row does, as printed amounts (ADR 0126 §6):
	// DeathPayoff on Blood Artist's "whenever a creature dies", a
	// Sweep on a saga chapter that destroys all creatures. Catalog data
	// projected onto the row's `ability_rows` entry; the engine never
	// reads it.
	Purpose Purpose

	// Exert marks a row as one of exert's (ADR 0130's amendment of
	// 2026-10-07): ExertRowLinked on the "when you do" trigger linked
	// to "You may exert this creature as it attacks" (CR 607.2h), and
	// ExertRowPayoff on "Whenever you exert a creature". Set by
	// effects.WhenExerted and effects.WheneverYouExert and nowhere
	// else; projected onto the row's `ability_rows` entry for the bot,
	// which prices an exert by these rows' purposes. The engine never
	// reads it.
	Exert ExertRow
}

// ExertRow is TriggeredAbility.Exert: which of exert's triggers a row
// is. The zero value is "not an exert row".
type ExertRow string

const (
	ExertRowLinked ExertRow = "linked"
	ExertRowPayoff ExertRow = "payoff"
)

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

	// Trade declares what answering "yes" DOES: it trades the source
	// permanent for the object the trigger is about — Perplexing
	// Chimera's "you may exchange control of this creature and that
	// spell" (ADR 0104). Nothing in the engine reads it. It rides the
	// prompt to the wire as `trade_for` (PendingChoice.TradeSubject), so
	// a bot can weigh the trade instead of saying yes to every trigger
	// it controls, which is the heuristic's default for a "you may".
	Trade bool
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
	// ADR 0093 PR 3, CR 603.6a: an enters-the-battlefield ability
	// "triggers when a permanent enters", looking at the permanent as it
	// exists on the battlefield — with the continuous effects that apply
	// to it. The one that matters here is a layer-6 GRANT: a creature
	// entering under Cryptolith Rite or Dionus already has the granted
	// abilities, including a granted "when this creature enters", and
	// its layer cache is still unset (nil — printed) at the moment the
	// entry announces itself. So the ETB harvest catches the layers up
	// first. Only for EventETB, and only when something is stale: the
	// entry has already landed every card it moves, so the board the
	// pass reads is the one the next reader would have recomputed
	// anyway. The recompute's own control-change events are emitted
	// after its store (recomputeLayersLocked), so a nested harvest finds
	// the cache clean.
	if ev.Kind == EventETB {
		g.RecomputeLayersIfStaleLocked()
	}
	pass := g.newHarvestPassLocked(ev)
	g.harvestDepth++
	defer func() { g.harvestDepth-- }()
	g.harvestFromZone(&pass, g.Battlefield)
	// #623 / CR 114.3: an emblem's triggered abilities function in the
	// command zone. One more zone into the same walk — see emblem.go.
	g.harvestFromEmblemsLocked(&pass)
	// #925: abilities that declared another zone (CR 113.6) — a
	// graveyard card's "when you cycle this card", suspend's exile
	// countdown. Indexed at Register, so an event kind nothing
	// watches from another zone costs one map lookup and no walk.
	// See trigger_zones.go.
	g.harvestFromDeclaredZones(&pass)
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
	// #663: event-conditioned delayed triggers — "when you next cast
	// an instant or sorcery spell this turn, copy that spell" (CR
	// 603.7b). Checked LAST, after every zone walk, so the delayed
	// list sees the event exactly once and the first match fires it
	// and removes it. One hook, one place, no per-card case. See
	// delayed.go and the 2026-09-18 amendment to ADR 0026.
	g.fireEventDelayedTriggersLocked(&pass)
	// ADR 0107 §1, CR 603.8 (owner decision 1): a state trigger
	// triggers as soon as the game state matches it, so the board this
	// event left behind is asked too — after every event trigger above
	// has been harvested. See state_triggers.go.
	g.stateTriggersAfterEventLocked()
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
	// The battlefield's abilities are a permanent's. The same walk over
	// an emblem zone (harvestFromEmblemsLocked) is not: CR 114.4.
	origin := triggerOfNonPermanent
	if z == g.Battlefield {
		origin = triggerOfPermanent
	}
	walked := len(z.Cards)
	for i := 0; i < walked; i++ {
		// The walk is read-only (#2608). If a trigger's AppliesTo or the
		// dispatch behind a match moved a permanent, the indices below no
		// longer line up with the cards the walk started with: stop here
		// rather than index past the slice, and say so loudly in tests.
		if len(z.Cards) != walked {
			harvestWalkMutated(z, walked, i)
			return
		}
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
		// TriggersForCard: a permanent under a CR 613.1f
		// ability-removing effect has no triggered abilities to
		// harvest, and neither has one whose designation gate is
		// unsatisfied — an unsolved Case's "Solved — whenever …" is
		// not a trigger that exists (ADR 0071). Off the battlefield
		// the key degrades to CatalogKey exactly.
		triggers := triggersOf(source)
		if len(triggers) == 0 {
			continue
		}
		lki := source.Effective()
		for _, t := range triggers {
			// #925: an ability that declared another zone does not
			// fire from the battlefield as well.
			if !TriggerWatchesFromZone(t, ZoneBattlefield) {
				continue
			}
			if !triggerWatches(t.Watches, ev.Kind) {
				continue
			}
			if t.AppliesTo != nil && !t.AppliesTo(ev, source, lki, g) {
				continue
			}
			g.harvestMatchLocked(pass, *source, lki, t, origin)
		}
	}
}

// harvestInvariantHook is the test hook for a harvest walk that changed
// its own zone (#2608). Production leaves it nil and logs; the package
// tests install a panic, as the card-index cross-check does.
var harvestInvariantHook func(msg string)

// SetHarvestInvariantHook installs the hook; nil restores log-only.
func SetHarvestInvariantHook(f func(msg string)) { harvestInvariantHook = f }

// harvestWalkMutated reports a zone that changed size while
// harvestFromZone was walking it. Trigger harvesting is read-only: an
// ability is queued, never carried out, so reaching this is an ordering
// bug in whatever ran under the walk, not something to absorb.
func harvestWalkMutated(z *Zone, walked, at int) {
	msg := fmt.Sprintf("harvestFromZone: zone changed from %d to %d cards during the walk (at index %d)", walked, len(z.Cards), at)
	if harvestInvariantHook != nil {
		harvestInvariantHook(msg)
		return
	}
	slog.Error(msg)
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
		lki := card.Effective()
		for _, t := range triggersOf(card) {
			if !t.FromStack || !triggerWatches(t.Watches, ev.Kind) {
				continue
			}
			if t.AppliesTo != nil && !t.AppliesTo(ev, card, lki, g) {
				continue
			}
			g.harvestMatchLocked(pass, *card, lki, t, triggerOfSpell)
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
//  3. Modal (Modes != nil) → mode_pick prompt (CR 603.3c); on the
//     answer the flow re-enters at step 4 via ResolveModePick.
//  4. Targeted → the pick_target walk, one prompt per clause of the
//     announcement; on the last pick, Build runs and the chosen refs
//     are stamped onto the item.
//  5. Otherwise Build → queue.
//
// Caller must hold g.mu. Added in S20 sub-PR 2; step 3 and the
// multi-clause walk in step 4 by #764.
func (g *Game) dispatchTriggerLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility) {
	if t.OncePerBatch && !g.oncePerBatchAllowsLocked(ev.Batch, source.InstanceID, oncePerBatchKeyLocked(t, ev, &source, g)) {
		return
	}
	g.dispatchTriggerInstanceLocked(ev, source, lki, t, doublerRef{})
}

// harvestMatchLocked fixes the additional-instance count at the moment this
// ability triggered, before prompts or items are queued.
//
// #1735: a suppressor is asked first (trigger_suppression.go). If the
// event does not cause this ability to trigger, there is nothing to
// count: the doublers are not asked, the once-per-batch slot is not
// spent, and no instance, prompt or target pick is made.
func (g *Game) harvestMatchLocked(pass *harvestPass, source Card, lki Characteristic, t TriggeredAbility, origin triggerOrigin) {
	// #2183: the batch is still growing, so the question "how much did
	// each player lose" is not answerable yet. Park the match; the
	// settle pass asks again with the final totals and comes back here
	// with pass.settled set.
	if t.AtBatchEnd != nil && !pass.settled && origin == triggerOfPermanent && t.Key != "" {
		g.stageBatchEndTriggerLocked(pass.ev, source.InstanceID, t.Key)
		return
	}
	if g.triggerSuppressedLocked(pass, source, lki, &t, origin) {
		return
	}
	if t.OncePerBatch && !g.oncePerBatchAllowsLocked(pass.ev.Batch, source.InstanceID, oncePerBatchKeyLocked(t, pass.ev, &source, g)) {
		return
	}
	occurrences := 1
	if t.PerCounter != "" {
		occurrences = g.counterPlacedDeltaLocked(pass.ev, t.PerCounter)
	}
	if t.PerCounterRemoved != "" {
		occurrences = g.counterRemovedDeltaLocked(pass.ev, t.PerCounterRemoved)
	}
	if occurrences < 1 {
		return
	}
	extra := g.triggerDoublersLocked(pass, source, lki, t, origin == triggerOfSpell)
	for n := 0; n < occurrences; n++ {
		g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, doublerRef{})
		for _, d := range extra {
			g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, d)
		}
	}
}

func (g *Game) dispatchTriggerInstanceLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef) {
	// #1223: the triggering event, read into a value HERE and
	// carried from here on. This is the last moment the CR 603.10
	// last-known-information snapshot of the object the event was
	// about is still in lastKnownBattlefield — harvestLTB deletes it
	// as its walk ends — so a context built any later would be
	// missing exactly the half a dies-trigger needs.
	tc := g.triggerContextLocked(ev)
	// CR 603.3d, checked before any prompt: an ability whose target
	// clause cannot be filled is never put on the stack, so nobody is
	// asked a question whose only answer is "nothing happens".
	// #764: every REQUIRED clause of the statement, not just the
	// first, and for a modal ability the question is instead whether
	// Min options remain choosable (choosableModeOptionsLocked).
	// #662: `source` is the value copy of the permanent whose ability
	// this is, which is the object CR 702.16b tests the quality
	// against — not its controller.
	src := SourceObject(source.Controller, &source)
	spec := g.triggerTargetsLocked(t, tc, &source)
	if t.Modes == nil && spec != nil &&
		g.anyClauseUnfillableLocked(src, AnnouncedClauses(spec, nil, nil)) {
		return
	}
	// ADR 0097: the ability's own "hasn't been chosen" memory is part
	// of the same question — an instance left with too few unused
	// modes is removed here, before any prompt (CR 700.2b).
	if t.Modes != nil &&
		!EnoughChoosableModes(len(g.choosableModeOptionsLocked(src, t.Modes, ModeAbilityOf(source, t.Key))), t.Modes) {
		return
	}
	if t.OptionalPrompt != nil {
		g.queueTriggerPromptLocked(tc, source, lki, t, doubledBy)
		return
	}
	g.buildOrPickTriggerLocked(tc, source, lki, t, doubledBy, nil)
}

// buildOrPickTriggerLocked is the post-"yes" half of the dispatch:
// the mode prompt for a modal trigger, then the target walk for a
// targeted one, then Build.
//
// `modes` is nil on the way in and carries the CR 603.3c answer on
// the way back through ResolveModePick — which is what makes the
// mode choice happen exactly once, before targets, and never for an
// ability that has none. Caller must hold g.mu.
func (g *Game) buildOrPickTriggerLocked(tc TriggerContext, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef, modes []int) {
	if !t.builds() {
		return
	}
	// CR 603.3c: modes first. An untargeted modal trigger (Black
	// Mark, Gala Greeters) takes this path too — the prompt is the
	// only thing between the harvest and the stack.
	if t.Modes != nil && modes == nil {
		if !g.queueModePickLocked(tc, source, lki, t, doubledBy) {
			// Fewer choosable options than Min: CR 603.3d removes the
			// ability rather than asking an unanswerable question.
			return
		}
		return
	}
	spec := g.triggerTargetsLocked(t, tc, &source)
	steps := AnnouncedClauses(spec, t.Modes, modes)
	if len(steps) > 0 {
		g.queuePickTargetLocked(tc, source, lki, t, doubledBy, modes, steps, spec)
		return
	}
	item := g.buildTriggerItemLocked(t, tc.Event, source, lki)
	if item == nil {
		return
	}
	item.Modes = append([]int(nil), modes...)
	item.modeSpec = t.Modes
	item.DoubledBy, item.DoubledByName = doubledBy.id, doubledBy.name
	stampTriggerContext(item, t, tc)
	stampTriggerSource(item, source, tc)
	g.queueHarvestedTriggerLocked(item)
}

// stampTriggerContext puts the triggering event on the item the
// harvester built (#1223).
//
// Here rather than inside Build, and that is the point of the field:
// Build is CATALOG code, ~2,200 declarations of it, and every one
// would have had to remember. The two sites that call Build call this
// immediately afterwards, so a card that declares a trigger gets the
// event on its item whether or not its author thought about it.
//
// An item that set its own context is left alone — a Build that
// constructs a trigger ABOUT a different event than the one that
// fired it has said something this cannot improve on.
//
// An ability that WATCHES NOTHING is left alone too, and that is the
// CR 603.12 reflexive trigger: QueueReflexiveTriggerForEffect hands
// the dispatch a synthetic EventResolve describing the resolution
// that created the trigger, precisely because there was no triggering
// event. Stamping that would tell a card its trigger fired off a
// resolution and hand it a snapshot of the parent's source, and both
// would be lies. Empty Watches is the fact rather than a flag: the
// harvester refuses to fire an ability that declares none, so
// anything that reaches here with none did not come from an event.
//
// The one context stamped with no event is a state trigger's about one
// of several objects (#1858, Bomb Squad's "Whenever a creature has four
// or more fuse counters on it"): its Object names the creature, which is
// both what the effect acts on and what the CR 603.8 latch reads
// (state_triggers.go).
func stampTriggerContext(item *StackItem, t TriggeredAbility, tc TriggerContext) {
	if item == nil || item.Trigger != nil {
		return
	}
	fromEvent := tc.Fired() && len(t.Watches) > 0
	aboutObject := t.State != "" && tc.Object != nil
	if !fromEvent && !aboutObject {
		return
	}
	item.Trigger = cloneTriggerContext(&tc)
}

// stampTriggerSource names the OBJECT a harvested trigger came from
// on the item Build returned (#1418, CR 400.7) — the sibling of
// stampTriggerContext, at the same two call sites and for the same
// reason: ~2,200 catalog Builds would each have had to remember.
//
// `source` is the dispatch's VALUE copy of the source card, taken
// when the ability triggered and carried through every prompt frame,
// so its epoch is the object's even if the card has moved while an
// optional or target prompt was open. The one case the copy cannot
// answer is a trigger about its source's own departure ("when this
// dies"): the harvest reads that card in its new zone, and the object
// the ability belongs to is the permanent that left. The event's
// object snapshot already names that permanent (objectSnapshotLocked
// reads its battlefield epoch off the record), so when the event is
// about the source, its ref is the answer.
//
// Left alone: an item that already names its source object (a
// reflexive trigger inherits its parent's, a delayed trigger carries
// the one it was scheduled with), and an item whose Build pointed it
// at a card other than `source` — queueHarvestedTriggerLocked stamps
// that one from the board.
func stampTriggerSource(item *StackItem, source Card, tc TriggerContext) {
	if item == nil || item.SourceObject.ID != uuid.Nil ||
		source.InstanceID == uuid.Nil || item.SourceCardID != source.InstanceID {
		return
	}
	if obj := tc.Object; obj != nil && obj.ID == source.InstanceID {
		item.SourceObject = obj.Ref()
		return
	}
	item.SourceObject = ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
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
	defer delete(g.lastKnownCounters, ev.CardID)
	card := g.findCardByIDLocked(ev.CardID)
	if card == nil {
		return
	}
	source := g.withLastKnownTriggerIdentityLocked(*card)
	if batch, ok := g.simultaneousExitCardLocked(ev.CardID); ok {
		source = batch
	}
	lki, ok := g.lastKnownBattlefield[ev.CardID]
	if batch, inBatch := g.simultaneousExitCardLocked(ev.CardID); inBatch {
		lki, ok = batch.Effective(), true
	}
	// S24 layer 6 / ADR 0093 Decision 2: the key is read off the LKI
	// SNAPSHOT, not off the card. CatalogAbilityKey cannot answer here
	// — the permanent has already left the battlefield and
	// clearEffectiveCacheLocked has dropped its layer cache — and
	// CR 603.10a says an LTB trigger is judged on what the permanent
	// looked like while it was still there. A creature that died under
	// a Kenrith's Transformation has none of its OWN dies-triggers,
	// and that stays true for the beat between the death and the Aura
	// falling off; one that died carrying a granted "when this
	// creature dies" still has that one (AbilityKeyFromLKI composes
	// both halves).
	oracle := CatalogKey(source)
	if ok {
		oracle = AbilityKeyFromLKI(source, lki)
	}
	// #2075 (ADR 0113 §4): the dies keyword triggers — undying and
	// persist — read off the LAST-KNOWN ability list, before any
	// catalog key is asked for, so a permanent with no catalog entry
	// (a vanilla persist creature, a token) still has them, a granted
	// instance it died wearing counts, and one that died having lost
	// all abilities has none (CR 603.10a). Only with a real snapshot:
	// the missing-LKI fallback below has no ability list worth reading.
	var triggers []TriggeredAbility
	if ok {
		triggers = ltbKeywordTriggersFor(lki)
	}
	// TriggersForKey, not TriggersForCard: this path has already
	// chosen its key, and the designation gate is evaluated against
	// the SNAPSHOT for the same CR 603.10 reason — a Case that was
	// solved when it died has its solved dies-trigger, one that was
	// not does not. ADR 0071.
	if oracle != "" {
		triggers = append(triggers, TriggersForKey(oracle, source)...)
	}
	if len(triggers) == 0 {
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
		// #925: an LTB trigger is a BATTLEFIELD ability read off the
		// permanent that just left (CR 603.10). An ability that
		// declared the graveyard is a graveyard ability, and the
		// declared-zone walk is the one that fires it — from the same
		// card, one event later in the same harvest, without the
		// battlefield LKI.
		if !TriggerWatchesFromZone(t, ZoneBattlefield) {
			continue
		}
		if !triggerWatches(t.Watches, ev.Kind) {
			continue
		}
		if t.AppliesTo != nil && !t.AppliesTo(ev, &source, lki, g) {
			continue
		}
		g.harvestMatchLocked(pass, source, lki, t, triggerOfPermanent)
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
	// #1418: a trigger that reached the queue without naming its
	// source object (a game-built item, or a Build that pointed it at
	// another card) names the object its source is now.
	g.stampSourceObjectLocked(item)
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
		g.lastKnownTriggerIdentity[cardID] = triggerIdentityLKI{
			OracleID:   c.OracleID,
			TokenKey:   c.TokenKey,
			ActiveFace: c.ActiveFace,
			AttachedTo: c.AttachedTo,
		}
		// #1218: The Ozolith's "if it had counters on it" — snapshotted
		// here, before MoveCard's battlefield-exit cleanup zeroes
		// Card.Counters a line later. Deep-copied so the map that
		// survives is not the one MoveCard is about to nil out from
		// under it. Only allocated when there is something to carry,
		// so a counterless departure (almost all of them) costs
		// nothing.
		if len(c.Counters) > 0 {
			if g.lastKnownCounters == nil {
				g.lastKnownCounters = make(map[uuid.UUID]map[string]int)
			}
			g.lastKnownCounters[cardID] = copyStringIntMap(c.Counters)
		}
		return
	}
}

// findCardByIDLocked finds the card with the given
// instance ID and returns a pointer into the slice (or nil if not
// found). Used by harvestLTB to locate a card whose zone is unknown
// to the harvester — the LTB destination depends on the cause
// (graveyard for destroy, exile for Path, hand for Unsummon).
//
// Caller must hold g.mu. Returns a pointer into the live slice;
// callers MUST NOT mutate the card through it (the harvester only
// reads).
func (g *Game) findCardByIDLocked(cardID uuid.UUID) *Card {
	// #1479: through the card index, not a walk of every zone.
	z, pos := g.locateCardLocked(cardID)
	if z == nil {
		return nil
	}
	return &z.Cards[pos]
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

// counterPlacedDeltaLocked is how many `kind` counters the
// EventCounterPlaced `ev` PUT on its target, or zero for a removal, a
// different kind or a different event. The event carries only the
// post-change total, so the previous total is the most recent
// EventCounterPlaced for the same card and kind, or zero when the card
// arrived on the battlefield (or as a token) more recently than that
// (CR 400.7: it came with no counters). The effects package's
// b33CountersPlacedDelta is the same walk for a card's own predicate.
//
// Caller must hold g.mu.
func (g *Game) counterPlacedDeltaLocked(ev Event, kind string) int {
	if ev.Kind != EventCounterPlaced || ev.Label != kind {
		return 0
	}
	before := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		prev := g.Events[i]
		if prev.Seq >= ev.Seq {
			continue
		}
		if (prev.Kind == EventETB || prev.Kind == EventTokenCreated) && prev.CardID == ev.Target {
			break
		}
		if prev.Kind == EventCounterPlaced && prev.Target == ev.Target && prev.Label == kind {
			before = prev.Amount
			break
		}
	}
	if ev.Amount <= before {
		return 0
	}
	return ev.Amount - before
}

// counterRemovedDeltaLocked is how many `kind` counters the
// EventCounterPlaced `ev` TOOK OFF its target, or zero for a placement,
// a different kind or a different event — counterPlacedDeltaLocked's
// mirror (#2466). The previous total is read the same way: the most
// recent EventCounterPlaced for the card and kind, stopping at the
// card's arrival (CR 400.7: a new object has no counters, so nothing
// can have been removed from it before its first placement).
//
// Caller must hold g.mu.
func (g *Game) counterRemovedDeltaLocked(ev Event, kind string) int {
	if ev.Kind != EventCounterPlaced || ev.Label != kind {
		return 0
	}
	before := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		prev := g.Events[i]
		if prev.Seq >= ev.Seq {
			continue
		}
		if (prev.Kind == EventETB || prev.Kind == EventTokenCreated) && prev.CardID == ev.Target {
			break
		}
		if prev.Kind == EventCounterPlaced && prev.Target == ev.Target && prev.Label == kind {
			before = prev.Amount
			break
		}
	}
	if ev.Amount >= before {
		return 0
	}
	return before - ev.Amount
}
