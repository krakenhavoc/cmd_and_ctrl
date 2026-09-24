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

	// EventResolve — a stack item successfully resolved. For a SPELL,
	// CardID (and Source) is the spell's card. For an ABILITY, CardID
	// is empty — the item has no card on the stack — Source is the
	// ability's source and Label the item's label; the public log
	// names the ability off those two (#1257).
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

	// EventCycle — Actor cycled CardID (CR 702.29b): they activated
	// its cycling ability and paid the cost, which discarded it.
	// Source and CardID are both the cycled card.
	//
	// Emitted AFTER the cost's EventDiscardCard, with the card
	// already in the graveyard, because that is where CR 702.29c
	// puts it for the watchers. Both events fire for one cycling and
	// that is the rule, not double-counting: a "cycles or discards"
	// clause (CR 702.29d) watches one of them, and the discard
	// payoffs in the catalog (Marauding Mako, Scrounging Skyray)
	// watch the other.
	//
	// A watcher on the BATTLEFIELD — Astral Slide, Drake Haven,
	// Fluctuator — sees this through the harvester's ordinary
	// battlefield scan. "When you cycle THIS card" (Magmakin
	// Artillerist) does not, and waits on a zone dimension for
	// triggered abilities: ADR 0062 Decision 7.
	EventCycle EventKind = "cycle"

	// EventSpecialAction — Actor took a CR 116.2 special action on
	// CardID. Label is the action as the card prints it ("Foretell
	// {2}", "Suspend 1—{R}").
	//
	// One event kind for every special action, the way the verb is
	// one verb: the rules group them because their contract is
	// identical, and a payoff that ever cares which one was taken
	// reads Label. Nothing in the catalog watches it today; it is
	// the public log's record that a card left a hand for exile
	// without being cast or discarded, which no other event says.
	//
	// That last sentence was a claim about a line that did not exist
	// until #1021: the log had no arm for this kind, so the table saw
	// the zone move and never the word "Foretell {2}". It is now
	// projected as the `special_action` entry, whose `label` is this
	// event's Label verbatim (protocol.LogSpecialAction).
	//
	// Emitted AFTER the action has been carried out, so the card is
	// already in exile when a watcher sees it. ADR 0062 Decision 4,
	// #658 / #659.
	EventSpecialAction EventKind = "special_action"

	// EventBecomesPlotted — CardID, a card in exile, became plotted
	// (CR 702.170c/d). Actor is the player who plotted it: the owner
	// for the plot special action (CR 702.170a), the resolving item's
	// controller for an effect ("exile target spell. It becomes
	// plotted" — Aven Interrupter plots an OPPONENT's spell, and the
	// actor is Aven's controller, not the spell's owner). Source is
	// what did it: the card itself for the keyword, the effect's
	// source card otherwise.
	//
	// Emitted from Game.PlotExiledCardForEffect, the one function both
	// routes end in, and only once the card is in exile with its
	// permission granted — a card that never landed (a commander whose
	// owner took the CR 903.9 offer) is not plotted and fires nothing.
	// A plain exile never emits it.
	//
	// "When this card becomes plotted" (Longhorn Sharpshooter, Aloe
	// Alchemist) watches it from EXILE — #925's TriggeredAbility.Zones,
	// the zone suspend's triggers watch from — through
	// effects.WhenThisBecomesPlotted. Added for #1382 (ADR 0062
	// amendment 2026-09-24).
	EventBecomesPlotted EventKind = "becomes_plotted"

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

	// EventAttachSkipped — an effect tried to attach something and
	// CR 701.3b's "the attachment doesn't happen" applied: the
	// attachment is not on the battlefield (its equip was activated
	// and it was sacrificed in response), the host is not, or the two
	// are the same permanent. CardID and Source are the would-be
	// attachment, Target the would-be host, and Label says which of
	// the three it was.
	//
	// Deliberately NOT EventEffectError, for the reason
	// EventPendingChoiceDropped is not one: nothing failed. An equip
	// whose Equipment has left still RESOLVES and simply does nothing
	// (CR 608.2, CR 702.6a), and reporting that as an effect error
	// both misreads the rule and fails the catalog soak, which treats
	// any effect error as a bug. The breadcrumb stays because a
	// silent no-op is undebuggable; it produces no public log
	// entry.
	// #812.
	EventAttachSkipped EventKind = "attach_skipped"

	// EventControlChanged — CardID changed controller (CR 613.1b).
	// Actor is the player who GAINED control, Target the player who
	// LOST it; Source is the card whose effect took it, and is
	// uuid.Nil when control REVERTED because the effect ended (the
	// baseline is nobody's effect — Act of Treason's creature going
	// home at cleanup).
	//
	// Emitted from the one materialise step at the end of the layer
	// recompute (materialiseControlLocked), which is where the delta
	// is known: control is layer 2's output, so a gain, an exchange,
	// an expiry and a Mind Control being destroyed are all the same
	// event from the same place. It rides the batch that was open
	// when the pass ran, so an exchange (CR 701.12) is two events in
	// one batch. Added for #930.
	EventControlChanged EventKind = "control_changed"

	// EventCounterPlaced — a counter of Label (see CounterKind /
	// KnownCardCounters) was placed on or removed from a card.
	// TARGET names the card — not CardID, which this one leaves
	// unset — and Amount is the count of that kind on it AFTER the
	// change, so a placement and a removal are the same event with a
	// different number and neither carries the delta.
	//
	// ACTOR is the player who PUT the counters, when the placement
	// came through the CR 614 counter window and named one
	// (ReplacementEvent.CounterPlacer) — the source's controller for a
	// CR 120.3d damage result, the proliferating player for CR 701.34.
	// SOURCE is the card whose effect or damage placed them. Both stay
	// uuid.Nil when the placement names neither, which is still most of
	// them: applyCounterLocked is also reached from a paid cost, a
	// trigger and the CR 704.5q cancel, and no single player is
	// responsible for those. A "whenever YOU put" reader takes Actor
	// first and falls back to its own heuristic on Nil (ADR 0056
	// Decision 5).
	//
	// Fires on AddCounter and on the SBA +1/+1 / -1/-1 cancel. The
	// public log narrates it as the `counters` entry, for every kind
	// but the two that are already a line somewhere else — see
	// protocol.counterKindIsNarrated (#1021).
	EventCounterPlaced EventKind = "counter_placed"

	// EventPlayerCounterPlaced — a counter of Label was placed on or
	// removed from a PLAYER. Target is the player, Actor the placer
	// (uuid.Nil when unknown), Source the card whose effect or damage
	// placed them, and Amount is THE SIGNED DELTA THAT LANDED — +3 for
	// three poison counters, -1 for one removed, and never zero,
	// because a placement that moves nothing emits nothing.
	//
	// A SEPARATE KIND from EventCounterPlaced, deliberately. The
	// card-counter trigger helpers compare ev.Target against card IDs
	// and walk the log backwards by Label to recover a delta
	// (batch33_helpers.go, batch12_helpers.go); a player UUID carrying
	// the label "poison" must never reach them, and one kind for both
	// would put it there.
	//
	// It carries the DELTA where the card event carries the new total,
	// which is the other half of the same lesson: the card event's
	// post-change total is exactly what forced three helpers to walk
	// the log backwards to find out how many counters had been placed.
	// The total is on Player.Counters for anyone who wants it.
	//
	// The layer listener bumps on it unconditionally (layer_listener.go):
	// a "corrupted" static reading an opponent's poison count, or a
	// "+1/+1 for each poison counter your opponents have", is a layer
	// input, and player counters change rarely enough that a gate would
	// save nothing measurable and go stale.
	//
	// ADR 0056 Decision 5. Its public log line — "Alice got 3 poison
	// counters (7/10)" — is Decision 6 and lands with the client PR;
	// until then it is a deliberate silence, recorded in
	// protocol.silentEventKinds.
	EventPlayerCounterPlaced EventKind = "player_counter_placed"

	// EventTokenCreated — a token was created under Actor's control.
	// CardID is the new instance.
	EventTokenCreated EventKind = "token_created"

	// EventCopyApplied — a CR 707 copy effect landed on CardID,
	// copying Source. Emitted as the permanent enters, between the
	// push and the zone-move / ETB events, so the log reads
	// "Clone entered as a copy of Llanowar Elves" in the order it
	// happened. Added in S16.5 (#159).
	EventCopyApplied EventKind = "copy_applied"

	// EventStorm — a storm trigger resolved and settled on its count
	// (CR 702.40a). Actor is the player who cast the storm spell,
	// Source and CardID are that spell, and Amount is how many other
	// spells were cast before it this turn — which is how many copies
	// the trigger is about to create, and is zero for the turn's
	// first spell.
	//
	// It exists because nothing else says the number. A spell COPY is
	// created and not cast, so createSpellCopyLocked deliberately
	// emits no EventCast (CR 707.10) and emits nothing else either;
	// and the ability's own EventResolve names the ability ("Grapeshot
	// — storm resolved", #1257) but not the count. The table would
	// otherwise watch N Grapeshots appear from nowhere
	// with nothing written down about why there are N. The count is
	// the card, so it gets a line — the same argument saga chapters
	// and Class levels get theirs. See ADR 0086 Decision 5.
	//
	// Emitted once per storm trigger, by StormCountForEffect
	// (storm.go), as the trigger resolves and before the first copy
	// is created. Nothing watches it.
	EventStorm EventKind = "storm"

	// EventSearchLibrary — Actor searched a library; Target names
	// WHOSE library it was. The two agree for an ordinary tutor, but
	// not for Bribery-style search of another player's library
	// (#1230) — Target is what tells "an opponent searched THEIR
	// library" (Actor == Target) apart from "an opponent searched
	// YOUR library" (Target == the watcher, Actor someone else),
	// which #1335 found EventSearchLibrary could not do at all before
	// this. Reserved for S14 catalog effects that fire
	// SearchLibrary; the log entry is the "you searched your library"
	// trigger source that S19 listens for (Panoptic Mirror,
	// Archivist of Oghma, etc.).
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

	// EventScry — Actor finished a scry (CR 701.22). Emitted after
	// the cards have been put back, with Amount = how many went to
	// the bottom, so a "whenever you scry" payoff sees a completed
	// scry rather than an in-flight one. Source is the card that
	// scried. Not emitted when the scry looked at nothing (an empty
	// library), because no scry happened.
	//
	// LookedAt is the SIZE of the scry — the "2" in "scry 2" — and
	// Amount is what moved. They are two different numbers and the
	// log says both; see the field.
	EventScry EventKind = "scry"

	// EventSurveil — Actor finished a surveil (CR 701.25). Same
	// shape as EventScry: emitted after the cards have been put
	// back, with Amount = how many went to the GRAVEYARD (not the
	// bottom — surveil has no bottom leg), so a "whenever you
	// surveil" payoff sees a completed surveil, and LookedAt = the
	// size of the surveil. Source is the card that surveilled. Not
	// emitted when the surveil looked at nothing (an empty library).
	//
	// Deliberately a distinct kind from EventScry rather than a
	// flag on it: the two are different keywords with different
	// payoffs, and a card that cares about surveil must not fire on
	// an ordinary scry. Added in S22.
	EventSurveil EventKind = "surveil"

	// EventSacrifice — a permanent was sacrificed (CR 701.21):
	// its controller moved it to the graveyard as a cost or as
	// part of an effect's instruction. Emitted immediately BEFORE
	// the zone move, so a listener sees the permanent still on the
	// battlefield; the ordinary EventLTB (NewZone == ZoneGraveyard)
	// follows, which means a sacrificed creature fires dies-triggers
	// too. Actor is the sacrificing player (the controller), CardID
	// the permanent.
	//
	// Sacrifice is NOT destruction: it ignores indestructible
	// (CR 702.12b) and regeneration (CR 701.19a), and "if a creature
	// would die" replacements that key on destruction don't see it —
	// the exit it takes carries no Destruction flag (#667). Aristocrats payoffs ("whenever
	// you sacrifice a permanent") watch this kind; "whenever a
	// creature dies" payoffs watch EventLTB as before. Added in S21
	// sub-PR 1.
	EventSacrifice EventKind = "sacrifice"

	// EventRegenerated — a regeneration shield replaced a destruction
	// (CR 701.19a). Emitted by the engine built-in that owns the rule
	// (regeneration.go), AFTER the shield has been spent and the
	// permanent has been tapped, cleaned of damage and taken out of
	// combat, and INSTEAD of the destruction: nothing left the
	// battlefield, so there is no EventZoneMove and no EventLTB, and
	// no dies-trigger fires.
	//
	// Actor is the permanent's controller, CardID and Source the
	// permanent itself. Nothing in the catalog watches it yet — it is
	// here because a regeneration is a thing that HAPPENED and the
	// game log has to be able to say so, and because "whenever a
	// permanent is regenerated" is a printed wording.
	//
	// It is NOT emitted when a shield is created: creating one changes
	// nothing a player can observe except the shield count itself,
	// which the wire carries on the card. Added in #667.
	EventRegenerated EventKind = "regenerated"

	// EventAttack — CardID was declared as an attacker. Actor is the
	// attacking creature's controller; Target is the player it is
	// attacking. Fires once per attacking creature, the way
	// EventDrawCard fires per card: a three-creature alpha strike
	// produces three events, so "whenever a creature you control
	// attacks" (Hellrider) triggers three times rather than once with
	// a count. Cards printed as "whenever one or more creatures you
	// control attack" therefore over-fire — the same "one or more" batching
	// gap EventETB already has, and no card in the catalog has that
	// wording.
	//
	// Emitted from commitAttackDeclarationLocked — the lock-in, not
	// the click (#859). CR 508.1 makes declaring attackers ONE
	// turn-based action, so the sandbox's per-creature
	// DeclareAttacker verb only stages the attack; nothing is
	// announced until the declaration is complete, which is the first
	// priority boundary of the declare-attackers step — still inside
	// that step and still before blockers exist. A creature
	// re-pointed at a different defender before that boundary
	// therefore announces ONCE, naming the defender it ends on, and a
	// creature re-pointed after it is not announced again at all
	// (CR 508.1: a creature is declared as an attacker once).
	//
	// A permanent PUT onto the battlefield attacking (CR 506.3c —
	// Parhelion II's Angels, Adeline's Humans, Legion Loyalty's
	// myriad copies) was never declared and gets no event; see
	// CreateTokensAttackingForEffect.
	//
	// Target is the DEFENDER: a player since S22, and since S27 a
	// planeswalker or a battle too (CR 506.2, 508.1d) —
	// classifyAttackTargetLocked is what widened it. Consumers that
	// want "the defending player" behind a planeswalker or battle go
	// through DefendingPlayerForAttackForEffect.
	//
	// Combat STATE is older and separate: DeclareAttacker stamps
	// Card.AttackingTarget and ClearCombat wipes it, so a spell that
	// reads the board at resolution (Aetherize) needs no event. This
	// kind exists for triggers, which need the moment rather than the
	// state. Added in S22.
	EventAttack EventKind = "attack"

	// EventBecomesTarget — an object or player became the target of
	// a spell or ability (CR 115.3). Actor is the controller of the
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
	// 115.3's per-instance-of-the-word-"target" reading.
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

	// EventActivateAbility — a player ACTIVATED an activated ability
	// (CR 602.2b), announced onto the stack. Actor is the activator,
	// Source and CardID are the permanent (or, since #660, the hand
	// card) whose ability it is, Label is the ability's printed label
	// — the same string the activation record is keyed by — and
	// Exhaust says whether the ability prints the exhaust keyword.
	//
	// A SECOND event beside the EventTrigger the announce has always
	// emitted, rather than fields bolted onto that one, and the
	// difference is the whole reason it exists (#1184): EventTrigger
	// is the "an item reached PendingTriggers" breadcrumb, shared by
	// triggered abilities, and triggerHarvester.OnEvent returns
	// immediately on it so that a trigger firing further triggers does
	// not re-enter the harvest. Nothing can watch it, by design. This
	// kind is the ANNOUNCEMENT, it carries the ability's identity, and
	// the harvester walks it like any other event.
	//
	// It is deliberately wider than the two cards that asked for it.
	// "Whenever you activate an exhaust ability" (Rangers' Refueler,
	// Afterburner Expert) reads Actor and Exhaust; the still-open
	// "whenever an OPPONENT activates an ability" watch (Harsh Mentor,
	// Runic Armasaur — docs/engine-seams.md) needs exactly this event
	// with ByAnOpponent in place of ByYou and no Exhaust test, so that
	// row closes on this shape rather than on a second one.
	//
	// CR 605's mana abilities take the other path and keep the other
	// kind: EventManaAbilityActivated carries the same Label and
	// Exhaust stamps (#1183 gave that path the same record), so a
	// watcher that means "any activated ability" watches both kinds.
	// Two kinds and not one because the two paths differ in what a
	// watcher may assume — a mana ability used no stack and granted
	// nobody priority.
	EventActivateAbility EventKind = "activate_ability"

	// EventEffectError — an effect primitive (S14 catalog) failed
	// to apply. Caller logs + keeps moving; the event is the
	// debugging breadcrumb. ErrorMsg carries the reason.
	EventEffectError EventKind = "effect_error"

	// EventPendingChoiceDropped — a choice was NOT queued (or was
	// swept from the queue) because its chooser has left the game.
	// Deliberately not EventEffectError: nothing failed — CR 800.4a
	// says the eliminated player's objects, and any decision left
	// for them, cease to exist along with them, so declining to ask
	// is the correct outcome, not a bug. Actor is the would-be
	// chooser, Source is the choice's card (when known), and Label
	// carries the PendingChoiceKind so a stall dump can tell a
	// dropped "pick_target" from a dropped "trigger_prompt". #864.
	EventPendingChoiceDropped EventKind = "pending_choice_dropped"

	// EventPendingChoiceReassigned — a choice owed by a player who has
	// left the game was handed to somebody else instead of being
	// dropped (CR 800.4g/h). Actor is the departed chooser, Target is
	// the player who inherits the prompt, Source is the choice's card
	// (when known), and Label carries the PendingChoiceKind — the same
	// three fields EventPendingChoiceDropped uses, so a stall dump can
	// read the two side by side and see which prompts moved and which
	// ended.
	//
	// The engine's event log is not on the wire (docs/protocol.md), and
	// this one is deliberately not projected into the public `log`
	// either: the reassignment is already visible to the table as the
	// prompt itself, which the next snapshot renders to its new
	// chooser. #902.
	EventPendingChoiceReassigned EventKind = "pending_choice_reassigned"

	// EventStepTransition is an engine-internal sentinel used only
	// by the S17 replacement-effect pipeline. Fired from the top of
	// runStepEntryHooksLocked so skip-step replacements (Stasis
	// cancels StepUntap) can intercept. Does NOT emit to the public
	// event log — cards read it only via ReplacementEffect.Watches.
	// Added in S17 sub-PR 2.
	EventStepTransition EventKind = "step_transition"

	// EventKeywordAction is the second engine-internal sentinel of
	// the same shape, and the replacement-watch key for a KEYWORD
	// ACTION with a count: proliferate (CR 701.34), scry (CR 701.22),
	// surveil (CR 701.25). Cards read it only via
	// ReplacementEffect.Watches; nothing emits it to the public log.
	//
	// One key for all three because RepEventKeywordAction is one
	// kind: which action it is lives on the event
	// (ReplacementEvent.KeywordAction) and the card-side helper
	// narrows on it, so a "if you would scry" replacement is not
	// woken by a proliferate. Reusing EventScry and EventSurveil here
	// would key a PRE-event window on the names of two POST-event
	// facts — those fire after the player has put the cards back,
	// with the count this window settled on — and there is no
	// EventProliferate at all. #976.
	EventKeywordAction EventKind = "keyword_action"

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
	// Not emitted on the turn-1 skipped draw step of a two-player
	// game's starting player (CR 103.8a; at three or more seats
	// CR 103.8c has nobody skip, so the draw and this event both
	// happen): that step still happens and still grants
	// priority, but this project's cursor returns before reaching
	// here. Howling Mine on turn 1 of the first player's turn is the
	// only case that notices, and it is not worth restructuring the
	// step hook for.
	//
	// Added in S22 for Howling Mine and the "each player's draw
	// step" family.
	EventBeginDrawStep EventKind = "begin_draw_step"

	// EventManaAbilityActivated — a mana-producing ability fired.
	// Actor = controller, Source and CardID = the permanent that
	// produced the mana, Label = the ability's printed label and
	// Exhaust = whether it prints the exhaust keyword (#1183). S15
	// sub-PR 2.
	//
	// CardID joined Source in #1210 so the two activation kinds carry
	// the SAME stamps: a watcher that looks up the ability's source
	// object (effects.WheneverAnOpponentActivates' `of` predicate)
	// must not have to know which kind it is holding, and reading a
	// uuid.Nil CardID matched nothing at all — a silent miss rather
	// than an error. Both emit sites set it: the hand click and the
	// AUTO-TAPPER's executor, which activates the ability too
	// (CR 605.3 — a mana ability is activated like any other).
	EventManaAbilityActivated EventKind = "mana_ability_activated"

	// EventManaAdded — one mana token landed in a player's pool.
	// Actor = pool owner, Source = the producing permanent (uuid.Nil
	// for non-card sources). Colors carries the one symbol added
	// (#763: it used to carry nothing, which is why nothing could
	// trigger off "mana of a particular color was added"); the
	// wire-side mana pool is still the canonical projection of the
	// pool itself. S15 sub-PR 2.
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
	//
	// #761: Amount is how many mana were spent and Colors the
	// distinct COLOURS among them, in WUBRG order — the facts
	// converge, sunburst and adamant read off the stack item, said
	// out loud in the log as well.
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
	// chapter (CR 714.2b). Target / CardID = the Saga, Actor = its
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

	// EventClassLevel — a Class permanent's level designation became
	// Amount (CR 716.2). Source / CardID / Target = the Class, Actor
	// = its controller, Amount = the new level.
	//
	// It is what "When this Class becomes level N" watches, and it is
	// also the layer-invalidation signal: a level change turns gated
	// statics on, and the layer listener bumps on this kind (ADR
	// 0071 decision 1).
	//
	// A dedicated kind rather than a predicate over
	// EventCounterPlaced, for the reason EventSagaChapter is one and
	// then some: a level is NOT a counter (CR 716.2b), so there is no
	// counter event to predicate over in the first place.
	//
	// Added in S46 (#757).
	EventClassLevel EventKind = "class_level"

	// EventCaseSolved — a Case became solved (CR 719.3). Source /
	// CardID / Target = the Case, Actor = its controller.
	//
	// Emitted once: a solved Case stays solved while it is on the
	// battlefield, and SolveCaseForEffect is idempotent, so nothing
	// watching this fires twice. Bumps the layer version for the same
	// reason EventClassLevel does.
	//
	// Added in S46 (#757).
	EventCaseSolved EventKind = "case_solved"

	// EventHarnessed — a permanent became harnessed (CR 701.64).
	// Source / CardID / Target = the permanent, Actor = its
	// controller.
	//
	// Emitted once: CR 701.64b says a harnessed permanent stays
	// harnessed while it is on the battlefield, and HarnessForEffect
	// is idempotent, so nothing watching this fires twice. Bumps the
	// layer version for the reason EventClassLevel / EventCaseSolved
	// do (ADR 0071 amendment, #1321).
	EventHarnessed EventKind = "harnessed"

	// EventTurnBegan — one real (not skipped) turn began. Actor is the
	// active player, Amount is Turn.Seq, and Label is "extra" for an
	// extra turn. Emitted after every per-turn reset and before the new
	// turn's first EventStepBegan. It is an engine boundary event, not a
	// second public-log line beside the step spine (ADR 0059).
	EventTurnBegan EventKind = "turn_began"

	// EventStepBegan — the turn cursor entered a step. Actor is the
	// active player, Step the step (typed), Amount the turn sequence,
	// Round the table-facing rotation, and Label the step name. Emitted
	// from runStepEntryHooksLocked AFTER
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

	// EventBlock — CardID blocks Target. Actor is the blocking
	// creature's controller, Target the attacker it is blocking. One
	// event per (blocker, attacker) pair of the FINAL block
	// declaration: this is "whenever this creature blocks"
	// (CR 509.3a) and "becomes blocked by a creature", and it is the
	// public game log's block entry.
	//
	// Emitted from commitBlockDeclarationLocked — the lock-in, not
	// the click (#830). CR 509.1 makes declaring blockers ONE
	// turn-based action, so the sandbox's per-pair DeclareBlocker
	// verb only stages the pairing; nothing is announced until the
	// declaration is complete, which is the first priority boundary
	// of the declare-blockers step. A blocker re-pointed from one
	// attacker to another before that boundary therefore announces
	// once, against the attacker it ends on, and the attacker it
	// left is never announced as blocked at all.
	//
	// Added in S31 sub-PR 0 for the public game log; moved to the
	// lock-in in #830.
	EventBlock EventKind = "block"

	// EventBecomesBlocked — the attacker named by Source / CardID /
	// Target became a blocked creature (CR 509.1h). Actor is the
	// defending player (the controller of its blockers).
	//
	// ONE event per blocked attacker, however many creatures block
	// it: CR 506.4 says an attacking creature is blocked once, at the
	// moment the declaration is complete, so a double block is one
	// "whenever this creature becomes blocked" and one afflict
	// trigger (CR 702.130). That is the whole reason this kind is
	// separate from EventBlock, which is per BLOCKER — before #830
	// each card that wanted the per-attacker reading deduplicated by
	// walking the event log back to the attacker's EventAttack, and a
	// block that was re-pointed away still counted in that walk.
	//
	// Source / CardID / Target are all the attacker, the way
	// EventBattleDefeated names the battle three ways: Target so a
	// predicate reads "the creature that became blocked" out of the
	// same field EventBlock puts the attacker in, CardID so the
	// generic card-shaped consumers find it, Source because the
	// attacker is what the event is about.
	//
	// Emitted from commitBlockDeclarationLocked, in the same event
	// batch as the EventBlock events of the same declaration, so a
	// OncePerBatch ability sees one occurrence. Added in #830.
	EventBecomesBlocked EventKind = "becomes_blocked"
	// EventBattleDefeated — a battle's last defense counter came off
	// (CR 310.12b). Source / Target / CardID = the battle, Actor = its
	// controller.
	//
	// Emitted from the state-based-action pass IMMEDIATELY BEFORE the
	// CR 704.5v/w move that puts the battle in the graveyard, so a
	// defeated trigger's source is still findable on the battlefield
	// when the harvester walks it. A dies-trigger shape (EventLTB
	// plus the LKI snapshot) would also work and would be lossier:
	// the point of a Siege's defeated trigger is to reach a card that
	// has just left, and the fewer hops between "defeated" and the
	// exile-and-cast that follows, the better. Added in S27.
	EventBattleDefeated EventKind = "battle_defeated"

	// EventTransform — CardID was turned over to its other face on the
	// battlefield (CR 701.27a). Actor is its controller, Amount is the
	// face index it turned TO, and Label is the name of the face it
	// turned FROM — which is the only place that name survives, since
	// the card's own Name is already the new face by the time anything
	// reads the event.
	//
	// NOT a zone change and deliberately not shaped like one (CR
	// 712.18: the permanent doesn't become a new object). Nothing
	// emits EventZoneMove, EventETB or EventLTB alongside it, so a
	// back face's "when this enters" trigger does not fire off a
	// transform — correctly, because nothing entered.
	//
	// Two consumers. layerVersionBump reads it to invalidate the layer
	// engine, which is not optional: a face change is a printed-value
	// change and Effective() serves a warm cache. And it is what makes
	// "whenever this transforms" (CR 701.27e) writable — no new
	// constructor was needed, because the harvester already watches
	// any kind a TriggeredAbility names. Added in S46 (ADR 0079, #343).
	EventTransform EventKind = "transform"

	// EventPhaseOut / EventPhaseIn — CardID phased out or in
	// (CR 702.26). Actor is its controller; Source, on a phase-out, is
	// the card whose effect said so (uuid.Nil for CR 502.1's
	// turn-based action, which has no source).
	//
	// NOT zone changes and deliberately not shaped like ones
	// (CR 702.26d: "Zone-change triggers don't trigger when a
	// permanent phases in or out"). Nothing emits EventZoneMove,
	// EventETB or EventLTB alongside them, which is what keeps an ETB
	// trigger silent on a phase-in and a dies trigger silent on a
	// phase-out.
	//
	// Two consumers, the same two EventTransform has. layerVersionBump
	// reads them to invalidate the layer engine, which is not optional:
	// what is ON the battlefield has just changed, so every "creatures
	// you control get +1/+1" and every AppliesTo has a new answer. And
	// they are what makes "whenever this phases in" writable with no
	// new constructor, because the harvester already watches any kind
	// a TriggeredAbility names — a phased-out permanent's own triggers
	// cannot fire, since the harvester walks the battlefield slice it
	// is no longer in, which is CR 702.26b.
	//
	// Added in S46 (ADR 0084, #1199).
	EventPhaseOut EventKind = "phase_out"
	EventPhaseIn  EventKind = "phase_in"

	// EventTurnedFaceUp — Actor turned the face-down permanent
	// CardID face up (CR 708.6, the CR 116.2g special action).
	// Source is the same card: a permanent turns ITSELF face up, and
	// there is no other object involved.
	//
	// A KIND OF ITS OWN, not a reuse of EventTransform or EventETB,
	// and the distinction is a rules one rather than a tidiness one.
	// Both of those mean "a different object is here now" and "an
	// object arrived"; CR 708.8 is the one transition in the game
	// that explicitly means NEITHER — the permanent does not become a
	// new object, and it does not enter anything. A card reading
	// "whenever this transforms" would fire on every morph under the
	// reuse.
	//
	// Two consumers, the same two EventTransform has. layerVersionBump
	// invalidates on it, because the permanent's printed
	// characteristics have just changed wholesale (the CR 708.2 body
	// for the real card) while it sits still. And it is what makes
	// "when this permanent is turned face up" (CR 708.8) writable,
	// with no new constructor — the harvester already watches any
	// kind a TriggeredAbility names, and by the time this is emitted
	// the permanent is face up and its catalog entry answers again.
	// Added in S43 (ADR 0082, #1194).
	EventTurnedFaceUp EventKind = "turned_face_up"

	// EventTurnedFaceDown — the permanent CardID was turned face down
	// by Source (CR 708.2a). Actor is the permanent's OWN controller,
	// not the effect's: it is the seat whose board just changed, and
	// Ixidron turns a whole table's creatures over in one batch.
	// Source is the object that did it, which is the only half of
	// "why" the event can carry — CR 708.7's other half, whether
	// there is a way back up, is TurnFaceUpOffer's to answer off the
	// card underneath.
	//
	// The twin of EventTurnedFaceUp, and a second kind rather than a
	// direction flag on the first, because the two are not the same
	// transition read backwards. CR 701.27b spells the distinction
	// out for transform — "abilities that trigger when a permanent is
	// turned face down won't trigger when that permanent transforms"
	// — and a card watching one direction must not fire on the other.
	//
	// Two consumers, the same two EventTurnedFaceUp has:
	// layerVersionBump invalidates on it, because the permanent's
	// printed characteristics have just been replaced wholesale by
	// the CR 708.2 body while it sits still; and a trigger on any
	// OTHER permanent can watch it with no new constructor.
	//
	// What it cannot do is carry the turned permanent's OWN "when
	// this is turned face down" trigger, because by the time it is
	// emitted that permanent has no text (CR 708.2a) and the
	// harvester reads a source's abilities through CatalogKey. That
	// is the rule's asymmetry rather than the engine's — no printed
	// card has such an ability — and it is written down in ADR 0082's
	// 2026-09-23 amendment, decision A4.
	// Added in S46 (ADR 0082 amendment, #1209).
	EventTurnedFaceDown EventKind = "turned_face_down"

	// EventRevealCards — Actor showed CardID to the whole table (CR
	// 701.20). Fires once per card, so "reveal the top five cards of
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

	// EventLoopSuspected — the CR 726 loop breaker fired: the
	// ability named by Source + Label has resolved Amount times this
	// turn with no player decision in between, and automatic passing
	// is now suspended for every seat. Actor is the ability's
	// controller — the player CR 726 would have name how many more
	// iterations to run. Emitted once per run, when Game.LoopNotice
	// is raised; the flag, not the event, is what the client reads.
	// Added for #628 (ADR 0055).
	EventLoopSuspected EventKind = "loop_suspected"

	// EventRollDie and EventFlipCoin are public random outcomes. One event is
	// emitted per die/coin; BatchSeq identifies the instruction that made them.
	EventRollDie  EventKind = "roll_die"
	EventFlipCoin EventKind = "flip_coin"

	// EventSettingsChanged — one table setting changed (ADR 0075
	// §2.3). Actor is the player who changed it (uuid.Nil for the
	// server admin), Label is the setting's key (SettingUndoLimit,
	// SettingAllowSpawn, …) and SettingOld / SettingNew are its
	// values before and after, as text, so the log can narrate
	// "set undos from 1 to 3 per turn". One event per field that
	// actually changed. Public. Added in S35 (#1032).
	EventSettingsChanged EventKind = "settings_changed"

	// EventSpawned — somebody put cards or tokens onto the table from
	// nowhere (ADR 0075 §2.4). Actor is who asked (uuid.Nil for the
	// server admin), Target the seat the cards belong to, NewZone
	// where they went, Label the card or token name and Amount how
	// many. ONE event per spawn request; the per-card EventETB and
	// EventTokenCreated events follow it.
	//
	// Public, and the reason the feature is allowed on a live table
	// at all: a spawned Treasure is indistinguishable from a real one
	// on the board, so the log is what tells the table where it came
	// from. The public projection names the zone but NOT the card for
	// a spawn into a hand or a library — see protocol.projectEvent.
	// Added in S35 (#1032).
	EventSpawned EventKind = "spawned"
)

// Event is a single entry in the per-game event log. Tagged union
// shape: Kind selects which of the payload fields are meaningful,
// the rest are zero. Chosen over per-kind interfaces so Clone is
// a simple slice value-copy and wire encoding is flat JSON.
//
// Seq is a monotonic per-Game counter stamped at emit time. First
// event has Seq == 1; zero is the un-stamped sentinel.
type Event struct {
	Seq uint64 `json:"seq"`

	// Batch names the EVENT BATCH this event belongs to: the run of
	// events the engine emitted as one occurrence, which is what
	// "whenever ONE OR MORE …" counts (CR 603.2c). Two events with
	// the same Batch happened at the same time as far as the rules
	// are concerned; two with different Batch values are two separate
	// occurrences, however close together they were and whatever is
	// still on the stack from the first.
	//
	// Stamped at emit time from Game.eventBatch, which advances at
	// exactly one boundary — see beginEventBatchLocked in
	// event_batch.go. Zero is the un-stamped sentinel, as with Seq.
	// Added for #829.
	Batch uint64 `json:"batch,omitempty"`

	Kind EventKind `json:"kind"`

	// BatchSeq groups per-die and per-coin events from one instruction. It is
	// the Seq of that instruction's first event.
	BatchSeq uint64 `json:"batch_seq,omitempty"`

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

	// Round is the table-facing round for EventStepBegan. Amount carries
	// the turn sequence on EventStepBegan and EventTurnBegan; keeping the
	// display value separate lets same-seat and extra turns retain distinct
	// identities.
	Round int `json:"round,omitempty"`

	// LookedAt is the SIZE of a finished keyword action that looks at
	// the top of a library — the "2" in "scry 2" — on EventScry and
	// EventSurveil, and zero on every other kind.
	//
	// A second number rather than a re-purposed Amount, because the
	// two facts are both wanted at once and neither implies the
	// other: Amount is how many cards MOVED (to the bottom, to the
	// graveyard) and a scry 2 that moves one card is not a scry 1.
	// Amount is also what "whenever you scry" payoffs and the
	// surveil tests read, and #1036 was explicit that its meaning
	// must not change.
	//
	// It is the size the action ACTUALLY had, which is two
	// adjustments away from the number printed on the card: the
	// CR 614 keyword-action window may have rewritten the count
	// before the prompt was queued (Crystal Ball's "scry that many
	// plus one" — see keyword_action.go), and a library shorter than
	// the count clamps it (CR 701.22a looks at as many as there are).
	// So it is len(PendingChoice.ScryCards) at the site that answers
	// the prompt, and never the argument ScryForEffect was called
	// with. Added for #1036.
	LookedAt int `json:"looked_at,omitempty"`

	// Label is a free-text qualifier: counter kind
	// ("+1/+1", "loyalty", "poison"), trigger label, error
	// classification. Kept as a string so adding new counter kinds
	// doesn't require a schema change.
	Label string `json:"label,omitempty"`

	// Colors are the distinct colours a payment spent, in WUBRG
	// order, on EventManaSpent (#761). Colourless is not a colour
	// (CR 105.1) and never appears; Amount carries the total number
	// of mana, colourless included.
	//
	// Empty on a payment the engine waived (permissive mode, a
	// ForceCast) and on a genuinely free cast, which the log tells
	// apart by the EventCostWarning that accompanies the first.
	//
	// On EventManaAdded (#763) it is the ONE symbol that token
	// carried, "C" included — one event per token, so there is never
	// more than one. Empty on every other event kind.
	Colors []string `json:"colors,omitempty"`

	// Sides is the die size for EventRollDie. Call and Won describe an
	// EventFlipCoin; Won is false for a face-only flip.
	Sides int    `json:"sides,omitempty"`
	Call  string `json:"call,omitempty"`
	Won   bool   `json:"won,omitempty"`

	// Step is the step that began, on EventStepBegan. Typed so a
	// trigger's predicate compares a constant rather than a string
	// (Label still carries the name for the public log). Added in
	// #588.
	Step Step `json:"step,omitempty"`

	// OldZone / NewZone are the zone kinds for ZoneMove-shaped
	// events. Empty string means "not applicable."
	OldZone ZoneKind `json:"old_zone,omitempty"`
	NewZone ZoneKind `json:"new_zone,omitempty"`

	// Played marks a battlefield entry as a PLAY (CR 305.1) rather
	// than an effect PUTTING the permanent onto the battlefield
	// (CR 305.4: "This isn't the same as 'playing a land' and doesn't
	// count as a land played during the current turn"). Set on
	// EventZoneMove and EventTokenCreated when the entry landed —
	// never on the earlier EventCast/EventTrigger that led to it.
	//
	// It mirrors the settled entry's internal landPlay flag
	// (replacements.go), stamped by announceEntryLocked from the
	// entryLanding record every battlefield entry produces
	// (entry_choice.go) — one finisher, so the marker can never drift
	// from the land-drop tally that flag already gates. False, the
	// zero value, is correct for every "put" path (a search, a
	// reanimation, PutFromHandOntoBattlefieldForEffect, an exile or
	// graveyard return with no play permission) AND for a token,
	// which comes from no zone at all and was never played.
	//
	// Meaningful only on a land's own entry — nothing stops it being
	// read off a nonland permanent, but nothing prints a nonland
	// clause that cares. Added for #1326.
	Played bool `json:"played,omitempty"`

	// DiscardCause is why a discard happened, on EventDiscardCard: an
	// effect's instruction, a cost, or the cleanup step's turn-based
	// action (CR 701.8a, 601.2h, 514.1). Empty on every other kind.
	//
	// It is the distinction the rules draw — ADR 0013 §10a withdrew
	// the voluntary/involuntary framing — and it rides the public event
	// so the log can say why a card was pitched and a payoff that
	// cares reads it beside Source, which names the card that asked.
	// Added with #650.
	DiscardCause DiscardCause `json:"discard_cause,omitempty"`

	// Cause, CauseController and CauseItem say WHAT moved a card, on
	// the events a routed zone change emits (EventZoneMove, EventMill,
	// EventDiscardCard, EventCounterSpell and the EventLTB beside
	// them): a resolving spell or ability and its controller, a cost
	// and its payer, a special action, a rule, or a manual sandbox
	// move. Empty when nothing was recorded. "A spell or ability you
	// control exiles one or more permanents" (Ranar the Ever-Watchful)
	// reads it — see move_cause.go and ExiledBySpellOrAbilityOf.
	// Engine-internal: protocol/log.go does not project it. Added with
	// #1320.
	Cause           MoveCauseKind `json:"cause,omitempty"`
	CauseController uuid.UUID     `json:"cause_controller,omitempty"`
	CauseItem       uuid.UUID     `json:"cause_item,omitempty"`

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

	// Exhaust marks an EventActivateAbility or
	// EventManaAbilityActivated whose ability prints the exhaust
	// keyword — "Activate each exhaust ability only once" (#1181,
	// #1183). False on every other event, and on the ordinary
	// activations that are nearly all of them.
	//
	// A bit on the event rather than a lookup a watcher does for
	// itself, because by the time a watcher runs the ability's source
	// may be gone (a SacrificeSelf cost) and its ability list may have
	// been renumbered or removed; the announcement is the only moment
	// the fact is reliably knowable. Added for #1184.
	Exhaust bool `json:"exhaust,omitempty"`

	// CombatStep names which combat damage step dealt a combat
	// EventDealDamage: CombatStepFirstStrike or CombatStepRegular
	// (CR 510.4 — when a creature has first strike or double strike
	// as combat damage begins, the phase has two combat damage steps).
	//
	// Set ONLY when the first-strike pass ran. A combat with no first
	// strike or double strike anywhere has one step, and its damage
	// stays untagged (""), so the tag's presence alone tells a
	// consumer there is something to sequence. Untagged is also what
	// an event written before the field existed decodes as, which is
	// the same answer. Never set on non-combat damage.
	//
	// Written in exactly one place, emitDealDamageLocked, from the
	// damage tail — so a paused event (a CR 616 ordering prompt, or a
	// CR 510.1c assignment prompt answered later) carries the step it
	// was created in, not the one current when it lands. Added for
	// #187 (ADR 0053 Decision 1).
	CombatStep string `json:"combat_step,omitempty"`

	// SettingOld / SettingNew are a table setting's value before and
	// after an EventSettingsChanged, formatted as text: an int as
	// decimal ("-1" is UndoUnlimited), a bool as "true"/"false", an
	// enum as its string value. Empty on every other kind. Added in
	// S35 (#1032, ADR 0075).
	SettingOld string `json:"setting_old,omitempty"`
	SettingNew string `json:"setting_new,omitempty"`
}

// The two values of Event.CombatStep (and DamageAssignmentFrame's
// CombatStep). Plain strings rather than a named type: they are wire
// values, copied verbatim onto protocol.LogEvent.combat_step.
const (
	// CombatStepFirstStrike is the first combat damage step: creatures
	// with first strike or double strike deal their damage (CR 510.4).
	CombatStepFirstStrike = "first_strike"
	// CombatStepRegular is the second combat damage step, run after a
	// first-strike step: creatures with double strike, and creatures
	// that had neither keyword as the first step began, deal theirs
	// (CR 510.4). The engine reads the keywords as the second pass
	// begins instead — that gap is #716, not this tag's concern.
	CombatStepRegular = "regular"
)

// emitBecameTargetLocked fans one EventBecomesTarget out per target
// slot in `targets`. Called from every site that finishes choosing
// targets for a spell or ability, and there are six: the cast path,
// the catalog activation (activated.go), the manual sandbox
// activation and the manual trigger announce (mutations.go), a
// trigger's CR 603.3d target pick (pending_choice.go) and a copy's
// re-target (spell_copy.go). One helper, so a card watching for
// "becomes the target" cannot see a different board depending on
// which verb announced (#968 was the sandbox activation missing).
//
// `actor` is the controller of the spell or ability, `source` is its
// source card and `itemID` is its stack item. TargetSelf /
// TargetNone slots are skipped — neither names an object anyone else
// chose.
//
// `source` and `itemID` are equal for a cast spell and differ for an
// ability; see Event.StackItemID for why both are carried. Every
// caller names an item: ward reads StackItemID to counter the object
// that targeted, and Source would be the permanent the ability came
// from.
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
	if ev.Kind == EventTokenCreated {
		g.noteCreatedSourceLocked(ev.CardID)
	}
	g.eventSeq++
	ev.Seq = g.eventSeq
	// #829: every event carries the batch that was open when it was
	// emitted. Stamped here rather than derived later so the log, an
	// undo and every consumer see the same grouping the trigger
	// harvester saw.
	ev.Batch = g.currentEventBatchLocked()
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
