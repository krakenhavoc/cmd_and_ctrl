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
// max-hand-size discards (those are simpler: chooser == owner,
// and the cursor auto-resumes when the map drains). Effect-
// driven choices that go through PendingChoices keep resolving
// asynchronously — the spell routes to graveyard immediately,
// the pick is made later by a resolve_choice action.
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
//	"mode_pick"         — chooser picks a mode index for a modal
//	                      spell.
//	"mill_reveal"       — chooser picks which of the revealed top-N
//	                      library cards go where (Brainstorm's
//	                      put-two-back half).
type PendingChoiceKind string

const (
	PendingChoiceDiscardFromHand PendingChoiceKind = "discard_from_hand"

	// PendingChoiceMana — the chooser picks one color from a fixed
	// option set (Birds of Paradise's WUBRG, Arcane Signet's
	// commander-identity subset). On resolve the picked color drops
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

	// ColorOptions is the legal-picks list for PendingChoiceMana.
	// Uppercase single-character entries (W/U/B/R/G/C). Unused for
	// other choice kinds. Populated server-side so the client
	// renders a color picker with exactly the right buttons
	// (commander-identity-filtered for Arcane Signet, the full five
	// for Birds of Paradise). Added in S15 sub-PR 2.
	ColorOptions []string

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

	// ManaAmounts is how many mana a PendingChoiceMana adds for each
	// colour in ColorOptions — "{T}: Add three mana of any one color"
	// (Gilded Lotus) is ONE pick minting three tokens, and Nyx Lotus's
	// amount differs per colour (its devotion to that colour). nil, or
	// a colour missing from the map, means one: every ordinary pick.
	// Parsed from the produced-mana grammar's "{W3|U3}" form (see
	// ParseProducedMana). Deep-copied by clone.go and carried by the
	// snapshot. Added for #742.
	ManaAmounts map[string]int

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

	// pickTargetResume is the server-only continuation for a
	// PendingChoicePickTarget: the captured event / source / LKI,
	// the Build closure, and the spec the pick is validated against
	// (and stamped onto the item for the resolution re-check).
	// Added in S20 sub-PR 2.
	pickTargetResume *pickTargetFrame

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
	confirmResume *confirmFrame

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
type pickTargetFrame struct {
	ev     Event
	source Card
	lki    Characteristic
	build  func(ev Event, source *Card, sourceLKI Characteristic, g *Game) *StackItem
	spec   *TargetSpec
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
	ability TriggeredAbility
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
func (g *Game) QueueChoiceForEffect(choice PendingChoice) uuid.UUID {
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
// On success, applies the kind-specific side effect (for
// discard_from_hand: move each pick to FromPlayer's graveyard,
// emit EventDiscardCard per pick), dequeues the entry, and
// returns nil.
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
		for _, id := range picks {
			if _, err := MoveCard(from.Hand, from.Graveyard, id); err != nil {
				return err
			}
			g.markCardKnownInZoneLocked(from.Graveyard, id)
			g.EmitEvent(Event{
				Kind:    EventDiscardCard,
				Actor:   from.ID,
				CardID:  id,
				OldZone: ZoneHand,
				NewZone: ZoneGraveyard,
			})
		}
	default:
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)
	return nil
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
	for k := 0; k < n; k++ {
		p.ManaPool.AddMana(ManaToken{
			Color:  color,
			Source: choice.Source,
			// The choice carried the ability's restrictions here so
			// the minted token gets them (#352). copyRestrictions
			// because the choice is about to be dequeued and the
			// token outlives it.
			Restrictions: copyRestrictions(choice.ManaRestrictions),
		})
		g.EmitEvent(Event{
			Kind:   EventManaAdded,
			Actor:  chooserID,
			Source: choice.Source,
		})
	}
	g.dequeueChoiceLocked(idx)
	return nil
}

// dequeueChoiceLocked drops the choice at index idx, preserving
// slice order for the rest. Caller must hold g.mu.
func (g *Game) dequeueChoiceLocked(idx int) {
	if idx < 0 || idx >= len(g.PendingChoices) {
		return
	}
	g.PendingChoices = append(g.PendingChoices[:idx], g.PendingChoices[idx+1:]...)
	if len(g.PendingChoices) == 0 {
		g.PendingChoices = nil
	}
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
	if err != nil {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	return g.finishSettledReplacementLocked(ev, out)
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
	case RepEventStepTransition:
		if ev.StepTransitionSeat >= 0 && ev.StepTransitionSeat < len(g.Seats) {
			return g.Seats[ev.StepTransitionSeat].ID
		}
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
	if err != nil {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}

	// Apply-loop settled. Dispatch the underlying mutation per
	// ev.Kind using the (possibly mutated) event payload. Added in
	// S17 sub-PR 3 so Doubling Season + Hardened Scales actually
	// land counters after the CR 616 prompt resolves.
	defer g.clearReplacementEventLocked(ev.ID)
	return g.finishSettledReplacementLocked(ev, out)
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
		return g.applyCounterLocked(ev.CounterTarget, ev.CounterName, ev.CounterDelta)
	case RepEventDraw:
		return g.actuallyDrawCardLocked(ev.DrawPlayer)
	case RepEventLife:
		// #482: a life change carries everything its resume needs on
		// the event itself — the player, the settled delta and the
		// source the log credits — so it is finished by exactly the
		// code the unpaused path runs. This branch used to be its own
		// copy of the tail, and it was already one field behind: it
		// emitted EventChangeLife with no Source, so a life change
		// that paused lost the card that caused it. See life_tail.go.
		err := g.applyResolvedLifeChangeLocked(ev)
		if errors.Is(err, ErrPlayerNotFound) {
			// The player left between the prompt and the answer. The
			// life change simply does not happen — but the choice is
			// already dequeued, so returning the error here would
			// fail the action AND take the prompt away with nothing
			// to show for it. Log it and move on.
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
		if errors.Is(err, ErrCardNotFound) || errors.Is(err, ErrPlayerNotFound) {
			// The target left between the prompt and the answer. The
			// damage simply does not happen — but the choice is
			// already dequeued, so returning the error here would
			// fail the player's action AND take their prompt away
			// with nothing to show for it. Log it and move on.
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
	case RepEventMove:
		// #529: a move that came through the shared exit primitive
		// carries everything its resume needs on the event itself, so
		// it is finished by the same code the unpaused path runs.
		// Checked first because it is the general case — the two
		// branches below are the older, hand-rolled resumes for the
		// battlefield entry and battlefield-leave paths.
		if ev.zoneRoute != nil {
			return g.executeZoneRouteLocked(ev)
		}
		// S17 sub-PR 6: resume path for battlefield-leave moves
		// after the CR 903.9 commander-zone Optional prompt. The
		// pipeline settled on ev.NewZone (either the original
		// destination if owner said "no", or ZoneCommand if they
		// said "yes"). Run the physical move through the shared
		// executeBattlefieldLeaveLocked helper.
		if ev.entryResumable && ev.NewZone == ZoneBattlefield && ev.OldZone != ZoneBattlefield {
			// A paused ENTRY — the shockland's pay-2-life prompt,
			// or a CR 616 ordering prompt between two enters-tapped
			// effects. The pipeline function bailed before moving
			// anything, so the push happens here. See
			// executeEntryToBattlefieldLocked.
			return g.executeEntryToBattlefieldLocked(ev)
		}
		if ev.OldZone != ZoneBattlefield {
			// Other non-LTB moves (graveyard → hand for a regrow,
			// say) don't have a resume path yet.
			return nil
		}
		var owner *Player
		if card, ok := g.LookupCardForEffect(ev.CardID); ok {
			owner = g.playerByIDLocked(card.Owner)
		}
		return g.executeBattlefieldLeaveLocked(ev.CardID, ev.NewZone, ev.NewZoneOwner, owner)
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
	if ev == nil || ev.Kind != RepEventStepTransition {
		return nil
	}
	return g.applyResolvedReplacementEventLocked(ev)
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
			ev:      ev,
			source:  source,
			lki:     lki,
			build:   ability.Build,
			ability: ability,
		},
	})
}

// queuePickTargetLocked queues the CR 603.3d target choice for a
// targeted trigger. The legal set is computed now (the caller
// already confirmed it's non-empty) and frozen onto the prompt;
// ResolvePickTarget re-validates the pick against the spec anyway,
// since the board can change while the prompt is open. Caller must
// hold g.mu. Added in S20 sub-PR 2.
func (g *Game) queuePickTargetLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility) {
	lt := g.legalTargetsLocked(source.Controller, t.Targets)
	label := t.Targets.Label
	if label == "" {
		label = "Choose a target"
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoicePickTarget,
		Chooser:           source.Controller,
		Count:             1,
		Source:            source.InstanceID,
		Reason:            label,
		PickTargetPlayers: lt.Players,
		PickTargetCards:   lt.Cards,
		PickTargetMin:     t.Targets.Min,
		PickTargetMax:     t.Targets.Max,
		pickTargetResume: &pickTargetFrame{
			ev:     ev,
			source: source,
			lki:    lki,
			build:  t.Build,
			spec:   t.Targets,
		},
	})
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
	if frame == nil || frame.build == nil || frame.spec == nil {
		g.dequeueChoiceLocked(idx)
		return nil
	}
	for _, t := range targets {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
	}
	if err := g.validateTargetsLocked(chooserID, frame.spec, targets); err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)
	source := frame.source
	item := frame.build(frame.ev, &source, frame.lki, g)
	if item == nil {
		return nil
	}
	item.Targets = append([]TargetRef(nil), targets...)
	item.targetSpec = frame.spec
	g.queueHarvestedTriggerLocked(item)
	// CR 603.3d / 115.7: a triggered ability's targets are chosen as
	// it is put on the stack, which is right here. Emitted after the
	// queue so a "becomes the target" trigger stacks above the
	// ability that targeted. Added in S22 for Monk Gyatso.
	g.emitBecameTargetLocked(item.Controller, item.SourceCardID, item.ID, targets)
	g.runStateChecksLocked()
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
	g.buildOrPickTriggerLocked(frame.ev, frame.source, frame.lki, frame.ability)
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
		Kind:    PendingChoicePayUnless,
		Chooser: chooser,
		Count:   1,
		Source:  source,
		Reason:  question,
		PayCost: cost,
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
	if !p.ManaPool.SpendMana(cost, 0) {
		return false
	}
	g.EmitEvent(Event{
		Kind:   EventManaSpent,
		Actor:  p.ID,
		Source: source,
	})
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
	g.dequeueChoiceLocked(idx)
	if err := g.sacrificePermanentLocked(cardID); err != nil {
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
			g.dequeueChoiceLocked(i)
			continue
		}
		c.SacrificeOptions = live
	}
}

// pausedZoneChangeStaleLocked reports whether the exit stashed on a
// queued prompt's resume frame can still happen.
//
// A paused exit moves NOTHING: the card sits in its old zone until
// the prompt is answered (see routeCardToZoneLocked). So a card that
// is no longer in the zone its paused move says it is leaving has
// already left by some other route, and that move will never happen.
// Answering the prompt then either fails — the battlefield-leave
// resume cannot find the card on the battlefield, which is the
// "card instance not found in zone" of #605 — or, worse, moves the
// card a SECOND time out of a zone nobody asked about, because the
// shared exit primitive moves it from wherever it finds it.
//
// Battlefield ENTRIES are excluded: executeEntryToBattlefieldLocked
// locates the card by scan and short-circuits when it is already on
// the battlefield, so an entry prompt is never stale in this sense.
//
// Caller must hold g.mu.
func (g *Game) pausedZoneChangeStaleLocked(frame *replacementResumeFrame) bool {
	if frame == nil || frame.ev == nil {
		return false
	}
	ev := frame.ev
	if ev.Kind != RepEventMove || ev.NewZone == ZoneBattlefield {
		return false
	}
	src := g.findCardZoneLocked(ev.CardID)
	return src == nil || src.Kind != ev.OldZone
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
		if ev == nil || ev.Kind != RepEventMove || ev.NewZone == ZoneBattlefield {
			continue
		}
		if ev.CardID == cardID {
			return true
		}
	}
	return false
}

// pruneStaleZoneChangeChoicesLocked drops every queued prompt whose
// paused exit can no longer happen, and releases the CR 614.5
// once-per-event bookkeeping the abandoned event was holding.
//
// Called at the two points a card actually lands — the shared exit
// primitive and the battlefield-leave resume — because that is the
// moment a sibling prompt about the same card goes stale, and the
// state-check loop that would otherwise notice is exactly what does
// NOT run while a choice is queued. Same reasoning, and the same call
// sites, as pruneSacrificeChoicesLocked.
//
// Caller must hold g.mu.
func (g *Game) pruneStaleZoneChangeChoicesLocked() {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || !g.pausedZoneChangeStaleLocked(c.replacementResume) {
			continue
		}
		g.clearReplacementEventLocked(c.replacementResume.ev.ID)
		g.dequeueChoiceLocked(i)
	}
}

// dropStaleReplacementResumeLocked is the answer-path half of the
// prune above: a belt-and-braces check that the move a just-answered
// prompt was asking about is still possible. Returns true when the
// frame was abandoned, in which case the caller returns nil — the
// prompt is already dequeued, and refusing the answer instead would
// leave the seat holding a prompt it can never discharge.
//
// Caller must hold g.mu.
func (g *Game) dropStaleReplacementResumeLocked(frame *replacementResumeFrame) bool {
	if !g.pausedZoneChangeStaleLocked(frame) {
		return false
	}
	g.clearReplacementEventLocked(frame.ev.ID)
	g.EmitEvent(Event{
		Kind: EventEffectError,
		ErrorMsg: "replacement prompt dropped: its card is no longer in the " +
			string(frame.ev.OldZone) + " the paused move would leave",
	})
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
	// Graveyard first: the cards that stay on top end up above a
	// library that no longer contains the ones that left, which is
	// the order the printed instruction reads in.
	for _, id := range graveyard {
		c := pulled[id]
		p.Graveyard.PushTop(c)
		g.markCardKnownInZoneLocked(p.Graveyard, c.InstanceID)
		g.EmitEvent(Event{
			Kind:    EventMill,
			Actor:   chooserID,
			Source:  choice.Source,
			CardID:  c.InstanceID,
			OldZone: ZoneLibrary,
			NewZone: ZoneGraveyard,
		})
	}
	// topOrder is top-first and PushTop appends to the top, so walk it
	// backwards: the last push lands on top and must be the caller's
	// first entry.
	for i := len(topOrder) - 1; i >= 0; i-- {
		p.Library.PushTop(pulled[topOrder[i]])
	}
	g.EmitEvent(Event{
		Kind:   EventSurveil,
		Actor:  chooserID,
		Source: choice.Source,
		Amount: len(graveyard),
	})
	// The rest of the effect, now that the library is in the order the
	// player chose.
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
