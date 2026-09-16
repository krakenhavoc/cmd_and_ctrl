package game

import (
	"log/slog"

	"github.com/google/uuid"
)

// events.go holds the per-game event log that the S14+ rules engine
// reads from. Events are emitted by every rules-visible mutation
// (damage dealt, card drawn, zone move, counter placement, spell
// cast / resolved, player eliminated) and appended to Game.Events
// in-order under the write lock. Pull-based — listeners registered
// via RegisterListener see each event synchronously after the
// underlying mutation, and S19+ card-effect triggers will harvest
// matching events from the log rather than being push-fired by
// each mutation site.
//
// The log is append-only per-Game, unbounded for the game's lifetime.
// A typical 30-minute Commander game produces a few thousand events;
// cloned across the 32-frame undo stack that's ~100-150k events in
// memory worst case, well under the per-room budget. Post-game
// cleanup runs on Game.End(), so unbounded growth is a non-issue.
//
// Introduced in S14 sub-PR 1 as infrastructure with zero production
// consumers; sub-PR 3 wires the effect-resolution path to emit, and
// S19 registers the first real listener (trigger harvester).

// EventKind discriminates what the event represents. Keep the set
// tight — it's an enum on the wire (when GameView.events ships in
// sub-PR 3) so every addition is a schema change. All rules-visible
// mutation sites map onto one of these.
type EventKind string

const (
	// EventCast — a spell was cast onto the stack. CardID is the
	// spell; Actor is the caster. OldZone is the zone the card was
	// cast FROM (hand, exile, command, graveyard) and NewZone is
	// always ZoneStack — "whenever you cast a spell from exile"
	// (Appa, Steadfast Guardian) reads OldZone, and CR 601.2a is the
	// reason it has to: the card is on the stack by the time the
	// event fires and carries no memory of where it came from.
	// Added in S22 alongside StackItem.CastFromZone, which is the
	// same fact recorded on the stack item.
	EventCast EventKind = "cast"

	// EventResolve — a stack item successfully resolved. CardID is
	// the card whose spell or ability resolved.
	EventResolve EventKind = "resolve"

	// EventFizzle — a spell or ability resolved but did nothing
	// (all targets illegal, or the effect short-circuited). CR
	// 608.2b's "countered by game rules."
	EventFizzle EventKind = "fizzle"

	// EventDealDamage — Amount damage was dealt from Source to Target.
	// Target may be a player or a card (distinguished by whether
	// Target resolves in PlayerByID vs the zone scan).
	EventDealDamage EventKind = "deal_damage"

	// EventChangeLife — a player's life total changed by Amount
	// (signed). Actor is who caused the change (uuid.Nil for admin
	// / SBA paths that have no single actor).
	EventChangeLife EventKind = "change_life"

	// EventDrawCard — Actor drew CardID. Fires per card, so "draw
	// 3" produces three events.
	EventDrawCard EventKind = "draw_card"

	// EventDiscardCard — Actor discarded CardID.
	EventDiscardCard EventKind = "discard_card"

	// EventMill — Actor milled CardID from the top of their library.
	// Fires per card.
	EventMill EventKind = "mill"

	// EventZoneMove — CardID moved from OldZone to NewZone. The
	// catch-all event for card motion that doesn't fit a more
	// specific kind. Draw / mill / discard / resolve all also
	// produce ZoneMove events — consumers can pick the granularity
	// they care about.
	EventZoneMove EventKind = "zone_move"

	// EventTapCard — CardID was tapped.
	EventTapCard EventKind = "tap_card"

	// EventUntapCard — CardID was untapped.
	EventUntapCard EventKind = "untap_card"

	// EventCounterPlaced — a counter of Label (see CounterKind /
	// KnownCardCounters) was placed on CardID. Amount is the new
	// count of that counter kind on the card. Fires on AddCounter
	// and on the SBA +1/+1 / -1/-1 cancel.
	// EventAttach fires when an Equipment or Aura becomes attached
	// to a permanent or player (CR 301.5c, CR 303.4). CardID and
	// Source are the attachment; Target is the host. Emitted by
	// AttachForEffect, which is the only writer of Card.AttachedTo
	// outside the state-based action.
	//
	// An event rather than a bare layerVersion bump because
	// attachment has more consumers than the layer engine — "becomes
	// attached" triggers read it, and the replay log and public game
	// log both want it greppable. Added in S24 (ADR 0036 decision 7).
	EventAttach EventKind = "attach"

	// EventUnattach is the other half: the link broke. Emitted by
	// UnattachForEffect and by the CR 704.5m/n state-based action,
	// in the latter case BEFORE an illegally attached Aura is routed
	// to its owner's graveyard, so a listener still sees what it was
	// attached to. Added in S24.
	EventUnattach EventKind = "unattach"

	EventCounterPlaced EventKind = "counter_placed"

	// EventTokenCreated — a token was created under Actor's control.
	// CardID is the new instance.
	EventTokenCreated EventKind = "token_created"

	// EventCopyApplied — a CR 707 copy effect landed on CardID,
	// copying Source. Emitted as the permanent enters, between the
	// push and the zone-move / ETB events, so the log reads
	// "Clone entered as a copy of Llanowar Elves" in the order it
	// happened. Added in S16.5 (#159).
	EventCopyApplied EventKind = "copy_applied"

	// EventSearchLibrary — Actor searched their library. Reserved
	// for S14 catalog effects that fire SearchLibrary; the log
	// entry is the "you searched your library" trigger source
	// that S19 listens for (Panoptic Mirror, etc.).
	EventSearchLibrary EventKind = "search_library"

	// EventCounterSpell — a stack item was countered (spell or
	// ability). Source is the counter itself; Target is the
	// countered item's ID. Routed to owner's graveyard by default,
	// or another zone per the counter's destination.
	EventCounterSpell EventKind = "counter_spell"

	// EventConcede — Actor conceded the game. Precedes the
	// eliminate-via-SBA path.
	EventConcede EventKind = "concede"

	// EventPlayerEliminated — Actor was eliminated from the game
	// (0 life, empty library draw, 21+ commander damage, poison >= 10,
	// concede, or manual eliminate). The single terminating event
	// for a player's participation in the game.
	EventPlayerEliminated EventKind = "player_eliminated"

	// EventETB — a permanent entered the battlefield. CardID is the
	// new permanent. Distinct from ZoneMove so listeners can key
	// off of ETB specifically without pattern-matching the zone
	// fields.
	EventETB EventKind = "etb"

	// EventLTB — a permanent left the battlefield (any reason).
	// NewZone carries the destination zone so listeners can tell a
	// "dies" (NewZone == ZoneGraveyard) apart from a bounce
	// (ZoneHand), exile (ZoneExile), or library tuck (ZoneLibrary).
	// CR 700.4: a permanent "dies" only when it goes to the
	// graveyard from the battlefield — a dies-trigger must gate on
	// ev.NewZone == ZoneGraveyard, not merely on EventLTB. Added in
	// S19 sub-PR 4.
	EventLTB EventKind = "ltb"

	// EventScry — Actor finished a scry (CR 701.18). Emitted after
	// the cards have been put back, with Amount = how many went to
	// the bottom, so a "whenever you scry" payoff sees a completed
	// scry rather than an in-flight one. Source is the card that
	// scried. Not emitted when the scry looked at nothing (an empty
	// library), because no scry happened.
	EventScry EventKind = "scry"

	// EventSurveil — Actor finished a surveil (CR 701.42). Same
	// shape as EventScry: emitted after the cards have been put
	// back, with Amount = how many went to the GRAVEYARD (not the
	// bottom — surveil has no bottom leg), so a "whenever you
	// surveil" payoff sees a completed surveil. Source is the card
	// that surveilled. Not emitted when the surveil looked at
	// nothing (an empty library).
	//
	// Deliberately a distinct kind from EventScry rather than a
	// flag on it: the two are different keywords with different
	// payoffs, and a card that cares about surveil must not fire on
	// an ordinary scry. Added in S22.
	EventSurveil EventKind = "surveil"

	// EventSacrifice — a permanent was sacrificed (CR 701.17):
	// its controller moved it to the graveyard as a cost or as
	// part of an effect's instruction. Emitted immediately BEFORE
	// the zone move, so a listener sees the permanent still on the
	// battlefield; the ordinary EventLTB (NewZone == ZoneGraveyard)
	// follows, which means a sacrificed creature fires dies-triggers
	// too. Actor is the sacrificing player (the controller), CardID
	// the permanent.
	//
	// Sacrifice is NOT destruction: it ignores indestructible and
	// regeneration, and "if a creature would die" replacements that
	// key on destruction don't see it. Aristocrats payoffs ("whenever
	// you sacrifice a permanent") watch this kind; "whenever a
	// creature dies" payoffs watch EventLTB as before. Added in S21
	// sub-PR 1.
	EventSacrifice EventKind = "sacrifice"

	// EventAttack — CardID was declared as an attacker. Actor is the
	// attacking creature's controller; Target is the player it is
	// attacking. Fires once per attacking creature, the way
	// EventDrawCard fires per card: a three-creature alpha strike
	// produces three events, so "whenever a creature you control
	// attacks" (Hellrider) triggers three times rather than once with
	// a count. Cards printed as "whenever one or more creatures you
	// control attack" therefore over-fire — the same CR 603.1 batching
	// gap EventETB already has, and no card in the catalog has that
	// wording.
	//
	// Emitted from DeclareAttacker at the moment the creature is
	// stamped — inside the declare-attackers step, before blockers
	// exist — and only on a creature's FIRST declaration. The sandbox
	// lets a player re-point an already-attacking creature at a
	// different defender (paper does not); re-pointing is not a second
	// attack and must not fire the trigger twice.
	//
	// Target is always a player. DeclareAttacker takes a player ID and
	// validates it against the seats — the engine has no
	// attack-a-planeswalker path at all — so "the player or
	// planeswalker it's attacking" collapses to the player. When
	// planeswalker defenders land, Target widens to "player or
	// permanent" the way EventDealDamage's already is, and existing
	// consumers keep working because they read Target as an opaque ID.
	//
	// Combat STATE is older and separate: DeclareAttacker stamps
	// Card.AttackingTarget and ClearCombat wipes it, so a spell that
	// reads the board at resolution (Aetherize) needs no event. This
	// kind exists for triggers, which need the moment rather than the
	// state. Added in S22.
	EventAttack EventKind = "attack"

	// EventBecomesTarget — an object or player became the target of
	// a spell or ability (CR 115.7). Actor is the controller of the
	// spell / ability, Source is its source card, Target is the
	// thing that was targeted, and CardID repeats Target when the
	// target is a card (uuid.Nil when it is a player) so a consumer
	// can tell the two apart without a zone scan.
	//
	// Fires once per target SLOT at the moment the targets are
	// chosen (CR 601.2c for a spell, 602.2b for an activated
	// ability, 603.3d for a triggered one) — which is the moment the
	// printed clause names, and is deliberately NOT resolution:
	// "becomes the target" still fires for a spell that is later
	// countered, and the triggers it produces go on the stack ABOVE
	// that spell. That ordering is the whole point of Monk Gyatso:
	// his trigger resolves first and removes the creature, so the
	// spell that targeted it fizzles.
	//
	// A two-target spell emits two events; a spell that targets the
	// same object twice (AllowSame) likewise emits two, matching CR
	// 115.7's per-instance-of-the-word-"target" reading.
	//
	// Known gap: an effect that CHANGES a spell's targets after
	// announce (Deflecting Swat, Redirect) does not re-emit, because
	// the engine has no change-targets path at all. When one lands,
	// it emits here too.
	//
	// Added in S22 for Monk Gyatso; "becomes the target of a spell
	// or ability" is a common Commander clause and this is the only
	// event that can serve any of it.
	EventBecomesTarget EventKind = "becomes_target"

	// EventTrigger — a triggered ability was announced onto
	// PendingTriggers. Legacy (manual) announce goes through
	// AnnounceTrigger in S13.1; S19's auto-announce will flow
	// through here too. The event lets listeners observe trigger
	// creation without having to poll PendingTriggers.
	EventTrigger EventKind = "trigger"

	// EventEffectError — an effect primitive (S14 catalog) failed
	// to apply. Caller logs + keeps moving; the event is the
	// debugging breadcrumb. ErrorMsg carries the reason.
	EventEffectError EventKind = "effect_error"

	// EventStepTransition is an engine-internal sentinel used only
	// by the S17 replacement-effect pipeline. Fired from the top of
	// runStepEntryHooksLocked so skip-step replacements (Stasis
	// cancels StepUntap) can intercept. Does NOT emit to the public
	// event log — cards read it only via ReplacementEffect.Watches.
	// Added in S17 sub-PR 2.
	EventStepTransition EventKind = "step_transition"

	// EventBeginUpkeep — the active player's upkeep step began.
	// Actor is the active player (whose upkeep it is). The S19
	// harvester fans this out to "at the beginning of your upkeep"
	// triggers; a card gates AppliesTo on ev.Actor == controller for
	// "your upkeep" (vs "each upkeep"). Emitted from
	// runStepEntryHooksLocked on entering StepUpkeep. Added in S19
	// sub-PR 5.
	EventBeginUpkeep EventKind = "begin_upkeep"

	// EventBeginEndStep — the active player's end step began. Actor
	// is the active player (whose end step it is). "At the beginning
	// of your end step" (Thassa, Y'shtola Rhul) gates AppliesTo on
	// ev.Actor == controller; "at the beginning of the end step"
	// (any player's) does not. Emitted from runStepEntryHooksLocked
	// on entering StepEnd, alongside the delayed-trigger drain —
	// both are "beginning of the end step" triggers and land on
	// PendingTriggers together. Added in S22.
	EventBeginEndStep EventKind = "begin_end_step"

	// EventBeginPrecombatMain — the active player's PRECOMBAT main
	// phase began. Actor is the active player, so "at the beginning
	// of your precombat main phase" (Hulking Raptor, Braids, Kiora)
	// gates AppliesTo on ev.Actor == controller.
	//
	// Precombat only, because that is the phase every card in the
	// family names: the printed wording is either "your precombat
	// main phase" or "your FIRST main phase", and on an ordinary turn
	// those are the same phase. Emitting on the postcombat main too
	// would fire a Raptor's ritual twice a turn. Added in S30
	// alongside ward, because a mana ritual gated on the phase was
	// the last line of Hulking Raptor standing between it and being
	// implemented as printed.
	EventBeginPrecombatMain EventKind = "begin_precombat_main"
	// EventBeginDrawStep — the active player's draw step began.
	// Actor is that player. Emitted AFTER the CR 504.1 turn-based
	// draw, which is the rules-correct order: the turn-based draw
	// does not use the stack and happens first, and "at the
	// beginning of the draw step" triggers go on the stack when a
	// player next receives priority (CR 504.2) — by which time the
	// card is already in hand.
	//
	// Not emitted on the turn-1 skipped draw step of the starting
	// player (CR 103.7c): that step still happens and still grants
	// priority, but this project's cursor returns before reaching
	// here. Howling Mine on turn 1 of the first player's turn is the
	// only case that notices, and it is not worth restructuring the
	// step hook for.
	//
	// Added in S22 for Howling Mine and the "each player's draw
	// step" family.
	EventBeginDrawStep EventKind = "begin_draw_step"

	// EventManaAbilityActivated — a mana-producing ability fired.
	// Actor = controller, Source = the permanent that produced the
	// mana. S15 sub-PR 2.
	EventManaAbilityActivated EventKind = "mana_ability_activated"

	// EventManaAdded — one mana token landed in a player's pool.
	// Actor = pool owner, Source = the producing permanent (uuid.Nil
	// for non-card sources). The token's color rides Amount as 0
	// (W/U/B/R/G/C carries no numeric weight) — the wire-side mana
	// pool is the canonical projection. S15 sub-PR 2.
	EventManaAdded EventKind = "mana_added"

	// EventManaPoolEmptied — a player's mana pool was cleared at a
	// step / phase boundary (CR 106.4) or by an admin reset. Actor =
	// pool owner. Amount = number of tokens that were dropped.
	// S15 sub-PR 2.
	EventManaPoolEmptied EventKind = "mana_pool_emptied"

	// EventManaSpent — the strict-mode S15 cost gate deducted mana
	// from a caster's pool to pay for a spell. Actor = caster,
	// Source = card being cast. Permissive / forced casts emit
	// EventCostWarning instead. S15 sub-PR 3.
	EventManaSpent EventKind = "mana_spent"

	// EventCostWarning — a cast_spell action proceeded under the S15
	// permissive-default cost path OR under a strict-mode override
	// (ForceCast=true) despite the pool not covering the cost.
	// Breadcrumb for the client's auto-toast UI and a debug signal
	// for operators that want to audit "did the caster actually pay
	// for this?" Actor = caster, Source = card being cast.
	// S15 sub-PR 3.
	EventCostWarning EventKind = "cost_warning"

	// EventSagaChapter — a lore counter advanced a Saga ONTO a
	// chapter (CR 714.2c). Target / CardID = the Saga, Actor = its
	// controller, Amount = the chapter number just reached. One
	// event per chapter crossed, in ascending order, so a Saga that
	// gains two lore counters at once triggers both chapters in the
	// printed order.
	//
	// A dedicated kind rather than a predicate over
	// EventCounterPlaced: that event is emitted for REMOVALS too and
	// carries only the post-change total, so "went from 1 to 2"
	// and "went from 3 to 2" are indistinguishable on it. A chapter
	// ability that fired when a lore counter was removed would be a
	// silent rules bug with no way for a card file to defend itself.
	// Added in S27.
	EventSagaChapter EventKind = "saga_chapter"

	// EventStepBegan — the turn cursor entered a step. Actor is the
	// active player, Step the step (typed), Amount the turn number and
	// Label the step name. Emitted from runStepEntryHooksLocked AFTER
	// the S17 skip-step replacement window has had its say, so a step
	// that Stasis cancelled never announces, and never while the
	// mulligan window holds the cursor at Untap, so it fires exactly
	// once per step that really begins.
	//
	// It is both the public game log's spine and, since #588, the
	// trigger event for every step: "at the beginning of combat",
	// "at the beginning of your postcombat main phase", "at end of
	// combat" watch this kind with a predicate on Step
	// (effects.StepBegan). The older per-step kinds — EventBeginUpkeep,
	// EventBeginDrawStep, EventBeginPrecombatMain, EventBeginEndStep —
	// are emitted a little later in the hook, after that step's
	// turn-based action, and stay for the cards that depend on that
	// ordering (the draw-step trigger must see the card already drawn).
	//
	// Added in S31 sub-PR 0; opened to the harvester in #588.
	EventStepBegan EventKind = "step_began"

	// EventBlock — CardID was declared as a blocker. Actor is the
	// blocking creature's controller, Target the attacker it is
	// blocking. The other half of EventAttack, and emitted under the
	// same rule: only on a creature's FIRST declaration against a
	// given attacker, so re-pointing a blocker in the sandbox does
	// not announce twice. Added in S31 sub-PR 0 for the public game
	// log — no card in the catalog reads "becomes blocked by" yet.
	EventBlock EventKind = "block"
	// EventBattleDefeated — a battle's last defense counter came off
	// (CR 310.9). Source / Target / CardID = the battle, Actor = its
	// controller.
	//
	// Emitted from the state-based-action pass IMMEDIATELY BEFORE the
	// CR 704.5p move that puts the battle in the graveyard, so a
	// defeated trigger's source is still findable on the battlefield
	// when the harvester walks it. A dies-trigger shape (EventLTB
	// plus the LKI snapshot) would also work and would be lossier:
	// the point of a Siege's defeated trigger is to reach a card that
	// has just left, and the fewer hops between "defeated" and the
	// exile-and-cast that follows, the better. Added in S27.
	EventBattleDefeated EventKind = "battle_defeated"

	// EventRevealCards — Actor showed CardID to the whole table (CR
	// 701.16). Fires once per card, so "reveal the top five cards of
	// your library" produces five events sharing one RevealSeq; the
	// wire projection groups them back into a single announcement.
	// Source is the card whose effect revealed; OldZone is the zone
	// the card was revealed FROM (it has not moved — a reveal is not
	// a zone change); Label is the one-line reason the client shows.
	//
	// The knowledge half of a reveal is not carried by this event: it
	// is the KnownBy set, which RevealForEffect writes before
	// emitting. The event exists because knowledge alone is silent —
	// nothing tells the other seats that a reveal just happened, or
	// what it was once a shuffle has taken the knowledge away again.
	// Added in S22.
	EventRevealCards EventKind = "reveal_cards"
)

// Event is a single entry in the per-game event log. Tagged union
// shape: Kind selects which of the payload fields are meaningful,
// the rest are zero. Chosen over per-kind interfaces so Clone is
// a simple slice value-copy and wire encoding is flat JSON.
//
// Seq is a monotonic per-Game counter stamped at emit time. First
// event has Seq == 1; zero is the un-stamped sentinel.
type Event struct {
	Seq  uint64    `json:"seq"`
	Kind EventKind `json:"kind"`

	// Actor is the player responsible for the event (caster,
	// controller, drawing player, etc.). uuid.Nil for admin / SBA
	// paths that have no single actor.
	Actor uuid.UUID `json:"actor,omitempty"`

	// Source is the card whose effect produced the event
	// (Lightning Bolt that dealt the damage, Counterspell that
	// countered, the ETB creature itself). uuid.Nil when the event
	// isn't tied to a specific source card (e.g. manual ChangeLife
	// via the player header).
	Source uuid.UUID `json:"source,omitempty"`

	// Target is the thing the event acts on: a player for life /
	// damage events, a card for zone moves / counters / tap / ETB.
	// Consumers key off of Kind to decide how to interpret.
	Target uuid.UUID `json:"target,omitempty"`

	// CardID names the card the event references. For ZoneMove /
	// DrawCard / Mill / Discard it's the moved card. For Cast /
	// Resolve it's the spell. For TokenCreated it's the new token.
	CardID uuid.UUID `json:"card_id,omitempty"`

	// StackItemID names the stack item the event is about, when the
	// event is about one. Set on EventBecomesTarget (S30, for ward).
	//
	// It is NOT the same as Source, and the difference is the whole
	// reason the field exists. For a SPELL the two coincide: the
	// StackItem's ID equals the spell card's InstanceID. For an
	// ACTIVATED or TRIGGERED ability they do not — the item carries
	// a freshly minted UUID and Source names the permanent the
	// ability came from, which stays on the battlefield. So an
	// effect that has to act on "the spell or ability that targeted
	// me" — counter it, copy it, redirect it — cannot get there from
	// Source alone, and ward is the first effect that has to.
	StackItemID uuid.UUID `json:"stack_item_id,omitempty"`

	// Amount is the signed / count payload: damage dealt, life
	// delta, number of cards, counter count after the change.
	Amount int `json:"amount,omitempty"`

	// Label is a free-text qualifier: counter kind
	// ("+1/+1", "loyalty", "poison"), trigger label, error
	// classification. Kept as a string so adding new counter kinds
	// doesn't require a schema change.
	Label string `json:"label,omitempty"`

	// Step is the step that began, on EventStepBegan. Typed so a
	// trigger's predicate compares a constant rather than a string
	// (Label still carries the name for the public log). Added in
	// #588.
	Step Step `json:"step,omitempty"`

	// OldZone / NewZone are the zone kinds for ZoneMove-shaped
	// events. Empty string means "not applicable."
	OldZone ZoneKind `json:"old_zone,omitempty"`
	NewZone ZoneKind `json:"new_zone,omitempty"`

	// ErrorMsg carries the failure reason on EventEffectError.
	ErrorMsg string `json:"error_msg,omitempty"`

	// RevealSeq groups the per-card EventRevealCards events of ONE
	// reveal, and is the Seq the first of them was stamped with.
	// "Reveal the top five cards of your library" is five events —
	// one per card, because CardID names exactly one — and the table
	// saw one thing, not five. protocol.publicRevealsOf keys on this
	// to put them back together.
	//
	// Deliberately an explicit key rather than inferred from
	// adjacency: two reveals in one resolution (each opponent reveals
	// their top card) are adjacent and are not the same
	// announcement. Deliberately a Seq rather than a fresh UUID so a
	// replayed game produces a byte-identical event stream.
	//
	// Zero on every other kind. Added in S22.
	RevealSeq uint64 `json:"reveal_seq,omitempty"`

	// Combat marks an EventDealDamage as combat damage (CR 510) —
	// dealt by an attacking or blocking creature in the combat
	// damage step, including trample overflow routed through the
	// damage-assignment prompt. False for spell / ability damage
	// (Lightning Bolt, Sulfuric Vortex). Combat-damage events also
	// carry the dealing creature's controller in Actor, captured at
	// emit time (the creature may die to simultaneous damage before
	// a downstream prompt is answered). "Whenever ~ deals combat
	// damage to a player" triggers read both. Added in S19 sub-PR 7.
	Combat bool `json:"combat,omitempty"`
}

// emitBecameTargetLocked fans one EventBecomesTarget out per target
// slot in `targets`. Called from every site that finishes choosing
// targets for a spell or ability: the cast path, the activated-
// ability announce, the triggered-ability target pick, and the
// manual sandbox announce.
//
// `actor` is the controller of the spell or ability, `source` is its
// source card and `itemID` is its stack item. TargetSelf /
// TargetNone slots are skipped — neither names an object anyone else
// chose.
//
// `source` and `itemID` are equal for a cast spell and differ for an
// ability; see Event.StackItemID for why both are carried. A caller
// with no item to name (the manual sandbox announce, before the item
// exists) may pass uuid.Nil.
//
// Emitted AFTER the item exists, so a trigger harvested off this
// event lands on PendingTriggers above the thing that targeted.
// Caller must hold g.mu.
func (g *Game) emitBecameTargetLocked(actor, source, itemID uuid.UUID, targets []TargetRef) {
	for _, t := range targets {
		switch t.Kind {
		case TargetCard:
			g.EmitEvent(Event{
				Kind:        EventBecomesTarget,
				Actor:       actor,
				Source:      source,
				StackItemID: itemID,
				Target:      t.ID,
				CardID:      t.ID,
			})
		case TargetPlayer:
			g.EmitEvent(Event{
				Kind:        EventBecomesTarget,
				Actor:       actor,
				Source:      source,
				StackItemID: itemID,
				Target:      t.ID,
			})
		}
	}
}

// EmitEvent appends ev to the game's event log under the existing
// write lock. Caller MUST hold g.mu — the emit is a mutation and
// participates in the same atomic write as the action that produced
// it. Stamps Seq monotonically; dispatches to registered listeners
// synchronously before returning, so a listener that wants to
// trigger a follow-on event sees state consistent with the event
// it's reacting to.
func (g *Game) EmitEvent(ev Event) {
	g.eventSeq++
	ev.Seq = g.eventSeq
	g.Events = append(g.Events, ev)
	// S17 sub-PR 6 diagnostic: effect-error events are otherwise
	// silent (no client toast yet). Surfacing them in the server log
	// so manual-test regressions have a visible breadcrumb. Keep
	// until the client learns to render effect_error as a toast.
	if ev.Kind == EventEffectError {
		slog.Warn("effect error",
			"seq", ev.Seq,
			"game", g.ID.String(),
			"source", ev.Source.String(),
			"error", ev.ErrorMsg,
		)
	}
	g.notifyListenersLocked(ev)
}
