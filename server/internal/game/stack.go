package game

import "github.com/google/uuid"

// stack.go holds the data model for the S13.1 stack: stack items
// (spells / activated abilities / triggered abilities), target
// references, and the cast-time / activation-time parameters that
// flow into them.
//
// The Stack zone (Game.Stack *Zone) keeps the Card slice as the
// source of truth for "which cards are on the stack." StackMeta
// (Game.StackMeta map[uuid.UUID]*StackItem, keyed by Card.InstanceID)
// runs in parallel and carries the announce-time choices: targets,
// modes, X, distribution, hold-priority, split-second.
//
// This dual representation keeps the existing zone plumbing
// (snapshot / clone / WS broadcast / FilterViewFor) untouched, while
// giving the new cast/resolve code a place to attach metadata
// without bloating the Card struct with stack-only fields.

// StackItemKind discriminates between the three things that can sit
// on the stack: cast spells, activated abilities, and triggered
// abilities. Spells reference an actual Card (in Game.Stack);
// abilities reference their *source* card (which stays where it is —
// usually the battlefield). Per CR 608.2n, abilities cease to exist
// on resolution; spells route to graveyard or battlefield.
type StackItemKind string

const (
	// StackItemSpell — a cast spell. Card is the spell itself; resolution
	// routes permanents to battlefield, instants/sorceries to graveyard.
	StackItemSpell StackItemKind = "spell"

	// StackItemActivated — an activated ability. SourceCardID points to
	// the card that produced the ability; the ability ceases to exist
	// on resolution.
	StackItemActivated StackItemKind = "activated"

	// StackItemTriggered — a triggered ability. Same shape as activated
	// for resolution semantics; flagged separately so the UI can label
	// them distinctly and APNAP ordering can apply.
	StackItemTriggered StackItemKind = "triggered"
)

// TargetRefKind discriminates the four shapes of an announce-time
// target reference per the cast/activate UI.
type TargetRefKind string

const (
	// TargetPlayer — a seated player.
	TargetPlayer TargetRefKind = "player"

	// TargetCard — a card by instance ID (battlefield / graveyard /
	// stack — any zone the announce UI surfaced).
	TargetCard TargetRefKind = "card"

	// TargetSelf — the spell or ability targets its own controller.
	// Used for "you draw a card" and similar self-targeting effects.
	TargetSelf TargetRefKind = "self"

	// TargetNone — no target. Used for spells like Wrath of God that
	// have no targeting clause. ID is uuid.Nil.
	TargetNone TargetRefKind = "none"
)

// TargetRef is a single target slot captured at announce time. The
// existence-only re-check on resolution (CR 608.2b) reads ID; if the
// referenced player or card has left the game, the slot is
// considered illegal.
type TargetRef struct {
	Kind TargetRefKind
	ID   uuid.UUID

	// Slot is the index of the target CLAUSE this ref answers,
	// within the clause list that governed the announcement
	// (TargetSpec.Clause(Slot)). Zero — the value every ref carried
	// before #764, in every snapshot and on every wire message — is
	// "the only clause", so nothing needed migrating.
	//
	// Announce order is clause order, so a positional reader
	// (item.Targets[0] is the biter, [1] the victim) is still
	// correct; the slot is what lets the CR 608.2b re-check run each
	// pick against its OWN predicate instead of a union. Added by
	// #764 (ADR 0065 §2).
	Slot int

	// Mode is the index into StackItem.Modes — the mode OCCURRENCE,
	// not the option — whose clause list Slot indexes. Zero for a
	// non-modal item and for the first chosen mode. A repeated mode
	// (CR 700.2d) is two occurrences with the same option index and
	// different Mode values, which is what gives each occurrence its
	// own target group. Added by #764.
	Mode int
}

// StackItem is the metadata for one item on the stack. Lives in
// Game.StackMeta keyed by the underlying Card's InstanceID for
// spells, or by a synthetic ID for abilities (which don't have a
// real card on the stack — the source card stays put).
type StackItem struct {
	// ID is the stack-item identity. For spells this equals the
	// underlying Card.InstanceID (and the card lives in Game.Stack).
	// For abilities this is a freshly-minted uuid.UUID and the source
	// card stays in its origin zone.
	ID uuid.UUID

	// Kind discriminates spell / activated / triggered. Spells route
	// to battlefield or graveyard on resolve; abilities cease to
	// exist (CR 608.2n).
	Kind StackItemKind

	// Controller is the player who cast / activated this item. They
	// retain priority after announcing (CR 117.3c). Counter actions
	// route to the controller's graveyard by default.
	Controller uuid.UUID

	// Owner is the player who owns the card (relevant for spells —
	// resolution sends the card to its owner's graveyard, not the
	// controller's). For abilities, equals Controller.
	Owner uuid.UUID

	// SourceCardID points to the card that produced this stack item.
	// For spells, equals ID (the spell IS the card). For abilities,
	// it's the card the ability came from — usually a battlefield
	// permanent that stays in place.
	SourceCardID uuid.UUID

	// SourceEpoch is Card.ObjectEpoch read off the source AS THE ITEM
	// WAS PUT ON THE STACK — the announce-time identity of the OBJECT
	// the ability came from, not of the card (CR 400.7).
	//
	// It exists for the one question SourceCardID cannot answer: is
	// the permanent on the battlefield right now the same permanent
	// whose ability this is? A Loxodon Warhammer bounced and replayed
	// while its equip is on the stack keeps its instance ID and is a
	// NEW OBJECT, and an ability that acts on its own source must
	// refuse the impostor exactly as it refuses a source that simply
	// left. AbilitySourceObjectGoneLocked is that read; equip
	// (AttachSourceForEffect) is its first caller.
	//
	// Stamped by the two announce paths that put an ABILITY of a
	// permanent on the stack — the catalog activation (activated.go)
	// and the manual sandbox one (ActivateAbility) — which are also
	// the only two places a StackItemActivated is built. Meaningless
	// on any other kind, and AbilitySourceGoneForEffect reads the
	// KIND rather than treating some value as a sentinel: zero is a
	// real epoch, since a token created straight onto the
	// battlefield has never changed zones.
	SourceEpoch int

	// Label is a free-text caller-provided string for ability items
	// ("Goblin Bombardment damage", "Counterspell ETB"). Empty for
	// spells (the card name is sufficient labeling).
	Label string

	// DoubledBy identifies the CR 603.2d effect that added this trigger
	// instance. It is empty on the original and on non-trigger items.
	DoubledBy     uuid.UUID
	DoubledByName string

	// Targets are the announce-time target slots. Empty slice means
	// "no targets" (TargetNone or no targeting clause). Re-checked at
	// resolution per CR 608.2b — see resolveTopOfStackLocked in
	// mutations.go.
	Targets []TargetRef

	// Payload is what the effect that CREATED this item had to tell
	// it — the cards that were revealed, the creature that was
	// sacrificed, the player that was chosen. Only a reflexive
	// trigger (CR 603.12, reflexive.go) carries one today; it is
	// empty on every cast spell and every harvested trigger, whose
	// Build reads the event instead.
	//
	// Separate from Targets because the two mean different things
	// and one of them is rewritten: Targets is the announce-time
	// TARGET choice, replaced wholesale when a pick_target prompt is
	// answered, so a payload parked there would be lost on exactly
	// the triggers that most need one. Nothing re-checks Payload
	// against the board — a payload is a record of what happened,
	// not a target, so hexproof and CR 608.2b are both irrelevant to
	// it and the Effect checks what it still needs itself.
	//
	// Added in #636.
	Payload []TargetRef

	// Trigger is what the TRIGGERING EVENT was, for an item the
	// harvester put here — the damage amount, the object that left
	// with its CR 603.10 last-known characteristics, the counters
	// placed, the spell cast, the ability activated. Nil on every
	// cast spell, every activated ability and every reflexive
	// trigger, which have no triggering event to carry.
	//
	// Stamped once, where the item is built, and read in two places
	// that are a priority round apart: the ability's TARGET CLAUSE as
	// it is put on the stack (CR 603.3d — Scrap Trawler's "with
	// lesser mana value" is lesser than a mana value only this field
	// remembers), and its EFFECT at resolution, through
	// effects.Context.Trigger().
	//
	// DATA rather than a value captured in the Effect closure, which
	// is where every card that needed one used to put it. A closure
	// reaches resolution and reaches nothing else: not the clause
	// that runs before it, not the snapshot (a func has no wire
	// form), and not — the reason this field exists at all — a CR
	// 707.10 COPY of the ability, which "copies any choices made when
	// it triggered" and must resolve against the same event the
	// original does. See trigger_event.go.
	//
	// Added in #1223.
	Trigger *TriggerContext

	// Modes is the list of mode indices chosen at announce time for
	// modal spells / abilities. Empty for non-modal items.
	Modes []int

	// XValue is the announce-time value of X for variable-cost spells.
	// Sandbox: the engine doesn't recompute mana cost — the player
	// already paid manually. Surfaced on the wire so opponents see X.
	XValue int

	// Distribution maps target ID → portion for "divide N among
	// targets" spells. Sandbox: just data capture, not enforcement.
	Distribution map[uuid.UUID]int

	// Paid is what this announcement actually cost: the counters
	// removed or added, the life paid, and — since #761 — the mana
	// tokens that left the pool, or the fact that the engine waived
	// the charge (PaidCost.OnPaper).
	//
	// DATA, not a closure, and stamped at announce for the same
	// reason XValue is: by the time the item resolves the counters
	// are gone, the mana has been spent and the Treasure that made
	// it may have been sacrificed, so nothing downstream could
	// recompute any of it. Converge, sunburst, adamant, "if no mana
	// was spent" and "for each counter removed this way" are all one
	// read of this field.
	//
	// A COPY of a spell carries an empty record (CR 707.10, the
	// Dawnglow Infusion ruling): mana is not an object, so nothing
	// was spent to cast the copy. That is a real zero rather than an
	// unknown — NoManaSpent is true of a copy, and it should be.
	//
	// Added in #789 (the counter half) and #761 (the mana half).
	Paid PaidCost

	// HoldPriority signals that the controller did NOT want priority
	// to rotate to the next seat after the announce. Used for
	// chaining casts (e.g. cast Lightning Bolt holding priority, then
	// cast another spell on top). PassPriority honours this flag and
	// keeps PriorityHolder unchanged for one cycle.
	HoldPriority bool

	// CastFromZone is the zone a spell was cast from — "hand", or
	// "command" for a commander, or "exile" under an impulse grant
	// (S21 sub-PR 6). Empty for ability items, which aren't cast from
	// anywhere.
	//
	// Stamped at announce because it stops being readable a moment
	// later: CR 601.2a moves the card to the stack, and nothing on
	// the card remembers where it came from. Wash Away's "target
	// spell that wasn't cast from its owner's hand" is the first
	// consumer; the source-zone question is a common one in
	// Commander (Appa's "whenever you cast a spell from exile" wants
	// the same fact on EventCast). Added in S22.
	CastFromZone ZoneKind

	// AltCost is the Key of the alternative cost paid to cast this
	// spell — "overload", "evoke", "cleave" — or empty when the
	// caster paid the printed mana cost (CR 118.9). Read back at
	// resolution through effects.Context.PaidAltCost, so a card whose
	// text changes with the cost branches on it: an overloaded
	// Cyclonic Rift bounces everything, a hard-cast one bounces one
	// thing, and they are the same StackItem shape apart from this
	// field. Also drives evoke's sacrifice-on-entry trigger and the
	// CR 608.2b re-check's choice of target clause. Added in S22.
	AltCost string

	// Foretold marks a spell cast from a FORETOLD CARD — CR 702.143c's
	// "if this spell was foretold", which Poison the Cup, Haunting
	// Voyage and Starnheim Unleashed read.
	//
	// Distinct from `AltCost == "foretell"`, and the difference is the
	// rule: a spell is foretold because the CARD was foretold, not
	// because the foretell cost was the one paid. An effect that let
	// its owner cast a foretold card some other way would still be
	// casting a foretold card.
	//
	// Stamped at announce from the card in its source zone, because
	// that is the last moment the fact is readable: CR 406.3a turns
	// the card face up as it is cast, and ADR 0069 decision 5 makes
	// that MoveCard's unconditional ClearFaceDown. Added for #658.
	Foretold bool

	// FaceDown is the CR 708.2 state a spell CAST FACE DOWN
	// (CR 708.4) resolves into — FaceDownMorphed for a morph or a
	// megamorph, FaceDownDisguised for a disguise, and the zero value
	// for every other spell in the game.
	//
	// It rides the item because the permanent is a different object
	// from the spell (CR 400.7) and MoveCard clears the face-down
	// state on every zone change (ADR 0069 decision 5) — so "it was
	// cast face down, therefore it enters face down" has to be
	// carried across the move rather than read off the card on the
	// far side of it. Stack resolution seeds it onto the entry event
	// and the ONE entry finisher applies it, which is how the CR 614
	// window, the pause-and-resume and the undo path come free.
	//
	// Distinct from `AltCost == "morph"` for the reason Foretold is
	// distinct from `AltCost == "foretell"`: the fact is about the
	// OBJECT, and an effect that cast a card face down some other way
	// would still produce a face-down permanent. Added for #1194
	// (ADR 0082 decision 3).
	FaceDown FaceDownKind

	// AltCostExiles is CR 702.34a's flashback clause as a FACT about
	// this stack object: "exile this card instead of putting it
	// anywhere else any time it would leave the stack".
	//
	// It rides the item rather than being re-derived from the catalog
	// because since ADR 0066 an alternative cost can be GRANTED — a
	// card Snapcaster gave flashback to was cast for a cost the
	// catalog has never heard of, and the permission that granted it
	// is usually gone by the time the spell leaves the stack. CR
	// 400.7g is the rule: the granted ability is part of the object on
	// the stack.
	//
	// False on a spell cast for its printed cost, and on a game
	// restored from a snapshot written before this field — for which
	// altCostExilesFromStack still falls back to the catalog, so a
	// printed flashback behaves exactly as it did.
	AltCostExiles bool

	// IsCopy marks a spell item that is a COPY of another spell
	// (CR 707.10) rather than a cast card — Reverberate's output,
	// Twincast's, the second half of a storm count. Set only on
	// StackItemSpell items created by CopySpellForEffect.
	//
	// One rule hangs off it and it is the one that matters: a copy
	// is not a card, so when it finishes resolving (or is countered
	// by game rules for illegal targets) it CEASES TO EXIST rather
	// than going to a graveyard. Route a copy to the graveyard
	// instead and a phantom Lightning Bolt accumulates in somebody's
	// yard, where it is countable by Tarmogoyf, castable by
	// flashback, and returnable by Regrowth. See
	// ceaseToExistLocked in spell_copy.go.
	//
	// It is NOT a "don't fire triggers" flag. A resolving copy deals
	// its damage, draws its cards and fires everything a cast spell
	// would; what a copy never does is trigger "whenever you CAST",
	// and that falls out of CopySpellForEffect emitting no
	// EventCast rather than out of a check on this field.
	//
	// Added in S30 (#95).
	IsCopy bool

	// SplitSecond marks an item as having split second (CR 702.61).
	// While any stack item has SplitSecond set, no further casts /
	// activations are legal except mana abilities and special
	// actions. Game.SplitSecondActive mirrors this for fast lookup.
	SplitSecond bool

	// Seq is the StackMeta insertion order, stamped via
	// nextStackSeqLocked when the item lands in the map. Ability
	// resolution picks the highest Seq so abilities resolve LIFO
	// (CR 608.1) instead of in map-iteration order. Items from
	// snapshots predating the field carry 0 and tie-break
	// arbitrarily among themselves.
	Seq uint64

	// Effect is the resolution callback for ability items (activated
	// / triggered). resolveTopAbilityLocked runs it after the CR
	// 608.2b target re-check passes and the EventResolve breadcrumb
	// is emitted; the item has already been removed from StackMeta
	// by then. Nil means the ability has no engine-side effect on
	// resolution — the S13.1 manual-sandbox shape, where players
	// resolve the ability by hand (AnnounceTrigger, ActivateAbility
	// for non-catalog cards).
	//
	// Runs under g.mu held in write mode: the callback MUST NOT call
	// public locking mutators — stay on the *ForEffect surface in
	// effect_api.go. It receives the live game rather than capturing
	// one so an item copied into an undo snapshot (cloneStackItem)
	// resolves against whichever Game it's restored into. For the
	// same reason it must not capture pointers into zone slices —
	// read the source / controller / targets off `item` instead.
	//
	// Not serialised: StackItem never crosses the wire (the protocol
	// package projects its own StackItemView) and games aren't
	// persisted, so a func field is safe here. Added in S19 — the
	// "triggers actually use the stack" completeness fix.
	Effect func(g *Game, item *StackItem) error

	// Ordered marks a pending trigger whose controller has already
	// answered a CR 603.3b ordering prompt covering it. The APNAP
	// drain re-prompts a seat only when it holds ≥2 differing
	// triggers of which at least one is not yet Ordered — so a
	// resolved prompt drains without asking again, while a trigger
	// that arrives afterwards re-opens the question with the full
	// list. Meaningless once the item is on the stack. Added in S19
	// sub-PR 8.
	Ordered bool

	// modeSpec is the ModeSpec an ability item was announced under,
	// so the CR 608.2b re-check can find the clause list of the mode
	// OCCURRENCE a TargetRef names. Nil for a spell (looked up by
	// oracle ID, like its targetSpec) and for every non-modal
	// ability. Unexported: set by the harvester and the activation
	// path, shared by cloneStackItem, re-derived on restore. Added
	// by #764.
	modeSpec *ModeSpec

	// targetSpec is the S20 TargetSpec a targeted ability item was
	// built against, so resolveTopAbilityLocked can run the same
	// predicate re-check spells get (CR 608.2b). Nil for spells
	// (their spec is looked up by oracle ID) and for free-form
	// abilities (existence check only). Unexported: set by the
	// harvester, copied by cloneStackItem. Added in S20 sub-PR 2.
	targetSpec *TargetSpec
}
