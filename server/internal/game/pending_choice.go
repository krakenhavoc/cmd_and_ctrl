package game

import (
	"errors"

	"github.com/google/uuid"
)

// pending_choice.go holds the S14+ generic "someone needs to make
// a pick" infrastructure. Replaces ad-hoc per-effect deferred-
// choice state with a unified queue that the client picker modal
// can drain one choice at a time.
//
// Shape: a choice names its Chooser (who picks), the source card
// that created it, what kind of pick it is, and — when the pick
// comes from somewhere other than the chooser's own zones — a
// FromPlayer reference. Thoughtseize is the canonical chooser !=
// discarder case: the caster is Chooser, the target is FromPlayer.
//
// The existing S13.4 DiscardPending map stays for cleanup-step
// max-hand-size discards and NOTHING else (#651): that one is a
// turn-based action (CR 514.1), chooser == owner, and the cursor
// auto-resumes when the map drains. An EFFECT's discard is part of
// the resolving effect (CR 608.2c) and goes through this queue like
// every other deferred decision — QueueDiscardChoiceForEffect. Two
// obligations, two mechanisms, no shared map. Effect-driven choices
// keep resolving asynchronously — the spell routes to graveyard
// immediately, the pick is made later by a resolve_choice action —
// but the table does not move on while one is open (choice_gate.go).
//
// Introduced in S14 sub-PR 5 as infrastructure for Thoughtseize.

// PendingChoiceKind discriminates what kind of pick is owed.
// Extensible — each kind is interpreted by the server when
// ResolvePendingChoice fires and by the client when rendering the
// picker. Values:
//
//	"discard_from_hand" — chooser picks Count card IDs from
//	                      FromPlayer's hand; those cards move to
//	                      FromPlayer's graveyard.
//
// Future kinds (reserved names, not implemented yet):
//
//	"mill_reveal"       — chooser picks which of the revealed top-N
//	                      library cards go where (Brainstorm's
//	                      put-two-back half).
type PendingChoiceKind string

const (
	PendingChoiceDiscardFromHand PendingChoiceKind = "discard_from_hand"

	// PendingChoiceMana — the chooser picks one color from a fixed
	// option set (Birds of Paradise's WUBRG with the commander's
	// identity listed first, Arcane Signet's commander-identity
	// subset). On resolve the picked color drops
	// into the chooser's pool as one ManaToken sourced from the
	// permanent that fired the ability. Added in S15 sub-PR 2.
	PendingChoiceMana PendingChoiceKind = "mana_pick"

	// PendingChoiceReplacementOrder — CR 616 affected-player-
	// chooses-order prompt queued when ≥2 replacement effects
	// apply to the same event. ReplacementEffectIDs carries the
	// gathered IDs; the client renders a drag-reorder list with
	// label + source-card context. On resolve the chooser submits
	// the order as an `order []string` payload of the same IDs in
	// their chosen order; ResolveReplacementOrder validates the
	// permutation and resumes the pipeline. Added in S17 sub-PR 2.
	PendingChoiceReplacementOrder PendingChoiceKind = "replacement_order"

	// PendingChoiceOptionalReplacement — CR 614.10 yes/no prompt
	// for an Optional replacement effect (today: CR 903.9
	// commander-zone). Owner answers yes → Replace runs; no →
	// effect is marked applied without running, event proceeds.
	// On resolve the chooser submits `{apply: true|false}`;
	// ResolveOptionalReplacement re-enters the pipeline. Added
	// in S17 sub-PR 6.
	PendingChoiceOptionalReplacement PendingChoiceKind = "optional_replacement"

	// PendingChoiceDamageAssignment — CR 510.1c prompt queued
	// when a multi-blocker combat damages step needs the
	// attacker's controller to assign damage across the ordered
	// blocker list. The chooser is the attacker's controller.
	// The client renders a drag-to-reorder blocker list with a
	// damage input per blocker (and, if the attacker has trample,
	// a "damage to defending player" input for overflow).
	//
	// On resolve the chooser submits
	// `{assignments: [{blocker_id, amount}, ...], trample_to_player: N}`.
	// ResolveDamageAssignment validates at-least-lethal prefix,
	// total = attacker power, and trample-only-if-trample. Added
	// in S18 sub-PR 3.
	PendingChoiceDamageAssignment PendingChoiceKind = "damage_assignment"

	// PendingChoiceTriggerPrompt — CR 603.5 "you may" yes/no
	// prompt queued by the S19 trigger harvester when a matching
	// TriggeredAbility has a non-nil OptionalPrompt. The chooser is
	// the source's controller (or an override defined on
	// OptionalPrompt.Chooser). On resolve `{apply: true}` the
	// stashed Build closure runs against the captured event + LKI
	// and the resulting StackItem appends to PendingTriggers; on
	// `{apply: false}` the trigger drops without effect.
	//
	// Modal-choice triggers ("draw a card OR gain 3 life") are NOT
	// covered here — sub-PR 2 ships the optional yes/no path only.
	// Added in S19 sub-PR 2.
	PendingChoiceTriggerPrompt PendingChoiceKind = "trigger_prompt"

	// PendingChoicePayUnless is the "unless that player pays {N}"
	// prompt (Rhystic Study, Smothering Tithe, Esper Sentinel — CR
	// 118.12 / 603.2). Queued by a triggered ability's Effect at
	// resolution; the chooser is the player being asked to pay,
	// not the ability's controller. `{apply: true}` attempts the
	// payment from the chooser's pool, auto-tapping their untapped
	// sources if the pool is short; `{apply: false}` — or a payment
	// that can't be made — runs the "unless" consequence. PayCost
	// carries the printed cost string for the client. Added in S19
	// sub-PR 6.
	PendingChoicePayUnless PendingChoiceKind = "pay_unless"

	// PendingChoiceTriggerOrder — CR 603.3b "you choose the order"
	// prompt for a player with two or more differing triggered
	// abilities waiting to go on the stack at the same time. The
	// APNAP drain holds every pending trigger until each such seat
	// has answered; the chooser submits `order []string` of the
	// trigger IDs in the order they should RESOLVE (first entry
	// resolves first), and the engine places them on the stack in
	// reverse. Identical triggers (same source, same label — two
	// Bident of Thassa draws) never prompt. Added in S19 sub-PR 8.
	PendingChoiceTriggerOrder PendingChoiceKind = "trigger_order"

	// PendingChoicePickTarget — S20 sub-PR 2: a targeted triggered
	// ability asks its controller to choose the target as it goes
	// on the stack (CR 603.3d). PickTargetPlayers / PickTargetCards
	// carry the legal set computed at trigger time; the chooser
	// answers with `{target: {kind, id}}`, ResolvePickTarget
	// validates it against the same TargetSpec, builds the item with
	// the ref stamped on, and drains it onto the stack. The client
	// renders this as the ordinary board-click targeting flow, not
	// a modal.
	PendingChoicePickTarget PendingChoiceKind = "pick_target"

	// PendingChoiceSacrifice — "each player sacrifices a creature"
	// (Grave Pact, Dictate of Erebos, Fleshbag Marauder — CR
	// 701.21a). One choice per affected player, addressed to that
	// player, carrying the permanents they may choose from.
	//
	// Deliberately NOT PendingChoicePickTarget. The effect does not
	// target, and the difference is observable: a creature with
	// hexproof, shroud, protection or "can't be the target of
	// spells or abilities" can still be sacrificed to Grave Pact,
	// and Grave Pact needs no legal target to resolve. Reusing the
	// targeting prompt would inherit all of those restrictions
	// silently.
	//
	// The choice is also not optional. A player with legal
	// permanents must pick one; only a player with none is skipped,
	// and they are skipped at queue time rather than prompted and
	// allowed to decline.
	PendingChoiceSacrifice PendingChoiceKind = "sacrifice_choice"

	// PendingChoiceScry — "look at the top N cards of your library.
	// Put any number of them on the bottom of your library and the
	// rest on top in any order." (CR 701.22)
	//
	// Private, not revealed: only the chooser becomes a knower of the
	// looked-at cards, so the wire redacts them for everyone else.
	// Getting that backwards would leak the top of a library to the
	// whole table, which is a real information advantage rather than
	// a cosmetic bug.
	//
	// Answered with {bottom: []string, top_order: []string}. Every
	// looked-at card must appear in exactly one of the two lists —
	// scry moves all of them, it does not leave any in place — and
	// top_order is top-first, so its first entry is the next card
	// drawn.
	PendingChoiceScry PendingChoiceKind = "scry"

	// PendingChoiceSurveil — "look at the top N cards of your
	// library. Put any number of them into your graveyard and the
	// rest on top of your library in any order." (CR 701.25)
	//
	// Structurally scry with the bottom-of-library leg replaced by
	// the graveyard, and it shares scry's plumbing: the looked-at
	// cards ride in ScryCards, the continuation in scryResume, and
	// the wire projects them through PendingChoiceView.Options with
	// the same chooser-only redaction. Surveil is "look at", not
	// "reveal".
	//
	// It is a SEPARATE KIND rather than a flag on the scry prompt
	// because the destination is the whole difference. A card put on
	// the bottom of a library is not a card in a graveyard: a
	// reanimator deck surveils to fill its graveyard, and a client
	// that rendered "to bottom" would be offering the wrong move.
	// The distinction is also observable downstream — EventSurveil
	// is not EventScry, and the cards that land in the graveyard
	// fire ordinary zone-change triggers.
	//
	// Answered with {graveyard: []string, top_order: []string}.
	// Every looked-at card must appear in exactly one of the two
	// lists, and top_order is top-first. Added in S22.
	PendingChoiceSurveil PendingChoiceKind = "surveil"

	// PendingChoiceLookAtTop — "look at the top N cards of your
	// library, then put them back in any order." (Ponder, Sensei's
	// Divining Top, Soothsaying.)
	//
	// The third member of the scry / surveil family and the
	// degenerate one: same look-at, same privacy, but the cards have
	// nowhere to go except back on top. There is no second
	// destination, so the answer is a permutation rather than a
	// partition.
	//
	// It is still a real choice and still a real prompt. "Put them
	// back in any order" with no way to express the order is the
	// same card as "do nothing", which is what makes Sensei's Top a
	// card rather than a blank.
	//
	// Answered with {top_order: []string} alone, top-first, naming
	// every looked-at card exactly once. Sending a `bottom` or
	// `graveyard` key with it is a client bug — those are the other
	// two keywords, and this one has no away lane. Added in S22.
	PendingChoiceLookAtTop PendingChoiceKind = "look_at_top"

	// PendingChoiceSearchLibrary — "search your library for ..."
	// (CR 701.23). The searcher picks which of the matching cards
	// they take; picking none is always legal ("you may fail to
	// find", CR 701.23b), so this prompt has a minimum of zero and a
	// maximum of the effect's limit.
	//
	// LOOK AT, not reveal — and the distinction is the whole reason
	// the prompt is safe to render at all. The candidates are marked
	// known to the CHOOSER alone, and the wire withholds the option
	// list from every other seat, so opponents learn neither which
	// cards matched nor how many did. Only the cards actually taken
	// by a "reveal those cards" effect become public, and the
	// post-search shuffle wipes the chooser's own positional
	// knowledge again.
	//
	// The chooser is always the library's owner — Assassin's Trophy
	// makes the VICTIM search their own library, not the caster —
	// so no information crosses the table.
	//
	// Answered with {card_ids: []string}; an empty or absent list is
	// "fail to find".
	PendingChoiceSearchLibrary PendingChoiceKind = "search_library"

	// PendingChoiceMayCast — "you may cast it without paying its
	// mana cost", asked while an ability is RESOLVING and about one
	// specific card it has just put somewhere the chooser can reach
	// (CR 702.85's cascade is the card that forced it).
	//
	// Deliberately not PendingChoicePayUnless with a "{0}" cost,
	// which would work mechanically: that prompt's whole client-side
	// vocabulary is "Pay {2}" / "Don't pay", and a dialog that reads
	// "Pay {0}?" about a free spell is the kind of copy that makes a
	// player answer the wrong question. The offered card rides the
	// choice (MayCastCard) so the prompt can show the card rather
	// than name it.
	//
	// Answered with the same {apply: bool} payload the other four
	// yes/no kinds use; `true` takes the offer. The engine grants a
	// free-cast permission rather than casting inline — see
	// cascade.go for why, and for the delayed cleanup that keeps the
	// grant from outliving the offer.
	PendingChoiceMayCast PendingChoiceKind = "may_cast"

	// PendingChoiceCoinCall asks a flipper to call heads or tails for a
	// won/lost flip. Stop is available only when CoinAllowStop is true.
	PendingChoiceCoinCall PendingChoiceKind = "coin_call"
)

// PendingChoice is one outstanding "someone needs to pick" entry
// in the game's queue. Serialized to the wire via
// protocol.PendingChoiceView with the Options slice materialized
// at view time.
type PendingChoice struct {
	// ID uniquely identifies the choice for address-by-ID in the
	// resolve_choice action. Minted fresh on queue.
	ID uuid.UUID

	// Kind determines how the picks[] payload is interpreted on
	// resolve_choice.
	Kind PendingChoiceKind

	// Chooser is the player responsible for submitting picks. Only
	// they see the picker modal.
	Chooser uuid.UUID

	// FromPlayer owns the pool the picks come from. For
	// discard_from_hand, this is the player whose hand the picks
	// live in. Equals Chooser for self-discard (Mind Rot-style)
	// choices; differs for Thoughtseize (caster picks from
	// target's hand).
	FromPlayer uuid.UUID

	// Count is the number of items the chooser must pick. 1 for
	// Thoughtseize; 2 for Mind Rot if we ever migrate it here.
	Count int

	// Source is the card that queued this choice (Thoughtseize's
	// instance ID). Empty uuid when the choice isn't card-
	// originated (e.g. admin-driven test harness).
	Source uuid.UUID

	// Reason is a free-text label for the picker modal's header
	// ("Thoughtseize", "Vendilion Clique"). Kept on the server so
	// the wire carries it; no localisation yet.
	Reason string

	// Coin-call prompt data. These are data (not the continuation), so the
	// protocol can render the exact choices and bot hint.
	CoinAllowStop     bool
	CoinCount         int
	CoinMaxUsefulWins int
	CoinWins          int

	// ColorOptions is the legal-picks list for PendingChoiceMana.
	// Uppercase single-character entries (W/U/B/R/G/C). Unused for
	// other choice kinds. Populated server-side so the client
	// renders a color picker with exactly the right buttons, in
	// order (commander-identity-filtered for Arcane Signet, the full
	// five for Birds of Paradise with the commander's identity listed
	// first — see manaPickOptionsFor). Added in S15 sub-PR 2.
	ColorOptions []string

	// ColorPurpose is what the card asking a PendingChoiceColor will do
	// with the answer — #780's card-side declaration. Unused for every
	// other kind. It is carried rather than derived because the only
	// thing that could derive it is the card's own text, and the
	// engine does not read card text. Projected onto the wire as
	// `color_purpose`; see ColorPurpose.
	ColorPurpose ColorPurpose

	// ManaRestrictions are the spend restrictions the token minted
	// by this PendingChoiceMana will carry — Delighted Halfling's
	// "spend this mana only to cast a legendary spell". Empty for
	// ordinary colour picks, which is nearly all of them.
	//
	// It lives on the choice rather than being re-derived at resolve
	// time because by then the ability is gone: ResolveManaChoice
	// sees a colour string and a source ID, and the source may have
	// been sacrificed as part of the activation cost. Losing the
	// restrictions here would put unrestricted mana in the pool,
	// which is the whole #259 failure this seam exists to avoid.
	//
	// Deep-copied by clone.go alongside ColorOptions. Added in the
	// S32 mana-pipeline pass (#352).
	ManaRestrictions []string

	// ManaSourceKinds is what the permanent producing this mana WAS
	// at the moment the pick was queued — snow, Treasure, creature,
	// land, artifact, enchantment (#1212, mana_source.go).
	//
	// Here for exactly the reason ManaRestrictions is, and the case
	// is sharper: a Treasure's mana ability sacrifices the Treasure,
	// so by the time the colour pick is answered the source is gone
	// and a Treasure TOKEN has ceased to exist (CR 111.7). The one
	// source that eighteen printed cards ask about is the one source
	// a resolve-time lookup could never answer, so the fact travels
	// on the choice.
	//
	// A plain value, so the clone and the snapshot carry it with the
	// rest of the struct.
	ManaSourceKinds ManaSourceKinds

	// ManaAmounts is how many mana a PendingChoiceMana adds for each
	// colour in ColorOptions — "{T}: Add three mana of any one color"
	// (Gilded Lotus) is ONE pick minting three tokens, and Nyx Lotus's
	// amount differs per colour (its devotion to that colour). nil, or
	// a colour missing from the map, means one: every ordinary pick.
	// Parsed from the produced-mana grammar's "{W3|U3}" form (see
	// ParseProducedMana). Deep-copied by clone.go and carried by the
	// snapshot. Added for #742.
	ManaAmounts map[string]int

	// ManaTapped marks a PendingChoiceMana that is part of TAPPING A
	// PERMANENT FOR MANA (CR 106.12a) — a mana ability with a {T} cost
	// whose colour the controller is still choosing.
	//
	// It is the one bit that tells ResolveManaChoice whether answering
	// this pick fires the CR 605.1b triggered mana abilities (#763,
	// ADR 0074 §3). For a Birds-of-Paradise-style source this is the
	// only moment the produced colour is known, so it is the only
	// place Mana Flare can read it. A pick queued by a resolving spell
	// (AddManaForEffect — Dark Ritual, Sanctum of Fruitful Harvest) or
	// by a triggered mana ability's own "any color" leaves it false:
	// neither of those tapped anything, and firing on them would break
	// CR 106.12a in one direction and recurse in the other.
	//
	// Carried by the value copy every clone starts from, and by the
	// snapshot.
	ManaTapped bool

	// ReplacementEffectIDs is the ordered set of applicable
	// replacement-effect IDs the chooser must reorder for a
	// PendingChoiceReplacementOrder entry. The resolve_choice
	// payload returns the same IDs in the chosen order. Unused
	// for other choice kinds. Added in S17 sub-PR 2.
	ReplacementEffectIDs []ReplacementEffectID

	// replacementResume is the server-only continuation frame for
	// a PendingChoiceReplacementOrder entry: the in-flight
	// ReplacementEvent + the gathered applicable list. Not
	// serialised to the wire. Consumed by ResolveReplacementOrder
	// on submit. Added in S17 sub-PR 2.
	replacementResume *replacementResumeFrame

	// DamageAssignment is the client-facing payload for a
	// PendingChoiceDamageAssignment entry: the attacker's instance
	// ID, ordered blocker instance IDs (in declared order; the
	// client may re-order), attacker's effective power, and
	// trample-allowed flag. Wire-serialised via
	// PendingChoiceView.DamageAssignment. Added in S18 sub-PR 3.
	DamageAssignment *DamageAssignmentFrame

	// NoLegalTarget marks a PendingChoiceTriggerPrompt whose effect
	// has no legal target / will pass without effect if the chooser
	// answers "Yes" (e.g. Reclamation Sage with no opponent artifact,
	// Eternal Witness with an empty graveyard). Computed at queue
	// time from the ability's HasLegalTarget predicate; serialised to
	// the wire so the client can warn the chooser. False for prompts
	// without a HasLegalTarget predicate. Added in S19 follow-up.
	NoLegalTarget bool

	// PickTargetPlayers / PickTargetCards are the legal-target set a
	// PendingChoicePickTarget offers, computed when the trigger fired.
	// Wire-serialised (as ID lists) so the client can highlight the
	// same surfaces the cast picker does. Added in S20 sub-PR 2.
	PickTargetPlayers []uuid.UUID
	PickTargetCards   []uuid.UUID
	// PickTargetMin / PickTargetMax are the clause's count (S20
	// sub-PR 5) so the client knows whether one click answers the
	// prompt or picks accumulate.
	PickTargetMin, PickTargetMax int

	// RetargetItem / RetargetPolicy / RetargetOptional /
	// RetargetSlot / RetargetReason are the whole of a
	// PendingChoiceRetarget (#1196, CR 115.7): which stack item is
	// being redirected, under which printed sentence, whether
	// declining is allowed, which of the item's target slots this
	// prompt is asking about, and the card's own header text for the
	// rest of the walk.
	//
	// DATA rather than a resume frame, and that is the whole point:
	// the object being rewritten is already on the stack, so the
	// answer needs no closure and the snapshot can carry the prompt
	// intact. The legal set rides the PickTarget* fields above, since
	// the QUESTION is the same one those were built for.
	RetargetItem     uuid.UUID
	RetargetPolicy   RetargetPolicy
	RetargetOptional bool
	RetargetSlot     int
	RetargetReason   string

	// pickTargetResume is the server-only continuation for a
	// PendingChoicePickTarget: the captured event / source / LKI,
	// the Build closure, and the spec the pick is validated against
	// (and stamped onto the item for the resolution re-check).
	// Added in S20 sub-PR 2.
	pickTargetResume *pickTargetFrame

	// ModeOptionIndex / ModeOptionLabel are the options a
	// PendingChoiceModePick offers, in printed order: the ModeSpec
	// index of each option and its oracle bullet. Only the choosable
	// ones are listed (an option whose clause has no legal target is
	// dropped, CR 603.3d), so the answer is validated against this
	// list rather than against the whole ModeSpec — and the CLIENT
	// cannot offer what the engine would refuse. Carried by the
	// snapshot: the options ARE the prompt. Added by #764.
	ModeOptionIndex []int
	ModeOptionLabel []string

	// ModeMin / ModeMax / ModeRepeatable are the ModeSpec's bounds,
	// repeated onto the prompt so the picker and the bot's
	// enumerator need nothing but the prompt to answer it.
	// Added by #764.
	ModeMin, ModeMax int
	ModeRepeatable   bool

	// modePickResume is the server-only continuation for a
	// PendingChoiceModePick: the captured trigger, so the answer
	// continues into the CR 603.3d target walk and then Build.
	// Added by #764.
	modePickResume *modePickFrame

	// copySpellResume is the other continuation a
	// PendingChoicePickTarget can carry (S30, #95): the CR 707.10
	// "you may choose new targets for the copy" prompt. It reuses
	// the pick_target prompt rather than getting a kind of its own
	// because the QUESTION is identical — here is a target clause,
	// here is its legal set, pick within Min..Max — and every
	// consumer (the wire projection, the client picker, the bot's
	// choice enumerator) then needs no change at all to answer it.
	// What differs is only what gets built on submit, which is what
	// the frame decides. Exactly one of pickTargetResume and
	// copySpellResume is set.
	copySpellResume *copySpellFrame

	// SacrificeOptions is the set of permanents a
	// PendingChoiceSacrifice's chooser may pick from — their own
	// permanents matching the effect's spec, computed when the
	// effect resolved. Wire-serialised so the client can highlight
	// exactly those cards. Re-checked on submit, since the board can
	// change between the prompt and the answer (an earlier player's
	// sacrifice can trigger something that removes a creature).
	SacrificeOptions []uuid.UUID

	// promptRun links this prompt to the RUN it is one leg of — the
	// printed instruction whose continuation waits for every seat it
	// asked (#1019, #1027, prompt_run.go). uuid.Nil on a prompt
	// nothing is waiting on, which is every sacrifice and every
	// discard the fire-and-forget entry points queue.
	//
	// TWO KINDS carry one: PendingChoiceSacrifice (one leg of a
	// prompted sacrifice, sacrifice_run.go) and the
	// PendingChoiceChooseCards a discard prompt is
	// (discard_run.go). One field rather than one per verb, because
	// the runs are one registry and "this prompt is a leg of that
	// run" is one fact; defaultDroppedChoiceLocked reads the FIELD
	// rather than the kind for the same reason, so a third verb needs
	// no third branch.
	//
	// A plain id rather than a pointer to the frame, and that is the
	// undo contract rather than a style choice: the prompts of one run
	// SHARE mutable state, and cloneLocked copies each PendingChoice by
	// value — a shared pointer would be duplicated per prompt and an
	// undo snapshot would hold as many half-finished runs as the run
	// had prompts. The runs themselves live on the Game
	// (Game.promptRuns) and are deep-copied once, so the counter and
	// the queue rewind together. Not serialised, like every other
	// continuation link.
	promptRun uuid.UUID

	// CopyOptions is the set of permanents a PendingChoiceCopyTarget
	// may be copied from — "any creature on the battlefield" for
	// Clone, "a creature or planeswalker you control" for Spark
	// Double. Wire-serialised so the picker renders card faces;
	// battlefield cards are public, so there is nothing to redact.
	// Re-checked on submit against the live board, because a
	// creature can leave between the prompt and the answer.
	// Added in S16.5 (#159).
	CopyOptions []uuid.UUID

	// ScryCards is the set of cards a PendingChoiceScry's chooser is
	// looking at, in top-to-bottom library order (so the first entry
	// is the card that would be drawn next if nothing moves).
	// Wire-serialised via PendingChoiceView.Options, redacted to the
	// chooser alone.
	// ScryCards is the set of cards a PendingChoiceScry — or a
	// PendingChoiceSurveil, which shares this plumbing — is asking
	// its chooser to look at, in top-to-bottom library order (so the
	// first entry is the card that would be drawn next if nothing
	// moves). Wire-serialised via PendingChoiceView.Options,
	// redacted to the chooser alone.
	ScryCards []uuid.UUID

	// TriggerOrderIDs is the set of pending-trigger item IDs a
	// PendingChoiceTriggerOrder entry asks the chooser to order.
	// Wire-serialised (with label + source per ID) so the client
	// renders the reorder list; the resolve payload returns the same
	// IDs in resolution order. Added in S19 sub-PR 8.
	TriggerOrderIDs []uuid.UUID

	// triggerResume is the server-only continuation frame for a
	// PendingChoiceTriggerPrompt entry: the captured event +
	// source-card value copy + LKI characteristics + the Build
	// closure to invoke on `apply: true`. Not serialised to the
	// wire. Consumed by ResolveTriggerPrompt on submit. Added in
	// S19 sub-PR 2.
	triggerResume *triggerResumeFrame

	// PayCost is the printed cost string ("{2}", "{1}", "{X}"
	// already substituted) for a PendingChoicePayUnless entry.
	// Wire-serialised so the client can label the "Pay" button.
	// Added in S19 sub-PR 6. Also carries the life payment ("2
	// life") for a PendingChoiceEntryPayLife entry — the label is
	// the only thing the client needs, and the authoritative number
	// stays on the effect's EntryLifeCost.
	PayCost string

	// payUnlessResume is the server-only continuation for a
	// PendingChoicePayUnless entry: the parsed cost plus the
	// "unless" consequence to run when the chooser declines or
	// can't pay. Not serialised. Added in S19 sub-PR 6.
	payUnlessResume *payUnlessFrame

	// SearchCards is the set of library cards a
	// PendingChoiceSearchLibrary's chooser may take, in library
	// order. Wire-serialised via PendingChoiceView.Options — and,
	// unlike every other Options-bearing kind, withheld from
	// non-choosers entirely, because the length of this list is
	// itself hidden information about a hidden zone.
	SearchCards []uuid.UUID

	// SearchMax is how many of SearchCards the chooser may take.
	// The minimum is always zero: CR 701.23b lets a player fail to
	// find however hard they looked.
	SearchMax int

	// searchResume is the server-only continuation for a
	// PendingChoiceSearchLibrary: the whole SearchLibrarySpec, which
	// carries the destination, the reveal / shuffle / tapped flags
	// and the rest of the effect (Fabled Passage's "then untap that
	// land", Gamble's random discard). Not serialised to the wire.
	searchResume *searchResumeFrame

	// MayCastCard is the one card a PendingChoiceMayCast is offering
	// — the cascade hit. Wire-serialised as the choice's single
	// Options entry so the client's prompt can show the card face
	// instead of quoting a name into a sentence.
	MayCastCard uuid.UUID

	// mayCastResume is the server-only continuation for a
	// PendingChoiceMayCast: what to do on each answer. Not
	// serialised. Added in S28.
	mayCastResume *mayCastFrame

	// LoopShortcutKey / LoopShortcutCount / LoopShortcutRepeat carry a
	// PendingChoiceLoopShortcut's question (#804, CR 726).
	//
	// Key is the TallyKey(source, label) the answer's allowance
	// attaches to — the ability that is repeating, named the way the
	// tally names everything. Count is how many times it had resolved
	// when the notice went up, which is the N in "has resolved N times
	// this turn". Repeat marks the SECOND ask of the turn for the same
	// loop: a shortcut ran to its end and the question came back, which
	// is what tells `internal/legal` to offer a bot seat nothing but
	// "stop" and so guarantees a bot-only table terminates.
	//
	// All three are plain data, carried by the clone and the snapshot:
	// the prompt has to mean the same thing after an undo, and the key
	// is the only way back to the run the answer is about.
	LoopShortcutKey    string
	LoopShortcutCount  int
	LoopShortcutRepeat bool

	// AcceptLabel / DeclineLabel are the two branch names a
	// PendingChoiceConfirm renders on its buttons — the card's own
	// words ("Pay 4 life" / "Put it on top"), because a chained
	// question is usually a choice between two things rather than a
	// yes/no. Empty renders as Yes / No. Wire-serialised. See
	// chained_choice.go.
	AcceptLabel, DeclineLabel string

	// LifeCost is the life a PendingChoiceConfirm's ACCEPT branch
	// charges — Sylvan Library's 4. Zero means the accept branch
	// costs no life, which is most of them.
	//
	// It is on the choice, and on the wire, for the reason #547 put
	// legal.MoveCost.Life there: a policy holding only the wire
	// payload otherwise prices "pay 4 life to keep this card" exactly
	// like "shuffle your library", and a bot at 4 life answers yes and
	// dies. The engine does NOT deduct it — the frame's accept branch
	// does, and re-checks affordability when it runs, because life can
	// move between the question and the answer.
	LifeCost int

	// confirmResume is the server-only continuation pair for a
	// PendingChoiceConfirm: one closure per branch. Not serialised.
	// See chained_choice.go.
	confirmResume  *confirmFrame
	coinFlipResume *coinFlipFrame

	// OwedInStep is the step THIS prompt has to be answered in: the
	// upkeep of "at the beginning of your upkeep, sacrifice this
	// unless you pay" (Stasis, Pact of Negation, every cumulative
	// upkeep). The zero value names no step, which is every other
	// prompt in the engine.
	//
	// While the cursor is still standing in that step the prompt
	// blocks the table, whatever its kind's classification says
	// (Game.ChoicePromptBlocksTable, choice_gate.go): CR 117.3 does
	// not let priority pass on with a required action outstanding and
	// CR 500.4 does not let the step end until it has been taken. Set
	// only by QueueUpkeepPayUnlessForEffect (upkeep_pay_unless.go),
	// which is the one door into this shape.
	//
	// It replaced a per-card "please block" boolean in #997, for the
	// reason #951 gives for deriving the other narrowing: the boolean
	// shipped with cumulative upkeep in #567, Stasis and Pact of
	// Negation were written afterwards with the same sentence printed
	// on them, and neither author found the switch.
	OwedInStep TurnStep

	// GuardsStackItem is the object on the stack whose fate THIS prompt
	// decides: the "that spell" of CR 118.12's "counter that spell
	// unless its controller pays {N}" — ward, Daze, Mana Leak.
	//
	// While that object is still on the stack the prompt blocks the
	// table, whatever its kind's classification says
	// (Game.ChoicePromptBlocksTable, choice_gate.go). It is the other
	// half of OwedInStep's rule and the same kind of fact: both are
	// facts about the GAME, re-read every time the gate is asked,
	// rather than a halt declared once when the prompt was queued.
	// That is the whole point — once the guarded object has left the
	// stack there is nothing left to counter, the question is
	// background again, and the halt lifts itself rather than needing
	// somebody to notice.
	//
	// Set only by QueueCounterUnlessPaidForEffect
	// (counter_unless_paid.go), which is the one door into this shape.
	// See #951.
	GuardsStackItem uuid.UUID

	// ChooseCards is the candidate set of a PendingChoiceChooseCards,
	// in the order the client should render them. Wire-serialised via
	// PendingChoiceView.Options and redacted per viewer like every
	// other Options-bearing kind — the candidates are frequently cards
	// in a hand.
	ChooseCards []uuid.UUID

	// ChooseMin / ChooseMax bound a PendingChoiceChooseCards pick.
	// Both are on the choice rather than derived at resolve time so
	// the legal-move enumerator can offer exactly the sets the
	// resolver will accept — the #544 lesson: an enumerator that
	// cannot see a constraint offers answers the engine refuses, and a
	// bot that keeps picking one holds the table forever.
	ChooseMin, ChooseMax int

	// chooseCardsResume is the server-only continuation for a
	// PendingChoiceChooseCards: what the picks mean, plus the zone
	// they are re-checked against. Not serialised. See
	// chained_choice.go.
	chooseCardsResume *chooseCardsFrame

	// PickOptions are the branches of a PendingChoiceOptionPick — the
	// card's own words for each, plus the cards it is about (a pile,
	// or nothing). Wire-serialised via PendingChoiceView.PickOptions,
	// with each option's card list projected through the same
	// per-viewer redaction as every other Options-bearing kind. See
	// option_pick.go.
	PickOptions []ChoiceOption

	// optionPickResume is the server-only continuation for a
	// PendingChoiceOptionPick: what the chosen index means. Not
	// serialised. See option_pick.go.
	optionPickResume *optionPickFrame

	// chooseColorResume is the continuation for a resolution-time
	// PendingChoiceColor (Wash Out's "return all permanents of the
	// color of your choice"). nil for the stored form, whose answer is
	// written onto the source permanent instead. Not serialised. See
	// color_choice.go.
	chooseColorResume *chooseColorFrame

	// scryResume is the continuation for a PendingChoiceScry: the
	// rest of the effect, which must not run until the player has
	// finished the scry. Preordain's "Scry 2, THEN draw a card" is the
	// case that forces it — the card you leave on top is the card you
	// draw, so a draw that happens before the reorder is a different
	// card. Not serialised to the wire.
	// scryResume is the continuation for a PendingChoiceScry or a
	// PendingChoiceSurveil: the rest of the effect, which must not
	// run until the player has finished the look-at. Preordain's
	// "Scry 2, THEN draw a card" is the case that forces it — the
	// card you leave on top is the card you draw, so a draw that
	// happens before the reorder is a different card. Not
	// serialised to the wire.
	scryResume func(g *Game) error
}

// payUnlessFrame carries a pay-unless prompt's parsed cost and the
// consequence of each answer. Both callbacks receive the live *Game
// (not a captured one) for the same undo-safety reason
// StackItem.Effect does. Added in S19 sub-PR 6.
//
// The two callbacks are the two shapes the same prompt serves:
//
//   - onDecline is CR 118.12 "unless that player pays {N}" — Rhystic
//     Study, Esper Sentinel. The effect happens when the payment does
//     NOT.
//   - onPay is "you may pay {N}. If you do, ..." — Hashaton. The
//     effect happens when the payment does.
//
// A card sets one or the other; nothing stops it setting both, and
// the prompt is identical either way, which is why they share a
// frame rather than getting a second PendingChoice kind.
type payUnlessFrame struct {
	cost      ParsedCost
	onDecline func(g *Game) error
	onPay     func(g *Game) error
}

// mayCastFrame carries the two branches of a PendingChoiceMayCast.
// Both receive the live *Game rather than a captured one, on the same
// undo-safety contract payUnlessFrame and StackItem.Effect follow.
//
// onDecline is not optional in practice: cascade's "no" branch still
// has to bottom the pile, and an offer whose refusal does nothing at
// all would not need a prompt. It is nil-checked anyway, because a
// future caller with a genuinely inert "no" shouldn't have to write
// an empty closure.
type mayCastFrame struct {
	onAccept  func(g *Game) error
	onDecline func(g *Game) error
}

// pickTargetFrame is the continuation for a targeted trigger's
// pick_target prompt. Added in S20 sub-PR 2.
//
// #764: a trigger's targets are a WALK over the announcement's
// clause steps, not a single clause. One prompt is one step; the
// frame accumulates the picks and re-queues itself for the next
// step until the list is exhausted, and only then does Build run.
// A single-clause trigger — every targeted trigger that predates
// #764 — is that walk with one step, which is one prompt, which is
// exactly what it was.
type pickTargetFrame struct {
	ev     Event
	source Card
	lki    Characteristic
	build  func(ev Event, source *Card, sourceLKI Characteristic, g *Game) *StackItem
	// spec is the ability's card-level clause statement (nil for a
	// modal ability, whose clauses come from the chosen options).
	spec *TargetSpec
	// modeSpec / modes are the ability's modes and the occurrences
	// chosen at CR 603.3c, stamped onto the built item so the
	// resolution re-check can find each ref's clause.
	modeSpec *ModeSpec
	modes    []int
	// steps is the announcement's clause list, step the cursor into
	// it, and picked the refs answered so far (each already stamped
	// with its Mode / Slot).
	steps  []AnnouncedClause
	step   int
	picked []TargetRef

	doubledBy doublerRef
}

// currentClause is the clause the open prompt is asking about, or
// nil when the walk is finished.
func (f *pickTargetFrame) currentClause() *TargetClause {
	if f == nil || f.step < 0 || f.step >= len(f.steps) {
		return nil
	}
	return &f.steps[f.step].Clause
}

// triggerResumeFrame stashes the per-trigger continuation data the
// harvester captured at OptionalPrompt-queue time. The Build closure
// fires on `apply: true` against the value-copy source + LKI; on
// `apply: false` the frame is discarded. Added in S19 sub-PR 2.
type triggerResumeFrame struct {
	ev     Event
	source Card
	lki    Characteristic
	build  func(ev Event, source *Card, sourceLKI Characteristic, g *Game) *StackItem
	// ability is the full declaration so a "yes" on a TARGETED
	// optional trigger can continue into the pick_target step
	// (S20 sub-PR 2) instead of building straight away.
	ability   TriggeredAbility
	doubledBy doublerRef
}

// TriggerDoubler returns the doubler attribution carried by a harvested
// trigger's optional or target prompt. Ordinary prompts return zero values.
func (c *PendingChoice) TriggerDoubler() (uuid.UUID, string) {
	if c == nil {
		return uuid.Nil, ""
	}
	if c.triggerResume != nil {
		return c.triggerResume.doubledBy.id, c.triggerResume.doubledBy.name
	}
	if c.pickTargetResume != nil {
		return c.pickTargetResume.doubledBy.id, c.pickTargetResume.doubledBy.name
	}
	return uuid.Nil, ""
}

// DamageAssignmentFrame is the payload for a
// PendingChoiceDamageAssignment pending-choice entry. Both
// client-facing (serialised to the wire projection) and
// server-private (used by ResolveDamageAssignment to apply damage
// after validation). Added in S18 sub-PR 3.
type DamageAssignmentFrame struct {
	// AttackerID is the combat-damage source (the attacker).
	AttackerID uuid.UUID
	// BlockerIDs lists the creatures blocking this attacker, in
	// their declared order. The client re-orders via drag; the
	// server validates the re-ordered permutation on submit.
	BlockerIDs []uuid.UUID
	// AttackerPower is the effective power of the attacker at the
	// time of prompt queue. Total assigned damage must equal this
	// value; with trample, the leftover spills to
	// trample_to_player.
	AttackerPower int
	// AllowTrample is true when the attacker has the trample
	// keyword. The client renders a "to player" input only when
	// set; the server accepts trample_to_player > 0 only when set.
	AllowTrample bool
	// HasDeathtouch is true when the attacker has deathtouch (CR
	// 702.2c — 1 damage is lethal). The server uses this to relax
	// the at-least-lethal prefix rule: 1 damage satisfies the
	// threshold regardless of the blocker's remaining toughness.
	HasDeathtouch bool
	// FirstStrike is true when the assignment prompt was queued
	// from the first-strike substep. The resume path needs this
	// to avoid re-routing damage through the regular substep
	// hook.
	FirstStrike bool

	// CombatStep is the Event.CombatStep value the pass that queued
	// this prompt stamps on its damage: CombatStepFirstStrike,
	// CombatStepRegular, or "" when no first-strike pass ran (#187,
	// ADR 0053 Decision 1). The resume paths tag the attacker's
	// assigned damage from it, through damageTailFromFrame.
	//
	// Not derivable from FirstStrike: a regular-pass prompt has
	// FirstStrike false in a combat that had a first-strike pass
	// ("regular") and in one that did not (""). Zero value "" means
	// untagged, which is also what a prompt restored from a snapshot
	// written before this field existed resumes as, so no snapshot
	// schema bump. Server-side only: not projected onto
	// DamageAssignmentView.
	CombatStep string `json:",omitempty"`

	// SourceLifelink is the cached lifelink state of the attacker
	// at prompt-queue time. Captured here because the attacker may
	// have been destroyed by blocker damage (which resolves in the
	// same substep before the prompt fires) — looking it up again
	// at resume time would miss the keyword.
	SourceLifelink bool

	// SourceController is the attacker's controller at prompt-queue
	// time. Captured for the same reason as SourceLifelink (lifelink
	// credits this player even if the attacker is no longer on the
	// battlefield).
	SourceController uuid.UUID

	// SourceLKI is the attacker's characteristics at prompt-queue
	// time, for the same died-before-resume reason as SourceLifelink:
	// CR 702.16e prevents damage from a source with the quality, and
	// the quality has to be read off the attacker as it was when it
	// assigned, not off a card in a graveyard. Rides onto the damage
	// event through damageTailFromFrame. #662.
	SourceLKI *Characteristic `json:",omitempty"`

	// SourceIsCommander is the attacker's commander flag at
	// prompt-queue time, cached for the same died-before-resume
	// reason as SourceLifelink. The trample-to-player resume path
	// uses it to accrue CR 903.10a commander damage on the defending
	// player, keyed by AttackerID above.
	//
	// (S25 (#77) dropped the companion SourceOwner field: since
	// CommanderDamage is keyed by commander instance ID rather than
	// by owning player, AttackerID is already the key and a cached
	// owner had no remaining reader.)
	SourceIsCommander bool
}

// replacementResumeFrame is the unexported per-prompt continuation
// stash. Holds the ReplacementEvent being processed + the gathered
// list so ResolveReplacementOrder can re-enter the apply-loop with
// the chosen order locked in. Added in S17 sub-PR 2.
//
// For PendingChoiceOptionalReplacement (sub-PR 6), `applicable`
// carries a single entry — the optional effect the prompt is
// asking about. The resume path either fires that effect's
// Replace (on yes) or skips it (on no), then re-enters the apply-
// loop for CR 616.1 iteration.
type replacementResumeFrame struct {
	ev         *ReplacementEvent
	applicable []activeReplacement
}

// QueueChoiceForEffect appends a PendingChoice to the game's queue.
// Caller must hold g.mu. Returns the generated ID so the caller
// can reference the choice downstream if needed.
//
// #864: a choice whose Chooser is not seated, or is seated but
// already Eliminated, is refused rather than queued. This is the
// generic backstop underneath the per-kind guards that already
// existed (QueuePayUnlessForEffect and friends) — every append to
// g.PendingChoices runs through here, so this is the one place that
// can promise the queue never holds an unanswerable prompt at the
// moment it's created. It does not by itself protect against a LIVE
// chooser who is eliminated later while their choice sits open —
// that's sweepEliminatedChoicesLocked's job (mutations.go), run at
// every runStateChecksLocked pass.
//
// Returns uuid.Nil on refusal, the same sentinel
// QueueDiscardChoiceForEffect already returns for its own
// zero-count short-circuit, so every caller in this codebase already
// treats "no choice, nothing to reference" as an ignorable return —
// audited caller by caller for #864. The one caller that needed more
// than "ignore the zero value" (drainPendingTriggersAPNAPLocked,
// which would otherwise treat the refusal as a still-open CR 603.3b
// ordering prompt and hold the whole APNAP drain forever) is fixed at
// its own call site to stop asking before it gets here.
//
// A dropped choice emits EventPendingChoiceDropped rather than
// failing silently, so a stalled table's event log shows why a seat
// never got prompted, and rather than EventEffectError because
// nothing failed — CR 800.4a means there was never anyone left to
// ask.
func (g *Game) QueueChoiceForEffect(choice PendingChoice) uuid.UUID {
	if p := g.playerByIDLocked(choice.Chooser); p == nil || p.Eliminated {
		g.EmitEvent(Event{
			Kind:   EventPendingChoiceDropped,
			Actor:  choice.Chooser,
			Source: choice.Source,
			Label:  string(choice.Kind),
		})
		return uuid.Nil
	}
	if choice.ID == uuid.Nil {
		choice.ID = uuid.New()
	}
	g.PendingChoices = append(g.PendingChoices, &choice)
	return choice.ID
}

// ResolvePendingChoice processes a resolve_choice action.
// Validates:
//   - the choice ID exists in the queue
//   - the chooserID matches the queue entry's Chooser
//   - len(picks) == entry.Count
//   - every pick is in the expected source zone (for
//     discard_from_hand, entry.FromPlayer's hand)
//
// On success, dequeues the entry and applies the kind-specific side
// effect (for discard_from_hand: hand the picks to discardCardsLocked,
// the one discard path — see discard.go).
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolvePendingChoice(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
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
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if len(picks) != choice.Count {
		return ErrInvalidParam
	}
	switch choice.Kind {
	case PendingChoiceMana:
		// Mana picks are resolved by ResolveManaChoice — the payload
		// is a color string, not a list of card IDs. Callers who
		// route here with card-ID picks hit a guard rail.
		return ErrInvalidParam
	case PendingChoiceDiscardFromHand:
		from := g.playerByIDLocked(choice.FromPlayer)
		if from == nil {
			// FromPlayer left the game mid-choice. Drop the entry
			// silently — the choice is moot.
			g.dequeueChoiceLocked(idx)
			return nil
		}
		// Pre-validate every pick sits in from.Hand.
		for _, id := range picks {
			if !from.Hand.Contains(id) {
				return ErrCardNotFound
			}
		}
		// Dequeued BEFORE the discard rather than after it: the
		// discard runs through the one discard path (discard.go),
		// which can queue prompts of its own and drop stale ones, and
		// an index into g.PendingChoices does not survive that.
		g.dequeueChoiceLocked(idx)
		return g.discardCardsLocked(from.ID, picks, discardOptions{
			cause:  DiscardCauseEffect,
			source: choice.Source,
		})
	default:
		return ErrInvalidParam
	}
}

// ResolveManaChoice processes a resolve_choice action for a
// PendingChoiceMana entry. The chooser picks one color from the
// entry's ColorOptions; the picked color drops into their pool as
// one ManaToken sourced from the permanent that fired the ability
// (choice.Source). Emits EventManaAdded.
//
// Distinct from ResolvePendingChoice because the payload shape is a
// color string, not a list of card IDs. The dispatcher decides
// which to call based on the action's payload.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveManaChoice(choiceID, chooserID uuid.UUID, color string) error {
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
	if choice.Kind != PendingChoiceMana {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	// Validate color is in the pre-materialised option set.
	matched := false
	for _, o := range choice.ColorOptions {
		if o == color {
			matched = true
			break
		}
	}
	if !matched {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(chooserID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// #742: "N mana of any one color" mints the picked colour's
	// amount; an ordinary pick has no entry and mints one.
	n := 1
	if v, ok := choice.ManaAmounts[color]; ok {
		n = v
	}
	tapped := choice.ManaTapped
	// #1222: through the one production body, which opens the
	// CR 106.12b window on the amount. This is the ONLY place a
	// Birds-of-Paradise-style source's colour is known, so it is the
	// only place Mana Reflection can double it — the same reason
	// ADR 0074 §3 fires the triggered mana abilities from here. The
	// choice carried the ability's restrictions so the minted tokens
	// get them (#352); copyRestrictions because the choice is about to
	// be dequeued and the tokens outlive it.
	colors := g.produceManaLocked(
		p, choice.Source,
		repeatColor(color, n),
		copyRestrictions(choice.ManaRestrictions),
		// #1212: the snapshot the CHOICE carried, not a live lookup —
		// the source may have been sacrificed to pay for the ability
		// whose colour is being answered here.
		choice.ManaSourceKinds,
		tapped,
		nil,
	)
	source := choice.Source
	g.dequeueChoiceLocked(idx)
	// #763, CR 605.1b / 605.4a: the second of the three production
	// sites, and the ONLY one that knows the colour a
	// Birds-of-Paradise-style source produced — which is what "one
	// mana of any type that land produced" needs (Mana Flare,
	// Mirari's Wake) and what Wild Growth on a dual land waits for.
	//
	// After the dequeue, so a trigger that queues a pick of its own
	// (Fertile Ground's "any color") does not land behind the answered
	// one in the queue. Fires only for a pick that came from TAPPING a
	// permanent for mana: see PendingChoice.ManaTapped.
	if tapped {
		if card := g.findCardByIDLocked(source); card != nil {
			g.fireManaTriggersLocked(ManaProduced{
				Source:     *card,
				Controller: chooserID,
				Colors:     colors,
			}, nil)
		}
	}
	return nil
}

// dequeueChoiceLocked drops the choice at index idx because its
// chooser ANSWERED it, preserving slice order for the rest. Every
// Resolve* path ends here.
//
// #628: answering a prompt is a player decision, so it restarts the
// CR 726 loop run. The engine's own prune paths call
// dropChoiceLocked instead — a choice the engine withdrew is not a
// decision anybody made, and counting it as one would let a loop
// that queues and prunes a prompt each iteration run forever.
//
// Caller must hold g.mu.
// The drop comes FIRST. notePlayerDecisionLocked withdraws the CR 726
// shortcut prompt (#804) as part of clearing the loop notice, which
// rewrites g.PendingChoices — and an index into that slice taken
// before it ran would then name the wrong entry.
func (g *Game) dequeueChoiceLocked(idx int) {
	if idx < 0 || idx >= len(g.PendingChoices) {
		return
	}
	g.removeChoiceAtLocked(idx)
	g.notePlayerDecisionLocked()
}

// dropChoiceLocked is the engine WITHDRAWING a prompt nobody answered
// — a sacrifice choice whose card has left, a stale zone-change
// prompt, a CR 726 shortcut whose notice was cleared, an option pick
// with no legal answer left. No player decision is recorded (that is
// dequeueChoiceLocked's job, above).
//
// THE WITHDRAWAL IS NOT ALWAYS THE END OF THE QUESTION (#1006). The
// departure table's second column says what a dropped prompt of this
// kind still has to do, and it is performed here, so every withdrawal
// path settles a kind the same way and the next prune cannot forget
// it: `option_pick` runs its continuation with the no-choice outcome,
// because the effect that asked is paused mid-resolution and its
// continuation is the rest of the card. Every other kind's action is
// the zero value and this costs a map lookup.
//
// The action runs AFTER the queue has been rewritten, so a
// continuation that queues the next link of a chain does not land
// behind the question it is replacing. A caller sweeping the queue by
// index should walk it BACKWARDS, as every existing one does.
//
// The departure sweep does NOT come through here — it rebuilds the
// slice in one pass and runs the same actions itself, behind the CR
// 800.4f/g gates that only a departure needs
// (dropChoicesForPlayerLocked, mutations.go). There is no path that
// reaches both.
//
// Caller must hold g.mu.
func (g *Game) dropChoiceLocked(idx int) {
	if idx < 0 || idx >= len(g.PendingChoices) {
		return
	}
	c := g.PendingChoices[idx]
	g.removeChoiceAtLocked(idx)
	g.runChoiceDropActionLocked(c)
}

// removeChoiceAtLocked takes the choice at idx out of the queue and
// does nothing else, preserving slice order for the rest. The two
// exits from the queue — an answer (dequeueChoiceLocked) and a
// withdrawal (dropChoiceLocked) — differ in what they do around this,
// and share the one line that does it. Caller must hold g.mu.
func (g *Game) removeChoiceAtLocked(idx int) {
	if idx < 0 || idx >= len(g.PendingChoices) {
		return
	}
	g.PendingChoices = append(g.PendingChoices[:idx], g.PendingChoices[idx+1:]...)
	if len(g.PendingChoices) == 0 {
		g.PendingChoices = nil
	}
}

// runChoiceDropActionLocked performs the departure table's second
// column for one prompt that has just left the queue unanswered —
// THE one place those actions are performed (choiceDepartureDecisions,
// leave_game.go).
//
// Two callers, and the split between them is deliberate: this says
// WHAT a dropped prompt of a kind still owes, and each caller says
// WHEN a prompt is dropped and what it has to check first. The
// departure sweep checks CR 800.4f/g's gates
// (departedChoiceActionAllowedLocked); an ordinary withdrawal checks
// nothing, because the chooser and the card are both still in the game
// and only the question has gone.
//
// Caller must hold g.mu, and must have taken the prompt out of the
// queue already.
func (g *Game) runChoiceDropActionLocked(c *PendingChoice) {
	if c == nil {
		return
	}
	switch choiceDepartureDecisions[c.Kind].onDrop {
	case dropDecline:
		g.declineDepartedChoiceLocked(c)
	case dropDefault:
		g.defaultDroppedChoiceLocked(c)
	}
}

// reassignChoiceLocked hands an open prompt to a new chooser — the
// performing half of CR 800.4g/h. The POLICY (which kinds move, and who
// inherits) is one file away, in leave_game.go with the rest of
// CR 800.4; this is the only code that rewrites a PendingChoice.Chooser
// after the choice has been queued.
//
// Reports whether the prompt moved. False means there is nothing left
// to ask and the caller drops it exactly as it did before #902.
//
// What it does, and deliberately no more:
//
//   - Prunes the candidates that leave the game with the old chooser.
//     CR 800.4a takes every object they OWN out of the game in the same
//     breath as this runs, so a card of theirs on the prompt is already
//     a card that will not be there when the answer arrives (the same
//     staleness #701 taught the zone-change resume to recognise, one
//     step earlier). A departed player is not offered as a target
//     either. If that empties the question, the prompt is not moved —
//     handing a seat a prompt with no answers is the #544 wedge.
//   - Rewrites Chooser, which IS the re-redaction: protocol's
//     redactChoiceCards decides what a prompt's card lists show from
//     PendingChoiceView.Chooser and .FromPlayer at VIEW time, on every
//     snapshot, per viewer. So the next broadcast already shows the new
//     chooser exactly what that seat may see of a pool that is not
//     theirs, and shows the departed seat nothing. There is no stored
//     redaction to re-run.
//   - Emits EventPendingChoiceReassigned, the twin of #868's
//     EventPendingChoiceDropped, so a stall dump says where the prompt
//     went instead of going quiet.
//
// What it does NOT touch is the CONTINUATION. Every resume frame on the
// choice — the trigger's Build closure, the pile's `then`, the target
// spec — is the rest of the CARD, and the card did not change because
// the player answering it did. The new chooser makes the choice; the
// effect still does what it printed, to whoever it printed it about.
// The bot enumerator needs nothing: choiceMoves keys on Chooser
// (internal/legal/choices.go), so the inheriting seat is offered the
// prompt's answers on its next window, and the client renders it from
// the same field.
//
// Caller must hold g.mu.
func (g *Game) reassignChoiceLocked(c *PendingChoice, newChooser uuid.UUID) bool {
	if c == nil || newChooser == uuid.Nil || newChooser == c.Chooser {
		return false
	}
	if p := g.playerByIDLocked(newChooser); p == nil || p.Eliminated {
		return false
	}
	gone := c.Chooser

	// A candidate is pruned when it is leaving the game with the old
	// chooser (they own it) or is already out of every zone.
	keep := func(ids []uuid.UUID) []uuid.UUID {
		if len(ids) == 0 {
			return nil
		}
		out := make([]uuid.UUID, 0, len(ids))
		for _, id := range ids {
			card := g.findCardByIDLocked(id)
			if card == nil || card.Owner == gone {
				continue
			}
			out = append(out, id)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}

	if isCardSetPickKind(c.Kind) {
		// Every kind carrying the choose-cards payload prunes the same
		// way (#1214): the candidates that leave with the old chooser
		// come off, the bounds move with the list through the one
		// helper the zone prune uses (setCardSetCandidates, #1045), and
		// a question with nothing left on it does not move at all —
		// handing a seat a prompt with no answers is the #544 wedge.
		live := keep(c.ChooseCards)
		if len(live) == 0 {
			return false
		}
		setCardSetCandidates(c, live)
	}
	switch c.Kind {
	case PendingChoicePickTarget:
		players := make([]uuid.UUID, 0, len(c.PickTargetPlayers))
		for _, id := range c.PickTargetPlayers {
			if id == gone {
				continue
			}
			players = append(players, id)
		}
		if len(players) == 0 {
			players = nil
		}
		c.PickTargetPlayers, c.PickTargetCards = players, keep(c.PickTargetCards)
		if len(c.PickTargetPlayers) == 0 && len(c.PickTargetCards) == 0 {
			return false
		}
	case PendingChoiceOptionPick:
		// The branches are labelled sentences and survive on their own;
		// only the piles they name, and the SEATS they name, can go.
		for i := range c.PickOptions {
			c.PickOptions[i].Cards = keep(c.PickOptions[i].Cards)
		}
		// #994 / CR 800.4a: a seat option naming the player who is
		// leaving is not an answer any more — they are not a player.
		// Pruned here as well as in pruneDepartedSeatOptionsLocked
		// because this path runs FIRST, before that sweep, and must
		// not hand the new chooser a list it would then have to
		// re-prune. If that empties the question the prompt does not
		// move, which is the same rule the card arms above keep:
		// handing a seat a prompt with no answers is the #544 wedge.
		if opts, ok := keepSeats(c.PickOptions, func(id uuid.UUID) bool { return id != gone }); ok {
			c.PickOptions = opts
		} else {
			return false
		}
	}

	c.Chooser = newChooser
	g.EmitEvent(Event{
		Kind:   EventPendingChoiceReassigned,
		Actor:  gone,
		Target: newChooser,
		Source: c.Source,
		Label:  string(c.Kind),
	})
	return true
}

// keepSeats filters the SEAT options of an option list, leaving every
// option that is not about a seat exactly where it is.
//
// The second return is "there is still a question here". False means
// every option named a seat and every one of those seats failed the
// test, so the prompt has no legal answer left; the caller decides what
// to do about it, because the two callers do different things (the
// reassignment refuses to move the prompt, the departure sweep drops
// it). It is never false for a list with a non-seat option in it, which
// is why a Fact or Fiction pile split cannot be emptied by this.
//
// `keep` is asked only about options that carry a seat (#994). An
// option with a zero Player is not about a player and is kept without
// being asked.
func keepSeats(in []ChoiceOption, keep func(uuid.UUID) bool) ([]ChoiceOption, bool) {
	dropped := 0
	for _, opt := range in {
		if opt.Player != uuid.Nil && !keep(opt.Player) {
			dropped++
		}
	}
	if dropped == 0 {
		return in, true
	}
	if dropped == len(in) {
		return nil, false
	}
	out := make([]ChoiceOption, 0, len(in)-dropped)
	for _, opt := range in {
		if opt.Player != uuid.Nil && !keep(opt.Player) {
			continue
		}
		out = append(out, opt)
	}
	return out, true
}

// queueOptionalReplacementPromptLocked queues a CR 614.10 yes/no
// prompt for a single optional replacement effect. The chooser is
// the effect's Controller (for CR 903.9 commander-zone: the
// commander's owner). The resume path in ResolveOptionalReplacement
// either fires the Replace (on yes) or marks it applied and skips
// (on no), then re-enters the apply-loop.
//
// Caller must hold g.mu.
func (g *Game) queueOptionalReplacementPromptLocked(ev *ReplacementEvent, chosen activeReplacement) {
	chooser := g.optionalReplacementChooserLocked(ev, chosen)
	reason := chosen.effect.PromptQuestion
	if reason == "" {
		reason = chosen.effect.Label
	}
	choice := PendingChoice{
		Kind:                 PendingChoiceOptionalReplacement,
		Chooser:              chooser,
		Count:                1,
		Reason:               reason,
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	}
	g.QueueChoiceForEffect(choice)
}

// optionalReplacementChooserLocked is who answers a CR 614.10 "may"
// prompt: the effect's Controller when it names one (the commander's
// owner for CR 903.9), else the event's affected player. Caller must
// hold g.mu.
func (g *Game) optionalReplacementChooserLocked(ev *ReplacementEvent, chosen activeReplacement) uuid.UUID {
	var chooser uuid.UUID
	if chosen.effect.Controller != nil {
		chooser = chosen.effect.Controller(ev, g, chosen.source)
	}
	if chooser == uuid.Nil {
		chooser = affectedPlayerForEvent(ev, []activeReplacement{chosen}, g)
	}
	return chooser
}

// offerOptionalReplacementLocked handles an applicable CR 614.10
// "may" replacement. Returns true when the yes/no prompt was queued
// and the caller must stop where it is — errReplacementPending from
// the apply-loop, a bare nil from the chosen-order resume — which is
// the pause-and-bail contract offerCopyChoiceLocked and
// offerEntryLifePaymentLocked already have.
//
// Returns false — the "may" is DECLINED, and marked applied so the
// apply-loop does not gather it a second time (CR 614.10: the
// decision is once per event) — in the two cases where there is no
// question to ask:
//
//   - the chooser has left the game, or cannot be identified. The
//     commander of an eliminated player going to the graveyard rather
//     than the command zone changes nothing for anyone.
//   - the event has nothing to resume it (#359, see
//     optionalReplacementResumableLocked). Pausing a battlefield entry
//     that cannot be finished afterwards strands the card in its old
//     zone; declining is weaker than printed and never stronger, the
//     posture #268 chose for the shockland's payment and #272
//     reaffirmed.
//
// It is a shared helper rather than two inline copies because the
// multi-effect order path had NO copy (#847): applyReplacementsLocked
// asked the question for a "may" that was the only applicable effect,
// and ResolveReplacementOrder fired the same "may" blind when a second
// effect had put it in a CR 616 ordering prompt — answering it in the
// direction that favours it, which is the bug the CopySelector branch
// beside it already existed to avoid.
//
// Caller must hold g.mu.
func (g *Game) offerOptionalReplacementLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	if !g.optionalReplacementResumableLocked(ev) ||
		g.chooserGoneLocked(g.optionalReplacementChooserLocked(ev, chosen)) {
		g.markReplacementAppliedLocked(ev, chosen.id)
		return false
	}
	g.queueOptionalReplacementPromptLocked(ev, chosen)
	return true
}

// ResolveOptionalReplacement processes a resolve_choice action
// for a PendingChoiceOptionalReplacement entry. `apply` is the
// owner's yes/no decision: true → fire the stashed Replace; false
// → mark applied without firing. Either way, re-enters the apply-
// loop so CR 616.1 can pick up any newly-applicable effects.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveOptionalReplacement(choiceID, chooserID uuid.UUID, apply bool) error {
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
	if choice.Kind != PendingChoiceOptionalReplacement {
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
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}
	// Mark applied regardless of yes/no so the apply-loop doesn't
	// re-evaluate this effect again for this event (CR 614.10: the
	// decision is once per event).
	g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
	if apply && chosen.effect.Replace != nil {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}

	// Re-enter the apply-loop for any newly-applicable effects.
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	// #808: the iteration cap lets the event through as-is, exactly as
	// the unpaused entry points treat it. Returning here instead would
	// drop the event AND its caller's continuation on a prompt that is
	// already dequeued.
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	return g.finishReplacementResumeLocked(ev, out)
}

// finishReplacementResumeLocked is the tail the two CR 614 / CR 616
// RESUME paths share: land the settled event, forget its bookkeeping,
// and then run the state checks the paused caller never got to run.
//
// That last line is the whole reason this is a helper (#1156). On the
// unpaused path the caller of the mutation runs them — moveCardByRefLocked
// calls runStateChecksLocked the moment routeCardToZoneLocked comes back
// unpaused, the SBA loop re-enters itself, resolveTopOfStackLocked has its
// own boundary. A prompt splits that in half: the caller returns with
// "nothing has moved, the resume lands the card", and the resume then
// landed the card and returned to the action layer with nobody left to
// look at the board. CR 117.5 puts the state-based-action pass and the
// APNAP trigger drain at the boundary where a player would next receive
// priority, and answering a replacement prompt is exactly such a
// boundary — the same one every other Resolve* entry point in this file
// already honours.
//
// What that cost, before this existed: a commander with an Aura on it
// left the battlefield, the CR 903.9 "put it in the command zone
// instead?" prompt paused the move, and the answer landed the commander
// in the command zone while the Aura sat on the battlefield attached to
// a card that is no longer there — CR 704.5m never re-checked, and the
// client drew the orphan in the enchantments row. Since #539 made the
// CR 903.9 window open on every exit "from anywhere", this is the
// ordinary way a commander leaves, not a corner.
//
// Not run on the two early returns above: errReplacementPending means
// another prompt is open and its own resume owns the boundary, and a
// stale resume frame moved nothing at all.
//
// Caller must hold g.mu.
func (g *Game) finishReplacementResumeLocked(ev, out *ReplacementEvent) error {
	err := func() error {
		defer g.clearReplacementEventLocked(ev.ID)
		return g.finishSettledReplacementLocked(ev, out)
	}()
	if err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}

// queueReplacementOrderPromptLocked queues a CR 616 order-choose
// prompt for ev. The chooser is the affected-player (inferred from
// ev.Kind's target field). ReplacementEffectIDs are emitted in the
// order the engine gathered them; the client renders a drag-
// reorder list and returns the same IDs in the chosen order. The
// resume frame stashes ev + applicable so ResolveReplacementOrder
// can re-enter the apply-loop.
//
// Caller must hold g.mu.
func (g *Game) queueReplacementOrderPromptLocked(ev *ReplacementEvent, applicable []activeReplacement) {
	ids := make([]ReplacementEffectID, 0, len(applicable))
	for _, a := range applicable {
		ids = append(ids, a.id)
	}
	chooser := affectedPlayerForEvent(ev, applicable, g)
	choice := PendingChoice{
		Kind:                 PendingChoiceReplacementOrder,
		Chooser:              chooser,
		Count:                len(ids),
		Reason:               "Order replacement effects",
		ReplacementEffectIDs: ids,
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: applicable,
		},
	}
	g.QueueChoiceForEffect(choice)
}

// affectedPlayerForEvent returns the chooser for a CR 616 prompt —
// the player whose event is being replaced (the affected player).
// Falls back to the first applicable replacement's Controller, then
// to uuid.Nil. Callers are expected to have ≥1 applicable entry.
func affectedPlayerForEvent(ev *ReplacementEvent, applicable []activeReplacement, g *Game) uuid.UUID {
	if ev == nil {
		return uuid.Nil
	}
	switch ev.Kind {
	case RepEventDraw:
		return ev.DrawPlayer
	case RepEventLife:
		return ev.LifePlayer
	case RepEventCounter:
		// ADR 0056 Decision 5. A counter on a PLAYER names that player
		// on the event, and CR 616.1 gives the ordering choice to the
		// affected one — the player the counters are going on, not the
		// controller of whichever replacement happened to be gathered
		// first. Vorinclex halving an opponent's poison beside a
		// doubler is the board where the fallback below would ask the
		// wrong seat.
		if ev.CounterPlayer != uuid.Nil {
			return ev.CounterPlayer
		}
		if card, ok := g.LookupCardForEffect(ev.CounterTarget); ok {
			return card.Controller
		}
	case RepEventDamage:
		if card, ok := g.LookupCardForEffect(ev.DamageTarget); ok {
			return card.Controller
		}
		return ev.DamageTarget
	case RepEventMove:
		if card, ok := g.LookupCardForEffect(ev.CardID); ok {
			return card.Controller
		}
	case RepEventDiscard:
		// #982. A discard is a move out of a hand, but the affected
		// player is not "the controller of the card" the move branch
		// above looks up — a card in a hand has no controller, and
		// CR 701.8a puts it into THAT PLAYER'S graveyard. The
		// discarding player is on the event; read it.
		//
		// Not observable while every discard replacement in the catalog
		// is controller-scoped, because the fallback below then names
		// the same player. Madness (#657) and the Obstinate Baloth
		// family read a discard an OPPONENT caused, and the first of
		// those to share a window with another discard replacement
		// would otherwise ask the wrong player.
		return ev.DiscardPlayer
	case RepEventCreateTokens:
		// #982, and the same rule read off the other event ADR 0061
		// added: "create one or more tokens UNDER YOUR CONTROL", so the
		// affected player is the one the tokens are created under the
		// control of.
		//
		// Primal Vigor is the printed shape that needs it. It is
		// deliberately symmetrical ("if one or more tokens would be
		// created"), so an opponent's Primal Vigor beside your own
		// Anointed Procession on YOUR creation used to put the CR 616
		// ordering prompt to the Primal Vigor player.
		return ev.TokenController
	case RepEventMill:
		// #569. CR 701.13a: the player who mills is the one whose
		// library is being read, and they are the affected player
		// whoever controls the replacements. Bruvac the Grandiloquent
		// and The Water Crystal both replace an OPPONENT's mill, so
		// with the two of them on one battlefield the fallback below
		// would name their controller rather than the opponent CR 616.1
		// gives the choice to. The first printed board where the
		// fallback is not accidentally right.
		return ev.MillPlayer
	case RepEventStepTransition:
		if ev.StepTransitionSeat >= 0 && ev.StepTransitionSeat < len(g.Seats) {
			return g.Seats[ev.StepTransitionSeat].ID
		}
	case RepEventProduceMana:
		// #1222. CR 106.12b: the mana is produced for a player, and
		// that player is the affected one whoever controls the
		// replacements — the same reading the mill arm above takes.
		//
		// The prompt it would order is never actually put to them: a
		// production sets mustSettleNow (CR 605.3a), so the apply-loop
		// applies the gathered order inline. The arm is still the
		// honest answer to "who would be asked", it is what the
		// eliminated-chooser branch reads before the mustSettleNow one,
		// and it is the value a future production replacement that CAN
		// pause would need.
		return ev.ManaPlayer
	case RepEventKeywordAction:
		// #976. The affected player of a keyword action is the player
		// TAKING it — the one who proliferates, the one who scrys —
		// which is Actor. Naming it here rather than falling through
		// to the first gathered effect's controller matters for the
		// same reason it does for a draw: the two are the same player
		// for every printed card today, and CR 616.1 gives the choice
		// to the affected one whoever controls the replacements.
		return ev.Actor
	}
	if len(applicable) > 0 && applicable[0].effect.Controller != nil {
		return applicable[0].effect.Controller(ev, g, applicable[0].source)
	}
	return uuid.Nil
}

// ResolveReplacementOrder processes a resolve_choice action for a
// PendingChoiceReplacementOrder entry. Validates:
//   - the choice ID exists in the queue
//   - the chooserID matches the entry's Chooser
//   - ordered is a permutation of the entry's ReplacementEffectIDs
//
// On success, re-enters the replacement apply-loop with the chosen
// order locked in for the current iteration. Subsequent iterations
// may queue another prompt (the chain unrolls asynchronously, one
// resolve_choice per branch-point). After the apply-loop settles,
// the pipeline function's resume helper re-invokes the underlying
// mutation with the (possibly mutated / canceled) event.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveReplacementOrder(choiceID, chooserID uuid.UUID, ordered []ReplacementEffectID) error {
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
	if choice.Kind != PendingChoiceReplacementOrder {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if len(ordered) != len(choice.ReplacementEffectIDs) {
		return ErrInvalidParam
	}
	// Permutation check: same multiset of IDs.
	want := make(map[ReplacementEffectID]int, len(choice.ReplacementEffectIDs))
	for _, id := range choice.ReplacementEffectIDs {
		want[id]++
	}
	for _, id := range ordered {
		if want[id] <= 0 {
			return ErrInvalidParam
		}
		want[id]--
	}
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil {
		return ErrInvalidParam
	}
	if g.dropStaleReplacementResumeLocked(frame) {
		return nil
	}

	// Apply ALL chosen effects in the submitted order (CR 616: the
	// affected player picks the order once; the engine fires them
	// in that order without re-prompting). Then re-enter the apply-
	// loop so CR 616.1 can pick up any newly-applicable effects
	// (effects that weren't applicable until one of these fired).
	//
	// Earlier drafts fired only ordered[0] and relied on the apply-
	// loop to re-queue a prompt for the remaining effects — that
	// mis-read 616.1 and forced the user to submit the same order
	// N times for N replacements. The right behavior is "one prompt
	// = one ordering decision, apply them all in sequence."
	applicableByID := make(map[ReplacementEffectID]activeReplacement, len(frame.applicable))
	for _, a := range frame.applicable {
		applicableByID[a.id] = a
	}
	ev := frame.ev
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}
	for _, id := range ordered {
		chosen, ok := applicableByID[id]
		if !ok {
			continue
		}
		if ev.Canceled {
			// A prior Cancel short-circuits the remaining chain.
			break
		}
		if !g.stillAppliesLocked(ev, chosen) {
			// CR 616.1f: after each applied effect the process repeats
			// "taking into account only replacement or prevention
			// effects that would now be applicable". An effect an
			// earlier one in the chosen order has switched off — a
			// "lose more than 3" gate after a halving — does not fire
			// on the strength of the gather the prompt was built from
			// (#808), and neither does one whose source stopped
			// applying between the prompt and the answer. It is left
			// unmarked, so the apply-loop re-entry below still picks
			// it up if a later effect switches it back on.
			continue
		}
		if chosen.effect.CopySelector != nil {
			// Same reason as the pay-life branch below: an effect
			// with a CHOICE inside it can't be fired blind. Clone's
			// controller still picks what it copies when a Kismet is
			// also replacing the entry. Firing Replace here instead
			// would mark the selector applied and silently drop the
			// copy — the permanent would enter as a 0/0 and nobody
			// would be asked anything.
			if g.offerCopyChoiceLocked(ev, chosen) {
				return nil
			}
			continue
		}
		if chosen.effect.EntryLifeCost > 0 {
			// An effect with a payment inside it can't be fired
			// blind — the shockland's controller still has to answer
			// "pay 2 life?" even when a Kismet is also replacing this
			// entry. Queue that prompt and bail; the effects later in
			// the chosen order are still unapplied, so the apply-loop
			// re-entry after the answer picks them up.
			if g.offerEntryLifePaymentLocked(ev, chosen) {
				return nil
			}
			continue
		}
		if chosen.effect.Optional {
			// #847: and a CR 614.10 "may" is the third of them. This
			// branch was missing, so a "may" ordered alongside any
			// other effect fired without ever being offered — the
			// engine said yes on its controller's behalf, which is
			// precisely what the two branches above exist to prevent.
			// Same pause-and-bail, same one resume: the chain
			// continues from this effect when
			// ResolveOptionalReplacement re-enters the apply-loop,
			// with the effects later in the chosen order still
			// unapplied and therefore still gatherable.
			if g.offerOptionalReplacementLocked(ev, chosen) {
				return nil
			}
			continue
		}
		g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
		if chosen.effect.Replace != nil {
			if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
				g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
			}
		}
	}

	// Resume the apply-loop to pick up any newly-applicable effects
	// (CR 616.1 — one of the applied replacements may have enabled
	// another that wasn't in the original prompt). Effects already
	// in replacementsAppliedThisEvent are skipped by gather.
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// Another prompt queued; unroll asynchronously.
		return nil
	}
	// #808: the iteration cap lets the event through as-is, as every
	// unpaused entry point does — see ResolveOptionalReplacement.
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}

	// Apply-loop settled. Dispatch the underlying mutation per
	// ev.Kind using the (possibly mutated) event payload. Added in
	// S17 sub-PR 3 so Doubling Season + Hardened Scales actually
	// land counters after the CR 616 prompt resolves. Through the
	// shared resume tail since #1156, so the CR 117.5 boundary this
	// answer is happens here too and not only on the CR 614 half.
	return g.finishReplacementResumeLocked(ev, out)
}

// applyResolvedReplacementEventLocked runs the underlying
// mutation for a fully-settled ReplacementEvent — called from the
// CR 616 resume path (ResolveReplacementOrder) after the
// replacement apply-loop finishes with no pending prompts. The
// event's payload may have been mutated by replacements (e.g.
// Doubling Season doubled CounterDelta; the commander-zone built-in
// rewrote NewZone). Pipeline functions' initial (non-paused) path inlines
// the same mutation; the resume path uses this central dispatcher.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedReplacementEventLocked(ev *ReplacementEvent) error {
	switch ev.Kind {
	case RepEventCounter:
		// ADR 0056 Decision 3. This arm used to be
		// `return g.applyCounterLocked(...)`, and it was the #694 bug
		// shape twice over.
		//
		// It returned the mutation's error, so when the target had left
		// between the prompt and the answer the ACTION failed — after
		// the choice had already been dequeued, which takes the prompt
		// away with nothing to show for it. The life and damage arms
		// below both learned to log and move on; this one had not.
		//
		// And it never swept. Answering a prompt is an action boundary
		// like every other Resolve* handler, and the unpaused counter
		// paths get their sweep from the resolution bookend they run
		// inside. A -1/-1 placement that PAUSED has no bookend to fall
		// back on, so the creature it brought to zero toughness has to
		// die here — which is exactly what ADR 0056's damage results
		// make routine, and why the hardening lands in this PR rather
		// than the one that branches the tail.
		err := g.applyResolvedCounterLocked(ev)
		if errors.Is(err, ErrCardNotFound) || errors.Is(err, ErrPlayerNotFound) || errors.Is(err, ErrPlayerEliminated) {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "counter placement dropped: its target is no longer in the game",
			})
			return nil
		}
		if err != nil {
			return err
		}
		g.runStateChecksLocked()
		return nil
	case RepEventDraw:
		// #1222: with the count the window settled on. CR 121.2 makes
		// those N one individual card draw each, so a paused draw and
		// an unpaused one cannot drift apart.
		return g.actuallyDrawCardsLocked(ev.DrawPlayer, ev.DrawCount)
	case RepEventProduceMana:
		// #1222: unreachable, and that is the decision rather than an
		// oversight. A production sets mustSettleNow (CR 605.3a — a
		// mana ability resolves as one indivisible step with no
		// priority window inside it, and the auto-tapper may raise no
		// prompt at all), so no event of this kind ever queues a
		// prompt and nothing ever resumes one. If one somehow did, the
		// mana it was carrying has already gone into the pool
		// unreplaced (replaceProducedManaLocked's error branch), so
		// adding it here would double it.
		return nil
	case RepEventLife:
		// #482: a life change carries everything its resume needs on
		// the event itself — the player, the settled delta and the
		// source the log credits — so it is finished by exactly the
		// code the unpaused path runs. This branch used to be its own
		// copy of the tail, and it was already one field behind: it
		// emitted EventChangeLife with no Source, so a life change
		// that paused lost the card that caused it. See life_tail.go.
		err := g.applyResolvedLifeChangeLocked(ev)
		if errors.Is(err, ErrPlayerNotFound) || errors.Is(err, ErrPlayerEliminated) {
			// The player left between the prompt and the answer. The
			// life change simply does not happen — but the choice is
			// already dequeued, so returning the error here would
			// fail the action AND take the prompt away with nothing
			// to show for it. Log it and move on.
			//
			// #793 / #808: the rest of the effect has already run,
			// with zero, inside applyResolvedLifeChangeLocked. A drain
			// whose second opponent conceded during the prompt gains
			// what the first one lost, not nothing — and not what the
			// conceded one would have lost either (CR 800.4a).
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "life change dropped: its player is no longer in the game",
			})
			return nil
		}
		if err != nil {
			return err
		}
		// Answering a prompt is an action boundary, like every other
		// Resolve* handler, so the sweep the tail deliberately skips
		// happens here — a life change that put somebody at 0 is a
		// CR 704.5a loss, and the unpaused paths get their sweep from
		// the resolution bookend they run inside.
		g.runStateChecksLocked()
		return nil
	case RepEventDamage:
		// #694: a damage event that came through any entry point
		// carries its own tail, so it is finished by exactly the code
		// the unpaused path runs — player life loss, the CR 120.3
		// planeswalker/battle split, CR 702.2c deathtouch, CR 702.15
		// lifelink and the CR 903.10a commander tally included. This
		// branch used to be a copy of the manual MarkDamage body,
		// which meant ordering two damage replacements marked
		// DamageMarked on everything and returned ErrCardNotFound for
		// a player. See damage_tail.go.
		err := g.applyResolvedDamageLocked(ev)
		if errors.Is(err, ErrCardNotFound) || errors.Is(err, ErrPlayerNotFound) || errors.Is(err, ErrPlayerEliminated) {
			// The target left between the prompt and the answer. The
			// damage simply does not happen — but the choice is
			// already dequeued, so returning the error here would
			// fail the player's action AND take their prompt away
			// with nothing to show for it. Log it and move on.
			//
			// #807: the caller's continuation has already been run
			// with zero by applyResolvedDamageLocked, for the same
			// reason the life side runs its tail here — a batch
			// adding up "the damage dealt this way" must not stall on
			// the opponent who conceded during the prompt.
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "damage dropped: its target is no longer in the game",
			})
			return nil
		}
		if err != nil {
			return err
		}
		// Answering a prompt is an action boundary, like every other
		// Resolve* handler, so the sweep the tail deliberately skips
		// happens here — which is also where it was before #694.
		g.runStateChecksLocked()
		return nil
	case RepEventMove, RepEventDiscard:
		// #529: a move that came through the shared exit primitive
		// carries everything its resume needs on the event itself, so
		// it is finished by the same code the unpaused path runs —
		// from WHATEVER zone the card was in when the window opened,
		// which is what makes CR 903.9's "from anywhere" resumable
		// rather than just askable. Checked first because it is the
		// general case: since #707 it covers every exit in the engine
		// but the destroy / sacrifice / SBA route below, the sandbox
		// move_card verb included. The two branches after it are the
		// older, hand-rolled resumes for the battlefield entry and
		// battlefield-leave paths.
		//
		// The one route that is NOT finished here is the destroy /
		// sacrifice / SBA exit, which keeps its own mover and carries
		// a route only to hold a continuation (#815, ViaBattlefieldLeave).
		// It falls through to the battlefield-leave branch below.
		if ev.zoneRoute != nil && !ev.zoneRoute.ViaBattlefieldLeave {
			return g.executeZoneRouteLocked(ev)
		}
		// S17 sub-PR 6: resume path for battlefield-leave moves
		// after the CR 903.9 commander-zone Optional prompt. The
		// pipeline settled on ev.NewZone (either the original
		// destination if owner said "no", or ZoneCommand if they
		// said "yes"). Run the physical move through the shared
		// executeBattlefieldLeaveLocked helper.
		if ev.entryResumable && ev.NewZone == ZoneBattlefield && ev.OldZone != ZoneBattlefield {
			// A paused ENTRY — the shockland's pay-2-life prompt, a
			// CR 616 ordering prompt between two enters-tapped effects,
			// Clone's "choose what to copy". The pipeline function
			// bailed before moving anything, so the push happens here,
			// through exactly the finisher the unpaused path runs
			// (executeEntryToBattlefieldLocked), and the effect's own
			// post-entry duties run off ev.entryTail with it: a search's
			// shuffle and EventSearchLibrary, the caller's Then, the new
			// object identity an exile return mints (#478).
			_, err := g.executeEntryToBattlefieldLocked(ev)
			if err != nil {
				return err
			}
			// Answering a prompt is an action boundary, like every other
			// Resolve* handler, so the sweep the finisher deliberately
			// skips happens here — the unpaused entry paths get theirs
			// from the resolution bookend they run inside. It runs after
			// the tail so a permanent that entered and then died does so
			// once, with everything else the effect did.
			g.runStateChecksLocked()
			return nil
		}
		if ev.OldZone != ZoneBattlefield {
			// What is left here is a battlefield ENTRY that is not
			// entryResumable: putOntoBattlefieldFromZoneLocked's batch,
			// which runs every card's pipeline against the pre-entry
			// board and then moves them together — a simultaneity a
			// per-card resume would break. It bails before moving
			// anything and is documented as not happening when it
			// pauses; see ReplacementEvent.entryResumable. Since #707 no
			// EXIT lands here: every one of them carries a zoneRoute or
			// comes off the battlefield.
			return nil
		}
		var owner *Player
		if card, ok := g.LookupCardForEffect(ev.CardID); ok {
			owner = g.playerByIDLocked(card.Owner)
		}
		// #815: through the shared finisher, so a destruction that
		// paused on the CR 903.9 prompt runs its caller's continuation
		// from exactly where an unpaused one runs it. A route reaches
		// this branch only when it is flagged ViaBattlefieldLeave —
		// every other exit went to executeZoneRouteLocked above.
		return g.finishBattlefieldLeaveLocked(ev, owner)
	case RepEventCreateTokens:
		// #762: the CR 701.7b window has settled on how many tokens of
		// which kinds this instruction makes. Creating them is the
		// same function the unpaused path runs, so a creation that
		// paused on an ordering prompt and one that did not cannot
		// drift apart — and each token then takes the ordinary
		// battlefield entry, which may pause again on its own.
		return g.applyResolvedTokenCreationLocked(ev)
	case RepEventKeywordAction:
		// #976: the CR 614 window has settled on how many times this
		// proliferate happens, or how many cards this scry looks at.
		// Taking the action is the same function the unpaused path
		// runs, so a keyword action that paused on an ordering prompt
		// and one that did not cannot drift apart.
		_, err := g.applyResolvedKeywordActionLocked(ev)
		if err != nil {
			return err
		}
		// Answering a prompt is an action boundary, like every other
		// Resolve* handler, so the sweep the effect-time helpers
		// deliberately skip happens here — a proliferated tenth poison
		// counter is a CR 704.5c loss, and the unpaused path gets its
		// sweep from the resolution bookend it runs inside.
		g.runStateChecksLocked()
		return nil
	case RepEventMill:
		// #569: the CR 614 window has settled on how many cards this
		// instruction mills. Planning and routing them is the same
		// function the unpaused path runs, so a mill that paused on an
		// ordering prompt and one that did not cannot drift apart — and
		// each card then takes the ordinary per-card exit, which may
		// pause again on CR 903.9.
		if _, err := g.applyResolvedMillLocked(ev); err != nil {
			return err
		}
		// Answering a prompt is an action boundary, like every other
		// Resolve* handler, so the sweep the effect-time helpers
		// deliberately skip happens here. Running a library out is NOT
		// a loss (CR 701.13b, #767), but the cards that landed can be —
		// a milled Aura loses its host and a mill payoff's trigger has
		// to be put on the stack against a swept board.
		g.runStateChecksLocked()
		return nil
	case RepEventStepTransition:
		// #710: finish the step entry the prompt interrupted, with
		// the same code the unpaused path runs. A cancelled event is
		// a skipped step (CR 500.11) — advance the cursor past it;
		// an uncancelled one runs the step's turn-based action.
		//
		// This branch used to return nil for both, while
		// runStepEntryHooksLocked fell through and ran the step
		// anyway, so two skip-step replacements under one controller
		// queued a prompt and then drew (or untapped) regardless.
		if !g.stepEntryStillPendingLocked(ev) {
			return nil
		}
		g.finishStepEntryLocked(ev.Canceled)
		return nil
	}
	return nil
}

// stepEntryStillPendingLocked reports whether the step entry ev
// paused is still the one the turn cursor is sitting on — i.e.
// whether finishing it now finishes the transition that queued the
// prompt rather than stamping an old answer onto a newer step.
//
// Nothing in the engine advances the cursor while a step-transition
// prompt is outstanding (the step never begins, so it grants no
// priority and runs no turn-based action), so this guard should never
// fire. It is here because the consequence of being wrong is the
// cursor jumping a step the table never saw, and because a resume
// whose prompt was answered out of a snapshot restore or a replay is
// exactly the shape #701 found for zone moves. Dropping the resume
// leaves the cursor where it is, which a player can always advance.
//
// Caller must hold g.mu.
func (g *Game) stepEntryStillPendingLocked(ev *ReplacementEvent) bool {
	return ev != nil &&
		g.Turn.Step == ev.StepTransitionStep &&
		g.Turn.ActiveSeat == ev.StepTransitionSeat
}

// finishSettledReplacementLocked is the single "the apply-loop has
// settled — now finish the event" call the CR 616 / CR 614.10 resumes
// make. `ev` is the event the resume has been holding; `out` is what
// applyReplacementsLocked handed back (nil when cancelled).
//
// Cancelling means "the mutation simply does not happen" for every
// event kind but one. A cancelled STEP TRANSITION is not nothing: by
// CR 500.11 it is a SKIPPED step, and skipping is an action — the
// turn cursor has to move past it, which only the step-entry resume
// can do. Returning nil for every cancelled event is what left
// two skip-step replacements queueing a prompt whose answer changed
// nothing (#710).
//
// Caller must hold g.mu.
func (g *Game) finishSettledReplacementLocked(ev, out *ReplacementEvent) error {
	if out != nil && !out.Canceled {
		return g.applyResolvedReplacementEventLocked(out)
	}
	if ev == nil {
		return nil
	}
	// One arm per ReplacementEventKind, and #982's enumeration test
	// (replacement_kind_gate_test.go) fails until a new kind has one.
	// This was an if-chain until then — the same code in the same
	// order, because every condition on it is exclusive on Kind — but a
	// chain is not something a scanner can read a kind out of, and
	// "nothing is owed here" is a decision worth writing down rather
	// than falling off the end of.
	switch ev.Kind {
	case RepEventLife:
		// #793: a cancelled LIFE change is still an answer to whoever
		// asked for it. "You gain life equal to the life lost this way"
		// gains nothing when the loss was replaced away — but a drain
		// adding up several players' losses has to be told so, or it
		// waits on this one forever.
		return g.runLifeTailLocked(ev, 0)
	case RepEventDamage:
		// #807: and the same for a cancelled DAMAGE event. "You gain
		// life equal to the damage dealt this way" gains nothing when a
		// Fog ate the damage, and the batch behind Creeping Bloodsucker
		// has to be told so before it can move on to the next opponent.
		return g.runDamageTailLocked(ev, 0)
	case RepEventCreateTokens:
		// #762: a cancelled or replaced-away token CREATION makes
		// nothing, and the rest of the card ("create a Treasure, then
		// sacrifice it") has to be told so rather than wait for tokens
		// that will never arrive.
		return g.abandonTokenCreationLocked(ev)
	case RepEventKeywordAction:
		// #976: and a cancelled KEYWORD ACTION. "Scry 2, then draw a
		// card" still draws when the scry was replaced away — the
		// sentence after "then" is not conditional on the action having
		// happened — so the continuation has to be told rather than
		// left waiting on a prompt that will never be queued.
		return g.abandonKeywordActionLocked(ev)
	case RepEventMill:
		// #569: and a cancelled MILL, whose caller is often reading the
		// list. "Mill three cards, then return a creature card milled
		// this way to your hand" returns nothing when the mill was
		// replaced away, but it has to be told that rather than wait
		// for cards that will never arrive.
		return g.abandonMillLocked(ev)
	case RepEventMove, RepEventDiscard:
		// #853: a cancelled EXIT that carries a route. The card stays
		// where it is, but a multi-card discard sequenced through the
		// route's continuation has to be told, or the rest of the batch
		// — and the "then draw two" behind it — never happens.
		//
		// #478: and a cancelled ENTRY that carries a tail. A search
		// whose fetched permanent the window cancelled still owes its
		// library a shuffle and its caller a "found nothing"; a leg of
		// a multi-card fetch has to be told before it can start the
		// next one. A move carries a route or an entry tail, never
		// both, so running both is one call and a no-op.
		//
		// #762: a cancelled ENTRY of a created token is the one move
		// that leaves an object behind. A token exists only on the
		// battlefield (CR 111.1), so one whose entry was replaced away
		// simply ceases to be; the rest of its batch still lands,
		// through the entry tail.
		g.dropEnteringTokenLocked(ev.CardID)
		if err := g.runRouteTailLocked(ev.zoneRoute); err != nil {
			return err
		}
		return g.runEntryTailLocked(ev, uuid.Nil)
	case RepEventStepTransition:
		// The one kind whose cancellation is not nothing — see the
		// doc comment above.
		return g.applyResolvedReplacementEventLocked(ev)
	case RepEventDraw, RepEventCounter, RepEventProduceMana:
		// Nothing is sequenced behind any of the three: a cancelled
		// draw, a cancelled counter placement and a production replaced
		// away simply do not happen, and no entry point carries a
		// continuation, so there is nobody to tell. A draw tail or a
		// counter tail, if one is ever added, belongs here.
		//
		// #1222: a production additionally cannot even reach this
		// function — it sets mustSettleNow, so it never pauses and
		// nothing resumes it. The arm is the written answer #982 asks
		// for rather than a fall-through, and it is right either way.
		return nil
	}
	return nil
}

// QueueDiscardFromRevealedHand is the Thoughtseize entry point.
// Reveals the target's hand to the chooser (sticky via S13.5
// KnownBy), then queues a discard_from_hand PendingChoice. The
// spell can return nil from its OnResolve immediately — the
// discard fires asynchronously when the chooser submits their
// pick via resolve_choice.
//
// Caller must hold g.mu.
func (g *Game) QueueDiscardFromRevealedHand(
	chooser, fromPlayer, source uuid.UUID,
	count int,
	reason string,
) uuid.UUID {
	// Reveal to the chooser specifically so their client-side KnownBy
	// lets redactCardForViewer keep the identity. The reveal is
	// sticky — cards the chooser saw stay revealed after the choice
	// resolves, same as any other reveal effect.
	if p := g.playerByIDLocked(fromPlayer); p != nil {
		for i := range p.Hand.Cards {
			p.Hand.Cards[i].AddKnower(chooser)
		}
	}
	// Cap count to available hand size so the chooser isn't stuck
	// on an impossible count (CR 609.3 "as many as you can").
	if p := g.playerByIDLocked(fromPlayer); p != nil && p.Hand.Size() < count {
		count = p.Hand.Size()
	}
	if count <= 0 {
		return uuid.Nil
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:       PendingChoiceDiscardFromHand,
		Chooser:    chooser,
		FromPlayer: fromPlayer,
		Count:      count,
		Source:     source,
		Reason:     reason,
	})
}

// DamageAssignmentEntry is one {blocker, amount} pair from a
// damage-assignment resolve. Added in S18 sub-PR 3.
type DamageAssignmentEntry struct {
	BlockerID uuid.UUID
	Amount    int
}

// ResolveDamageAssignment processes a resolve_choice action for a
// PendingChoiceDamageAssignment entry. Validates:
//   - the choice ID exists in the queue
//   - the chooserID matches the attacker's controller
//   - ordered is a permutation of the blocker IDs from the frame
//   - sum(amounts) + trampleToPlayer == attacker power
//   - each ordered-prefix blocker is assigned at-least-lethal before
//     the next one receives any damage (CR 510.1c). Deathtouch
//     relaxes the threshold to 1.
//   - trampleToPlayer > 0 only when the attacker has trample
//
// On success, applies damage via the regular combat-damage path so
// replacement (Fog), lifelink (mark source's controller), and
// deathtouch (flag target) all fire. Added in S18 sub-PR 3.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveDamageAssignment(
	choiceID, chooserID uuid.UUID,
	ordered []DamageAssignmentEntry,
	trampleToPlayer int,
) error {
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
	if choice.Kind != PendingChoiceDamageAssignment {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.DamageAssignment
	if frame == nil {
		return ErrInvalidParam
	}
	if len(ordered) != len(frame.BlockerIDs) {
		return ErrInvalidParam
	}
	// Permutation check.
	want := make(map[uuid.UUID]int, len(frame.BlockerIDs))
	for _, id := range frame.BlockerIDs {
		want[id]++
	}
	for _, e := range ordered {
		if want[e.BlockerID] <= 0 {
			return ErrInvalidParam
		}
		want[e.BlockerID]--
		if e.Amount < 0 {
			return ErrInvalidParam
		}
	}
	// Trample gate — the client may send 0 even without trample,
	// which is fine; nonzero without trample is rejected.
	if trampleToPlayer < 0 {
		return ErrInvalidParam
	}
	if trampleToPlayer > 0 && !frame.AllowTrample {
		return ErrInvalidParam
	}
	// Sum check.
	total := trampleToPlayer
	for _, e := range ordered {
		total += e.Amount
	}
	if total != frame.AttackerPower {
		return ErrInvalidParam
	}
	// Prefix-lethal check: every blocker before the last non-zero
	// assignment must have received at-least-lethal damage. Lethal
	// threshold = max(1, blocker.CurrentToughness - blocker.DamageMarked).
	// Deathtouch collapses the threshold to 1.
	assigned := make(map[uuid.UUID]int, len(ordered))
	for i, e := range ordered {
		lethal := 1
		if !frame.HasDeathtouch {
			blk := findBattlefieldCard(g, e.BlockerID)
			if blk == nil {
				// Blocker left the battlefield mid-prompt. Treat as
				// lethal satisfied (no target).
				lethal = 0
			} else {
				remaining := blk.CurrentToughness() - blk.DamageMarked
				if remaining > 1 {
					lethal = remaining
				} else if remaining <= 0 {
					// Already lethal-marked; damage still applies but
					// threshold is satisfied.
					lethal = 0
				}
			}
		}
		// Earlier blockers must be at-least-lethal before this one
		// receives any damage (CR 510.1c "assigns damage in order").
		// Only a blocker that actually receives damage constrains its
		// predecessors: with three blockers and two power, [2, 0, 0]
		// is the only legal split, and checking the zero entries'
		// priors rejected it — no assignment could ever be accepted
		// and the table wedged. Found by the S31 bot fuzzer.
		if e.Amount > 0 {
			for j := 0; j < i; j++ {
				prior := ordered[j]
				priorLethal := 1
				if !frame.HasDeathtouch {
					pblk := findBattlefieldCard(g, prior.BlockerID)
					if pblk != nil {
						priorRemaining := pblk.CurrentToughness() - pblk.DamageMarked
						if priorRemaining > 1 {
							priorLethal = priorRemaining
						} else if priorRemaining <= 0 {
							priorLethal = 0
						}
					}
				}
				if prior.Amount < priorLethal {
					return ErrInvalidParam
				}
			}
		}
		// If this blocker got less than lethal AND anything downstream
		// got non-zero damage OR trample spilled, reject.
		if e.Amount < lethal {
			for j := i + 1; j < len(ordered); j++ {
				if ordered[j].Amount > 0 {
					return ErrInvalidParam
				}
			}
			if trampleToPlayer > 0 {
				return ErrInvalidParam
			}
		}
		assigned[e.BlockerID] = e.Amount
	}
	g.dequeueChoiceLocked(idx)

	// Apply damage using the frame-cached source keywords: the
	// attacker may have been destroyed by blocker damage (which
	// resolved in the same substep before this prompt fires), so
	// looking up HasKeyword on the attacker at resume time would
	// miss deathtouch / lifelink. The frame captured those at
	// queue time.
	for _, e := range ordered {
		if e.Amount <= 0 {
			continue
		}
		g.markCombatDamageFromFrameLocked(e.BlockerID, e.Amount, frame)
	}
	if trampleToPlayer > 0 {
		// Determine the defending player: find the attacker if
		// still on battlefield; otherwise look at the prompt's
		// original intent — trample-to-player was only allowed
		// for a live attacker that had declared a target. Fall
		// back to any seat that isn't the controller (best-effort).
		//
		// S27: the attacker may have been attacking a planeswalker or
		// a battle, in which case trample overflow goes to THAT, not
		// to its defending player (CR 702.19b — excess damage is
		// assigned to the player or permanent the creature is
		// attacking). So the target id is used as-is and the routing
		// happens in dealCombatDamageToAttackTargetLocked.
		var defenderID uuid.UUID
		if atkCard := findBattlefieldCard(g, frame.AttackerID); atkCard != nil {
			defenderID = atkCard.AttackingTarget
		}
		if defenderID != uuid.Nil && g.classifyAttackTargetLocked(defenderID) != AttackTargetPlayer {
			g.markCombatDamageFromFrameLocked(defenderID, trampleToPlayer, frame)
			defenderID = uuid.Nil
			trampleToPlayer = 0
		}
		if trampleToPlayer > 0 && defenderID == uuid.Nil {
			for _, s := range g.Seats {
				if s.ID != frame.SourceController {
					defenderID = s.ID
					break
				}
			}
		}
		if defenderID != uuid.Nil {
			g.markCombatDamageToPlayerFromFrameLocked(defenderID, trampleToPlayer, frame)
		}
	}
	// Blockers also deal their power back (simultaneous damage —
	// CR 510.1d). The attacker loop in assignAndDealCombatDamageLocked
	// already marked blocker→attacker damage before queuing the
	// prompt, so we don't re-fire it here.
	//
	// Run SBAs so deaths from this assignment land before the next
	// substep / step advance.
	g.runStateChecksLocked()
	return nil
}

// queueTriggerPromptLocked queues a CR 603.5 yes/no prompt for an
// optional triggered ability that just matched. Captures the event,
// a value copy of the source, and the LKI snapshot — all of which
// the resume path will pass back into the Build closure on `apply:
// true`. The chooser is the source's controller, unless the
// ability's OptionalPrompt overrides it for opponent-prompted
// triggers.
//
// Caller must hold g.mu. Added in S19 sub-PR 2.
func (g *Game) queueTriggerPromptLocked(
	ev Event,
	source Card,
	lki Characteristic,
	ability TriggeredAbility,
	doubledBy doublerRef,
) {
	chooser := source.Controller
	if ability.OptionalPrompt != nil && ability.OptionalPrompt.Chooser != nil {
		if override := ability.OptionalPrompt.Chooser(ev, &source, g); override != uuid.Nil {
			chooser = override
		}
	}
	question := ""
	if ability.OptionalPrompt != nil {
		question = ability.OptionalPrompt.Question
	}
	// Evaluate the legal-target predicate (if any) at queue time so
	// the client can warn the chooser that "Yes" will pass without
	// effect. nil predicate => assume the effect always does
	// something (NoLegalTarget stays false).
	noLegalTarget := false
	if ability.HasLegalTarget != nil {
		noLegalTarget = !ability.HasLegalTarget(ev, &source, lki, g)
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:          PendingChoiceTriggerPrompt,
		Chooser:       chooser,
		Count:         1,
		Source:        source.InstanceID,
		Reason:        question,
		NoLegalTarget: noLegalTarget,
		triggerResume: &triggerResumeFrame{
			ev:        ev,
			source:    source,
			lki:       lki,
			build:     ability.Build,
			ability:   ability,
			doubledBy: doubledBy,
		},
	})
}

// queuePickTargetLocked queues the CR 603.3d target choice for a
// targeted trigger. The legal set is computed now and frozen onto the
// prompt; ResolvePickTarget re-validates the pick against the spec
// anyway, since the board can change while the prompt is open, and
// refreshTargetChoicesLocked re-reads the frozen set at the next
// priority-grant boundary (#809). Caller must hold g.mu. Added in S20
// sub-PR 2.
func (g *Game) queuePickTargetLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef, modes []int, steps []AnnouncedClause) {
	g.queuePickTargetStepLocked(&pickTargetFrame{
		ev:        ev,
		source:    source,
		lki:       lki,
		build:     t.Build,
		spec:      t.Targets,
		modeSpec:  t.Modes,
		modes:     append([]int(nil), modes...),
		steps:     steps,
		doubledBy: doubledBy,
	})
}

// queuePickTargetStepLocked opens the prompt for the frame's CURRENT
// step, skipping steps whose optional clause ("up to one target") has
// nothing to offer, and finishing the walk when the steps run out.
//
// CR 603.3d: a REQUIRED clause with no legal target means the ability
// is removed from the stack and does nothing.
// dispatchTriggerInstanceLocked already checked that at harvest time,
// but the OPTIONAL path comes back through here from
// ResolveTriggerPrompt an arbitrary time later, and a prompt with an
// empty option list is a prompt nobody can answer — which since #791
// is a table that cannot move.
//
// Caller must hold g.mu.
func (g *Game) queuePickTargetStepLocked(f *pickTargetFrame) {
	for f.step < len(f.steps) {
		clause := f.currentClause()
		// The frame's source is a VALUE COPY taken when the trigger
		// was harvested, which is exactly the source CR 702.16b wants:
		// the permanent may have left between the trigger and the
		// prompt, and its qualities then are what count.
		lt := g.legalTargetsLocked(SourceObject(f.source.Controller, &f.source), clause)
		lt = withoutPicked(lt, f.picked, clause.Distinct)
		if len(lt.Players) == 0 && len(lt.Cards) == 0 {
			if clause.Min > 0 {
				// The whole ability is removed (CR 603.3d), together
				// with the picks already made for earlier clauses —
				// nothing was put on the stack, so nothing is undone.
				return
			}
			// "Up to one target" with nothing to point at: the step is
			// answered by choosing nothing.
			f.step++
			continue
		}
		label := clause.Label
		if label == "" {
			label = "Choose a target"
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:              PendingChoicePickTarget,
			Chooser:           f.source.Controller,
			Count:             1,
			Source:            f.source.InstanceID,
			Reason:            label,
			PickTargetPlayers: lt.Players,
			PickTargetCards:   lt.Cards,
			PickTargetMin:     clause.Min,
			PickTargetMax:     clause.Max,
			pickTargetResume:  f,
		})
		return
	}
	g.finishPickTargetLocked(f)
}

// finishPickTargetLocked runs Build with every step answered and
// queues the item, stamping the announcement onto it so the CR
// 608.2b re-check can find each ref's clause.
//
// Caller must hold g.mu.
func (g *Game) finishPickTargetLocked(f *pickTargetFrame) {
	if f.build == nil {
		return
	}
	source := f.source
	item := f.build(f.ev, &source, f.lki, g)
	if item == nil {
		return
	}
	item.Targets = append([]TargetRef(nil), f.picked...)
	item.Modes = append([]int(nil), f.modes...)
	item.targetSpec = f.spec
	item.modeSpec = f.modeSpec
	item.DoubledBy, item.DoubledByName = f.doubledBy.id, f.doubledBy.name
	g.queueHarvestedTriggerLocked(item)
	// CR 603.3d / 115.7: a triggered ability's targets are chosen as
	// it is put on the stack, which is right here. Emitted after the
	// queue so a "becomes the target" trigger stacks above the
	// ability that targeted. Added in S22 for Monk Gyatso.
	g.emitBecameTargetLocked(item.Controller, item.SourceCardID, item.ID, item.Targets)
	g.runStateChecksLocked()
}

// withoutPicked drops from a legal set every object already picked
// for an EARLIER clause, when this clause is Distinct ("a second
// target permanent you control"). A non-Distinct clause gets the set
// untouched — CR 601.2c lets one object fill two different instances
// of the word "target".
func withoutPicked(lt LegalTargets, picked []TargetRef, distinct bool) LegalTargets {
	if !distinct || len(picked) == 0 {
		return lt
	}
	taken := make(map[uuid.UUID]bool, len(picked))
	for _, p := range picked {
		taken[p.ID] = true
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

// ResolvePickTarget processes the controller's target choice for a
// PendingChoicePickTarget. The ref is validated against the spec
// against the CURRENT board (ErrIllegalTarget if it no longer
// qualifies — the chooser's client will re-render with the fresh
// legal set), then Build runs, the ref is stamped onto the item
// along with the spec for the resolution re-check, and the item
// drains onto the stack.
//
// Caller must NOT hold g.mu — this method takes the write lock.
// Added in S20 sub-PR 2.
func (g *Game) ResolvePickTarget(choiceID, chooserID uuid.UUID, target TargetRef) error {
	return g.ResolvePickTargets(choiceID, chooserID, []TargetRef{target})
}

// ResolvePickTargets is the multi-slot form (S20 sub-PR 5): the
// chooser's refs are validated together against the prompt's spec —
// count within Min..Max, each legal, distinct — and stamped on the
// built item in the order given.
func (g *Game) ResolvePickTargets(choiceID, chooserID uuid.UUID, targets []TargetRef) error {
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
	if choice.Kind != PendingChoicePickTarget {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	// S30: the same prompt kind serves the CR 707.10 spell-copy
	// re-target. Handled before the trigger frame because the two
	// are mutually exclusive and the copy path builds something
	// that is not a triggered ability.
	if cf := choice.copySpellResume; cf != nil {
		return g.resolveCopySpellTargetsLocked(idx, cf, targets)
	}
	frame := choice.pickTargetResume
	step := frame.currentClause()
	if frame == nil || frame.build == nil || step == nil {
		g.dequeueChoiceLocked(idx)
		return nil
	}
	for _, t := range targets {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
	}
	// #764: the refs answer THIS step, so they are validated against
	// this clause alone and stamped with its (Mode, Slot) — the walk
	// asks one clause at a time and the item ends up carrying the
	// whole announcement.
	cur := frame.steps[frame.step]
	stamped := make([]TargetRef, 0, len(targets))
	for _, t := range targets {
		t.Mode, t.Slot = cur.Mode, cur.Slot
		stamped = append(stamped, t)
	}
	if err := g.validateAnnouncedTargetsLocked(SourceObject(chooserID, &frame.source), frame.steps[frame.step:frame.step+1], stamped); err != nil {
		return err
	}
	if step.Distinct {
		for _, t := range stamped {
			for _, p := range frame.picked {
				if p.ID == t.ID {
					return ErrInvalidParam
				}
			}
		}
	}
	g.dequeueChoiceLocked(idx)
	frame.picked = append(frame.picked, stamped...)
	frame.step++
	g.queuePickTargetStepLocked(frame)
	return nil
}

// ResolveTriggerPrompt processes the controller's yes/no answer for
// a PendingChoiceTriggerPrompt entry. On `apply: true` the stashed
// Build closure runs against the captured event + LKI; the resulting
// StackItem (if non-nil) appends to PendingTriggers via the same
// queueHarvestedTriggerLocked the mandatory path uses and is drained
// onto the stack right away. On `apply: false` the entry is dropped
// silently — the trigger is treated as having never been declared.
//
// Caller must NOT hold g.mu — this method takes the write lock.
// Added in S19 sub-PR 2.
func (g *Game) ResolveTriggerPrompt(choiceID, chooserID uuid.UUID, apply bool) error {
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
	if choice.Kind != PendingChoiceTriggerPrompt {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.triggerResume
	g.dequeueChoiceLocked(idx)
	if !apply || frame == nil || frame.build == nil {
		return nil
	}
	// S20: a targeted optional trigger continues into the target
	// pick; an untargeted one builds straight away.
	g.buildOrPickTriggerLocked(frame.ev, frame.source, frame.lki, frame.ability, frame.doubledBy, nil)
	// The prompt is answered outside any priority-wrap, so nothing
	// downstream would drain the queue until the next pass around
	// the table — and an empty stack at that wrap would advance the
	// step first, stranding the trigger a step late. Drain now:
	// answering "yes" is the moment the ability is put on the stack
	// (CR 603.3), and SBAs run at the same boundary.
	g.runStateChecksLocked()
	return nil
}

// QueuePayUnlessForEffect queues an "unless that player pays
// <cost>" prompt for `chooser`. Called from a triggered ability's
// Effect at resolution (Rhystic Study: "you may draw a card unless
// that player pays {1}") — the ability has resolved, and what's
// left is the payer's decision. `question` is the dialog header;
// `source` attributes the prompt to the card for the client.
// onDecline runs when the chooser answers "no" OR answers "yes"
// but cannot produce the mana (pool + auto-tap).
//
// A chooser who is no longer seated / is eliminated can't be
// asked; the consequence runs immediately (the "unless" clause
// fails when there is nobody to pay). An unparseable cost is a
// programming error surfaced as EventEffectError, and the
// consequence runs so the card still does something.
//
// Caller must hold g.mu. Added in S19 sub-PR 6.
func (g *Game) QueuePayUnlessForEffect(
	chooser, source uuid.UUID,
	cost, question string,
	onDecline func(g *Game) error,
) error {
	return g.queuePayUnlessLocked(chooser, source, cost, question, onDecline, TurnStep{}, uuid.Nil)
}

// queuePayUnlessLocked is the one body behind every pay-unless.
//
// `owed` is the step the prompt has to be answered in, or the zero
// TurnStep for one the table may walk past (#997,
// upkeep_pay_unless.go); `guards` is the stack object whose fate the
// decline decides, or uuid.Nil for the detached Rhystic shape (#951,
// counter_unless_paid.go). They are separate because they are facts
// about different parts of the board — the cursor and the stack — and
// no caller sets both. Both are read LIVE by the gate, so neither can
// hold the table after the thing it is about has gone.
func (g *Game) queuePayUnlessLocked(
	chooser, source uuid.UUID,
	cost, question string,
	onDecline func(g *Game) error,
	owed TurnStep,
	guards uuid.UUID,
) error {
	parsed, err := ParseCost(cost)
	if err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   source,
			ErrorMsg: "pay-unless: unparseable cost " + cost + ": " + err.Error(),
		})
		if onDecline != nil {
			return onDecline(g)
		}
		return nil
	}
	if p := g.playerByIDLocked(chooser); p == nil || p.Eliminated {
		if onDecline != nil {
			return onDecline(g)
		}
		return nil
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:            PendingChoicePayUnless,
		Chooser:         chooser,
		Count:           1,
		Source:          source,
		Reason:          question,
		PayCost:         cost,
		OwedInStep:      owed,
		GuardsStackItem: guards,
		payUnlessResume: &payUnlessFrame{
			cost:      parsed,
			onDecline: onDecline,
		},
	})
	return nil
}

// QueueMayPayForEffect queues a "you may pay <cost>. If you do,
// ..." prompt for `chooser` — the same dialog QueuePayUnlessForEffect
// raises, with the consequence hanging off the other answer.
// Hashaton, Scarab's Fist ("whenever you discard a creature card,
// you may pay {2}{U}. If you do, create a token that's a copy of
// that card") is the shape it was written for.
//
// onPay runs only when the chooser answers "Pay" AND the mana
// actually materialises (pool + auto-tap). A "yes" the chooser can't
// fund degrades to a decline exactly as it does for pay-unless, and
// with no onDecline that simply means nothing happens — which is the
// right reading of "if you do".
//
// A chooser who is no longer seated / is eliminated is not asked and
// onPay does not run. Note this is the OPPOSITE default from
// pay-unless, and correct for both: the consequence is attached to
// the answer that can no longer be given.
//
// Caller must hold g.mu.
func (g *Game) QueueMayPayForEffect(
	chooser, source uuid.UUID,
	cost, question string,
	onPay func(g *Game) error,
) error {
	parsed, err := ParseCost(cost)
	if err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   source,
			ErrorMsg: "may-pay: unparseable cost " + cost + ": " + err.Error(),
		})
		return nil
	}
	if p := g.playerByIDLocked(chooser); p == nil || p.Eliminated {
		return nil
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoicePayUnless,
		Chooser: chooser,
		Count:   1,
		Source:  source,
		Reason:  question,
		PayCost: cost,
		payUnlessResume: &payUnlessFrame{
			cost:  parsed,
			onPay: onPay,
		},
	})
	return nil
}

// QueueMayCastForEffect queues a "you may cast <card> without paying
// its mana cost" prompt for `chooser`, asked from inside a resolving
// ability (cascade). `source` attributes the prompt to the card whose
// text is asking; `question` is the dialog header.
//
// A chooser who is no longer seated / is eliminated is not asked and
// `onDecline` runs instead — the same default pay-unless takes, and
// right for the same reason: the consequence attached to the answer
// that can no longer be given is the one that happens.
//
// Caller must hold g.mu. Added in S28.
func (g *Game) QueueMayCastForEffect(
	chooser, source, cardID uuid.UUID,
	question string,
	onAccept, onDecline func(g *Game) error,
) error {
	if p := g.playerByIDLocked(chooser); p == nil || p.Eliminated {
		if onDecline != nil {
			return onDecline(g)
		}
		return nil
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoiceMayCast,
		Chooser:     chooser,
		Count:       1,
		Source:      source,
		Reason:      question,
		MayCastCard: cardID,
		mayCastResume: &mayCastFrame{
			onAccept:  onAccept,
			onDecline: onDecline,
		},
	})
	return nil
}

// ResolveMayCast processes the chooser's answer to a
// PendingChoiceMayCast entry. `apply: true` takes the offer.
//
// Exactly one branch runs, the prompt is dequeued either way, and the
// SBA / trigger drain runs afterwards because either branch can move
// cards between zones.
//
// Caller must NOT hold g.mu — this method takes the write lock.
// Added in S28.
func (g *Game) ResolveMayCast(choiceID, chooserID uuid.UUID, apply bool) error {
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
	if choice.Kind != PendingChoiceMayCast {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.mayCastResume
	g.dequeueChoiceLocked(idx)
	if frame == nil {
		return nil
	}
	branch := frame.onDecline
	if apply {
		branch = frame.onAccept
	}
	if branch != nil {
		if err := branch(g); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    chooserID,
				Source:   choice.Source,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runStateChecksLocked()
	return nil
}

// ResolvePayUnless processes the chooser's answer to a
// PendingChoicePayUnless entry. `apply: true` means "I pay": the
// cost is deducted from the chooser's pool, auto-tapping their
// untapped mana sources first if the pool is short (same planner
// CastSpell's AutoTap uses, no exclusions). If neither covers it
// the answer degrades to a decline — the player said yes to a bill
// they can't settle — and the "unless" consequence runs. `apply:
// false` runs the consequence directly.
//
// Either way the prompt is dequeued and the SBA / trigger drain
// runs, since the consequence (a draw, a token) may itself have
// triggered something.
//
// Caller must NOT hold g.mu — this method takes the write lock.
// Added in S19 sub-PR 6.
func (g *Game) ResolvePayUnless(choiceID, chooserID uuid.UUID, apply bool) error {
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
	if choice.Kind != PendingChoicePayUnless {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.payUnlessResume
	g.dequeueChoiceLocked(idx)
	if frame == nil {
		return nil
	}
	paid := false
	if apply {
		if p := g.playerByIDLocked(chooserID); p != nil {
			paid = g.payCostLocked(p, frame.cost, choice.Source)
		}
	}
	// Exactly one of the two branches runs. `paid` is the real
	// question, not `apply`: a chooser who said "Pay" and could not
	// fund it has declined, so the pay-unless consequence fires and
	// the may-pay one does not.
	consequence := frame.onDecline
	if paid {
		consequence = frame.onPay
	}
	if consequence != nil {
		if err := consequence(g); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    chooserID,
				Source:   choice.Source,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runStateChecksLocked()
	return nil
}

// declineDepartedChoiceLocked runs the "no" branch of a prompt whose
// chooser left the game still owing it. It is the dropDecline action
// of the departure table (choiceDepartureDecisions, leave_game.go) and
// the second half of CR 800.4f: the departed player's cost is not
// paid, and "unless that player pays" is a sentence about what happens
// when it is not. Rhystic Study still draws; Smothering Tithe still
// makes its Treasure (#961).
//
// This is the SAME continuation ResolvePayUnless runs for an answered
// "no", reached from the elimination sweep rather than from an answer
// — the #808 shape: one more terminal outcome of a paused pipeline,
// and leaving the game is the one being added. Two things make it safe
// to run from inside the sweep, and both are somebody else's invariant
// rather than a check here:
//
//   - it cannot re-queue a prompt to the departed seat, because
//     QueueChoiceForEffect refuses an eliminated chooser (#864). A
//     prompt it queues to a SURVIVOR is fine and stays queued;
//   - it is only reached for a prompt whose object a survivor still
//     controls (departedChoiceObjectLocked), so the branch is written
//     against material CR 800.4a has not just taken off the table.
//
// A may-pay frame (QueueMayPayForEffect: "you may pay {2}. If you do,
// ...") carries no decline branch and nothing happens, which is the
// right reading of "if you do" for a player who no longer can.
//
// Caller must hold g.mu.
func (g *Game) declineDepartedChoiceLocked(c *PendingChoice) {
	if c == nil || c.payUnlessResume == nil || c.payUnlessResume.onDecline == nil {
		return
	}
	if err := c.payUnlessResume.onDecline(g); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    c.Chooser,
			Source:   c.Source,
			ErrorMsg: err.Error(),
		})
	}
}

// S32 (#352): this one keeps the zero spend context deliberately. A
// pay-unless / may-pay cost is not a cast and not an activation — it
// is a cost demanded by a resolving effect — so no "spend only to
// cast X" token may fund it. Conservative in the weaker-than-printed
// direction, and correct for every restriction the catalog writes
// today, all of which name casting or activating.
//
// payCostLocked deducts `cost` (no X) from p's pool, auto-tapping
// untapped mana sources into the pool first when it's short.
// Returns false — with nothing tapped or spent — when the cost
// can't be met. Emits EventManaSpent on success so the client's
// pool display and the event log line up with the cast path.
// Caller must hold g.mu.
func (g *Game) payCostLocked(p *Player, cost ParsedCost, source uuid.UUID) bool {
	if !p.ManaPool.CanPay(cost, 0) {
		plan, ok := g.autoTapLocked(p.ID, cost, 0, nil)
		if !ok {
			return false
		}
		g.materializePlanLocked(p, plan, cost)
	}
	spent, ok := p.ManaPool.SpendManaFor(cost, 0, ManaSpendContext{})
	if !ok {
		return false
	}
	// #761: the log line carries what was spent, like every other
	// payment path. There is no stack item to hang the record on —
	// this pays a PROMPT's cost, which is already resolving — so
	// nothing else records it.
	g.EmitEvent(manaSpentEvent(p.ID, source, spent))
	return true
}

// ResolveTriggerOrder processes the chooser's answer to a
// PendingChoiceTriggerOrder entry. `resolutionOrder` lists the
// prompt's trigger IDs in the order the chooser wants them to
// RESOLVE (first resolves first). The engine reorders that seat's
// pending triggers so that placement puts the last-to-resolve item
// on the stack first, marks them Ordered, dequeues the prompt, and
// re-runs the APNAP drain — which proceeds if no other seat is
// still being asked.
//
// Triggers that arrived after the prompt was queued (not in the
// prompt's ID set) keep their harvest position after the ordered
// block and stay un-Ordered, so the next drain asks again with the
// full list.
//
// Caller must NOT hold g.mu — this method takes the write lock.
// Added in S19 sub-PR 8.
func (g *Game) ResolveTriggerOrder(choiceID, chooserID uuid.UUID, resolutionOrder []uuid.UUID) error {
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
	if choice.Kind != PendingChoiceTriggerOrder {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if len(resolutionOrder) != len(choice.TriggerOrderIDs) {
		return ErrInvalidParam
	}
	want := make(map[uuid.UUID]int, len(choice.TriggerOrderIDs))
	for _, id := range choice.TriggerOrderIDs {
		want[id]++
	}
	for _, id := range resolutionOrder {
		if want[id] <= 0 {
			return ErrInvalidParam
		}
		want[id]--
	}
	g.dequeueChoiceLocked(idx)

	// Rebuild PendingTriggers: other seats' items keep their
	// positions; this seat's prompted items are replaced, in place
	// of the first of them, by the chosen order reversed (placement
	// order = reverse of resolution order); this seat's un-prompted
	// items keep their relative order after the block.
	byID := make(map[uuid.UUID]*StackItem, len(g.PendingTriggers))
	for _, t := range g.PendingTriggers {
		byID[t.ID] = t
	}
	placement := make([]*StackItem, 0, len(resolutionOrder))
	for i := len(resolutionOrder) - 1; i >= 0; i-- {
		if t, ok := byID[resolutionOrder[i]]; ok {
			t.Ordered = true
			placement = append(placement, t)
		}
	}
	inPrompt := make(map[uuid.UUID]bool, len(resolutionOrder))
	for _, id := range resolutionOrder {
		inPrompt[id] = true
	}
	rebuilt := make([]*StackItem, 0, len(g.PendingTriggers))
	inserted := false
	for _, t := range g.PendingTriggers {
		if inPrompt[t.ID] {
			if !inserted {
				rebuilt = append(rebuilt, placement...)
				inserted = true
			}
			continue
		}
		rebuilt = append(rebuilt, t)
	}
	g.PendingTriggers = rebuilt
	g.runStateChecksLocked()
	return nil
}

// PendingChoiceKindFor returns the kind of the queue entry with the
// given ID, or empty + false when no such entry exists. Used by the
// resolve_choice action dispatcher to route a yes/no payload to the
// right resolve method (S17 PendingChoiceOptionalReplacement vs S19
// PendingChoiceTriggerPrompt — both consume `{apply: bool}` so the
// dispatcher can't disambiguate from the payload shape alone).
//
// Caller must NOT hold g.mu — takes the read lock. Added in S19
// sub-PR 2.
func (g *Game) PendingChoiceKindFor(choiceID uuid.UUID) (PendingChoiceKind, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			return c.Kind, true
		}
	}
	return "", false
}

// findBattlefieldCard returns a pointer to the battlefield card
// with the given instance ID, or nil. Caller must hold g.mu.
func findBattlefieldCard(g *Game, id uuid.UUID) *Card {
	if g.Battlefield == nil {
		return nil
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// ResolveSacrificeChoice answers a PendingChoiceSacrifice: the chooser
// names one of their own permanents and it is sacrificed (CR 701.21a).
//
// The option list is re-checked rather than trusted. Grave Pact
// prompts every other player at once, and an earlier answer can change
// a later player's board — a Blood Artist drain that kills something,
// a sacrifice that triggers a Butcher of Malakir. A card that has left
// the battlefield, or is no longer theirs, is refused; the choice
// stays queued so they can pick again.
//
// A choice whose chooser has no legal permanent left is dropped rather
// than left blocking the queue: the requirement is "sacrifice a
// creature if you can", and they no longer can.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveSacrificeChoice(choiceID, chooserID, cardID uuid.UUID) error {
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
	if choice.Kind != PendingChoiceSacrifice {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	offered := false
	for _, id := range choice.SacrificeOptions {
		if id == cardID {
			offered = true
			break
		}
	}
	if !offered {
		return ErrInvalidParam
	}
	// Still theirs, still on the battlefield?
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return ErrCardNotFound
	}
	if c.Controller != chooserID {
		return ErrCardCallerMismatch
	}
	// The run this prompt is one leg of, read BEFORE the dequeue: the
	// entry is about to leave the queue and the continuation below
	// closes over the id rather than over the entry.
	run, seat := choice.promptRun, choice.Chooser
	g.dequeueChoiceLocked(idx)
	// #1019: the picked permanent goes through the sacrifice's own
	// CONTINUATION rather than the fire-and-forget call, for the two
	// reasons ADR 0013 §5x gives. It is what lets a run wait for a leg
	// the CR 903.9 window has merely PAUSED — a sacrificed commander
	// is still on the battlefield while its owner answers, and a run
	// that settled on this line would pay out with the permanent still
	// in play — and what lands in the run is
	// sacrificedThisWayLocked's answer rather than a re-read of the
	// board on the next line.
	//
	// The source stays uuid.Nil because the fire-and-forget form
	// passed none (sacrificePermanentLocked): EventSacrifice carries
	// no source for a prompted sacrifice today, and giving the leg a
	// continuation must change the sequencing and nothing else.
	if err := g.SacrificeThenForEffect(uuid.Nil, cardID, func(g *Game, sacrificed bool) error {
		var landed []uuid.UUID
		if sacrificed {
			landed = []uuid.UUID{cardID}
		}
		return g.settleRunLegLocked(run, seat, landed)
	}); err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}

// pruneSacrificeChoicesLocked drops any queued sacrifice choice whose
// chooser has run out of legal permanents, and trims option lists to
// what is still on the battlefield under that chooser's control.
//
// Called after each sacrifice resolves, because one player's answer
// can empty another player's board. Without it a Grave Pact whose
// last creature died to a drain would leave a prompt nobody can
// answer, wedging the queue.
//
// Caller must hold g.mu.
func (g *Game) pruneSacrificeChoicesLocked() {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || c.Kind != PendingChoiceSacrifice {
			continue
		}
		live := make([]uuid.UUID, 0, len(c.SacrificeOptions))
		for _, id := range c.SacrificeOptions {
			card := findBattlefieldCard(g, id)
			if card != nil && card.Controller == c.Chooser {
				live = append(live, id)
			}
		}
		if len(live) == 0 {
			g.dropChoiceLocked(i)
			continue
		}
		c.SacrificeOptions = live
	}
}

// pruneCardSetChoicesLocked is the prune above for the other prompt
// family that names CARDS: choose_cards. It trims every open pick to
// the candidates still in the zone the pick is about, and DROPS one
// whose candidates have all gone (#1045).
//
// ONE FUNCTION KEYED BY ZONE, not one per verb. A discard is a
// choose_cards over the discarding player's own HAND, Thoughtseize's
// pick is one over somebody else's hand, a graveyard pick names a
// GRAVEYARD and a reveal-and-take names a LIBRARY — and the question
// "is this candidate still there" has exactly one answer for all of
// them, `chooseCardsFrame.zone`. Writing a pruneDiscardChoicesLocked
// beside a pruneGraveyardChoicesLocked would be three copies of one
// rule, which is the shape the catalog's clone gate exists to refuse.
//
// It asks the SUBMIT PATH's question (pickStillInPickZoneLocked,
// shared with checkChooseCardsPicksLocked), so the candidate list the
// client and the bot enumerator are shown can never offer a card the
// resolver would refuse. That drift IS the bug: the list is frozen at
// queue time and the zone is re-read on submit, so a prompt whose
// candidates have all left the zone has no legal answer, and one with
// a floor of one cannot be discharged at all — the #544 wedge, with
// the whole table behind it (#791's gate) and, since #1027, a run leg
// that never settles either.
//
// A prompt with no zone on its frame is left alone. Zero means "no
// live re-check" (see chooseCardsFrame.zone): the continuation owns
// what a pick means, the resolver refuses nothing, and there is
// nothing for this to be right about.
//
// An emptied prompt is DROPPED rather than left unanswerable, through
// dropChoiceLocked — which runs the departure table's action for the
// kind, so a prompt that is one leg of a prompted RUN settles that leg
// with "nothing discarded" (#1016's dropDefault, ADR 0013 §5y item 5)
// and the rest of the printed instruction still happens. A pick that
// is no run's leg has no continuation to run, exactly as it has none
// when its chooser leaves the game; that is the departure table's
// existing answer for the kind and this does not widen it.
//
// The bounds move with the list (setCardSetCandidates): a pick of two
// from a hand that now holds one is CR 701.8a's "as many as you can",
// and a floor left above the candidate count is the same wedge one
// card later.
//
// The drop is announced (EventPendingChoiceDropped, #868) for
// pruneDepartedSeatOptionsLocked's reason: this prompt was blocking
// the table, and a stall dump has to say where it went.
//
// untap_choice carries the same payload (isCardSetPickKind) and is
// deliberately NOT swept. Its continuation is the rest of the UNTAP
// STEP (finishUntapStepLocked), its departure row is dropDiscard, so a
// withdrawal here would end the question by stranding the step — and
// it cannot need this: CR 502.3's determination is about the active
// player's own permanents during their own untap step, where nothing
// has priority to move them.
//
// #1198's entry_reveal_from_hand is out for the same reason and one
// of its own. The drop would strand a paused CR 614 ENTRY, which is
// worse than the untap step it would strand above; and the prompt
// cannot need the prune, because its floor is zero — "reveal nothing"
// is an answer no emptied candidate list can take away, which is
// exactly the property untap_choice lacks and a choose_cards with a
// floor of one lacks. A stale candidate is still refused on submit by
// pickStillInPickZoneLocked, so the worst an un-pruned list costs is
// one rejected click.
//
// Called where pruneSacrificeChoicesLocked is called, and for its
// reason: a queued choice stops priority from passing, so the
// state-check loop is exactly what does NOT run while one is open.
//
// And at a fourth site those three do not reach (#1069): the ENTRY
// side, through pruneChoicesAfterArrivalLocked. Those three are exits
// and a departure, and a card that leaves a hand, a library or a
// graveyard FOR THE BATTLEFIELD takes none of them — a candidate
// reanimated or put onto the battlefield under an open prompt left the
// pick's zone exactly as a discarded one does. See
// battlefield_entry.go for the three landings that share that door.
//
// THE WHOLE QUEUE IS EXAMINED BEFORE ANY PROMPT IS DROPPED, which is
// pruneStaleZoneChangeChoicesLocked's shape and its reason: a drop runs
// the kind's drop action, a run leg settling runs the rest of the
// card, and that continuation can queue the next prompt, withdraw
// another one, or re-enter this very function through an exit it
// starts. Walking the slice by index while a callee rewrites it reads
// the wrong entry at best and indexes past the end at worst. The first
// pass only rewrites candidate lists, which move nothing; the second
// re-finds each emptied prompt by ID, so one that a continuation has
// already withdrawn is simply gone.
//
// Caller must hold g.mu.
func (g *Game) pruneCardSetChoicesLocked() {
	var emptied []*PendingChoice
	for _, c := range g.PendingChoices {
		// Every kind carrying the choose-cards payload, not just
		// choose_cards itself (#1214): a their_permanents prompt over
		// a board that empties under it is exactly the prompt this
		// sweep exists for, and untap_choice sets no zone so it falls
		// out at the next line.
		if c == nil || !isCardSetPickKind(c.Kind) {
			continue
		}
		frame := c.chooseCardsResume
		if frame == nil || frame.zone == "" {
			continue
		}
		live := make([]uuid.UUID, 0, len(c.ChooseCards))
		for _, id := range c.ChooseCards {
			if g.pickStillInPickZoneLocked(c, frame.zone, id) {
				live = append(live, id)
			}
		}
		if len(live) == len(c.ChooseCards) {
			continue
		}
		if len(live) == 0 {
			emptied = append(emptied, c)
			continue
		}
		setCardSetCandidates(c, live)
	}
	for _, c := range emptied {
		idx, _ := g.findChoiceLocked(c.ID)
		if idx < 0 {
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventPendingChoiceDropped,
			Actor:  c.Chooser,
			Source: c.Source,
			Label:  string(c.Kind),
		})
		g.dropChoiceLocked(idx)
	}
}

// setCardSetCandidates replaces a choose-cards prompt's candidate list
// and re-clamps the bounds to it — THE one place a shortened list and
// its Min / Max / Count are kept honest, shared by the zone prune
// above and by reassignChoiceLocked's departure prune.
//
// Two callers, one rule: an enumerator that offers a bigger set than
// the resolver accepts is the #544 wedge, and so is a floor no
// remaining set can reach. Clamping down is CR 701.8a's "as many as
// you can" for a discard and the same arithmetic for every other
// pick.
//
// `live` must be non-empty; an emptied prompt is the caller's
// decision, and the two callers make different ones (this prune drops
// it, the reassignment refuses to move it).
func setCardSetCandidates(c *PendingChoice, live []uuid.UUID) {
	c.ChooseCards = live
	if c.ChooseMax > len(live) {
		c.ChooseMax = len(live)
	}
	if c.ChooseMin > c.ChooseMax {
		c.ChooseMin = c.ChooseMax
	}
	if c.Count > c.ChooseMax {
		c.Count = c.ChooseMax
	}
}

// pruneDepartedSeatOptionsLocked takes the seats that have left the
// game off every open option-pick prompt (#994, CR 800.4a).
//
// THE OTHER HALF of reassignChoiceLocked's seat prune, and the reason
// both exist. That one runs when the CHOOSER leaves and is about the
// prompt changing hands; this one runs when ANYBODY leaves and is about
// the prompt's own option list, including the far commoner case — a
// prompt owed by a survivor that offers a seat somebody else has just
// vacated. Nothing ran for that case before: eligibility was filtered
// once, at queue time (eligibleChosenPlayersLocked), and a seat that
// left afterwards stayed on the buttons. "Choose a player" could then
// name somebody who was not a player, which CR 800.4a says it cannot.
//
// It prunes and does not re-ask. A player choice is a choice among
// seats, so a shorter list is the same question with one fewer answer;
// the question is the same one and the card is not re-read.
//
// An emptied prompt is DROPPED rather than left unanswerable. That is
// the state QueueChoosePlayerForEffect refuses to queue in the first
// place — no eligible seat, no question — and an option_pick blocks the
// table (choice_gate.go), so a prompt with nothing on it is the #544
// wedge rather than a harmless leftover.
//
// The drop goes through dropChoiceLocked, which since #1006 also runs
// the departure table's action for the kind: option_pick is
// dropDefault, so the frame runs with "nobody chose" and the rest of
// the card — the sentence printed after "choose a player" — still
// happens. This path is the reason that fix is not only about
// departures: here the CHOOSER is still at the table and it is the
// question that has gone.
//
// Card options are NOT touched here. A card that left with its owner is
// removeObjectsOwnedByLocked's business and reaches the prompts through
// the prunes beside this one; keepSeats only ever asks about an option
// that names a seat.
//
// Caller must hold g.mu.
func (g *Game) pruneDepartedSeatOptionsLocked() {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || c.Kind != PendingChoiceOptionPick || len(c.PickOptions) == 0 {
			continue
		}
		opts, ok := keepSeats(c.PickOptions, func(id uuid.UUID) bool {
			p := g.playerByIDLocked(id)
			return p != nil && !p.Eliminated
		})
		if !ok {
			g.EmitEvent(Event{
				Kind:   EventPendingChoiceDropped,
				Actor:  c.Chooser,
				Source: c.Source,
				Label:  string(c.Kind),
			})
			g.dropChoiceLocked(i)
			continue
		}
		c.PickOptions = opts
	}
}

// pausedZoneChangeStaleLocked reports whether the move stashed on a
// queued prompt's resume frame can still happen.
//
// A paused move moves NOTHING: the card sits in its old zone until
// the prompt is answered (see routeCardToZoneLocked, and
// enterBattlefieldThroughPipelineLocked for the entry side). So a card
// that is no longer in the zone its paused move says it is leaving has
// already left by some other route, and that move will never happen.
// Answering the prompt then either fails — the battlefield-leave
// resume cannot find the card on the battlefield, which is the
// "card instance not found in zone" of #605 — or, worse, moves the
// card a SECOND time out of a zone nobody asked about, because the
// shared primitive moves it from wherever it finds it.
//
// #478 brought battlefield ENTRIES under the same rule. They used to be
// excluded on the grounds that a paused entry could only ever be
// answered into a harmless double-push, which was true while nothing
// that pauses an entry could be overtaken: a fetched permanent's entry
// prompt is now open while the card sits in a LIBRARY, and a mill, a
// draw or an opponent's exile can take it out from under the question.
// A card already ON the battlefield is not stale — the entry it is
// waiting on simply short-circuits — and neither is one still in its
// source zone.
//
// Caller must hold g.mu.
func (g *Game) pausedZoneChangeStaleLocked(frame *replacementResumeFrame) bool {
	if frame == nil || frame.ev == nil {
		return false
	}
	ev := frame.ev
	if !isExitMove(ev.Kind) {
		return false
	}
	src := g.findCardZoneLocked(ev.CardID)
	if src == nil {
		// #762: a CREATED TOKEN whose entry is paused is in no zone at
		// all — minted, staged, not yet pushed. That is not a card that
		// has left by another route; it is exactly where the paused
		// entry left it.
		_, staged := g.enteringTokenLocked(ev.CardID)
		return !staged
	}
	if ev.NewZone == ZoneBattlefield && src == g.Battlefield {
		return false
	}
	return src.Kind != ev.OldZone
}

// zoneChangePausedLocked reports whether an exit for cardID is already
// waiting on a player's answer — the CR 903.9 "send your commander to
// the command zone instead?" prompt, today the only one that pauses a
// card on its way off the battlefield.
//
// The SBA sweep consults it because a paused exit leaves the doomed
// permanent ON the battlefield with the condition that doomed it
// intact, so the next sweep dooms it all over again (#605).
//
// Caller must hold g.mu.
func (g *Game) zoneChangePausedLocked(cardID uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c == nil || c.replacementResume == nil {
			continue
		}
		ev := c.replacementResume.ev
		if ev == nil || !isExitMove(ev.Kind) || ev.NewZone == ZoneBattlefield {
			continue
		}
		if ev.CardID == cardID {
			return true
		}
	}
	return false
}

// pruneStaleZoneChangeChoicesLocked drops every queued prompt whose
// paused exit can no longer happen, releases the CR 614.5
// once-per-event bookkeeping the abandoned event was holding, and runs
// each abandoned route's continuation.
//
// Called at the two points a card actually lands — the shared exit
// primitive and the battlefield-leave resume — because that is the
// moment a sibling prompt about the same card goes stale, and the
// state-check loop that would otherwise notice is exactly what does
// NOT run while a choice is queued. Same reasoning, and the same call
// sites, as pruneSacrificeChoicesLocked.
//
// #865: the prune is a TERMINAL outcome of the route it withdraws, so
// the continuation the route was carrying runs —
// abandonZoneRouteLocked, shared with the drop path. The whole queue
// is swept before any tail runs, for two reasons: a tail may queue the
// next leg's prompt, and it re-enters this function through
// executeZoneRouteLocked when that leg lands, so walking the slice by
// index while a callee mutates it is exactly the bug this shape
// avoids. finishDroppedReplacementLocked's caller collects its frames
// the same way and for the same reason.
//
// Caller must hold g.mu.
func (g *Game) pruneStaleZoneChangeChoicesLocked() {
	var abandoned []*replacementResumeFrame
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || !g.pausedZoneChangeStaleLocked(c.replacementResume) {
			continue
		}
		abandoned = append(abandoned, c.replacementResume)
		g.dropChoiceLocked(i)
	}
	for _, frame := range abandoned {
		if err := g.abandonZoneRouteLocked(frame); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "route continuation failed: " + err.Error(),
			})
		}
	}
}

// dropStaleReplacementResumeLocked is the answer-path half of the
// prune above: a belt-and-braces check that the move a just-answered
// prompt was asking about is still possible. Returns true when the
// frame was abandoned, in which case the caller returns nil — the
// prompt is already dequeued, and refusing the answer instead would
// leave the seat holding a prompt it can never discharge.
//
// #865: abandoning is terminal, so it goes through the same
// abandonZoneRouteLocked the prune and the drop do and the route's
// continuation runs. Reaching here at all means the prune above missed
// the frame, but a batch that stalls is a batch that stalls however it
// got there.
//
// Caller must hold g.mu.
func (g *Game) dropStaleReplacementResumeLocked(frame *replacementResumeFrame) bool {
	if !g.pausedZoneChangeStaleLocked(frame) {
		return false
	}
	g.EmitEvent(Event{
		Kind: EventEffectError,
		ErrorMsg: "replacement prompt dropped: its card is no longer in the " +
			string(frame.ev.OldZone) + " the paused move would leave",
	})
	if err := g.abandonZoneRouteLocked(frame); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			ErrorMsg: "route continuation failed: " + err.Error(),
		})
	}
	return true
}

// ResolveScry answers a PendingChoiceScry (CR 701.22): `bottom` are the
// looked-at cards going to the bottom of the library, `topOrder` are the
// ones staying on top, listed top-first.
//
// Every looked-at card must appear in exactly one list. Scry moves all
// of the cards it looked at — a card the chooser "leaves alone" is
// still being put back on top, which is a choice about order, so an
// answer that omits one is a client bug rather than a shorthand for
// "leave it".
//
// Relative order within `bottom` is not observable (the cards go under
// the whole library, and nothing in this sandbox reads library-bottom
// order), so it is applied as given rather than being validated.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveScry(choiceID, chooserID uuid.UUID, bottom, topOrder []uuid.UUID) error {
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
	if choice.Kind != PendingChoiceScry {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	p := g.playerByIDLocked(chooserID)
	if p == nil {
		return ErrPlayerNotFound
	}
	want, err := partitionLookedAtCards(p, choice.ScryCards, bottom, topOrder)
	if err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)

	// Pull all of them out first, then put them back, so the
	// intermediate state can't depend on removal order.
	pulled := make(map[uuid.UUID]Card, len(want))
	for id := range want {
		c, err := p.Library.Remove(id)
		if err != nil {
			return err
		}
		pulled[id] = c
	}
	for _, id := range bottom {
		p.Library.PushBottom(pulled[id])
	}
	// topOrder is top-first and PushTop appends to the top, so walk it
	// backwards: the last push lands on top and must be the caller's
	// first entry.
	for i := len(topOrder) - 1; i >= 0; i-- {
		p.Library.PushTop(pulled[topOrder[i]])
	}
	g.EmitEvent(Event{
		Kind:   EventScry,
		Actor:  chooserID,
		Source: choice.Source,
		Amount: len(bottom),
		// The size of the scry, not the size of the motion: the cards
		// the prompt was asking about, which is what the CR 614
		// keyword-action window settled on clamped to the library
		// (see Event.LookedAt). Both numbers are on the event because
		// the log says both.
		LookedAt: len(choice.ScryCards),
	})
	// The rest of the effect, now that the library is in the order the
	// player chose. Preordain's draw happens here, which is what makes
	// "then" mean what it says.
	if choice.scryResume != nil {
		if err := choice.scryResume(g); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    chooserID,
				Source:   choice.Source,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runStateChecksLocked()
	return nil
}

// IsLookAtTopKind reports whether a choice kind belongs to the scry
// family: the prompts that show their chooser the top N cards of
// their own library and ask where they go.
//
// The three share PendingChoice.ScryCards, scryResume, the
// chooser-only wire redaction and the client dialog — so the places
// that care about the family rather than the individual keyword (the
// view projection, the legal-move enumerator, the dispatcher) ask
// here instead of listing the kinds and going stale when a fourth
// arrives. Added in S22.
func IsLookAtTopKind(k PendingChoiceKind) bool {
	switch k {
	case PendingChoiceScry, PendingChoiceSurveil, PendingChoiceLookAtTop:
		return true
	}
	return false
}

// ResolveLookAtTop answers a PendingChoiceLookAtTop: `topOrder` is
// every looked-at card, listed top-first, put back on the library in
// that order.
//
// The degenerate member of the scry / surveil family — there is no
// away lane, so the answer is a permutation rather than a partition.
// It still runs through the same validation, with an empty away list,
// because "every looked-at card exactly once" is the part that has to
// hold either way.
//
// No event is emitted. Nothing observable happened: the same cards are
// in the same zone, and "the order of your library changed" is not a
// thing another player is entitled to know about.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveLookAtTop(choiceID, chooserID uuid.UUID, topOrder []uuid.UUID) error {
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
	if choice.Kind != PendingChoiceLookAtTop {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	p := g.playerByIDLocked(chooserID)
	if p == nil {
		return ErrPlayerNotFound
	}
	want, err := partitionLookedAtCards(p, choice.ScryCards, nil, topOrder)
	if err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)

	pulled := make(map[uuid.UUID]Card, len(want))
	for id := range want {
		c, err := p.Library.Remove(id)
		if err != nil {
			return err
		}
		pulled[id] = c
	}
	// topOrder is top-first and PushTop appends to the top, so walk it
	// backwards: the last push lands on top.
	for i := len(topOrder) - 1; i >= 0; i-- {
		p.Library.PushTop(pulled[topOrder[i]])
	}
	if choice.scryResume != nil {
		if err := choice.scryResume(g); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    chooserID,
				Source:   choice.Source,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runStateChecksLocked()
	return nil
}

// partitionLookedAtCards validates the answer to a scry or a surveil:
// `away` (bottom of library / graveyard) and `topOrder` must together
// name every card in `lookedAt` exactly once, and every one of them
// must still be in the library.
//
// Both keywords move ALL of the cards they looked at — a card the
// chooser "leaves alone" is still being put back on top, which is a
// choice about order — so an answer that omits one is a client bug
// rather than a shorthand for "leave it".
//
// The library re-check is belt-and-braces: both resolve atomically, so
// nothing should have moved, but it turns a desync into a rejection
// rather than a half-applied reorder. Returns the looked-at set as a
// lookup so the caller can pull the cards out without rebuilding it.
//
// Caller must hold g.mu.
func partitionLookedAtCards(p *Player, lookedAt, away, topOrder []uuid.UUID) (map[uuid.UUID]bool, error) {
	want := make(map[uuid.UUID]bool, len(lookedAt))
	for _, id := range lookedAt {
		want[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(want))
	for _, list := range [][]uuid.UUID{away, topOrder} {
		for _, id := range list {
			if !want[id] || seen[id] {
				return nil, ErrInvalidParam
			}
			seen[id] = true
		}
	}
	if len(seen) != len(want) {
		return nil, ErrInvalidParam
	}
	for id := range want {
		if !p.Library.Contains(id) {
			return nil, ErrCardNotFound
		}
	}
	return want, nil
}

// ResolveSurveil answers a PendingChoiceSurveil (CR 701.25):
// `graveyard` are the looked-at cards going to the chooser's
// graveyard, `topOrder` are the ones staying on top of the library,
// listed top-first.
//
// Scry's shape with the bottom-of-library leg replaced by the
// graveyard, and the difference is the whole point of the keyword —
// the cards that go down are in a zone the chooser can reanimate,
// delve and escape from, not buried under sixty other cards.
//
// The graveyard leg emits EventMill per card, the same event an
// ordinary mill emits, because it is the same zone change from the
// same zone: a "whenever a card is put into your graveyard from your
// library" payoff must not care how it got there. The completed
// surveil then emits EventSurveil once, for the payoffs that watch
// the keyword itself.
//
// #931: the graveyard leg goes through the shared exit primitive
// (routeAllThenLocked with millRoute, the template it shares with an
// ordinary library → graveyard move) rather than pushing the cards
// into the pile by hand. So the CR 614 window opens on each of them —
// "if a card would be put into a graveyard from anywhere, exile it
// instead" (Rest in Peace, Leyline of the Void) — and a surveilled
// commander is offered the command zone (CR 903.9).
//
// Which means the leg can PAUSE, so everything after it moved into the
// batch's continuation: the EventSurveil announcement and the rest of
// the effect run once every card has settled, not on the next line
// with a prompt still open.
//
// The cards that are going to the graveyard are put BACK on top of the
// library first, above the ones that stay, and then routed out of it.
// They have to be in a zone for the window to replace a move out of
// it, and the order is what leaves the library correct on the far
// side: once the graveyard-bound cards have left, the kept ones are
// the top of the library in the order the player chose. A leg the
// window cancels outright is the one visible consequence — that card
// stays on top of the library rather than under the kept ones, which
// is a position the rules do not name for a move that never happened.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveSurveil(choiceID, chooserID uuid.UUID, graveyard, topOrder []uuid.UUID) error {
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
	if choice.Kind != PendingChoiceSurveil {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	p := g.playerByIDLocked(chooserID)
	if p == nil {
		return ErrPlayerNotFound
	}
	want, err := partitionLookedAtCards(p, choice.ScryCards, graveyard, topOrder)
	if err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)
	// Read off the choice before it goes out of scope: the
	// continuation below may run an action later, from the resume of a
	// CR 903.9 prompt, and must not close over a dequeued frame.
	resume := choice.scryResume
	// The size of the surveil, read off the frame for the same reason
	// the resume is: the continuation below runs later, possibly from
	// the resume of a CR 903.9 prompt, with the frame dequeued.
	lookedAt := len(choice.ScryCards)

	// Pull all of them out first, then place them, so the
	// intermediate state can't depend on removal order.
	pulled := make(map[uuid.UUID]Card, len(want))
	for id := range want {
		c, err := p.Library.Remove(id)
		if err != nil {
			return err
		}
		pulled[id] = c
	}
	// The kept cards go back first and the graveyard-bound ones on top
	// of them, so that once the route has taken the latter out of the
	// library the former are its top in the order the player chose.
	// topOrder is top-first and PushTop appends to the top, so walk it
	// backwards: the last push lands on top and must be the caller's
	// first entry.
	for i := len(topOrder) - 1; i >= 0; i-- {
		p.Library.PushTop(pulled[topOrder[i]])
	}
	for i := len(graveyard) - 1; i >= 0; i-- {
		p.Library.PushTop(pulled[graveyard[i]])
	}
	r := millRoute(chooserID, ZoneGraveyard)
	r.Source = choice.Source
	source := choice.Source
	err = g.routeAllThenLocked(r, graveyard, func(g *Game, milled []uuid.UUID) error {
		g.EmitEvent(Event{
			Kind:   EventSurveil,
			Actor:  chooserID,
			Source: source,
			Amount: len(milled),
			// `milled` is what the route actually put in the
			// graveyard — a leg a Rest in Peace exiled instead is not
			// in it — while the size of the surveil is the whole
			// looked-at set and does not move with the replacement.
			LookedAt: lookedAt,
		})
		// The rest of the effect, now that the library is in the order
		// the player chose and every card that was going to a graveyard
		// has settled.
		if resume != nil {
			if err := resume(g); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    chooserID,
					Source:   source,
					ErrorMsg: err.Error(),
				})
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}
