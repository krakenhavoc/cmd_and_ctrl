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
// usually the battlefield). Per CR 608.2m, abilities cease to exist
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
	// exist (CR 608.2m).
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

	// Label is a free-text caller-provided string for ability items
	// ("Goblin Bombardment damage", "Counterspell ETB"). Empty for
	// spells (the card name is sufficient labeling).
	Label string

	// Targets are the announce-time target slots. Empty slice means
	// "no targets" (TargetNone or no targeting clause). Re-checked at
	// resolution per CR 608.2b — see resolveTopOfStackLocked in
	// mutations.go.
	Targets []TargetRef

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

	// HoldPriority signals that the controller did NOT want priority
	// to rotate to the next seat after the announce. Used for
	// chaining casts (e.g. cast Lightning Bolt holding priority, then
	// cast another spell on top). PassPriority honours this flag and
	// keeps PriorityHolder unchanged for one cycle.
	HoldPriority bool

	// SplitSecond marks an item as having split second (CR 702.79).
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
}
