package game

import (
	"errors"
	"slices"
	"strconv"

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

	// PendingChoiceOptionalReplacement — "may" yes/no prompt
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
	// cards ride in ScryCards, the continuation in scryThen, and
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

	// PendingChoicePutInLibrary — "put these cards into their owners'
	// libraries in the order you choose" (CR 401.4, ADR 0088). The
	// placement half of the library-top row (#996): "put the rest on
	// the bottom of your library in any order" (Goblin Ringleader,
	// Impulse), "put two cards from your hand on top of your library
	// in any order" (Brainstorm), "its owner puts it on their choice
	// of the top or bottom of their library" (Aetherspouts).
	//
	// The fourth member of the scry family: same ScryCards, same
	// chooser-only projection, same dialog, same answer keys. It is
	// its own kind because it is not a look at the top of a library —
	// the cards may be in a hand, on the battlefield or on the stack,
	// and LibraryPlacement says which lanes the answer may use.
	//
	// Answered with {top_order, bottom}, both TOP-FIRST (the first
	// entry of `bottom` is the one nearest the top of the pile that
	// goes under the library). Every card appears in exactly one list,
	// and a lane the placement does not open must be empty.
	PendingChoicePutInLibrary PendingChoiceKind = "put_in_library"

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

	// ControlPurpose is what giving the entering permanent away does to
	// its recipient, for a PendingChoiceEntryController (ADR 0102).
	// Unused for every other kind. Carried for ColorPurpose's reason —
	// only the card knows — and projected as `control_purpose`.
	ControlPurpose ControlPurpose

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

	// ManaRiders are the spend riders the token minted by this
	// PendingChoiceMana will carry (#1547) — Cavern of Souls' "and that
	// spell can't be countered", Path of Ancestry's scry. On the choice
	// for ManaRestrictions' reason: the token is minted later, in
	// ResolveManaChoice, when the ability shape is out of scope. The
	// Production id is stamped at that mint, not here.
	//
	// Deep-copied by clone.go and mirrored by the snapshot.
	ManaRiders []ManaSpendRider

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

	// ManaDifferent is N for a PendingChoiceMana that names N DIFFERENT
	// colours one answer at a time — "Add two mana of different colors"
	// (#2558, ProducedManaEntry.Distinct). Zero for every ordinary pick.
	// Each answer is struck from ColorOptions and recorded in
	// ManaChosen, and the choice stays open, with a fresh ID, until the
	// N-th answer; only then is any mana added, all N at once (CR
	// 605.3b), and the triggered mana abilities fire once (CR 106.12a).
	// Carried by the snapshot, so a restore with the pick half answered
	// resumes it.
	ManaDifferent int

	// ManaChosen is the colours already named for a ManaDifferent pick,
	// in answer order. None of them is in ColorOptions. Deep-copied by
	// clone.go and carried by the snapshot.
	ManaChosen []string

	// ManaLabel is the ability's own label for a ManaDifferent pick, the
	// stem Reason is rebuilt from after each answer.
	ManaLabel string

	// ManaSettled and ManaPicksLeft make one mana ability's whole output
	// ONE production (#3939, CR 106.12b): the mana a replacement such as
	// Damping Sphere's "if a land is tapped for two or more mana" has to
	// see as a unit. When an activation has colours still to pick, the
	// colours it already knows wait in ManaSettled instead of reaching
	// the pool, the first pick is this choice, and the rest of the picks
	// wait in ManaPicksLeft. Each answer moves its colours into
	// ManaSettled and, while picks remain, rewrites the choice in place
	// for the next one (a fresh ID, as a ManaDifferent pick does); the
	// last answer produces ManaSettled in one window. Both are empty for
	// an ordinary pick. Deep-copied by clone.go and carried by the
	// snapshot.
	ManaSettled   []string
	ManaPicksLeft []ManaPickSlot

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
	// on submit. Added in S17 sub-PR 2. Written to a restore point as
	// data when its event holds no closure (#4253,
	// replacement_resume_keyed.go).
	replacementResume *replacementResumeFrame

	// costCommanderResume is the parked announcement behind a CR 903.9
	// prompt asked BEFORE a cost is paid (#1397, cost_commander_choice.go).
	// The prompt is an ordinary PendingChoiceOptionalReplacement on the
	// wire; this frame is what tells ResolveOptionalReplacement to re-run
	// the announcement with the answer rather than resume a paused event.
	// Mutually exclusive with replacementResume. Not serialised.
	costCommanderResume *costCommanderFrame

	// DamageAssignment is the client-facing payload for a
	// PendingChoiceDamageAssignment entry: the attacker's instance
	// ID, ordered blocker instance IDs (in declared order; the
	// client may re-order), attacker's effective power, and
	// trample-allowed flag. Wire-serialised via
	// PendingChoiceView.DamageAssignment. Added in S18 sub-PR 3.
	DamageAssignment *DamageAssignmentFrame

	// DivideShield is the payload of a PendingChoiceDivideShield (ADR
	// 0108 §7 decision 6, divide_shield.go): the charged shield, its
	// charge and the damage events it meets. Wire-serialised via
	// PendingChoiceView.DivideShield. Written with the prompt, with the
	// shields after it and the damage it lands, as data (#4361 S16).
	DivideShield *DivideShieldPrompt

	// PayAmount is the payload of a PendingChoicePayAmount (ADR 0129 §3,
	// energy_payment.go): the bounds of the answer, the card's own
	// threshold and what each counter buys. Wire-serialised via
	// PendingChoiceView.PayAmount. Plain data, written to a restore point
	// (#4289, #4294): the keyed "choose X" of a MayPayX (MayPayX below), a
	// keyed number prompt (PickThen is a RegisterNumberThen key,
	// number_then.go), and since #4361 (which deleted the closure frame) a
	// payment nothing follows.
	PayAmount *PayAmountPrompt

	// MayPayX makes a pay_amount the "choose X" half of a KEYED
	// "you may pay {X}{R}. If you do, ..." (#4289, pay_x_on_resolution.go):
	// the printed cost, the words of the next prompt and whether it holds
	// the step. Answering it queues the may-pay for the cost with X
	// settled, whose paid branch is PickThen's, handed the X (PickNumber,
	// PayUnlessAnswer.Number). Plain data: the prompt holds no frame, so a
	// table waiting on X is a restore point. Nil on every other prompt.
	MayPayX *MayPayXClause

	// PickNumber is a number a keyed pay-unless's branches are handed
	// (#4289): the X a MayPayX settled before it asked whether to pay.
	// Zero on every other prompt.
	PickNumber int

	// AttackSplit makes a keyed number prompt a question of a
	// tokens-attacking split (CR 508.4, attack_split.go): the template, the
	// things to attack and the shares fixed so far. Plain data, written to
	// a restore point; nil on every other prompt.
	AttackSplit *AttackSplit

	// PickWords and PickNums are the words and numbers a keyed number
	// prompt's branch is handed beside OptionCarry (#4294, number_then.go):
	// the kinds of counters still to ask about, and how many were chosen
	// so far. Empty on every other prompt.
	PickWords []string
	PickNums  []int

	// NoLegalTarget marks a PendingChoiceTriggerPrompt whose effect
	// has no legal target / will pass without effect if the chooser
	// answers "Yes" (e.g. Reclamation Sage with no opponent artifact,
	// Eternal Witness with an empty graveyard). Computed at queue
	// time from the ability's HasLegalTarget predicate; serialised to
	// the wire so the client can warn the chooser. False for prompts
	// without a HasLegalTarget predicate. Added in S19 follow-up.
	NoLegalTarget bool

	// OptionalItem is, on a trigger_prompt asked as an optional trigger
	// resolves (CR 603.5, #2494), the ID of that resolving item, which
	// the parked resolution tail holds; uuid.Nil on the harvest-time
	// prompt (miracle's reveal) and every other kind. Plain data, so
	// the prompt is a restore point. OptionalTradeFor is the item's
	// OptionalTrigger.TradeFor, for TradeSubject.
	OptionalItem     uuid.UUID
	OptionalTradeFor uuid.UUID

	// NoGain marks a resolution-time "you may" whose row says "yes"
	// would leave the chooser no better off as things stand (ADR 0197
	// §10, CatalogActivatedMayHelps): Arni Brokenbrow's new base power
	// would be no higher than its old one. The client warns and the
	// bot declines; the answer is still the player's. Plain data.
	NoGain bool

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
	// RetargetTo is a PINNED retarget's destination (#1743,
	// RetargetOffer.To): "change a target of target spell or ability
	// to this creature". Set, it turns the prompt's question from
	// "where does this slot go" into "which slot goes there" — the
	// PickTarget* options are the objects currently in the eligible
	// slots, and the answer names one of them. Zero on every free
	// retarget.
	RetargetTo TargetRef

	// RetargetNext carries the uncommitted prefix of a related-target walk.
	RetargetNext []TargetRef

	// pickTargetResume is the server-only continuation for a
	// PendingChoicePickTarget: the captured event / source / LKI,
	// the Build closure, and the spec the pick is validated against
	// (and stamped onto the item for the resolution re-check).
	// Added in S20 sub-PR 2. Written to a restore point as data since
	// #4361 S14b (pendingChoices[].targetWalk, trigger_frame_keyed.go):
	// the ability by its catalog row, its clauses re-derived from the
	// row, and the walk's answers and bound amounts carried.
	pickTargetResume *pickTargetFrame

	// PickTargetFor is, on a pick_target an opponent answers for an "of
	// an opponent's choice" clause (#2055), the player whose spell or
	// ability it is: the controller, who is still the player targeting
	// (CR 601.2c). uuid.Nil on every prompt the controller answers
	// themselves. Read by the wire (a bot choosing for somebody else's
	// effect prices the pick the other way round) and never by a rule.
	PickTargetFor uuid.UUID

	// PickTargetHarmful is, on a pick_target, the clause's Harmful flag
	// (#4249): the effect hurts the permanent picked, so a bot choosing
	// among its own gives up the least valuable. Read by the wire only.
	PickTargetHarmful bool

	// opponentTargetResume is the parked announcement behind an "of an
	// opponent's choice" clause (#2055, opponent_target.go): on the
	// opponent's pick_target it is what the answer makes again; on the
	// controller's choose-an-opponent option pick it is only a marker
	// (the option pick's own continuation runs the answer). Plain data.
	// Not serialised: a table waiting on it is not a restore point.
	opponentTargetResume *opponentTargetFrame

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

	// ModeUsedIndex / ModeUsedLabel are the options the ability has
	// ALREADY chosen under its "that hasn't been chosen" restriction
	// (ADR 0097), in printed order — not on offer, and never a legal
	// answer, but carried so the picker can show them greyed out
	// ("already chosen this turn"). ModeNotChosen says which of the
	// two restrictions applies. All three empty for an ordinary modal
	// trigger.
	ModeUsedIndex []int
	ModeUsedLabel []string
	ModeNotChosen ModeMemory

	// modePickResume is the server-only continuation for a
	// PendingChoiceModePick: the captured trigger, so the answer
	// continues into the CR 603.3d target walk and then Build.
	// Added by #764.
	modePickResume *modePickFrame

	// copyResume is the other continuation a
	// PendingChoicePickTarget can carry (S30, #95): the CR 707.10
	// "you may choose new targets for the copy" prompt. It reuses
	// the pick_target prompt rather than getting a kind of its own
	// because the QUESTION is identical — here is a target clause,
	// here is its legal set, pick within Min..Max — and every
	// consumer (the wire projection, the client picker, the bot's
	// choice enumerator) then needs no change at all to answer it.
	// What differs is only what gets built on submit, which is what
	// the frame decides. Exactly one of pickTargetResume and
	// copyResume is set. Written as data (#4361 S15,
	// pendingChoices[].copyWalk, copy_walk_keyed.go) when its item can
	// be written whole.
	copyResume *copyFrame

	// SacrificeOptions is the set of permanents a
	// PendingChoiceSacrifice's chooser may pick from — their own
	// permanents matching the effect's spec, computed when the
	// effect resolved. Wire-serialised so the client can highlight
	// exactly those cards. Re-checked on submit, since the board can
	// change between the prompt and the answer (an earlier player's
	// sacrifice can trigger something that removes a creature).
	SacrificeOptions []uuid.UUID

	// DiscardOptions is the set of cards in FromPlayer's revealed hand
	// a PendingChoiceDiscardFromHand's chooser may pick (ADR 0116 §3):
	// "you choose a nonland card from it". The card's filter is run
	// once, at the reveal, and only the matching instance IDs are kept
	// — a func cannot be written to a restore point, and a prompt with
	// no closure on it stays one. Re-checked on submit against the
	// hand.
	//
	// nil means every card in the hand. Only a restore point written
	// before ADR 0116 carries that; every prompt
	// QueueDiscardFromRevealedHand queues has a non-empty list.
	DiscardOptions []uuid.UUID

	// CastPile is a resolution_cast's cards passed over (#4243):
	// cascade's and discover's exiled cards that were not the hit. They
	// go to the bottom of the library in a random order once the prompt
	// is answered, together with the hit when it was not cast (CR
	// 702.85a). Plain data, never written in place, so a table paused on
	// the prompt is a restore point. Nil for the copies' prompt.
	CastPile []uuid.UUID

	// DiscardLabel names what DiscardOptions holds, the way the card
	// prints it — "nonland card", "card with mana value 3 or greater".
	// Client copy only; never read by the resolver.
	DiscardLabel string

	// PickDestination, PickOptional, PickFromGraveyard and PickThen
	// are the rest of a PendingChoiceRevealedHandPick (ADR 0116's
	// 2026-10-05 amendment, #2115): where the chosen card goes, whether
	// "you may choose" lets the chooser choose nothing, whether the
	// candidates include the revealing player's graveyard, and the KEY
	// of the registered continuation that runs once the pick is made
	// (revealed_hand_pick.go). All four are plain data, so a table
	// waiting on the pick is still a restore point. Zero on every
	// other kind, and on a discard_from_hand.
	//
	// PickThen is also a KEYED option_pick's continuation (#2854,
	// option_pick_keyed.go), looked up in the option-pick registry for
	// that kind. One field rather than a second one on purpose: every
	// v7 binary since #2115 refuses a restore point whose pickThen it
	// does not know, so a binary from before #2854, handed a file with
	// a keyed option pick open, refuses it and keeps it (the rollback
	// case) instead of restoring a question whose answer runs nothing.
	//
	// PickThen is also a keyed choose_cards pick's continuation (#2040),
	// a keyed outside-the-game collection pick's (#4256,
	// RegisterOutsidePick: the key names its filter too) and a keyed
	// pay-unless prompt's decline (#4256, RegisterPayUnlessElse).
	PickDestination   PickDestination
	PickOptional      bool
	PickFromGraveyard bool
	PickThen          string
	// PickMatchArg is the one word a keyed outside-the-game collection
	// pick's filter is handed with each typed card (#4256,
	// OutsidePickPrompt.MatchArg): Spawnsire of Ulamog's subtype, which
	// a text-changing effect may have changed. Plain data. Empty on
	// every other prompt.
	//
	// It is also the one word a keyed confirm hands its answer
	// (ConfirmPrompt.Word, #4291) and a keyed choose_cards pick's
	// CarryWord.
	PickMatchArg string
	// ChooseLiveZone is the zone a KEYED choose_cards pick's candidates are
	// re-checked against, on submit and by the sweep that withdraws a pick
	// whose candidates have all left it (#4291, ChooseCardsPrompt.LiveZone).
	// An entry card choice and a permanent pick read theirs off their kind
	// (liveZone). Empty on every other prompt, which means no live re-check.
	ChooseLiveZone ZoneKind
	// PickMeasures is RevealedHandDiscard.Measure's reading of each
	// candidate, in DiscardOptions order, taken as the pick was raised.
	// Nil when the card measures nothing.
	PickMeasures []int
	// PickExiledWith is the source OBJECT a PickExile pick links its
	// exiled card to (CR 607.2a, #3516): Null Summoner's "exile that
	// card ... you may cast the exiled card". Captured when the pick
	// goes up, because the prompt is answered after the ability has
	// resolved and the source may have changed by then (CR 400.7). The
	// zero value links nothing, which is every prompt raised before the
	// field existed. Plain data, so the prompt stays a restore point.
	PickExiledWith PermissionCardRef

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

	// ordered is the state of a KEYED ordered collection this prompt is one
	// seat's turn of (ordered_picks_keyed.go, #4290, ADR 0175): the seats
	// still to ask, the picks so far and the registered continuation. Plain
	// data, written to a restore point (pendingChoices[].ordered) and
	// cloned for an undo; nil on every other prompt.
	ordered *orderedPicks

	// CopyOptions is the set of permanents a PendingChoiceCopyTarget
	// may be copied from — "any creature on the battlefield" for
	// Clone, "a creature or planeswalker you control" for Spark
	// Double. Wire-serialised so the picker renders card faces;
	// battlefield cards are public, so there is nothing to redact.
	// Re-checked on submit against the live board, because a
	// creature can leave between the prompt and the answer.
	// Added in S16.5 (#159).
	CopyOptions []uuid.UUID

	// ExchangesText marks a PendingChoiceCopyTarget whose pick is the
	// creature to exchange text boxes with rather than the object to
	// copy (CopySelector.ExchangeText, ADR 0160 Decision 6). The answer
	// is the same; only the wording differs. Wire `exchanges_text`.
	ExchangesText bool

	// ExcludedTypes are creature types a PendingChoiceCreatureType may
	// NOT be answered with (ADR 0160's 2026-10-15 amendment): Artificial
	// Evolution's "The new creature type can't be Wall", and the word
	// being replaced, since a word can't be changed to itself. Canonical
	// spelling. The resolver refuses them, the enumerator and the wire's
	// type_options leave them out. Empty on every other prompt.
	ExcludedTypes []string

	// SecondName marks a PendingChoiceCardName whose answer is the
	// source's SECOND chosen name (Card.ChosenSecondName): Null
	// Chamber's opponent's name (#1997). False on every other prompt,
	// whose answer is ChosenName.
	SecondName bool

	// NotBasicLandName marks a PendingChoiceCardName that may not be
	// answered with a basic land card's name: Null Chamber's "a card
	// name other than a basic land card name" (CR 201.4a, the names
	// IsBasicLandCardName knows). The resolver refuses one, and the
	// enumerator and the wire's name_options leave them out.
	NotBasicLandName bool

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

	// LibraryPlacement is which lanes a PendingChoicePutInLibrary's
	// answer may use — top, bottom, or either (ADR 0088). Plain data,
	// carried by the clone and the snapshot: the prompt has to mean
	// the same thing after an undo and after a restart. Empty on every
	// other kind.
	LibraryPlacement LibraryPlacement

	// LibraryTopCount / LibraryTopDepth refine a put_in_library's top
	// lane (#1298, ADR 0088's 2026-09-23 amendment): the count is
	// EXACTLY how many cards the answer's top lane must hold (Cream of
	// the Crop's "put one of those cards on top"; zero is any number),
	// and the depth is where that lane starts, counted from the top
	// (Temporal Cleansing's "second from the top"; zero and one are
	// the top). Plain data, carried by the clone and the snapshot like
	// LibraryPlacement. Zero on every other kind.
	LibraryTopCount int
	LibraryTopDepth int

	// TriggerOrderIDs is the set of pending-trigger item IDs a
	// PendingChoiceTriggerOrder entry asks the chooser to order.
	// Wire-serialised (with label + source per ID) so the client
	// renders the reorder list; the resolve payload returns the same
	// IDs in resolution order. Added in S19 sub-PR 8.
	TriggerOrderIDs []uuid.UUID

	// TriggerOrderWhy is, for a PendingChoiceTriggerOrder, the one
	// sentence that says why the prompt was asked, in the player's
	// terms (#3015, trigger_order_why.go). Display text only: nothing
	// reads it back, and it never changes whether the seat is asked.
	// Empty on every other kind.
	TriggerOrderWhy string

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

	// PayActionClause is a pay-unless's non-mana payment (#4264):
	// "discard a card", "sacrifice two lands", "tap an untapped
	// creature". Plain data, written to a restore point, so the prompt
	// carries the payment without a frame (payUnlessFrameOf reads it
	// back). Nil on every mana prompt.
	PayActionClause *PayAction

	// PayTap is a pay-unless's waterbend clause (#4264): the kind of tap
	// clause and the part of the cost its taps may cover, which
	// RegisterTapCost's builder turns back into the clause
	// (payUnlessFrameOf, PayTapCost). Plain data. Nil on every plain
	// payment.
	PayTap *PayTapSpec

	// OptionObjects are the objects, with the epoch each had, that a KEYED
	// pay-unless or any-player prompt's branches are handed beside
	// OptionCarry's IDs (#4264): a permanent the card named as the offer
	// was made, so a branch can tell it from a new object the same card
	// has become (CR 400.7). Plain data.
	OptionObjects []ObjectRef

	// Asked is the stack item that queued a KEYED pay-unless or
	// any-player prompt, as plain data (#4264, asked_item.go): its
	// controller, the card and object it came from, its targets, its
	// triggering event and its X. A keyed branch is handed it as the item
	// it runs against, so it reads what a closure over the item used to
	// (CR 400.7 included: the object's epoch rides it). Nil on a framed
	// prompt and on a keyed one whose branches read only the payer, the
	// card and the carried IDs.
	Asked *AskedItem

	// SearchCards is the set of library cards a
	// PendingChoiceSearchLibrary's chooser may take, in library
	// order. Wire-serialised via PendingChoiceView.Options — and,
	// unlike every other Options-bearing kind, withheld from
	// non-choosers entirely, because the length of this list is
	// itself hidden information about a hidden zone.
	SearchCards []uuid.UUID

	// SearchGraveyardCards is the subset of SearchCards that sits in
	// the searcher's GRAVEYARD, for the "search your library and/or
	// graveyard" cards (#3217, ADR 0015's 2026-10-13 amendment). Empty
	// for a library-only search, which is every search but those.
	//
	// It is what makes the shuffle zone-aware: a take wholly from the
	// graveyard means the library was never searched. It does not raise
	// the floor ("and/or" lets the searcher look only in the library, so
	// a stated-quality search may still find nothing); it also tells the
	// client which tiles to caption. Wire-serialised as PendingChoiceView.
	// SearchGraveyard, and withheld from non-choosers with the rest of
	// the search prompt.
	SearchGraveyardCards []uuid.UUID

	// SearchHandCards is the subset of SearchCards that sits in the
	// searched player's HAND, for "search that player's graveyard, hand,
	// and library" (#3988, Kotose, the Silent Spider). Empty for every
	// other search. A candidate is checked against the zone it was
	// offered from, as for SearchGraveyardCards. Wire-serialised as
	// PendingChoiceView.SearchHand, chooser-only.
	SearchHandCards []uuid.UUID

	// SearchMax is how many of SearchCards the chooser may take.
	SearchMax int

	// SearchMin is how many of SearchCards the chooser MUST take
	// (#3053). Zero — the common case — is CR 701.23b: a search of a
	// hidden zone for a card with a stated quality may find nothing
	// however many matches there are. It is above zero only for an
	// unconditional search (CR 701.23d: "a card", "three cards"), where
	// the searcher has to find that many, or as many as the library
	// holds. See SearchLibrarySpec.floorFor.
	SearchMin int

	// searchResume is the server-only continuation for a
	// PendingChoiceSearchLibrary: the whole SearchLibrarySpec's data,
	// which carries the destination, the reveal / shuffle / tapped
	// flags and the rest of the effect as a keyed tail (Fabled
	// Passage's "then untap that land", Gamble's random discard). Not
	// serialised to the wire; written to a restore point
	// (pendingChoices[].searchPrompt). Data only since #4361 S4.
	searchResume *searchResumeFrame

	// MayCastCard is the one card a PendingChoiceMayCast is offering
	// — the cascade hit. Wire-serialised as the choice's single
	// Options entry so the client's prompt can show the card face
	// instead of quoting a name into a sentence.
	MayCastCard uuid.UUID

	// MayCastKeyword names the rule a PendingChoiceMayCast is asked
	// under — "cascade", "discover", "suspend", "madness" — so the
	// client can name it and word the decline (ADR 0099 §7). Empty for
	// an offer whose card text is the whole story (hideaway, Malcolm).
	// The branch names ride AcceptLabel / DeclineLabel, as a confirm's
	// do. Wire-serialised.
	MayCastKeyword string

	// A PendingChoiceMayCast's answer is a registered key on PickThen
	// (may_cast_keyed.go, #4361), with OptionCarry its data.

	// LoopShortcutKey / LoopShortcutCount / LoopShortcutRepeat carry a
	// PendingChoiceLoopShortcut's question (#804, CR 732).
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
	// costs no life, which is most of them. A PendingChoiceEntryPayLife
	// carries its payment here too (a shockland's 2, #2390). PayUnless
	// uses it for mandatory life paid alongside mana (#2945), beside the
	// PayCost text the client shows.
	//
	// It is on the choice, and on the wire, for the reason #547 put
	// legal.MoveCost.Life there: a policy holding only the wire
	// payload otherwise prices "pay 4 life to keep this card" exactly
	// like "shuffle your library", and a bot at 4 life answers yes and
	// dies. The engine does NOT deduct it — the frame's accept branch
	// does (or the shared pay-unless payment does), and re-checks
	// affordability when it runs, because life can move between the question and the answer.
	LifeCost int

	// PoisonCost is the poison counters a PendingChoiceConfirm's ACCEPT branch
	// gives the chooser (#4613, "Ward—Get five poison counters"). It is
	// LifeCost's sibling, on the choice and the wire for the same reason: a
	// policy holding only the wire payload would otherwise price "get five
	// poison counters" like "shuffle your library". The engine does not
	// place them here; the frame's accept branch does. Zero is no counters.
	PoisonCost int

	// DeclineEndsLoop marks a confirm whose ACCEPT would continue a loop
	// the table is already in (#2040): Channel Harm asked to deal damage
	// that it prevented from another Channel Harm. Declining is how the
	// ruling says such a loop ends. It changes nothing the engine
	// accepts; it is a hint on the wire for a policy that would
	// otherwise take every offer and ping-pong forever.
	DeclineEndsLoop bool

	// private marks a prompt whose WORDS name a card in a hidden zone
	// (ADR 0133: the question an opening-hand action asks names the
	// card in the chooser's hand). The wire keeps the prompt — the
	// table has to see that a seat owes an answer — and strips what it
	// says from every viewer but the chooser (PrivateText, view.go).
	// Serialised as Private (#4298): the opening-hand offer is a keyed
	// confirm, so a restore point holds it, and a restored table must not
	// name the card in a hand to every viewer.
	private bool

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

	// CollectionSuggestions are the card NAMES a
	// PendingChoiceOutsideCollection suggests (ADR 0198): cards in the
	// database that satisfy the effect's filter, in the collection's
	// suggestion order. Plain data. The client shows them before the
	// player searches, and the enumerator offers them as the answers a
	// bot can give. Empty for every other kind.
	CollectionSuggestions []string

	// ChoosePlayers are the SEATS a PendingChoiceProliferate offers
	// alongside ChooseCards (CR 701.34a: "permanents and/or players").
	// A seat is named in the answer's card_ids by its player ID, so
	// the one payload every card-set pick shares still carries it.
	// Empty for every other kind.
	ChoosePlayers []uuid.UUID

	// ChooseSuggested is the engine's recommended answer, a subset of
	// ChooseCards and ChoosePlayers (#2525). It is a default and not a
	// rule: the client pre-selects it, the enumerator always offers it
	// as one move, and the bot answers with it. Empty for every kind
	// that has no recommendation.
	ChooseSuggested []uuid.UUID

	// ChooseDestination says where a PendingChoiceChooseCards sends
	// what is named, when the card says so (#2680): "battlefield" for
	// "put a land card from your hand onto the battlefield", and
	// "battlefield_tapped" for the same clause with "tapped". Empty is
	// a prompt that declares nothing, which is every other choose_cards
	// (a discard, a loot's half, a library look). Plain data for the
	// bot, the same shape as ColorPurpose and #2115's PickDestination:
	// a policy that cannot see where a named card goes prices putting
	// a land onto the battlefield as discarding it. The resolver never
	// reads it; the continuation is what moves the card.
	ChooseDestination ChooseDestination

	// PickOptions are the branches of a PendingChoiceOptionPick — the
	// card's own words for each, plus the cards it is about (a pile,
	// or nothing). Wire-serialised via PendingChoiceView.PickOptions,
	// with each option's card list projected through the same
	// per-viewer redaction as every other Options-bearing kind. See
	// option_pick.go.
	PickOptions []ChoiceOption

	// OptionCarry is the plain data a KEYED option pick's continuation
	// is handed with the answer (#2854, option_pick_keyed.go); the key
	// itself is PickThen. Written to a restore point. Since #4361 S14b
	// every option pick is keyed, so none holds a closure. Empty on
	// every other prompt.
	OptionCarry []uuid.UUID

	// PayCauseController and PayCauseItem are the cause (#4279) a
	// KEYED pay-unless was queued under: the controller and the stack
	// item ID of the spell or ability whose resolution asked, captured
	// while that resolution was open. A pay-unless does not stop the
	// table, so it can be answered after Game.resolving has moved on;
	// the decline stamps THIS cause on a discard or a sacrifice, not
	// whatever the slot holds then. The source card is Source. Zero
	// when the prompt was queued outside a resolution.
	//
	// #4281: any prompt a keyed branch queues while it runs under such a
	// cause takes it too (QueueChoiceForEffect), and its answer (a keyed
	// confirm, option pick or choose-cards, a discard or a sacrifice)
	// runs under it, so a chain of follow-ups keeps the first cause.
	PayCauseController, PayCauseItem uuid.UUID

	// CouncilVote makes an option pick a BALLOT (CR 701.38, ADR 0146,
	// council_vote.go): the vote so far, which the answer is cast into
	// before the next voter is asked. Plain data, written to a restore
	// point and shown to every seat; nil on every other prompt.
	CouncilVote *CouncilVote

	// AnyPlayerOffer makes a pay_unless one LINK of an "unless any
	// player pays" offer (ADR 0150, any_player_pays.go): whose effect
	// it is, the order the table is asked in and who has declined.
	// Plain data, shown to every seat; nil on every other prompt.
	AnyPlayerOffer *AnyPlayerOffer

	// Assist makes a pay_amount an assist helper's prompt (ADR 0150 §4,
	// assist.go): who is casting what, and the generic it has. Plain
	// data, shown to every seat (the caster's waiting line reads it).
	// Written to a restore point (#4294) with the parked cast it re-runs;
	// the prompt names the engine's own branch (assistHelpThen).
	Assist *AssistPrompt

	// TurnSkipOrder makes a prompt the which-skip question of a paused
	// turn change (ADR 0188 Decision 12, turn_skip_order.go). Plain
	// data, written to a restore point and shown to every seat; nil on
	// every other prompt.
	TurnSkipOrder *TurnSkipOrder

	// scryThen is the continuation for a PendingChoiceScry or a
	// PendingChoiceSurveil (#4253 slice 3): a registered tail run once the
	// look-at is done, which must not run until the player has finished
	// it. Preordain's "Scry 2, THEN draw a card" is the case that forces
	// it — the card you leave on top is the card you draw, so a draw that
	// happens before the reorder is a different card. A scry or surveil
	// with a "then" is a restore point while it waits. Written to a
	// restore point (pendingChoiceSnapshot.ScryThen). Set one or neither.
	scryThen KeyedTail

	// pickTail is a keyed choose_cards pick's registered tail as data
	// (#4253 slice 8): the rest of a keyword action the pick is the middle
	// of (amass's "choose an Army" hands CR 701.47c's clause on). Written
	// to a restore point (pendingChoiceSnapshot.PickTail); it never touches
	// PickThen, which is the pick's own key.
	pickTail KeyedTail
	// nameThen is a resolution-time card-name prompt's "then" as data
	// (#4253 slice 10c): a registered tail told the chosen name as its
	// Word. Written to a restore point (pendingChoiceSnapshot.NameThen), its
	// key riding PickThen. The zero tail is the stored form (the name lands
	// on the source). Since the sweep's second pass (#4253) a
	// resolution-time colour or creature-type prompt carries one too
	// (value_then_keyed.go); since #4361 always, valueChosenNothing when
	// nothing follows the answer.
	nameThen KeyedTail

	// sourceShield is a choose_source prompt's consequence when it is the
	// engine's own (#4253 slice 10c, choose_source_keyed.go): the shield the
	// card described, completed with the source once it is pinned. Written
	// to a restore point (pendingChoiceSnapshot.SourceShield), its
	// continuation's key riding PickThen. Nil on every other prompt.
	sourceShield *SourceShield

	// libraryOrderKeyed is a PendingChoicePutInLibrary's continuation as
	// data (#4301): how the cards move and the registered tail that is the
	// rest of the effect. Handed the validated answer (both lanes
	// top-first), it places the cards, then runs that tail. Written to a
	// restore point (pendingChoiceSnapshot.LibraryOrder), its key riding
	// PickThen. (#4361 S17 deleted the closure form, libraryOrderResume.)
	// See library_order.go.
	libraryOrderKeyed *libraryOrderKeyed

	// dice is a paused dice roll (#4307, dice_roll.go): the roll, the
	// rolls still in play and the choosers still to answer, plain data
	// beside the dice flow's keyed option pick. Written to a restore point
	// (pendingChoiceSnapshot.DiceRoll). Never mutated once queued, so a
	// clone shares it.
	dice *diceRollPause

	// midResolution says the prompt was queued while a stack item was
	// resolving and is part of that resolution (#1289, CR 608.2 /
	// CR 704.3): while it is open and blocks the table, the resolution
	// has not finished, so state-based actions and the trigger drain
	// wait for it. Stamped by QueueChoiceForEffect, never by a caller.
	// Carried by the snapshot. See resolution_pause.go.
	midResolution bool

	// AutoAnswerKey is the key a standing answer to this prompt is
	// filed under (ADR 0127 §2), or "" when the prompt is not covered
	// (§1). Derived by the server when the prompt is queued and opaque
	// to the client. AutoAnswerCard and AutoAnswerPrompt are its
	// display copies: the asking card's name and the question, which
	// the client stores beside the rule. All three are on the
	// chooser's view only. Carried by the clone and the snapshot.
	AutoAnswerKey    string
	AutoAnswerCard   string
	AutoAnswerPrompt string

	// AskedByHand marks a prompt with a standing answer that is asked
	// by hand instead, and says why (ADR 0127 §4): the room never
	// answers it automatically again. Empty on every other prompt.
	AskedByHand AskedByHand
}

// payUnlessFrame carries a pay-unless prompt's parsed cost and the
// consequence of each answer. Both callbacks receive the live *Game
// (not a captured one) for the same undo-safety reason
// StackItem.Effect does. Added in S19 sub-PR 6.
//
// No prompt holds one (#4361): payUnlessFrameOf builds it from the
// prompt's data each time the prompt is read, its branches looked up by
// the prompt's key, so a table waiting on any pay-unless is a restore
// point.
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

	// tap is the CR 701.67 "you can tap your artifacts and creatures
	// to help" clause of a waterbend payment — "Ward—Waterbend {4}"
	// (The Unagi of Kyoshi Island, #1311). Nil for every other
	// pay-unless, which is nearly all of them. Its Extra names the
	// part of `cost` the taps may cover; see waterbend_cost.go.
	tap *TapPermanentsCost

	// action is a non-mana payment — "discard a card", "sacrifice two
	// lands" (ADR 0108 §5, pay_unless_action.go). Nil for a mana
	// payment. When set, `cost` is empty and the prompt's PayCost is
	// the payment's printed words.
	action *PayAction
}

// PrivateText reports whether the prompt's words name a card in a
// hidden zone (ADR 0133), so the protocol view strips them from every
// viewer but the chooser.
func (c *PendingChoice) PrivateText() bool { return c != nil && c.private }

// PayTapCost is the waterbend clause of a PendingChoicePayUnless —
// the permanents the chooser may tap instead of paying some of the
// generic mana (#1311) — or nil when the payment is mana only. Read by
// the protocol view (the picker) and the legal-move enumerator (the
// bot's payment); ResolvePayUnlessWithTaps is the one validator.
func (c *PendingChoice) PayTapCost() *TapPermanentsCost {
	if c == nil {
		return nil
	}
	// The prompt rebuilds its clause from PayTap (#4264).
	return tapCostOf(c.PayTap)
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
	// ability is the declaration being put on the stack: its Build
	// and Effect make the item once every step is answered, its
	// Watches decide whether the item is stamped with its triggering
	// event (see stampTriggerContext on why an ability that watches
	// nothing — a CR 603.12 reflexive trigger — must not be), and its
	// catalog row names the item (ADR 0041 P9, tier 4-2).
	ability TriggeredAbility
	// tc is the triggering event (#1223), carried rather than
	// re-derived: the CR 603.10 snapshot it holds is deleted from
	// lastKnownBattlefield as the harvest ends, which is before this
	// prompt can possibly be answered. tc.Event is the `ev` this
	// frame used to hold.
	tc     TriggerContext
	source Card
	lki    Characteristic
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
	// dist is the division announced so far (#1563, CR 601.2d via CR
	// 603.3d): each answered divided step's shares, settled by the
	// same gate a cast uses, stamped onto the item as its
	// Distribution when the walk finishes.
	dist map[uuid.UUID]int

	// choosers are the opponents named for the walk's "of an opponent's
	// choice" steps so far (#2055, opponent_target.go), stamped onto the
	// item as its TargetChoosers when the walk finishes. Never written
	// into steps, which an undo clone shares: boundSteps binds a copy.
	choosers []TargetChooser

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

// boundSteps is the walk's steps with the opponents named so far bound
// onto their clauses (#2055): a fresh copy, since steps is shared.
func (f *pickTargetFrame) boundSteps() []AnnouncedClause {
	if len(f.choosers) == 0 {
		return f.steps
	}
	out := append([]AnnouncedClause(nil), f.steps...)
	bindStepChoosers(out, f.choosers)
	return out
}

// stepChooser is the opponent named for step i, or uuid.Nil.
func (f *pickTargetFrame) stepChooser(i int) uuid.UUID {
	if i < 0 || i >= len(f.steps) {
		return uuid.Nil
	}
	for _, ch := range f.choosers {
		if ch.Mode == f.steps[i].Mode && ch.Slot == f.steps[i].Slot {
			return ch.Player
		}
	}
	return uuid.Nil
}

// triggerResumeFrame stashes the per-trigger continuation data the
// harvester captured at OptionalPrompt-queue time. The Build closure
// fires on `apply: true` against the value-copy source + LKI; on
// `apply: false` the frame is discarded. Added in S19 sub-PR 2.
type triggerResumeFrame struct {
	// tc is the triggering event (#1223) — see pickTargetFrame.tc
	// for why it is carried whole rather than re-read on the answer.
	tc     TriggerContext
	source Card
	lki    Characteristic
	// ability is the full declaration so a "yes" on a TARGETED
	// optional trigger can continue into the pick_target step
	// (S20 sub-PR 2) instead of building straight away — and so the
	// item it builds can be named by its catalog row (ADR 0041 P9).
	ability   TriggeredAbility
	doubledBy doublerRef
}

// TradeSubject is the object a "you may" trigger would trade its
// source for — the triggering event's card, on a prompt whose
// ability declares TriggerOptionalPrompt.Trade (Perplexing Chimera's
// spell, ADR 0104). uuid.Nil on every other prompt.
func (c *PendingChoice) TradeSubject() uuid.UUID {
	if c == nil {
		return uuid.Nil
	}
	if c.OptionalItem != uuid.Nil {
		return c.OptionalTradeFor
	}
	if c.triggerResume == nil {
		return uuid.Nil
	}
	p := c.triggerResume.ability.OptionalPrompt
	if p == nil || !p.Trade {
		return uuid.Nil
	}
	return c.triggerResume.tc.Event.CardID
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

// SharedTargetKey is the identity of the trigger a pick_target prompt
// asks about, as card and ability rather than as source object (#3817,
// ADR 0127's 2026-10-15 amendment): the source's catalog key, the
// triggered row's index and label, and the clause being asked. Every
// copy of one card (twenty Blood Artists, a Clone of one) shares it;
// two abilities of one card, or two clauses of one ability, do not.
// The client offers "same target for all of this card's triggers" over
// the prompts that share it and the same chooser.
//
// "" when the prompt is not a trigger's target pick, when the row is not
// a catalog row (an engine-built or reflexive trigger), or when the
// source is face down, so nothing is shared across an unknown identity.
func (c *PendingChoice) SharedTargetKey() string {
	if c == nil || c.Kind != PendingChoicePickTarget || c.copyResume != nil {
		return ""
	}
	f := c.pickTargetResume
	if f == nil || f.source.FaceDown {
		return ""
	}
	t := f.ability
	if t.row.key == "" || t.Key == "" {
		return ""
	}
	return t.row.key + "|" + AbilitySlotTriggered + "|" + strconv.Itoa(t.row.index) + "|" + t.Key +
		"|step" + strconv.Itoa(f.step)
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
	// BlockerDivides marks the CR 510.1d prompt (#1706): the source is
	// a BLOCKER that blocks two or more attackers, AttackerID names the
	// blocker, BlockerIDs the attackers it blocks, and the chooser is
	// the blocker's controller. Its damage is "divided as its controller
	// chooses among them", and trample does not apply. Zero in every attacker's prompt,
	// and in a frame restored from before the field, which is what
	// those frames were.
	BlockerDivides bool `json:",omitempty"`
	// AsThoughUnblocked marks a prompt whose attacker may assign all of its
	// damage "as though it weren't blocked" (CR 510.1c, Wolverine, Claws
	// Out): read off Characteristic.AssignsAsThoughUnblocked as the prompt
	// was queued. Besides the ordinary division among BlockerIDs the
	// answer may put the whole AttackerPower on the player, planeswalker
	// or battle the attacker is attacking: every blocker 0 and
	// trample_to_player the full power, with no lethal damage owed to any
	// blocker and no trample needed. BlockerIDs may be empty (every
	// blocker has left combat): the attacker then assigns nothing or all
	// of it to what it attacks. Zero in every other prompt, and in a
	// frame restored from before the field, which is what those were.
	AsThoughUnblocked bool `json:",omitempty"`
	// HasDeathtouch is true when the attacker has deathtouch (CR
	// 702.2c — 1 damage is lethal). The server uses this to relax
	// trample's lethal-to-every-blocker rule (CR 702.19b): 1 damage
	// satisfies the threshold regardless of the blocker's remaining
	// toughness.
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

	// SourceInfect, SourceWither and SourceToxic are the attacker's
	// ADR 0056 damage-result keywords at prompt-queue time (CR 702.90,
	// 702.80, and the CR 702.164b toxic total), cached for the same
	// died-before-resume reason as SourceLifelink: an infect attacker
	// killed by first-strike damage before its assignment is answered
	// still puts -1/-1 counters and gives poison. Copied onto the tail
	// by damageTailFromFrame.
	//
	// Zero in a frame restored from a snapshot written before the
	// fields existed, which resumes as ordinary damage — the fallback
	// CombatStep has, and why there is no schema bump. Server-side
	// only: not projected onto DamageAssignmentView.
	SourceInfect bool `json:",omitempty"`
	SourceWither bool `json:",omitempty"`
	SourceToxic  int  `json:",omitempty"`
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
	// #4281: a prompt queued by a keyed branch that runs under a carried
	// cause (a keyed pay-unless's decline, or a follow-up of one) takes
	// that cause with it, so the answer, given after the table moved on,
	// still names the spell or ability that asked.
	if choice.PayCauseController == uuid.Nil && g.carriedCause != nil && g.carriedCause.Controller != uuid.Nil {
		choice.PayCauseController, choice.PayCauseItem = g.carriedCause.Controller, g.carriedCause.Item
	}
	// #1289: a prompt a resolution asks is part of that resolution.
	choice.midResolution = g.resolutionOpen && choiceBelongsToResolution(&choice)
	// ADR 0127 §2: a covered pay_unless or confirm is filed under the
	// key of the resolution or branch that asks it.
	g.keyResolutionPromptLocked(&choice)
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
	if choice.Kind == PendingChoiceRevealedHandPick {
		// #2115: an optional pick takes an empty answer, so it is
		// routed ahead of the exact-count check below.
		return g.resolveRevealedHandPickLocked(idx, choice, picks)
	}
	if choice.Kind == PendingChoiceControlledPlay {
		// ADR 0166 §4: the empty answer, "it can't be played". The
		// play itself is ControlledPlayCast.
		return g.resolveControlledPlayLocked(idx, choice, picks)
	}
	if choice.Kind == PendingChoiceResolutionCast {
		// #2947: the empty answer, "cast no more of them". The casts
		// are CastForResolutionChoice.
		return g.resolveResolutionCastLocked(idx, choice, picks)
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
		// ADR 0116 §5: every pick must be one the card allows — the
		// server never relies on the client or the bot to have offered
		// only those — and no card may be named twice.
		if !discardOptionsAllow(choice, picks) {
			return ErrInvalidParam
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
		cause := carriedCauseOfChoice(choice) // #4281
		g.dequeueChoiceLocked(idx)
		return g.withCarriedCauseLocked(cause, func() error {
			return g.discardCardsLocked(from.ID, picks, discardOptions{
				cause:  DiscardCauseEffect,
				source: choice.Source,
			})
		})
	default:
		return ErrInvalidParam
	}
}

// nextManaPickLocked rewrites an answered pick in place for the next
// pick its activation owes (#3939): `produced` joins ManaSettled, and the
// choice takes the next slot's options, amounts and different-colours
// count, with a fresh ID so every seat sees a new question — the way
// advanceDifferentColorsPickLocked re-asks a different-colours pick.
//
// Caller must hold g.mu.
func (g *Game) nextManaPickLocked(choice *PendingChoice, produced []string) {
	choice.ManaSettled = append(append([]string(nil), choice.ManaSettled...), produced...)
	next := choice.ManaPicksLeft[0]
	choice.ManaPicksLeft = copyManaPickSlots(choice.ManaPicksLeft[1:])
	choice.ColorOptions = append([]string(nil), next.Options...)
	choice.ManaChosen = nil
	choice.ManaDifferent = next.Different
	if next.Different > 0 {
		choice.ManaAmounts = nil
		choice.Reason = differentColorsReason(choice.ManaLabel, nil, next.Different)
	} else {
		choice.ManaAmounts = copyManaAmounts(next.Amounts)
		choice.Reason = choice.ManaLabel
	}
	choice.ID = uuid.New()
	g.notePlayerDecisionLocked()
}

// ManaPickSlot is one colour pick an activation still owes after the
// pick open now (PendingChoice.ManaPicksLeft, #3939): the options the
// activation offered, already narrowed and ordered, the per-colour
// amounts of a "N mana of any one color" pick, and N for an "N mana of
// different colors" pick.
type ManaPickSlot struct {
	Options   []string       `json:"options"`
	Amounts   map[string]int `json:"amounts,omitempty"`
	Different int            `json:"different,omitempty"`
}

// copyManaPickSlots deep-copies a pick queue for the clone and the
// snapshot.
func copyManaPickSlots(in []ManaPickSlot) []ManaPickSlot {
	if len(in) == 0 {
		return nil
	}
	out := make([]ManaPickSlot, len(in))
	for i, s := range in {
		out[i] = ManaPickSlot{
			Options:   append([]string(nil), s.Options...),
			Amounts:   copyManaAmounts(s.Amounts),
			Different: s.Different,
		}
	}
	return out
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
	produced := repeatColor(color, n)
	// #2558: "N mana of different colors" mints nothing until its last
	// colour is named, and then all N together.
	if choice.ManaDifferent > 0 {
		all, done, err := g.advanceDifferentColorsPickLocked(choice, color)
		if err != nil {
			return err
		}
		if !done {
			return nil
		}
		produced = all
	}
	// #3939, CR 106.12b: one activation is one production. While the
	// activation still owes picks, this answer's colours join the ones
	// it already knows and the choice is asked again for the next pick;
	// the last answer produces the whole output in one window.
	if len(choice.ManaPicksLeft) > 0 {
		g.nextManaPickLocked(choice, produced)
		return nil
	}
	if len(choice.ManaSettled) > 0 {
		produced = append(append([]string(nil), choice.ManaSettled...), produced...)
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
		produced,
		copyRestrictions(choice.ManaRestrictions),
		// #1547: and its spend riders, carried the same way.
		choice.ManaRiders,
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
			}, nil, nil)
		}
	}
	return nil
}

// dequeueChoiceLocked drops the choice at index idx because its
// chooser ANSWERED it, preserving slice order for the rest. Every
// Resolve* path ends here.
//
// #628: answering a prompt is a player decision, so it restarts the
// CR 732 loop run. The engine's own prune paths call
// dropChoiceLocked instead — a choice the engine withdrew is not a
// decision anybody made, and counting it as one would let a loop
// that queues and prunes a prompt each iteration run forever.
//
// Caller must hold g.mu.
// The drop comes FIRST. notePlayerDecisionLocked withdraws the CR 732
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
// prompt, a CR 732 shortcut whose notice was cleared, an option pick
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
	case dropSettle:
		g.settleDroppedDivideShieldLocked(c)
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

// queueOptionalReplacementPromptLocked queues a "may" yes/no
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
	if q := g.redirectQuestionLocked(ev, chosen); q != "" {
		// #2056: a "may" redirection says which damage it is asking
		// about, as the event stands now.
		reason = q
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
	if (chosen.effect.entryKeyword != "" || chosen.effect.FromGraveyard) && chosen.source != nil {
		// Unleash (#1556): the question is about the entering card, so
		// the prompt names it, as entry_riot's does. A dredge offer is
		// about the card in the graveyard (#2127).
		choice.Source = chosen.source.InstanceID
	}
	if chosen.effect.commanderZone && ev != nil && ev.CardID != uuid.Nil {
		// CR 903.9b (ADR 0115 decision 2): the question is about the
		// commander that would go to a hand or a library, and the
		// built-in has no source card of its own, so the prompt names
		// the moving card, as the cost path's does.
		choice.Source = ev.CardID
	}
	g.QueueChoiceForEffect(choice)
}

// optionalReplacementChooserLocked is who answers a "may"
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

// offerOptionalReplacementLocked handles an applicable
// "may" replacement. Returns true when the yes/no prompt was queued
// and the caller must stop where it is — errReplacementPending from
// the apply-loop, a bare nil from the chosen-order resume — which is
// the pause-and-bail contract offerCopyChoiceLocked and
// offerEntryLifePaymentLocked already have.
//
// Returns false — the "may" is DECLINED, and marked applied so the
// apply-loop does not gather it a second time (CR 614.5: the
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
	if parked := choice.costCommanderResume; parked != nil {
		// #1397: a CR 903.9 question asked BEFORE a cost is paid.
		// There is no paused event behind it — the announcement it is
		// about has not been made yet — so the answer re-makes it.
		g.dequeueChoiceLocked(idx)
		return g.resolveCostCommanderChoiceLocked(parked, chooserID, apply)
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
	// re-evaluate this effect again for this event (CR 614.5: the
	// decision is once per event).
	g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
	// #3482: a "yes" fires the effect of the object that offered it,
	// found again by that object rather than by its battlefield slot;
	// if it has left, its effect has gone with it.
	if live, present := g.liveReplacementSourceLocked(chosen); apply && present {
		g.runReplaceLocked(ev, live)
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
// CR 903.9 window open on every exit "from anywhere", this was the
// ordinary way a commander left, not a corner. After ADR 0115 a
// commander that dies leaves first and is asked afterwards, so this
// boundary still matters for a bounce or a tuck (CR 903.9b) and for
// any other prompt that pauses an exit.
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
	case RepEventRollDice:
		// #4307. The player who would roll (CR 616.1). The window sets
		// mustSettleNow, so this is never asked; the dice flow asks the
		// roller the order of the choosers itself (dice_roll.go).
		return ev.RollPlayer
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
		// CR 701.9a puts it into THAT PLAYER'S graveyard. The
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
		// #569. CR 701.17a: the player who mills is the one whose
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
	case RepEventCounterSpell:
		// #4056. CR 616.1: the affected object is the spell about to be
		// countered, and its controller chooses, not the controller of
		// whichever "if you would counter" replacement gathered first
		// (that player is the one doing the countering).
		if item := g.StackMeta[ev.CounterSpellID]; item != nil {
			if item.Controller != uuid.Nil {
				return item.Controller
			}
			return item.Owner
		}
	case RepEventExtraTurn:
		// #2529. The player whose extra turn it is. The prompt this
		// would order is never put to them (the window sets
		// mustSettleNow), so the arm is the honest answer to "who would
		// be asked" and the value a pausing replacement would need.
		if ev.ExtraTurnSeat >= 0 && ev.ExtraTurnSeat < len(g.Seats) {
			return g.Seats[ev.ExtraTurnSeat].ID
		}
	case RepEventProduceMana:
		// #1222. CR 106.12b: the mana is produced for a player, and
		// that player is the affected one whoever controls the
		// replacements — the same reading the mill arm above takes.
		//
		// The prompt it would order is never actually put to them: a
		// production sets mustSettleNow (CR 605.3b), so the apply-loop
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
	var inst DamageInstance
	if c := g.pendingChoiceByIDLocked(choiceID); c != nil && c.replacementResume != nil && c.replacementResume.ev != nil {
		inst = c.replacementResume.ev.DamageInstance
	}
	if err := g.resolveReplacementOrderLocked(choiceID, chooserID, ordered); err != nil {
		return err
	}
	// #2066: the same answer settles the instance's other events.
	g.settleSiblingRedirectOrdersLocked(inst, chooserID, ordered)
	return nil
}

// pendingChoiceByIDLocked is the queued choice with the given ID, or nil.
//
// Caller must hold g.mu.
func (g *Game) pendingChoiceByIDLocked(id uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == id {
			return c
		}
	}
	return nil
}

// resolveReplacementOrderLocked is ResolveReplacementOrder under the lock.
//
// Caller must hold g.mu (write).
func (g *Game) resolveReplacementOrderLocked(choiceID, chooserID uuid.UUID, ordered []ReplacementEffectID) error {
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
	g.rememberRedirectOrderLocked(ev, frame.applicable, chooserID, ordered)
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}
	for _, id := range ordered {
		// Expanded draws are individual replacement descendants. Settle
		// the first before another ordered effect examines its siblings.
		g.splitExpandedDrawLocked(ev)
		chosen, ok := applicableByID[id]
		if !ok {
			continue
		}
		if ev.Canceled {
			// A prior Cancel short-circuits the remaining chain.
			break
		}
		// #3482: the source as it is now, found by the object it was
		// gathered from, not by the battlefield slot the gather read it
		// at. One that has left takes its effect with it; a
		// card that came back is a new object (CR 400.7), gathered
		// afresh by the apply-loop re-entry below if it applies.
		chosen, ok = g.liveReplacementSourceLocked(chosen)
		if !ok {
			continue
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
		// An effect with a CHOICE inside it can't be fired blind: a
		// Clone's controller still picks what it copies, a shockland's
		// still answers "pay 2 life?", a reveal-land's and a Mox
		// Diamond's still pick from hand, and a "may" (#847) is still
		// offered, when a Kismet is also replacing the entry. Queue
		// that prompt and bail; the effects later in the chosen order
		// are still unapplied, so the apply-loop re-entry after the
		// answer picks them up. One dispatcher with the apply-loop
		// (offerOwnQuestionLocked), because this list had already lost
		// the reveal-land branch once (ADR 0098 gap 4).
		if pending, handled := g.offerOwnQuestionLocked(ev, chosen); handled {
			if pending {
				return nil
			}
			continue
		}
		g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
		g.runReplaceLocked(ev, chosen)
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
		//
		// #1282: through the tail-running exit, so a caller sequenced
		// behind the placement (amass's "the Army you amassed", a
		// Saga's chapter check) runs now that the counters are really
		// there, and is told zero when the target has gone.
		err := g.applyResolvedCounterThenLocked(ev)
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
		//
		// Then the rest of the instruction (CR 121.6b): a "draw three"
		// paused on its first card still owes two (draw_instead.go).
		return g.settleResumedDrawLocked(ev, false)
	case RepEventCounterSpell:
		// #4056: the window has settled on what happens to the spell
		// (countered as asked, or exiled instead). Finishing it is the
		// same function the unpaused path runs, so a counter that paused
		// on an ordering prompt and one that did not cannot drift
		// apart.
		return g.finishCounterSpellLocked(ev)
	case RepEventExtraTurn:
		// #2529: unreachable. The window sets mustSettleNow (the
		// rotation seam has no resume), so no event of this kind ever
		// queues a prompt, and the turn it describes is begun or
		// skipped by popExtraTurnLocked itself.
		return nil
	case RepEventRollDice:
		// #4307: unreachable. The window sets mustSettleNow, so no roll
		// queues a CR 616 prompt; the dice flow asks its own keyed
		// questions after the window settles (dice_roll.go).
		return nil
	case RepEventProduceMana:
		// #1222: unreachable, and that is the decision rather than an
		// oversight. A production sets mustSettleNow (CR 605.3b — a
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
		//
		// #1322: one card of a SIMULTANEOUS entry is not landed here.
		// Its window has settled, so the settled event goes back to the
		// batch, which opens the next card's window and lands the whole
		// batch together once the last one settles (entry_batch.go).
		// First, because a batch card is an entry the branches below
		// would otherwise finish on its own.
		if b, i, ok := ev.takeEntryBatch(); ok {
			if err := g.resumeEntryBatchLocked(b, i, ev); err != nil {
				return err
			}
			// An action boundary, as for the single entry below.
			g.runStateChecksLocked()
			return nil
		}
		if ev.zoneRoute != nil && !ev.zoneRoute.ViaBattlefieldLeave {
			return g.executeZoneRouteLocked(ev)
		}
		// S17 sub-PR 6: resume path for battlefield-leave moves
		// after the CR 903.9 commander-zone Optional prompt. The
		// pipeline settled on ev.NewZone (either the original
		// destination if owner said "no", or ZoneCommand if they
		// said "yes"). Run the physical move through the shared
		// executeBattlefieldLeaveLocked helper.
		if ev.entryResumable && ev.OldZone != ZoneBattlefield {
			// A paused ENTRY — the shockland's pay-2-life prompt, a
			// CR 616 ordering prompt between two enters-tapped effects,
			// Clone's "choose what to copy". Since ADR 0098 also one the
			// window REDIRECTED (Mox Diamond's "if you don't, put it into
			// its owner's graveyard"): the finisher moves it there
			// (moveRedirectedEntryLocked) rather than dropping it, which
			// stranded a resolving Mox on the stack. The pipeline function
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
			// What is left here is an entry that is not entryResumable
			// at all — since #1322 only the sandbox move_card verb,
			// which never pauses. (A redirected resumable entry is
			// finished by the branch above since ADR 0098.) The
			// caller's continuation is still told nothing entered.
			// Since #707 no EXIT lands here: every one of them carries
			// a zoneRoute or comes off the battlefield.
			return g.runEntryTailLocked(ev, uuid.Nil)
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
		// a loss (CR 701.17b, #767), but the cards that landed can be —
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
// settled — now finish the event" call the CR 616 / "may" replacement
// resumes make. `ev` is the event the resume has been holding; `out` is what
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
		g.restoreEntryControllerLocked(ev)
		g.dropEnteringTokenLocked(ev.CardID)
		// #1985, CR 608.3e: a resumed permanent spell's entry refused
		// (a Clone told to copy a land under Worms of the Earth) goes to
		// its owner's graveyard rather than staying on the stack.
		if err := g.settleRefusedStackEntryLocked(ev); err != nil {
			return err
		}
		// ADR 0178: a life change a replacement declared before the
		// cancel still happens, and the route waits for it.
		if err := g.runMoveTailLocked(ev); err != nil {
			return err
		}
		return g.runEntryTailLocked(ev, uuid.Nil)
	case RepEventStepTransition:
		// The one kind whose cancellation is not nothing — see the
		// doc comment above.
		return g.applyResolvedReplacementEventLocked(ev)
	case RepEventCounter:
		// #1282: a cancelled COUNTER placement places nothing, but the
		// rest of the effect behind it ("amass, then the Army deals
		// damage equal to its power") still runs, told zero.
		return g.runCounterTailLocked(ev, 0)
	case RepEventDraw:
		// #2168: a cancelled draw owes its REPLACEMENT's body (dredge,
		// Underrealm Lich, Forbidden Crypt) and the rest of its
		// instruction ("draw three" has two more to go), which the
		// body pays through `done`. draw_instead.go.
		return g.settleResumedDrawLocked(ev, true)
	case RepEventCounterSpell:
		// #4056: a cancelled counter leaves the spell on the stack, and
		// the caller's continuation ("if that spell is countered this
		// way") is told it was not.
		return g.runRouteTailLocked(ev.zoneRoute)
	case RepEventExtraTurn:
		// #2529: a cancelled extra turn is a SKIPPED one, and
		// popExtraTurnLocked acts on that verdict itself, inline. The
		// window never pauses, so nothing resumes into here.
		return nil
	case RepEventRollDice:
		// #4307: RollDiceThenForEffect reads a cancelled roll inline (no
		// dice, and the tail told so). The window never pauses, so
		// nothing resumes into here.
		return nil
	case RepEventProduceMana:
		// #1222: a production cannot even reach this function — it
		// sets mustSettleNow, so it never pauses and nothing resumes
		// it, and a production replaced away simply does not happen.
		// The arm is the written answer #982 asks for rather than a
		// fall-through, and it is right either way.
		return nil
	}
	return nil
}

// RevealedHandDiscard describes "target player reveals their hand. You
// choose a <kind of> card from it. That player discards that card."
// (ADR 0116). A struct because the filter and its label travel
// together and every card in the shape names both.
type RevealedHandDiscard struct {
	// Chooser picks; FromPlayer reveals and discards. They may be the
	// same player (Thoughtseize can target its caster).
	Chooser, FromPlayer uuid.UUID
	// Source is the card whose effect asks. Its name heads the reveal.
	Source uuid.UUID
	// Count is the printed number of cards to choose — one on every
	// card in the shape but Gruesome Discovery's two.
	Count int
	// Reason is the prompt header, written the way the card is
	// ("Thoughtseize").
	Reason string
	// Filter is the printed restriction, evaluated against each card
	// in the hand as it is in the hand (CR 712.8a, 709.4b, 202.3e).
	// nil is "a card": every card qualifies.
	Filter func(Card) bool
	// Label is what Filter admits, for the client ("nonland card").
	Label string

	// The variants (ADR 0116's 2026-10-05 amendment, #2115). Any one of
	// them set makes the pick a PendingChoiceRevealedHandPick, raised
	// through RevealedHandPickForEffect (revealed_hand_pick.go).

	// Destination is where the chosen card goes. The zero value and
	// PickDiscard discard it; PickExile exiles it, which is not a
	// discard (CR 701.9a), so madness and "whenever a player discards"
	// never see it (Appetite for Brains).
	Destination PickDestination
	// Optional is "You may choose": the chooser may choose nothing
	// (Nightsnare, Extract the Truth).
	Optional bool
	// Revealed says the hand was already revealed this resolution (Cloak
	// and Dagger, Entwined reveals it, then asks whether to take a card
	// from it), so the pick does not reveal it a second time. Read when
	// the pick is raised; never stored.
	Revealed bool
	// Look is "Look at target opponent's hand" rather than "reveals
	// their hand": only the chooser learns the hand (Word of Command,
	// ADR 0166).
	Look bool
	// FromGraveyard is "or a card from their graveyard" (Agonizing
	// Remorse): every card in FromPlayer's graveyard is a candidate as
	// well as the hand's matches. Filter applies to the hand only.
	FromGraveyard bool
	// Then is the registered continuation told what was chosen: "if
	// you do / if you don't", a clause that reads the chosen card, or
	// anything else printed after the pick that must wait for it.
	Then RevealedPickThen
	// Measure reads a number off each candidate as the pick is raised,
	// while the spell is still resolving: Talara's Bane's "that creature
	// card's toughness". The continuation is handed the chosen cards'
	// numbers (RevealedPick.Measures). Read here rather than in the
	// continuation because the continuation runs when the pick is
	// answered, and by then the resolved spell has been put into its
	// graveyard (CR 608.2n), which a toughness that counts graveyards
	// would see; nothing else can change in between, since the prompt
	// stops the table. Like Filter, the func is never stored.
	Measure func(Card) int
	// ExiledWith links a PickExile pick's exiled card to this source
	// OBJECT (CR 607.2a, #3516): the card lands in exile with
	// Card.ExiledWith naming it, through the one linked-exile route
	// (ExileLinkedKeyedForEffect), so a standing cast permission filtered
	// on PermissionFilter.ExiledWithSource finds it. The zero ObjectRef
	// links nothing. Refused with any other Destination: a discard or a
	// card left in the hand is not exiled with anything.
	ExiledWith ObjectRef
}

// isVariant reports whether d needs the variant prompt rather than the
// plain discard_from_hand.
func (d RevealedHandDiscard) isVariant() bool {
	return (d.Destination != "" && d.Destination != PickDiscard) ||
		d.Optional || d.FromGraveyard || d.Then.key != "" || d.Measure != nil || d.Look
}

// QueueDiscardFromRevealedHand is the revealed-hand pick (ADR 0116,
// ADR 0010 §10). In order:
//
//  1. FromPlayer reveals their hand: every seated player becomes a
//     knower of every card in it and the table gets one reveal run
//     (RevealForEffect, CR 701.20a — Unmask's ruling: a player who
//     targets themselves reveals to everyone else too). The reveal is
//     sticky, like any other.
//  2. The filter runs once over the hand, in hand order, and the
//     matching cards become the prompt's DiscardOptions. The set is
//     fixed here: the card is chosen from the hand that was revealed
//     (CR 608.2c), so a card that reaches the hand afterwards cannot
//     be chosen.
//  3. No matching card: nothing is chosen and nothing is discarded,
//     and no prompt goes up (CR 609.3). The hand was still revealed,
//     and the rest of the resolving effect still happens — Thoughtseize
//     still costs 2 life (its 2020-08-07 ruling).
//  4. Otherwise a discard_from_hand prompt is queued for the chooser,
//     with Count capped at the number of matching cards.
//
// The prompt carries no continuation, so it stays a restore point. A
// clause printed after the discard runs on the caller's next line; the
// open resolution holds state-based actions and the trigger drain until
// the pick is answered (#1289), so the order cannot be observed.
//
// Returns the prompt's ID, or uuid.Nil when nothing was queued.
//
// A variant (Destination, Optional, FromGraveyard or Then set) is
// handed to RevealedHandPickForEffect, which can also return an error
// from a continuation it ran at once; here that error is reported as an
// effect error, the way an answered prompt's would be. A card calls
// RevealedHandPickForEffect itself to see it.
//
// Caller must hold g.mu.
func (g *Game) QueueDiscardFromRevealedHand(d RevealedHandDiscard) uuid.UUID {
	if d.isVariant() {
		id, err := g.RevealedHandPickForEffect(d)
		if err != nil {
			g.emitChoiceEffectErrorLocked(d.Chooser, d.Source, err)
		}
		return id
	}
	p := g.playerByIDLocked(d.FromPlayer)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return uuid.Nil
	}
	if g.effectDiscardBlockedLocked(d.FromPlayer) {
		// #2178: an opponent's effect can't make them discard.
		return uuid.Nil
	}
	hand := make([]uuid.UUID, 0, p.Hand.Size())
	options := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		hand = append(hand, c.InstanceID)
		if d.Filter == nil || d.Filter(c) {
			options = append(options, c.InstanceID)
		}
	}
	reason := d.Reason
	if reason == "" && d.Source != uuid.Nil {
		if card, ok := g.LookupCardForEffect(d.Source); ok {
			reason = card.Name
		}
	}
	g.RevealForEffect(RevealSpec{
		Player: d.FromPlayer,
		Source: d.Source,
		Reason: revealHandReason(reason),
		Cards:  hand,
	})
	count := d.Count
	if count > len(options) {
		// CR 609.3: choose as many as you can — none, with no match.
		count = len(options)
	}
	if count <= 0 {
		return uuid.Nil
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:           PendingChoiceDiscardFromHand,
		Chooser:        d.Chooser,
		FromPlayer:     d.FromPlayer,
		Count:          count,
		Source:         d.Source,
		Reason:         reason,
		DiscardOptions: options,
		DiscardLabel:   d.Label,
	})
}

// revealHandReason is the reveal banner's line for a revealed-hand
// pick: "Thoughtseize — reveals their hand".
func revealHandReason(source string) string {
	if source == "" {
		return "reveals their hand"
	}
	return source + " — reveals their hand"
}

// discardOptionsAllow reports whether every pick is one the prompt
// offered and none is named twice (ADR 0116 §5). A nil DiscardOptions
// — a restore point from before ADR 0116 — offers the whole hand, which
// the caller checks separately.
func discardOptionsAllow(c *PendingChoice, picks []uuid.UUID) bool {
	seen := make(map[uuid.UUID]bool, len(picks))
	for _, id := range picks {
		if seen[id] {
			return false
		}
		seen[id] = true
		if c.DiscardOptions != nil && !slices.Contains(c.DiscardOptions, id) {
			return false
		}
	}
	return true
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
//   - trampleToPlayer > 0 only when the attacker has trample, and
//     only when every blocker is assigned lethal damage (CR 702.19b).
//     Deathtouch relaxes the threshold to 1.
//
// Any other split is legal: CR 510.1c divides the damage among the
// blockers as the controller chooses, with no order (#2692).
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
	// CR 510.1c, Wolverine, Claws Out: "as though it weren't blocked" is
	// the whole power on what the attacker attacks and nothing on any
	// blocker, which needs neither trample nor lethal damage to anyone.
	asUnblocked := frame.AsThoughUnblocked && trampleToPlayer == frame.AttackerPower
	for _, e := range ordered {
		if e.Amount != 0 {
			asUnblocked = false
		}
	}
	if trampleToPlayer > 0 && !frame.AllowTrample && !asUnblocked {
		return ErrInvalidParam
	}
	// Sum check.
	total := trampleToPlayer
	for _, e := range ordered {
		total += e.Amount
	}
	// With no blocker left to divide among, the same attacker may also
	// decline (assign nothing): the only other thing it could do.
	declined := frame.AsThoughUnblocked && len(frame.BlockerIDs) == 0 && total == 0
	if total != frame.AttackerPower && !declined {
		return ErrInvalidParam
	}
	// CR 510.1c: a blocked creature's damage is "divided as its
	// controller chooses among" its blockers. There is no damage
	// assignment order and no lethal-first rule, so any split that adds
	// up is legal (#2692; the engine enforced the old ordered rule until
	// then). CR 510.1d is the same for a blocker dividing among the
	// attackers it blocks (#1706).
	//
	// Trample is the one constraint (CR 702.19b): damage reaches the
	// player only once every blocker has been assigned lethal damage,
	// max(1, toughness less damage marked), and 1 from deathtouch (CR
	// 702.2c). A blocker that has left the battlefield, or that already
	// carries lethal damage, needs nothing.
	if trampleToPlayer > 0 && !asUnblocked {
		for _, e := range ordered {
			lethal := 1
			blk := findBattlefieldCard(g, e.BlockerID)
			switch {
			case blk == nil:
				lethal = 0
			case !frame.HasDeathtouch:
				// #3665: and the damage other creatures have been
				// assigned to it in this step, held until it is dealt.
				lethal = max(blk.CurrentToughness()-blk.DamageMarked-g.heldCombatDamageToLocked(blk.InstanceID), 0)
				if lethal == 0 {
					break
				}
				lethal = max(lethal, 1)
			}
			if e.Amount < lethal {
				return ErrInvalidParam
			}
		}
	}
	g.dequeueChoiceLocked(idx)

	// #3665, CR 510.2: the step this prompt belongs to is holding its
	// damage. The answer joins it, and the last answer deals the whole
	// step at once; the state-based actions see all of it together.
	if g.heldCombatDamage != nil {
		g.holdDamageAssignmentAnswerLocked(frame, ordered, trampleToPlayer)
		if g.settleHeldCombatDamageLocked() {
			g.runStateChecksLocked()
		}
		return nil
	}

	// A prompt with no hold behind it was restored from a snapshot
	// written before #3665, whose step dealt the rest of its damage
	// when the prompt was queued: the answer is dealt now, as it was.
	//
	// ADR 0108 §7: the assigned damage is one group of the step's
	// instance, collected while a charged shield is live so the shield
	// can be divided among it before any lands (divide_shield.go).
	var prevStage *damageStage
	staged := false
	if g.anyChargedShieldLocked() {
		prevStage, staged = g.openDamageStageLocked(g.combatDamageInstanceLocked(), true, false)
	}
	g.applyDamageAssignmentLocked(frame, ordered, trampleToPlayer)
	if staged {
		g.closeDamageStageLocked(prevStage)
	}
	g.runStateChecksLocked()
	return nil
}

// applyDamageAssignmentLocked deals one answered damage-assignment
// prompt's division: each share to its blocker (or, for a blocker
// dividing, to its attacker), and a trampler's excess to what it is
// attacking.
//
// The source keywords come from the frame, which snapshotted them as
// the prompt was queued: the source may have left the battlefield
// since (a player leaving the game, or a pre-#3665 step that dealt the
// blockers' damage first), so looking up HasKeyword on it now would
// miss deathtouch / lifelink.
//
// The caller opens any damage stage and runs the state-based actions.
//
// Caller must hold g.mu.
func (g *Game) applyDamageAssignmentLocked(frame *DamageAssignmentFrame, ordered []DamageAssignmentEntry, trampleToPlayer int) {
	for _, e := range ordered {
		if e.Amount <= 0 {
			continue
		}
		g.markCombatDamageFromFrameLocked(e.BlockerID, e.Amount, frame)
	}
	if trampleToPlayer <= 0 {
		return
	}
	// Determine the defending player: find the attacker if still on
	// battlefield; otherwise look at the prompt's original intent —
	// trample-to-player was only allowed for a live attacker that had
	// declared a target. Fall back to any seat that isn't the
	// controller (best-effort).
	//
	// S27: the attacker may have been attacking a planeswalker or a
	// battle, in which case trample overflow goes to THAT, not to its
	// defending player (CR 702.19b — excess damage is assigned to the
	// player or permanent the creature is attacking). So the target id
	// is used as-is and the routing happens in
	// dealCombatDamageToAttackTargetLocked.
	var defenderID uuid.UUID
	if atkCard := findBattlefieldCard(g, frame.AttackerID); atkCard != nil {
		defenderID = atkCard.AttackingTarget
	}
	if defenderID != uuid.Nil && g.classifyAttackTargetLocked(defenderID) != AttackTargetPlayer {
		g.markCombatDamageFromFrameLocked(defenderID, trampleToPlayer, frame)
		return
	}
	if defenderID == uuid.Nil {
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
	tc TriggerContext,
	source Card,
	lki Characteristic,
	ability TriggeredAbility,
	doubledBy doublerRef,
) {
	chooser := source.Controller
	if ability.OptionalPrompt != nil && ability.OptionalPrompt.Chooser != nil {
		if override := ability.OptionalPrompt.Chooser(tc.Event, &source, g); override != uuid.Nil {
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
		noLegalTarget = !ability.HasLegalTarget(tc.Event, &source, lki, g)
	}
	// ADR 0127 §2: an optional trigger's standing answer is filed under
	// its catalog row.
	key, card, prompt := triggerPromptKey(source, ability), "", ""
	if key != "" {
		card = source.Name
		prompt = question
		if prompt == "" {
			prompt = ability.Key
		}
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceTriggerPrompt,
		Chooser:          chooser,
		Count:            1,
		Source:           source.InstanceID,
		Reason:           question,
		NoLegalTarget:    noLegalTarget,
		AutoAnswerKey:    key,
		AutoAnswerCard:   card,
		AutoAnswerPrompt: prompt,
		triggerResume: &triggerResumeFrame{
			tc:        tc,
			source:    source,
			lki:       lki,
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
func (g *Game) queuePickTargetLocked(tc TriggerContext, source Card, lki Characteristic, t TriggeredAbility, doubledBy doublerRef, modes []int, steps []AnnouncedClause, spec *TargetSpec) {
	// #1657, CR 603.3d → CR 601.2d: a trigger's divided amount read
	// off the board ("X is the number of lands you control") or off its
	// source's last-known state ("damage equal to its power") is fixed
	// as the ability is put on the stack — which is this walk — and
	// the prompt, the view and the gate all read the fixed number.
	lkiCopy := lki
	g.bindDivideAmountsLocked(steps, DivideAmountArgs{
		Controller: source.Controller, Source: source.InstanceID, SourceLKI: &lkiCopy,
	})
	g.queuePickTargetStepLocked(&pickTargetFrame{
		tc:        tc,
		source:    source,
		lki:       lki,
		ability:   t,
		spec:      spec,
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
		src := SourceObject(f.source.Controller, &f.source).WithTrigger(f.tc)
		// #2055: an "of an opponent's choice" step is answered by the
		// opponent the controller names (CR 603.3d with CR 601.2c), its
		// legal set still judged for the controller.
		chooser, forPlayer := f.source.Controller, uuid.Nil
		if clause.ChosenByTriggerActor {
			// #1997: "of their choice" — the triggering event's player
			// chooses (The Abyss's upkeep player), its legal set still
			// judged for the controller. A walk with no such player, or
			// one who has left, has nobody to choose: CR 603.3d removes
			// the ability for a required clause.
			who := triggerActorOf(src)
			if who == uuid.Nil || !g.playerInGameLocked(who) {
				if clause.Min > 0 {
					return
				}
				f.step++
				continue
			}
			chooser = who
			if who != f.source.Controller {
				forPlayer = f.source.Controller
			}
		}
		if clause.ChosenByOpponent {
			named := f.stepChooser(f.step)
			if named == uuid.Nil {
				eligible := g.opponentChoosersForStepLocked(src, f.boundSteps(), f.step, f.picked)
				switch len(eligible) {
				case 0:
					if clause.Min > 0 {
						// No opponent could choose a legal target: the
						// ability is removed (CR 603.3d).
						return
					}
					f.step++
					continue
				case 1:
					named = eligible[0]
					f.choosers = append(f.choosers, TargetChooser{Mode: f.steps[f.step].Mode, Slot: f.steps[f.step].Slot, Player: named})
				default:
					g.queueTriggerOpponentLocked(f, eligible)
					return
				}
			}
			chooser, forPlayer = named, f.source.Controller
		}
		steps := f.boundSteps()
		clause = &steps[f.step].Clause
		lt := g.LegalTargetsForAnnouncedClauseForEffect(src, steps, f.step, append([]TargetRef{}, f.picked...))
		lt = withoutPicked(lt, f.picked, clause.Distinct)
		// #1559: a clause with a set rule can have candidates and
		// still no legal SET of Min picks ("two target creatures
		// controlled by different players" with every creature on one
		// side). That is CR 603.3d's "no legal targets" too, and a
		// prompt for it could never be answered. #1807: so can one
		// whose picks must share a key, when no key group is big
		// enough. #4616: and one whose picks must stay within a total,
		// when not enough of them fit together.
		if (clause.Different != nil || clause.Same != nil || clause.Total != nil) && clause.Min > 0 && g.fillableCountLocked(clause, lt) < clause.Min {
			return
		}
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
		if clause.AtRandom {
			// #1815: "target … chosen at random" — the engine draws it
			// now, as the trigger goes on the stack, and nobody is
			// asked.
			f.picked = append(f.picked, g.drawTriggerTargetLocked(f, lt))
			f.step++
			continue
		}
		label := clause.Label
		if label == "" {
			label = "Choose a target"
		}
		if forPlayer != uuid.Nil {
			label = g.seatLabelLocked(forPlayer) + "'s " + f.source.Name + ": choose " + clauseLabel(clause)
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:              PendingChoicePickTarget,
			Chooser:           chooser,
			Count:             1,
			Source:            f.source.InstanceID,
			Reason:            label,
			PickTargetPlayers: lt.Players,
			PickTargetCards:   lt.Cards,
			PickTargetMin:     clause.Min,
			PickTargetMax:     clause.Max,
			PickTargetFor:     forPlayer,
			PickTargetHarmful: clause.Harmful,
			pickTargetResume:  f,
		})
		return
	}
	g.finishPickTargetLocked(f)
}

// queueTriggerOpponentLocked asks a trigger's controller which opponent
// chooses the open step's targets (#2055). The answer continues the
// walk on a copy of the frame, with that opponent named. The option
// pick carries the frame as pickTargetResume too: it is part of the
// trigger's announcement, so the drain holds the batch behind it
// (triggerAnnouncementOpenLocked), an ended turn drops it with the
// walk, and it belongs to no resolution. Its answer is the keyed
// triggerOpponentChooserThen (#4361 S14b), which reads the walk back
// off the prompt (OptionPicked.walk): the prompt holds no closure, and
// a restore point writes it with the walk as data. ResolvePickTargets
// never sees it (its kind is option_pick).
//
// Caller must hold g.mu.
func (g *Game) queueTriggerOpponentLocked(f *pickTargetFrame, eligible []uuid.UUID) {
	step := f.steps[f.step]
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  f.source.Controller,
		Source:   f.source.InstanceID,
		Question: f.source.Name + ": choose an opponent to choose " + clauseLabel(&step.Clause),
		Options:  g.seatChoiceOptionsLocked(eligible),
		walk:     f,
		ThenKey:  triggerOpponentChooserThen,
	})
}

// triggerOpponentChooserThen is the controller's answer to "which
// opponent chooses" in a trigger's walk (#2055; keyed by #4361 S14b):
// the walk goes on, on a copy of the frame, with that opponent named
// for the open step. The walk is the prompt's own pickTargetResume, so
// the key captures nothing. A dropped prompt (the controller has left)
// does nothing: the trigger is not put on the stack (CR 800.4d). An
// on-disk key: never renamed, never reused.
var triggerOpponentChooserThen OptionPickThen

// Registered from init: the continuation reaches code that names the
// key, which a package-level initializer cannot.
func init() {
	triggerOpponentChooserThen = RegisterOptionPickThen("option-pick/trigger-opponent-chooser", triggerOpponentChooserAnswered)
}

func triggerOpponentChooserAnswered(g *Game, r OptionPicked) error {
	chosen, frame := chosenSeat(r), r.walk
	if chosen == uuid.Nil || frame == nil || frame.currentClause() == nil {
		return nil
	}
	step := frame.steps[frame.step]
	next := clonePickTargetFrame(frame)
	next.choosers = append(next.choosers, TargetChooser{Mode: step.Mode, Slot: step.Slot, Player: chosen})
	g.queuePickTargetStepLocked(next)
	if !g.triggerAnnouncementOpenLocked() {
		g.runStateChecksLocked()
	}
	return nil
}

// finishPickTargetLocked runs Build with every step answered and
// queues the item, stamping the announcement onto it so the CR
// 608.2b re-check can find each ref's clause.
//
// Caller must hold g.mu.
func (g *Game) finishPickTargetLocked(f *pickTargetFrame) {
	if !f.ability.builds() {
		return
	}
	item := g.buildTriggerItemLocked(f.ability, f.tc.Event, f.source, f.lki)
	if item == nil {
		return
	}
	stampTriggerContext(item, f.ability, f.tc)
	stampTriggerSource(item, f.source, f.tc)
	item.Targets = append([]TargetRef(nil), f.picked...)
	item.Modes = append([]int(nil), f.modes...)
	item.Distribution = cloneDistributionLocked(f.dist)
	item.TargetChoosers = copyTargetChoosers(f.choosers)
	item.targetSpec = f.spec
	item.modeSpec = f.modeSpec
	item.DoubledBy, item.DoubledByName = f.doubledBy.id, f.doubledBy.name
	// CR 603.3d / 115.3: a triggered ability's targets are chosen as
	// it is put on the stack, and that placement is the drain's, not
	// this answer's: the batch may still be held behind an ordering
	// prompt or a batch-mate's announcement (#1529). So the "becomes
	// the target" event is owed, not emitted, and the drain that
	// places the item emits it. A trigger it sets off then waits for
	// the whole batch and goes on above it (CR 603.3b), never inside
	// the batch that targeted (#1539).
	item.TargetsAnnouncePending = len(item.Targets) > 0
	g.queueHarvestedTriggerLocked(item)
	// #2608: this is also reached straight from the harvest (a trigger
	// whose every clause is "up to" and has nothing to point at needs no
	// prompt, so it finishes at once). A state-based sweep there ran
	// under the battlefield walk and removed a departed player's cards
	// from it. The event that is being harvested belongs to a mutation
	// whose caller runs the boundary, exactly as for an untargeted
	// trigger, so only the answer path runs it here.
	if g.harvestDepth == 0 {
		g.runStateChecksLocked()
	}
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
	return g.ResolvePickTargetsDivided(choiceID, chooserID, targets, nil)
}

// ResolvePickTargetsDivided is ResolvePickTargets with the division
// announced alongside the targets (#1563, CR 601.2d via CR 603.3d) —
// Fury's "4 damage divided as you choose among any number of target
// creatures and/or planeswalkers". `dist` is keyed by target id and
// is settled against the prompt's clause exactly as a cast's is: each
// target at least 1, the shares summing to the amount, a lone target
// given the whole amount when no share is named. Nil for a clause
// that divides nothing, which is every other pick_target prompt.
func (g *Game) ResolvePickTargetsDivided(choiceID, chooserID uuid.UUID, targets []TargetRef, dist map[uuid.UUID]int) error {
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
	if cf := choice.copyResume; cf != nil {
		return g.resolveCopyTargetsLocked(idx, cf, targets)
	}
	// #2055: an opponent choosing a target for a parked cast or
	// activation. The answer makes the announcement again.
	if of := choice.opponentTargetResume; of != nil {
		if len(dist) > 0 {
			return ErrInvalidParam
		}
		return g.resolveOpponentTargetsLocked(idx, choice, of, chooserID, targets)
	}
	frame := choice.pickTargetResume
	step := frame.currentClause()
	if frame == nil || !frame.ability.builds() || step == nil {
		g.dequeueChoiceLocked(idx)
		// #1529: the batch this trigger belonged to was held for it.
		g.runStateChecksLocked()
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
	all := append(append([]TargetRef(nil), frame.picked...), stamped...)
	// #2055: judged for the trigger's CONTROLLER, who is not the chooser
	// on an "of an opponent's choice" step, with the opponents named so
	// far bound onto their clauses. The answering opponent is recorded
	// as the step's chooser: a departure may have handed the prompt on
	// (CR 800.4g).
	if cur.Clause.ChosenByOpponent {
		frame.choosers = withStepChooser(frame.choosers, cur, chooserID)
	}
	if err := g.validateAnnouncedTargetsWithLocked(SourceObject(frame.source.Controller, &frame.source).WithTrigger(frame.tc), frame.boundSteps()[:frame.step+1], all, func(i int, _ TargetRef) bool { return i < len(frame.picked) }); err != nil {
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
	// #1563: a trigger announces no X, so the amount is the clause's
	// fixed one. Settled before anything moves, so a refused division
	// leaves the prompt open for another answer.
	settled, err := settleDistribution(frame.steps[frame.step:frame.step+1], stamped, dist, 0)
	if err != nil {
		return err
	}
	for id := range settled {
		if _, dup := frame.dist[id]; dup {
			return errDivision("the same target is in two divided clauses")
		}
	}
	for id, v := range settled {
		if frame.dist == nil {
			frame.dist = make(map[uuid.UUID]int, len(settled))
		}
		frame.dist[id] = v
	}
	g.dequeueChoiceLocked(idx)
	frame.picked = append(frame.picked, stamped...)
	frame.step++
	g.queuePickTargetStepLocked(frame)
	// #1529: the drain holds the whole trigger queue while this walk
	// is open. finishPickTargetLocked runs the boundary when the item
	// joins; a walk that ended because CR 603.3d removed the ability
	// (a later required clause with nothing left to pick) has to run
	// it too, or the batch waits for an unrelated action. A no-op
	// while a further step's prompt is open.
	if !g.triggerAnnouncementOpenLocked() {
		g.runStateChecksLocked()
	}
	return nil
}

// ResolveTriggerPrompt processes the chooser's yes/no answer for a
// PendingChoiceTriggerPrompt entry. Almost every such prompt is an
// optional trigger's CR 603.5 "you may", asked as it resolves (#2494):
// `apply: true` runs the item's effect, `apply: false` lets it finish
// doing nothing (answerOptionalTriggerLocked). The one asked before the
// stack (TriggerOptionalPrompt.BeforeStack, miracle's reveal) still
// works as it always did: on `apply: true` the stashed Build closure
// runs against the captured event + LKI and the item is queued for the
// drain; on `apply: false` the trigger is never put on the stack.
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
	g.answerTriggerPromptLocked(idx, choice, apply, false)
	return nil
}

// answerTriggerPromptLocked is ResolveTriggerPrompt's body, shared with
// the automatic answer (ADR 0127 §4): `auto` takes the prompt out of the
// queue without counting it as a player decision. Caller must hold g.mu.
func (g *Game) answerTriggerPromptLocked(idx int, choice *PendingChoice, apply, auto bool) {
	if choice.OptionalItem != uuid.Nil {
		// CR 603.5 (#2494): the "you may" of a resolving item.
		g.answerOptionalTriggerLocked(idx, choice, apply, auto)
		return
	}
	frame := choice.triggerResume
	g.answerChoiceLocked(idx, auto)
	if !apply || frame == nil || !frame.ability.builds() {
		// #1529: declining still releases the batch the drain was
		// holding for this trigger (CR 603.3b).
		g.runStateChecksLocked()
		return
	}
	// S20: a targeted optional trigger continues into the target
	// pick; an untargeted one builds straight away.
	g.buildOrPickTriggerLocked(frame.tc, frame.source, frame.lki, frame.ability, frame.doubledBy, nil)
	// The prompt is answered outside any priority-wrap, so nothing
	// downstream would drain the queue until the next pass around
	// the table — and an empty stack at that wrap would advance the
	// step first, stranding the trigger a step late. Drain now:
	// answering "yes" is the moment the ability is put on the stack
	// (CR 603.3), and SBAs run at the same boundary.
	g.runStateChecksLocked()
}

// QueuePayUnlessForEffect queues an "unless that player pays <cost>"
// prompt for `chooser` with nothing hanging off either answer: the
// PayUnless primitive with no Else. `question` is the dialog header;
// `source` attributes the prompt to the card for the client. A card
// with a consequence names it as registered branches instead
// (QueuePayUnlessKeyedForEffect, QueuePayUnlessPromptKeyedForEffect).
//
// The prompt carries no closure (#4361): its key is the engine's own
// payUnlessNothing, so a table waiting on it is a restore point. A
// chooser who is no longer seated is not asked, and an unparseable cost
// is surfaced as EventEffectError.
//
// Caller must hold g.mu. Added in S19 sub-PR 6.
func (g *Game) QueuePayUnlessForEffect(chooser, source uuid.UUID, cost, question string) error {
	return g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
		Chooser: chooser, Source: source, Cost: cost, Question: question,
		Then: payUnlessNothing,
	})
}

// MayCastPrompt is a may_cast's question with its words: the rule it is
// asked under and the names of its two branches (ADR 0099 §3). Every
// field but the chooser and the card is optional. The answer is a
// registered key (QueueMayCastKeyedForEffect, #4361).
type MayCastPrompt struct {
	Chooser, Source, Card uuid.UUID
	// Question is the dialog header.
	Question string
	// Keyword is the rule's name — PendingChoice.MayCastKeyword.
	Keyword string
	// AcceptLabel / DeclineLabel name the two buttons. Empty renders
	// the client's defaults.
	AcceptLabel, DeclineLabel string
	// Private marks a prompt whose words name a card only the chooser
	// has looked at (a "look at the top card" offer, #3801): the wire
	// strips the question, the labels and the card from every other
	// viewer (PendingChoice.PrivateText).
	Private bool
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
	branch := mayCastBranchOf(choice, apply)
	g.dequeueChoiceLocked(idx)
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
	return g.ResolvePayUnlessWithTaps(choiceID, chooserID, apply, nil)
}

// ResolvePayUnlessWithTaps is ResolvePayUnless with the permanents the
// chooser taps to help pay a waterbend cost (#1311, CR 701.67a) —
// "Ward—Waterbend {4}". Each tapped untapped artifact or creature the
// chooser controls pays {1} of the cost's generic mana; the rest is
// paid from the pool and the auto-tapper exactly as before, with the
// tapped permanents kept out of the auto-tapper's reach (CR 118.3).
//
// A malformed tap list is REFUSED, with the prompt left in place,
// rather than read as a decline: tap IDs on a prompt with no waterbend
// clause, on a "Don't pay" answer, or naming a permanent that cannot
// pay (not the chooser's, already tapped, not an artifact or creature,
// named twice, more than the generic). That is the posture every other
// cost validator takes — a client bug must not cost the payer their
// spell — and the rule the tap_ids of a cast and the waterbend_ids of
// an activation follow.
//
// A well-formed payment that still cannot be funded (the remaining
// mana isn't there) degrades to a decline with NOTHING tapped: the
// mana half is settled first and the taps only happen once it has
// been.
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolvePayUnlessWithTaps(choiceID, chooserID uuid.UUID, apply bool, tapIDs []uuid.UUID) error {
	return g.resolvePayUnless(choiceID, chooserID, apply, tapIDs, nil, 0, nil)
}

// ResolvePayUnlessWithLife is ResolvePayUnlessWithTaps with the number
// of the cost's symbols the chooser pays 2 life each for instead of the
// mana (CR 107.4f, ADR 0131 §2): a ward {B} or a Rhystic Study tax paid
// under K'rrik, Son of Yawgmoth, or a printed {B/P} in a pay-unless
// cost. The same count, the same ceiling and the same CR 119.4 / 119.8
// gate a cast or an activation announces (strikePhyrexianLifeLocked).
//
// A claim on a "Don't pay", on a non-mana payment, or past what the
// cost lets life pay is REFUSED with the prompt left in place, the
// posture the tap and card lists take. A claim whose remaining mana
// cannot be funded degrades to a decline with no life paid, as an
// unfundable mana "yes" always has.
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolvePayUnlessWithLife(choiceID, chooserID uuid.UUID, apply bool, tapIDs []uuid.UUID, phyrexianLife int) error {
	return g.resolvePayUnless(choiceID, chooserID, apply, tapIDs, nil, phyrexianLife, nil)
}

// ResolvePayUnlessWithCards is ResolvePayUnless for a prompt whose
// payment is not mana (ADR 0108 §5, pay_unless_action.go): `cardIDs`
// are the cards the chooser discards or the permanents they sacrifice
// to pay "discard a card" or "sacrifice two lands".
//
// A malformed payment is REFUSED with the prompt left in place, the
// posture ResolvePayUnlessWithTaps takes for a malformed tap list: card
// IDs on a mana prompt or on a "Don't pay", the wrong count, a card
// named twice, or one that cannot pay (not in the hand, not a permanent
// of the clause's kind the chooser controls). A "Pay" naming nothing
// when the chooser cannot make the payment at all is a decline, as an
// unfundable mana "yes" is; naming nothing when they can is refused.
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolvePayUnlessWithCards(choiceID, chooserID uuid.UUID, apply bool, cardIDs []uuid.UUID) error {
	return g.resolvePayUnless(choiceID, chooserID, apply, nil, cardIDs, 0, nil)
}

// ResolvePayUnlessWithOption is ResolvePayUnlessWithCards for an either
// payment (ADR 0187 §5, PayActionEither): `option` is the index into the
// payment's branches the chooser takes, and `cardIDs` name the cards that
// branch pays with when it is a card pick ("sacrifice two lands"). A
// branch that names nothing ("take 5 damage") sends no cards.
//
// Every malformed answer is REFUSED with the prompt left in place, the
// posture of the other pay-unless resolvers: an `option` out of range,
// an `option` on a prompt that is not an either payment or on a "Don't
// pay", `apply: true` on an either payment with no `option` (no branch
// is ever guessed), and cards that are not exactly the branch's payment.
// A branch the chooser cannot pay at all is a decline, as an unfundable
// mana "yes" is (CR 118.3).
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolvePayUnlessWithOption(choiceID, chooserID uuid.UUID, apply bool, option *int, cardIDs []uuid.UUID) error {
	return g.resolvePayUnless(choiceID, chooserID, apply, nil, cardIDs, 0, option)
}

// resolvePayUnless is the one body behind the three pay-unless
// resolvers.
func (g *Game) resolvePayUnless(choiceID, chooserID uuid.UUID, apply bool, tapIDs, cardIDs []uuid.UUID, phyrexianLife int, option *int) error {
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
	frame := payUnlessFrameOf(choice)
	if len(tapIDs) > 0 {
		if !apply || frame == nil {
			return ErrInvalidParam
		}
		budget := WaterbendBudget(frame.tap, frame.cost, 0)
		if err := g.validateTapPermanentsCostLocked(chooserID, frame.tap, tapIDs, budget); err != nil {
			return err
		}
	}
	// ADR 0108 §5: a non-mana payment names what it discards or
	// sacrifices. Validated before the prompt is dequeued, so a refused
	// answer leaves it to be answered again.
	var action *PayAction
	if frame != nil {
		action = frame.action
	}
	// ADR 0187 §5: an either payment is answered with the branch it
	// takes (`pay_option`), and the rest of the validation is that
	// branch's, as its own payment. `pays` is the payment this answer
	// makes: the chosen branch, or the prompt's own payment.
	pays := action
	switch {
	case action != nil && action.Kind == PayActionEither:
		if apply {
			// No branch is ever guessed, and none past the list.
			if option == nil {
				return ErrInvalidParam
			}
			if pays = EitherBranch(action, *option); pays == nil {
				return ErrInvalidParam
			}
		} else if option != nil {
			return ErrInvalidParam
		}
	case option != nil:
		// A `pay_option` on anything but an either payment.
		return ErrInvalidParam
	}
	if len(cardIDs) > 0 && (!apply || pays == nil || pays.NamesNothing() || pays.Kind == PayActionEither) {
		return ErrInvalidParam
	}
	// ADR 0131 §2: the life half of a mana payment, announced beside the
	// "pay". Validated against the cost as this chooser would pay it
	// before the prompt is dequeued, so a malformed or unpayable claim
	// leaves it to be answered again.
	if phyrexianLife != 0 {
		if !apply || frame == nil || action != nil || phyrexianLife < 0 {
			return ErrInvalidParam
		}
		if p := g.playerByIDLocked(chooserID); p != nil {
			cost := g.costAsPaidByLocked(chooserID, ManaSpendContext{}, WaterbendReduced(frame.cost, 0, len(tapIDs)), 0)
			_, life, err := g.strikePhyrexianLifeLocked(p, g.cardNameLocked(choice.Source), cost, ManaSpendContext{}, phyrexianLife, nil)
			if err != nil {
				return err
			}
			if !g.CanPayLifeLocked(p, choice.LifeCost+life) {
				return ErrInvalidParam
			}
		}
	}
	payable := false
	if apply && pays.NamesNothing() {
		// ADR 0129 §3: an energy payment names nothing. "Pay" is a
		// payment exactly when the energy is there (CR 118.3). ADR
		// 0150 §3: so do life (CR 119.4, 119.8), damage and nothing.
		payable = g.payActionPayableLocked(chooserID, pays)
	} else if apply && pays != nil {
		if len(cardIDs) > 0 || len(g.payActionOptionsLocked(chooserID, pays)) >= pays.Count {
			if err := g.validatePayActionLocked(chooserID, pays, cardIDs); err != nil {
				return err
			}
			payable = true
		}
	}
	var branch *PayAction
	if apply && action != nil && action.Kind == PayActionEither {
		branch = pays
	}
	g.answerPayUnlessLocked(idx, choice, chooserID, apply, tapIDs, cardIDs, phyrexianLife, payable, false, branch)
	return nil
}

// answerPayUnlessLocked is resolvePayUnless's body once the answer has
// been validated, shared with the automatic answer (ADR 0127 §4):
// `phyrexianLife` is the symbols paid with 2 life (ADR 0131 §2; an
// automatic answer pays mana only), `payable` is whether a non-mana
// payment can be made, and `auto` takes
// the prompt out of the queue without counting it as a player decision.
// The branch runs as the answered prompt's branch, so a prompt it
// queues is keyed from this one's (§2). Caller must hold g.mu.
func (g *Game) answerPayUnlessLocked(idx int, choice *PendingChoice, chooserID uuid.UUID, apply bool, tapIDs, cardIDs []uuid.UUID, phyrexianLife int, payable, auto bool, branch *PayAction) {
	frame := payUnlessFrameOf(choice)
	var action *PayAction
	if frame != nil {
		action = frame.action
	}
	// ADR 0187 §5: an either payment pays the branch the answer named
	// (`branch`); the automatic answer names none, and is never payable.
	if branch != nil {
		action = branch
	}
	g.answerChoiceLocked(idx, auto)
	if frame == nil {
		return
	}
	// #4566: a mana branch of an either payment (Titania's "Ward—Discard
	// a card or pay {2}") is paid as the prompt's own mana is, at the
	// branch's cost, so the pool, the auto-tapper and an unfundable
	// "yes" all behave as they do for a plain mana pay-unless.
	manaCost := frame.cost
	if action != nil && action.Kind == PayActionMana {
		parsed, err := ParseCost(action.Mana)
		if err != nil {
			apply = false
		}
		manaCost, action = parsed, nil
	}
	if choice.GuardsStackItem != uuid.Nil && g.StackItemForEffect(choice.GuardsStackItem) == nil {
		g.runStateChecksLocked()
		return
	}
	runBranch := func(fn func(g *Game) error) {
		g.runPromptBranchLocked(choice.AutoAnswerKey, func() {
			g.runPayUnlessBranchLocked(choice, chooserID, fn)
		})
	}
	if action != nil {
		// CR 118.12a: "sacrifice it unless you discard a card" — the
		// discard is the payment, and the "yes" branch runs once it has
		// been made. An unpayable "yes" is the decline.
		if !payable {
			runBranch(frame.onDecline)
			g.runStateChecksLocked()
			return
		}
		// #4361 S8: the "yes" waits for the payment as data, the
		// answered prompt's own (payUnlessPaidTail runs its onPay as
		// runBranch would).
		if err := g.payActionLocked(chooserID, choice.Source, action, cardIDs, payUnlessPaidTailOf(choice)); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, Actor: chooserID, Source: choice.Source, ErrorMsg: err.Error()})
		}
		g.runStateChecksLocked()
		return
	}
	paid := false
	if apply {
		if p := g.playerByIDLocked(chooserID); p != nil {
			cost := WaterbendReduced(manaCost, 0, len(tapIDs))
			paid = g.payCostWithLifeLocked(p, cost, choice.Source, unionIDs(tapIDs), phyrexianLife, choice.LifeCost)
			if paid {
				g.payTapPermanentsCostLocked(chooserID, tapIDs)
			}
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
	runBranch(consequence)
	g.runStateChecksLocked()
}

// runPayUnlessBranchLocked runs one branch of an answered pay-unless,
// surfacing its error on the log rather than refusing an answer that
// has already been taken. Caller must hold g.mu.
func (g *Game) runPayUnlessBranchLocked(choice *PendingChoice, chooserID uuid.UUID, branch func(g *Game) error) {
	if branch == nil {
		return
	}
	if err := branch(g); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    chooserID,
			Source:   choice.Source,
			ErrorMsg: err.Error(),
		})
	}
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
//     controls, so the branch is written against material CR 800.4a has
//     not just taken off the table. #4389 audited that for every caller
//     of runChoiceDropActionLocked, the only way here:
//     the departure sweep (dropChoicesForPlayerLocked) asks
//     departedChoiceActionAllowedLocked, which for pay_unless is
//     departedChoiceObjectLocked, read again for each prompt just before
//     its action runs; and dropChoiceLocked, the other caller, is never
//     handed a pay_unless (every withdrawal site names its kinds: trigger
//     announcements, loop notices, mode, sacrifice, card-set, option and
//     target picks, and a paused exit's replacement prompt, and none of
//     those carries a pay_unless frame). What the gate asks is that the
//     prompt's Source card is still in some zone, controlled (else owned)
//     by a seated player who is not the chooser, and that the prompt is
//     not about the chooser's own pool. It does not ask for the stack or
//     the battlefield, and must not: the ability, not its source, is
//     the object that requires the payment (CR 113.7a), so Rhystic Study
//     destroyed in response still draws on a decline (CR 800.4f).
//     A new withdrawal site that can reach a pay_unless has to pass the
//     same gate first.
//
// A may-pay frame ("you may pay {2}. If you do, ...") carries no decline
// branch and nothing happens, which is the right reading of "if you do"
// for a player who no longer can.
//
// Caller must hold g.mu.
func (g *Game) declineDepartedChoiceLocked(c *PendingChoice) {
	// #4256: a keyed prompt's frame is built from its data.
	frame := payUnlessFrameOf(c)
	if frame == nil || frame.onDecline == nil {
		return
	}
	g.runPromptBranchLocked(c.AutoAnswerKey, func() {
		g.runPayUnlessBranchLocked(c, c.Chooser, frame.onDecline)
	})
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
//
// `excluded` is sources the auto-tapper may not reach for — the
// permanents a waterbend payment is tapping instead of paying mana
// (#1311), which cannot also tap for the rest (CR 118.3). Nil for
// every other pay-unless.
//
// Caller must hold g.mu.
func (g *Game) payCostLocked(p *Player, cost ParsedCost, source uuid.UUID, excluded map[uuid.UUID]bool) bool {
	return g.payCostWithLifeLocked(p, cost, source, excluded, 0, 0)
}

// payCostWithLifeLocked is payCostLocked for a payment that also
// owes fixedLife and announces `phyrexianLife` symbols paid with 2 life each (CR 107.4f,
// ADR 0131 §2). The strike and the life payment are the helper pair a
// cast and an activation run (strikePhyrexianLifeLocked,
// payPhyrexianLifeLocked), so a ward {B} under K'rrik and a spell's {B}
// are one code path. The life is paid only once the mana half is known
// payable and before any land taps, because the life is the fallible
// half: false leaves the life total, the pool and every land alone. The
// claim was validated when the answer arrived; one that no longer holds
// here is a decline.
func (g *Game) payCostWithLifeLocked(p *Player, cost ParsedCost, source uuid.UUID, excluded map[uuid.UUID]bool, phyrexianLife, fixedLife int) bool {
	// #1600, CR 609.4b: Chromatic Orrery's "you may spend mana as
	// though it were mana of any color" reaches a ward or a Rhystic
	// Study tax too — every cost the player pays. A narrowed grant
	// (Oath of Nissa's planeswalker spells) does not: the zero context
	// names no spell.
	cost = g.costAsPaidByLocked(p.ID, ManaSpendContext{}, cost, 0)
	cost, lifeOwed, err := g.strikePhyrexianLifeLocked(p, g.cardNameLocked(source), cost, ManaSpendContext{}, phyrexianLife, nil)
	if err != nil {
		return false
	}
	if fixedLife < 0 || !g.CanPayLifeLocked(p, lifeOwed+fixedLife) {
		return false
	}
	lifeOwed += fixedLife
	var (
		plan  tapPlan
		short ParsedCost
		tap   bool
	)
	if !p.ManaPool.CanPay(cost, 0) {
		// ADR 0118 §1: plan only what the floating pool is missing.
		// #2461: a plan that would not fund the cost taps nothing.
		var ok bool
		plan, short, ok = g.autoTapFundedLocked(p, cost, 0, ManaSpendContext{}, excluded, 0)
		if !ok {
			return false
		}
		tap = true
	}
	if err := g.payPhyrexianLifeLocked(source, p.ID, lifeOwed); err != nil {
		return false
	}
	if tap {
		g.materializePlanLocked(p, plan, short)
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
	return g.ResolveTriggerOrderRemember(choiceID, chooserID, resolutionOrder, nil)
}

// ResolveTriggerOrderRemember is ResolveTriggerOrder with the "Use this
// order again" toggle (ADR 0018's eighth amendment, #3016): nil changes
// nothing, true files the answer's ranking under the prompt's key in
// the same commit, false forgets any order filed there. A true on a
// prompt with no key, or on an answer that interleaves copies of one
// row, is refused and nothing changes.
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolveTriggerOrderRemember(choiceID, chooserID uuid.UUID, resolutionOrder []uuid.UUID, remember *bool) error {
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
	if remember != nil {
		if err := g.rememberTriggerOrderLocked(choice, resolutionOrder, *remember); err != nil {
			return err
		}
	}
	g.applyTriggerOrderLocked(idx, resolutionOrder, false)
	return nil
}

// applyTriggerOrderLocked takes the trigger-order prompt at idx out of
// the queue — as a player decision, or for an automatic answer without
// restarting the CR 732 loop run (ADR 0127 §7) — and puts its items in
// resolutionOrder, then runs the drain again. Caller must hold g.mu.
func (g *Game) applyTriggerOrderLocked(idx int, resolutionOrder []uuid.UUID, auto bool) {
	g.answerChoiceLocked(idx, auto)

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
	cause := carriedCauseOfChoice(choice) // #4281: a branch's cause rides the answer
	g.dequeueChoiceLocked(idx)
	// #1019: the picked permanent goes through the sacrifice's own
	// CONTINUATION rather than the fire-and-forget call, for the two
	// reasons ADR 0013 §5x gives. It is what lets a run wait for a leg
	// a replacement's window has merely PAUSED — a sacrificed
	// permanent is still on the battlefield while its owner answers,
	// and a run that settled on this line would pay out with the
	// permanent still in play — and what lands in the run is
	// sacrificedThisWayLocked's answer rather than a re-read of the
	// board on the next line.
	//
	// The source stays uuid.Nil because the fire-and-forget form
	// passed none (sacrificePermanentLocked): EventSacrifice carries
	// no source for a prompted sacrifice today, and giving the leg a
	// continuation must change the sequencing and nothing else.
	if err := g.withCarriedCauseLocked(cause, func() error {
		return g.SacrificeAllKeyedForEffect(uuid.Nil, []uuid.UUID{cardID}, KeyedTail{
			Ref: sacrificeLegSettled, Player: seat, Carry: []uuid.UUID{run},
		})
	}); err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}

// sacrificeLegSettled settles a sacrifice prompt's leg of its run (#4253):
// the leg's seat (Player) and run (Carry), told the permanent that really
// left the battlefield (nothing when it did not). An on-disk key: never
// renamed, never reused.
var sacrificeLegSettled = RegisterTail("engine/sacrifice-leg-settled", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return g.settleRunLegLocked(o.Tail.Carry[0], o.Tail.Player, o.IDs)
})

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
// them, the prompt's live zone (liveZone). Writing a pruneDiscardChoicesLocked
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
// A prompt with no live zone is left alone. Zero means "no live
// re-check" (see liveZone): the continuation owns
// what a pick means, the resolver refuses nothing, and there is
// nothing for this to be right about.
//
// An emptied prompt is DROPPED rather than left unanswerable, through
// dropChoiceLocked — which runs the departure table's action for the
// kind, so a prompt that is one leg of a prompted RUN settles that leg
// with "nothing discarded" (#1016's dropDefault, ADR 0013 §5y item 5)
// and the rest of the printed instruction still happens. A pick that
// is no run's leg has no run to settle, but since #1225 it still runs
// the rest of the card: its frame's continuation is called with
// nothing picked, exactly as it is when its chooser leaves the game,
// because a mid-card choose_cards is asked from inside a resolution
// that is paused waiting for it and the drop must not be where the
// card stops (Torment of Hailfire's sacrifice branch is the case).
//
// The bounds move with the list (setCardSetCandidates): a pick of two
// from a hand that now holds one is CR 701.9a's "as many as you can",
// and a floor left above the candidate count is the same wedge one
// card later.
//
// The drop is announced (EventPendingChoiceDropped, #868) for
// pruneDepartedSeatOptionsLocked's reason: this prompt was blocking
// the table, and a stall dump has to say where it went.
//
// untap_choice carries the same payload (isCardSetPickKind) and is
// deliberately NOT swept — excluded BY KIND, below, since #1263.
// (Before #1263 this comment claimed the exclusion was automatic
// because untap_choice "sets no zone", which was never true —
// queueUntapChoiceLocked has stamped Zone: ZoneBattlefield on it
// since #826, so the loop examined and could WITHDRAW it exactly like
// any other card-set pick, contradicting this very paragraph and ADR
// 0018 §6's #1045 amendment.) Its continuation is the rest of the
// UNTAP STEP (finishUntapStepLocked), its departure row is
// dropDiscard, so a withdrawal here would end the question by
// stranding the step — and it cannot need this: CR 502.3's
// determination is about the active player's own permanents during
// their own untap step, where nothing has priority to move them.
//
// #1198's entry_reveal_from_hand is out for the same reason and one
// of its own (also excluded by kind, below, since #1263 — its frame's
// zone is ZoneHand, not empty either). The drop would strand a paused
// CR 614 ENTRY, which is worse than the untap step it would strand
// above; and the prompt cannot need the prune, because its floor is
// zero — "reveal nothing" is an answer no emptied candidate list can
// take away, which is exactly the property untap_choice lacks and a
// choose_cards with a floor of one lacks. A stale candidate is still
// refused on submit by pickStillInPickZoneLocked, so the worst an
// un-pruned list costs is
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
		// sweep exists for.
		if c == nil || !isCardSetPickKind(c.Kind) {
			continue
		}
		// #1263: untap_choice and entry_reveal_from_hand are excluded
		// BY KIND, not by an absent zone — both stamp a real Zone on
		// their frame (ZoneBattlefield, ZoneHand), so the zone check
		// below does not exclude them and never did. See the doc
		// comment above for why neither needs (and neither can
		// survive) this sweep: both would be WITHDRAWN by it, and
		// each one's drop action is dropDiscard, which strands the
		// paused untap step or the paused CR 614 entry it is
		// pausing — forever, with nothing to settle it.
		if c.Kind == PendingChoiceUntapChoice || isEntryCardChoiceKind(c.Kind) {
			continue
		}
		zone := c.liveZone()
		if zone == "" {
			continue
		}
		live := make([]uuid.UUID, 0, len(c.ChooseCards))
		for _, id := range c.ChooseCards {
			if g.pickStillInPickZoneLocked(c, zone, id) {
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

// liveZone is the zone a card-set pick's candidates are re-checked against:
// for an entry card choice and a permanent pick, the zone its kind names
// (#4361 S7b: the closure frame that held it is gone); for a keyed pick, the
// one it was queued with (ChooseLiveZone, #4291). Empty means no live
// re-check.
func (c *PendingChoice) liveZone() ZoneKind {
	switch {
	case isEntryCardChoiceKind(c.Kind):
		return entryCardChoiceZone(c.Kind)
	case isPermanentPickKind(c.Kind):
		return ZoneBattlefield
	case isKeyedChooseCards(c):
		return c.ChooseLiveZone
	}
	return ""
}

// setCardSetCandidates replaces a choose-cards prompt's candidate list
// and re-clamps the bounds to it — THE one place a shortened list and
// its Min / Max / Count are kept honest, shared by the zone prune
// above and by reassignChoiceLocked's departure prune.
//
// Two callers, one rule: an enumerator that offers a bigger set than
// the resolver accepts is the #544 wedge, and so is a floor no
// remaining set can reach. Clamping down is CR 701.9a's "as many as
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
		if c == nil || (c.Kind != PendingChoiceOptionPick && c.Kind != PendingChoiceEntryController) || len(c.PickOptions) == 0 {
			continue
		}
		opts, ok := keepSeats(c.PickOptions, func(id uuid.UUID) bool {
			p := g.playerByIDLocked(id)
			return p != nil && !p.Eliminated
		})
		if !ok && c.Kind == PendingChoiceEntryController {
			// ADR 0102: every opponent the permanent could have been
			// given to has left. The entry still happens — under its
			// would-be controller, the no-opponent rule reached late —
			// so the paused event is resumed rather than dropped.
			g.removeChoiceAtLocked(i)
			g.settleEntryControllerWithoutChoiceLocked(c)
			continue
		}
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
// waiting on a player's answer — a replacement's "may" prompt, such as
// CR 903.9b's "send your commander to the command zone instead?" for a
// bounce or a tuck, which pauses a card on its way off the battlefield.
// A death or an exile is not paused by CR 903.9 (ADR 0115).
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
	g.runScryAfterLocked(chooserID, choice.Source, choice.scryThen, len(topOrder))
	g.runStateChecksLocked()
	return nil
}

// runScryAfterLocked runs a look-at's keyed tail (#4253 slice 3), reporting
// a failure as an effect error. The tail is told how many cards went back
// on top (TailOutcome.Amount): "for each card you put on top of your
// library" (Starving Revenant). Zero when nothing was looked at.
//
// Caller must hold g.mu.
func (g *Game) runScryAfterLocked(chooser, source uuid.UUID, keyed KeyedTail, kept int) {
	if err := g.runKeyedTailLocked(keyed, nil, kept); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    chooser,
			Source:   source,
			ErrorMsg: err.Error(),
		})
	}
}

// surveilRouted is a surveil's continuation once every card the player
// put into the graveyard has settled (#4361 S11c): the EventSurveil, then
// the rest of the effect (Next) with how many cards stayed on top. Player
// is the surveilling player, N the size of the surveil (the whole
// looked-at set, which does not move with a replacement) and Nums[0] the
// number kept on top. The IDs are what the route really put in the
// graveyard: a leg a Rest in Peace exiled instead is not in them. An
// on-disk key: never renamed, never reused.
var surveilRouted = RegisterTail("engine/surveil-routed", func(g *Game, o TailOutcome) error {
	t := o.Tail
	g.EmitEvent(Event{
		Kind:     EventSurveil,
		Actor:    t.Player,
		Source:   t.Source,
		Amount:   len(o.IDs),
		LookedAt: t.N,
	})
	kept := 0
	if len(t.Nums) > 0 {
		kept = t.Nums[0]
	}
	// The rest of the effect, now that the library is in the order the
	// player chose and every card that was going to a graveyard has
	// settled.
	g.runScryAfterLocked(t.Player, t.Source, nextTailValue(t.Next), kept)
	return nil
})

// IsLookAtTopKind reports whether a choice kind belongs to the scry
// family: the prompts that show their chooser a set of cards and ask
// where in a library they go. Three of them are a look at the top of
// the chooser's own library; the fourth, put_in_library (ADR 0088),
// orders cards on their way into a library from wherever they are.
//
// The four share PendingChoice.ScryCards, the chooser-only wire
// redaction and the client dialog (put_in_library has its own resume
// slot, because its continuation is handed the answer) — so the places
// that care about the family rather than the individual keyword (the
// view projection, the legal-move enumerator, the dispatcher) ask
// here instead of listing the kinds. Added in S22; the fourth member
// in #996.
func IsLookAtTopKind(k PendingChoiceKind) bool {
	switch k {
	case PendingChoiceScry, PendingChoiceSurveil, PendingChoiceLookAtTop, PendingChoicePutInLibrary:
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
	g.runScryAfterLocked(chooserID, choice.Source, choice.scryThen, len(topOrder))
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
	want, err := partitionChoiceCards(lookedAt, away, topOrder)
	if err != nil {
		return nil, err
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
// The graveyard leg emits an EventZoneMove per card, library to
// graveyard, and NOT the EventMill a mill emits (#2809, ADR 0013 §5q
// amendment 2026-10-23): surveil is its own keyword action (CR
// 701.25a), not a mill (CR 701.17a), so "whenever one or more cards
// are milled" (Mirelurk Queen, Zellix) does not see it. A "whenever a
// card is put into your graveyard from your library" payoff watches
// both kinds and reads the zone pair, so it still does. The completed
// surveil then emits EventSurveil once, for the payoffs that watch
// the keyword itself.
//
// #931: the graveyard leg goes through the shared exit primitive
// (a route batch with millRoute, the template it shares with an
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
	keyed := choice.scryThen
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
	// Surveil is not a mill (CR 701.25a): the leg is an ordinary
	// library-to-graveyard zone move (#2809).
	r.Mill = false
	r.Source = choice.Source
	// #4361 S11c: the rest of the surveil is data (surveilRouted), so a
	// leg paused on a CR 616 prompt is a restore point.
	after := KeyedTail{Ref: surveilRouted, Player: chooserID, Source: choice.Source, N: lookedAt, Nums: []int{len(topOrder)}}
	if !keyed.IsZero() {
		next := copyKeyedTail(keyed)
		after.Next = &next
	}
	err = g.routeAllKeyedLocked(r, graveyard, after)
	if err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}
