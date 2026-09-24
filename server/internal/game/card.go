package game

import "github.com/google/uuid"

// Card is a single instance of a Magic card inside a running game. One
// physical card = one Card value; if a player plays two copies of the
// same printed card from their library, there are two Card instances,
// each with its own InstanceID.
//
// Card carries the printed data stamped at deck import (Scryfall
// fields, faces, keywords) alongside the per-instance runtime state
// the rules engine keeps (tap, combat, damage, knowledge, layers,
// attachments). Fields are grouped by the mechanic that added them,
// except the bools, which share one block at the end for alignment.
// TestCardHasNoInteriorPadding (card_layout_test.go) guards the
// layout; put a new bool in that block.
type Card struct {
	// InstanceID uniquely identifies this physical card within the game.
	// Generated when the card enters play or when a deck is imported.
	InstanceID uuid.UUID

	// Name is the printed card name. Placeholder in S02; replaced by a
	// lookup key into the Scryfall cache in S04.
	Name string

	// ScryfallID is the Scryfall UUID for the card's printing, stamped
	// by the deck importer at seat time (S05). Empty only for
	// placeholder cards (e.g. the demo game seeded via
	// CMDCTRL_SEED_DEMO). The client uses this to resolve image URIs
	// by hitting GET /cards/{id}/image.
	ScryfallID string

	// OracleID is the Scryfall oracle-level card identity, stable
	// across printings (every printing of Lightning Bolt shares one
	// oracle_id). Used by the S14+ card-effect catalog as the
	// lookup key so a deck importing a specific printing still
	// matches the catalog entry. Empty for placeholder / demo-seed
	// cards. Added in S14 sub-PR 4.
	OracleID string

	// TokenKey is a TOKEN template's synthetic catalog key
	// ("token:treasure", "token:food") — the identity a token has
	// instead of an oracle ID, so that its printed abilities can live
	// in the catalog like every other object's (#521).
	//
	// Empty for every card, and for a token whose template declares
	// no abilities at all (a vanilla 1/1 Soldier has nothing to look
	// up). Set by the token template constructors in
	// cards/effects/tokens.go, which register the matching catalog
	// entry at boot; see game/token_key.go for the namespace and why
	// it cannot collide with an oracle ID.
	//
	// It is part of the COPIABLE VALUES (PrintedValues.TokenKey): CR
	// 707.2 makes a copy of a Treasure token a Treasure, ability
	// included. A copy of a printed CARD carries that card's oracle
	// ID and an empty key, which is what keeps CR 707.2's ordinary
	// direction working — CatalogKey prefers the oracle ID and reads
	// this only when there is none.
	//
	// A name, never a closure, so the snapshot carries it and an undo
	// deep-copies it — which is the whole point: a Treasure on the
	// battlefield no longer blocks every restore point in the game.
	TokenKey string

	// TypeLine is Scryfall's type line ("Legendary Creature — Human
	// Wizard", "Land", "Sorcery", etc.). Stamped at deck-import time
	// (S08) so combat-rule gates can check whether a card is a
	// creature without a round-trip back to the cards index. Empty
	// for placeholder cards (the demo seed) and for any card the
	// importer was unable to resolve type info for.
	TypeLine string

	// Power and Toughness are the printed creature stats, parsed
	// from Scryfall's strings at deck-import time. Zero for non-
	// creatures and for any card whose printed stats are non-numeric
	// (e.g. "*" for cards like Mortivore — handled manually until
	// rules enforcement grows). Used by ResolveCombatDamage to
	// auto-apply unblocked attacker damage. Added in S08.
	//
	// A Toughness of 0 is therefore ambiguous on its own: it is a
	// printed 0/0 on a Hangarback Walker and "no number here" on a
	// Mortivore, a token template or a test fixture. VariableToughness
	// records which, and ToughnessIsKnown is the one place that
	// decides — do not re-derive the answer from this field.
	Power     int
	Toughness int

	// ManaCost is the printed casting cost, copied from Scryfall at
	// deck-import time — e.g. "{1}{R}", "{X}{B}{B}", "{W/U}". Empty
	// for lands and for placeholder / demo-seed cards. The S15
	// cost validator parses this at cast time into a ParsedCost.
	// Carried on every game.Card instance so the view layer can
	// surface it onto CardView.ManaCost without a round-trip back
	// to the cards index. Added in S15 sub-PR 1.
	ManaCost string

	// ProducedMana lists the mana colors this permanent can produce
	// via any of its mana abilities. Entries are uppercase single-
	// character letters from {"W","U","B","R","G","C"}. Empty for
	// non-producers. The S15 auto-tapper uses this to prune the
	// candidate set; basic lands get a synthetic mana ability
	// derived from their TypeLine regardless of what Scryfall says
	// here. Added in S15 sub-PR 1.
	ProducedMana []string

	// Colors is the card's printed color list — uppercase letters
	// from {"W","U","B","R","G"} — as Scryfall computes it (color
	// indicators and Devoid included). Empty means colorless OR
	// "not stamped" (tokens, test fixtures); HasColor / IsColorless
	// fall back to deriving colors from ManaCost in that case. S20
	// targeting predicates read this. Added in S20 sub-PR 1.
	Colors []string

	// ColorIdentity is the card's Commander colour identity (CR
	// 903.4) — uppercase letters from {"W","U","B","R","G"}, copied
	// verbatim from Scryfall's top-level `color_identity` at deck
	// import. Distinct from Colors: identity folds in mana symbols
	// in rules text, colour indicators, and — crucially — BOTH
	// faces of a double-faced card, which is why it is the only
	// colour data that survives Scryfall's null top-level
	// mana_cost / colors on a `transform` or `modal_dfc` record.
	//
	// Issue #276: commanderIdentityFor used to derive identity from
	// Effective().Colors, falling back to the printed mana cost.
	// Both are empty for a DFC commander, so the identity came back
	// empty and every "any colour in your commander's identity"
	// pipe (Command Tower, Arcane Signet, Fellwar Stone) skipped
	// narrowing and offered all five colours. deck/validate.go has
	// always read this field off cards.Card correctly — it simply
	// had no path onto game.Card. Empty for tokens and fixtures,
	// where the pre-#276 derivation still applies.
	ColorIdentity []string

	// StartingLoyalty is the printed loyalty a planeswalker enters
	// the battlefield with (CR 306.5b), parsed from Scryfall's
	// `loyalty` string at deck-import time. Zero for every other
	// card type, and for planeswalkers whose printed loyalty is
	// non-numeric (Chandra, Fire of Kaladesh's back face prints a
	// number, but X-loyalty walkers and tokens do not).
	//
	// This lives on the card — not in the effect catalog — because
	// it is printed data like Power / Toughness / ManaCost, not
	// card-effect data. Issue #274: while the only source was
	// effects.Spec.StartingLoyalty, every planeswalker outside the
	// opt-in catalog entered with zero loyalty counters and was
	// immediately moved to the graveyard by the CR 704.5i SBA.
	// The catalog value survives as a fallback for cards with no
	// printed data (tokens, fixtures) — see CatalogStartingLoyalty.
	StartingLoyalty int

	// Keywords are printed keyword abilities carried on the card
	// object itself rather than looked up in the catalog by oracle
	// ID. Tokens are the reason this exists: a token has no oracle
	// ID, so CatalogPrintedKeywords can never find it, and before
	// S21 every token's flying / deathtouch was cosmetic. The token
	// template declares them here and printedCharacteristic folds
	// them in, so the layer engine treats them like any other
	// printed keyword. Added in S21 sub-PR 1.
	//
	// Since #317 / #319 / #320 this is also the road ORDINARY cards
	// travel: the deck importer stamps Scryfall's `keywords` array
	// here (lowercased and filtered to the keywords the engine
	// enforces — see deck.printedKeywords and CanonicalKeyword).
	// Before that, a printed keyword only existed if someone had
	// hand-written a catalog Spec for the card, which left
	// vigilance, flash, flying and the rest inert on roughly 7,500
	// cards — the bird that tapped when it attacked, the flash
	// creature the server refused at instant speed. The catalog's
	// Spec.PrintedKeywords survives alongside it, merged and
	// deduped by printedCharacteristic, for cards that never go
	// through deck import.
	//
	// Empty for cards from neither source (test fixtures) and for
	// the ~78% of real cards that print no keyword at all.
	Keywords []string

	// GrantedAbilities are the catalog keys of ability bundles a
	// CR 707.9a copy effect granted to this object — Phantasmal
	// Image's "and it has 'when this creature becomes the target of a
	// spell or ability, sacrifice it'", Sakashima's "{2}{U}{U}:
	// return it".
	//
	// It sits here, in the flat printed fields, for the same reason
	// Keywords and TypeLine do: a granted ability is part of the
	// COPIABLE VALUES (CR 707.9a's second sentence), and
	// PrintedValues is the portable snapshot of exactly these
	// fields. CopiableValuesOf reads it, applyCopy writes it, and a
	// copy of a copy therefore inherits the grant with no code of its
	// own. CatalogKey folds it into the composite catalog key that
	// makes the abilities findable (copy_grants.go).
	//
	// Names, never closures, so the snapshot carries it and an undo
	// deep-copies it. Nil for every card that was never granted
	// anything, which is all of them but a handful.
	GrantedAbilities []string

	// NeedsEffect lives in the bool block at the end of Card, for alignment.

	// ManaAbilities are mana abilities carried on the card object,
	// for the same reason as Keywords: a Treasure token's "{T},
	// Sacrifice this artifact: Add one mana of any color" can't come
	// from a catalog lookup. ManaAbilitiesForCard prefers these over
	// the catalog and the synthetic basic-land shape. Added in S21
	// sub-PR 1.
	ManaAbilities []ManaAbilityShape

	// ActivatedAbilities are CR 602 activated abilities carried on
	// the card object — the third and last of the catalog hooks a
	// token can't reach, since all three key on oracle ID. Food,
	// Clue and Blood tokens are defined entirely by their activated
	// ability ("{2}, Sacrifice this artifact: Draw a card"), so
	// without this they'd be blank artifacts. ActivatedAbilitiesForCard
	// prefers these over the catalog. Added in S21 sub-PR 4.
	ActivatedAbilities []ActivatedAbilityShape

	// Owner is the player who brought this card to the game. Ownership
	// is fixed at deck-build time and never changes.
	Owner uuid.UUID

	// Controller is the player who currently controls the card. May
	// differ from Owner for stolen permanents, Control Magic effects,
	// etc. Always equal to Owner for cards not on the battlefield.
	Controller uuid.UUID

	// Tapped lives in the bool block at the end of Card, for alignment.

	// BattleX, BattleY are the normalised position of a card on the
	// battlefield, as fractions of the battlefield area (each in the
	// range [0, 1]; the server clamps on write). Only meaningful on the
	// battlefield — cleared when the card leaves, alongside Tapped and
	// Counters. Normalised so a rendering resolution change doesn't
	// invalidate saved snapshots. Cards entering the battlefield
	// default to (0, 0) until the client stamps a drag-release.
	BattleX float64
	BattleY float64

	// Counters is a generic per-card counter map (+1/+1, -1/-1, loyalty,
	// charge, fade, etc.). nil means no counters. S02 does not interpret
	// counters; they're just storage until rules enforcement grows.
	Counters map[string]int

	// IsCommander lives in the bool block at the end of Card, for alignment.

	// AttackingTarget is the player ID this card has been declared to
	// attack. uuid.Nil means "not declared as attacker". Set by
	// DeclareAttacker, cleared by ClearCombat or zone exit. Only
	// meaningful on the battlefield. Added in S08.
	AttackingTarget uuid.UUID

	// BlockingTarget is the attacker instance ID this card has been
	// declared to block. uuid.Nil means "not declared as blocker".
	// Set by DeclareBlocker, cleared by ClearCombat or zone exit.
	// Only meaningful on the battlefield. Added in S08.
	BlockingTarget uuid.UUID

	// GoadedBy is the player ID who goaded this creature. uuid.Nil
	// means "not goaded". A goaded creature must attack each combat
	// (and not the goader) under MTG rules; the sandbox surfaces the
	// marker but doesn't enforce the must-attack constraint until rules
	// graft work lands. Cleared on zone exit alongside Tapped /
	// AttackingTarget. Added in S10.
	GoadedBy uuid.UUID

	// PhasedOutBy is the player who controlled this permanent AT THE
	// MOMENT IT PHASED OUT (CR 702.26a), and it is meaningful only
	// while the card sits in Game.PhasedOut. uuid.Nil on every
	// phased-in permanent.
	//
	// Not Controller, and not derivable from it: CR 502.1 phases in
	// "all phased-out permanents that the active player controlled
	// when they phased out", and CR 702.26f lets a control-changing
	// continuous effect expire while a permanent is phased out. An Act
	// of Treason creature phased out under YOUR control phases in
	// during YOUR untap step even though control has reverted by then.
	//
	// See phasing.go and ADR 0084. Added in S46 (#1199).
	PhasedOutBy uuid.UUID

	// PhaseInLockedBy names the object whose presence on the
	// battlefield is this phase-out's duration — Oubliette's "phases
	// out until this enchantment leaves the battlefield", Out of
	// Time's "phase out until this enchantment leaves the
	// battlefield".
	//
	// uuid.Nil is ordinary phasing: the permanent phases in during its
	// controller's next untap step (CR 502.1). Set, it phases in the
	// moment the named object stops being on the battlefield, which is
	// what the cards do — Out of Time's whole point is that the
	// creatures come back at once when the last time counter goes,
	// mid-turn.
	//
	// See phasing.go and ADR 0084. Added in S46 (#1199).
	PhaseInLockedBy uuid.UUID

	// DamageMarked is the damage currently noted on the creature this
	// turn, used by the lethal-damage state-based action (CR 704.5g).
	// Combat damage and direct-damage spells (resolved manually) write
	// into this field; the cleanup-step turn-based action zeroes it.
	// Only meaningful for creatures on the battlefield. Added in S13.1.
	DamageMarked int

	// RegenerationShields is how many regeneration shields (CR
	// 701.19a) this permanent is carrying. Each one replaces the next
	// destruction of this permanent THIS TURN — tap it, remove all
	// damage from it, remove it from combat — and is used up doing so.
	// Cleared by the cleanup step alongside DamageMarked, and by the
	// battlefield exit, because a shield belongs to the permanent that
	// was given one and the card in the next zone is a new object
	// (CR 400.7).
	//
	// A COUNT rather than a flag because "regenerate it twice" is two
	// shields and survives two destructions (CR 701.19a's "the next
	// time"), and a count rather than a Duration-scoped entry
	// (ADR 0063) because a shield is not a continuous effect: it
	// changes no characteristic, it is consumed rather than expiring
	// when it applies, and it belongs to ONE object — which is exactly
	// what a per-object integer says. The engine reads it from one
	// place, the CR 701.19 built-in replacement in
	// builtin_replacements.go. See regeneration.go.
	//
	// It is NOT a counter (CR 122): nothing that reads, removes,
	// doubles or proliferates counters may see it, so it deliberately
	// does not live in Card.Counters. Added in #667.
	RegenerationShields int

	// FaceDown lives in the bool block at the end of Card, for alignment.

	// KnownBy is the per-instance "who currently knows this card's
	// identity" set (S13.5). Sticky across zone moves: once a player
	// sees a card face-up, they stay in the set until a knowledge-
	// clearing event (shuffle, mulligan-into-library, library-bottom)
	// removes them. Initialised always-non-nil by NewCard so the
	// helpers don't have to allocate defensively. Drives the hub
	// filter's per-viewer redaction in S13.5.
	KnownBy map[uuid.UUID]bool

	// EnteredBattlefieldAt is the Unix-nano timestamp when the card
	// most recently entered the battlefield. Drives the S16 layer-
	// engine's CR 613 timestamp-ordering: when two static abilities
	// affect the same characteristic, the one whose source has the
	// earlier timestamp applies first. Stamped by the layer listener
	// on every EventZoneMove with NewZone == battlefield; cleared
	// (zeroed) on battlefield-leave so a re-entered permanent gets
	// a fresh timestamp. Zero ⇒ never on the battlefield in this
	// game's lifetime. Added in S16 sub-PR 3.
	EnteredBattlefieldAt int64

	// ObjectEpoch counts how many times this card has changed zones,
	// and so is the identity of the OBJECT rather than of the card
	// (CR 400.7: "an object that moves from one zone to another
	// becomes a new object with no memory of its previous
	// existence"). Bumped once per move in MoveCard — every zone
	// change, in both directions, because the rule has no exceptions
	// — and it never resets: an epoch is a serial number, not a
	// count of anything a player can see.
	//
	// The engine keeps several per-object registries keyed by
	// instance ID, which is the CARD's identity and survives a zone
	// change. Those that must forget are dropped at the one
	// battlefield exit (battlefield_exit.go). TurnTally's per-ability
	// counts could not be, because the CR 726 loop breaker shares
	// their key and a blink loop would reset its own run every
	// iteration (#936) — so they are keyed by (instance, epoch)
	// instead: the returning object reads a key nothing has written,
	// while the breaker keeps counting the card. See
	// ObjectTallyKey in turn_tally.go.
	//
	// Not a characteristic and not on the wire: nothing renders it
	// and no rule reads the number itself, only whether two readings
	// of it are equal.
	ObjectEpoch int

	// SummonedThisTurn and MarkedLethalByDeathtouch live in the bool
	// block at the end of Card, for alignment.

	// Layout is Scryfall's printing layout, copied verbatim at deck
	// import: "normal", "transform", "modal_dfc", "adventure",
	// "split", "prepare", … The cast path branches on it to decide
	// what a face CHOICE means — see CastableFaces and faceOnResolve
	// in face.go. Empty for tokens, fixtures and the demo seed,
	// which are all single-faced. Added by ADR 0034.
	Layout string

	// Faces is every printed face of a multi-face card, front first.
	// nil for the ~33,000 single-faced oracle IDs, which keep the
	// pre-ADR-0034 behaviour to the byte.
	//
	// The flat printed fields above (Name, TypeLine, ManaCost,
	// Colors, Power, Toughness, StartingLoyalty) are the
	// MATERIALISATION of Faces[ActiveFace], not an independent copy
	// of Scryfall's top-level record. That is deliberate: making
	// them methods would have rewritten 293 TypeLine: struct-literal
	// sites and 389 Card{} literals, whereas leaving them as fields
	// means all 74 Is*() call sites keep compiling and START being
	// right, since "the characteristics of the face that's currently
	// up" is exactly CR 712.8a.
	Faces []Face

	// ActiveFace indexes Faces.
	//
	// INVARIANT: the flat printed fields equal Faces[ActiveFace].
	// Maintained by SetFace and by nothing else — never assign this
	// field directly, or the card desynchronises and no test will
	// catch it. AssertFaceInvariant (face_test.go) walks a finished
	// game and checks exactly this.
	ActiveFace int

	// BaseController is the controller this permanent reverts to when
	// every control-changing continuous effect on it ends (CR 613.1b)
	// — the player who controlled it when it entered the
	// battlefield.
	//
	// Captured LAZILY by the layer recompute (which runs before
	// anything can read a control-changed value, because every
	// battlefield entry bumps the layer version) and cleared by
	// MoveCard on battlefield exit, so it is zero exactly when "the
	// current controller IS the base" holds. That is why no write
	// site had to learn about it: all ~15 places that assign
	// Card.Controller do so as a permanent ENTERS, before the
	// capture.
	//
	// Meaningless off the battlefield. Added in S24 with the layer-2
	// control change (Mind Control).
	BaseController uuid.UUID

	// AttachedTo is the CR 301.5c / CR 303.4 attachment relation,
	// stored on the ATTACHED object (the Equipment or the Aura) and
	// pointing at its host. Zero value (Kind == "") means
	// "unattached", which is every card in every zone but a handful
	// of battlefield permanents.
	//
	// TargetRef rather than a bare uuid.UUID because a Curse
	// enchants a PLAYER and an Aura or Equipment enchants a CARD,
	// and player IDs and card instance IDs are both UUIDs with
	// nothing to tell them apart. TargetRef already carries that
	// discrimination, it is already what the targeting pipeline
	// produces, and it is already on the wire as TargetRefView — so
	// the aura attach path assigns item.Targets[0] verbatim.
	//
	// A VALUE type, not a pointer, deliberately: cloneCard starts
	// `out := c` and deep-copies only the slice / map fields by
	// hand, so a pointer here would alias between the live game and
	// every undo snapshot. Same reason AttackingTarget /
	// BlockingTarget / GoadedBy are uuid.Nil-sentinel values.
	//
	// Cleared on battlefield exit by MoveCard alongside Tapped and
	// the combat relations. The REVERSE direction (a host that
	// left) is not swept eagerly — a dangling ref fails the CR
	// 704.5m/n state-based action on the next pass, which is
	// exactly why those rules are state-based actions. Added in
	// S24, per ADR 0036 decision 1.
	AttachedTo TargetRef

	// AttachedAt is the CR 613.7e timestamp: an Equipment's or
	// Aura's continuous effect gets a NEW timestamp when it becomes
	// attached, not the one it got when it entered the
	// battlefield. The layer engine prefers this over
	// EnteredBattlefieldAt when non-zero. Zero means "not attached"
	// (or attached before this field mattered), and the engine
	// falls back to the entry stamp.
	//
	// Unobservable for every card in S24's first cut — every
	// attachment static in the catalog is a layer 6 grant or a 7c
	// modify, and both are commutative. It exists so that the first
	// 7b "set" that meets a 7c "modify" is right by construction.
	// Added in S24, per ADR 0036 decision 2.
	AttachedAt int64

	// FaceTurnedAt is the CR 613.7f timestamp: "a permanent receives
	// a new timestamp each time it turns face up or face down". The
	// layer engine orders this permanent's own static abilities by
	// the LATEST of EnteredBattlefieldAt, AttachedAt and this
	// (layerTimestamp), so a morph turned face up applies its statics
	// after everything already on the battlefield.
	//
	// A third field rather than a re-stamp of EnteredBattlefieldAt,
	// because that one is also the permanent's IDENTITY pin — every
	// "until end of turn" and control effect is keyed on
	// (InstanceID, EnteredBattlefieldAt), and CR 708.8 says turning
	// face up is not a new object — and it drives summoning
	// sickness's "entered this turn". Stamped by turnFaceUpLocked and
	// TurnFaceDownForEffect (both directions, #1271), zeroed on
	// battlefield exit, carried by clone and snapshot. Zero means
	// "never turned since it entered".
	FaceTurnedAt int64
	// NamedTribe is the creature type chosen for this permanent by an
	// "as this enters, choose a creature type" instruction (CR
	// 614.12) — Cavern of Souls, Door of Destinies, Vanquisher's
	// Banner, Adaptive Automaton. Empty means no type has been
	// chosen, which is both "this card has no such instruction" and
	// the transient state between the permanent entering and its
	// controller answering the prompt.
	//
	// Per-INSTANCE, not per-card: two Caverns name two different
	// tribes, and the static abilities that read it take the source
	// card, never the catalog Spec. It is the first piece of chosen
	// state the engine keeps on a permanent, which is why it is a
	// plain string rather than a map — a second one (a named colour,
	// for Iona or Painter's Servant) can be a second field, and a
	// map would only pay for itself at four or five.
	//
	// Cleared when the permanent leaves the battlefield, alongside
	// Tapped and Counters: a Cavern that is bounced and replayed
	// chooses again (CR 614.12 fires on each entry). Added in S26.
	NamedTribe string

	// ChosenColor is the colour chosen for this permanent by an "as
	// this enters, choose a color" instruction (CR 105.4) — Coldsteel
	// Heart, Heraldic Banner, the Thriving lands. One uppercase letter
	// (W/U/B/R/G), or empty when none has been chosen. Same lifecycle
	// as NamedTribe: per instance, carried by the snapshot, cleared
	// when the permanent leaves the battlefield. Added for #742; see
	// color_choice.go.
	ChosenColor string

	// ChosenPlayer is the player chosen for this permanent by an "as
	// this enters, choose a player" instruction (CR 614.12) —
	// True-Name Nemesis. uuid.Nil when none has been chosen, which is
	// both "this card has no such instruction" and the window between
	// the permanent entering and its controller answering the prompt.
	//
	// Same lifecycle as NamedTribe and ChosenColor: per INSTANCE (two
	// Nemeses name two different players), carried by the snapshot and
	// by clone, and cleared when the permanent leaves the battlefield
	// (CR 400.7) — a Nemesis that is bounced and recast chooses again,
	// and one in a graveyard is protected from nobody.
	//
	// NOT a copiable value (CR 707.2), and that falls out of where it
	// lives rather than out of a rule anybody has to remember:
	// CopiableValuesOf projects printed characteristics and never looks
	// here, so a Clone of a Nemesis chooses its own player as IT enters.
	//
	// The one reader is protection.go: "protection from the chosen
	// player" (CR 702.16k) resolves against this field, comparing the
	// SOURCE'S CONTROLLER rather than any characteristic of it — the
	// one quality in the grammar that does. Read it with
	// ChosenPlayerOf; written by ResolveOptionPick through the
	// as-enters prompt in choose_player.go and by nothing else.
	// Added for #980; see ADR 0072 §7.
	ChosenPlayer uuid.UUID

	// ChosenName is the CARD NAME chosen for this permanent by an "as
	// this enters, choose a card name" instruction (CR 614.12) —
	// Pithing Needle, Phyrexian Revoker, Sorcerous Spyglass, Meddling
	// Mage, Nevermore. Empty means no name has been chosen, which is
	// both "this card has no such instruction" and the window between
	// the permanent entering and its controller answering the prompt.
	//
	// FREE TEXT, and that is the one way it differs from its three
	// siblings. NamedTribe is validated against CR 205.3m's 345-word
	// vocabulary and ChosenColor against five letters; CR 201.2 lets
	// a player name ANY card name, including one in no deck at the
	// table and one this server has never seen, so there is nothing
	// to validate against. ResolveCardNameChoice checks the SHAPE
	// (trimmed, non-empty, length-capped) and nothing else.
	//
	// Same lifecycle as NamedTribe, ChosenColor and ChosenPlayer: per
	// INSTANCE (two Needles name two different cards), carried by the
	// snapshot and by clone, and cleared when the permanent leaves
	// the battlefield (CR 400.7) — a Needle that is bounced and
	// recast names again.
	//
	// NOT a copiable value (CR 706.2), and that falls out of where it
	// lives rather than out of a rule anybody has to remember:
	// CopiableValuesOf projects printed characteristics and never
	// looks here, so a Clone of a Pithing Needle names its own card
	// as IT enters.
	//
	// Compare it with CardNameMatches, never with ==: the comparison
	// asks every FACE (CR 201.2b) and is case-insensitive on trimmed
	// strings. Read it with ChosenNameOf; written by
	// ResolveCardNameChoice and by nothing else. Added for #1210; see
	// choose_card_name.go.
	ChosenName string

	// Provenance is what this permanent remembers about the SPELL it
	// came from — CR 400.7d, "an ability of a permanent can reference
	// information about the spell that became that permanent as it
	// resolved, including what costs were paid". Zero value means
	// "this permanent did not come from a spell, or came from one cast
	// for its mana cost and nothing else".
	//
	// It exists because an entering permanent's own trigger cannot
	// reach the stack item. "When Gatekeeper of Malakir enters, IF IT
	// WAS KICKED" and "When Phlage enters, sacrifice it UNLESS IT
	// ESCAPED" are both TriggeredAbilities whose Build receives the
	// game, the source and the event — and by the time EventETB is
	// emitted the item is out of StackMeta. So the entry finisher
	// stamps the record here, in the one moment that holds both the
	// landed permanent and the item.
	//
	// ONE RECORD, not one per fact. #653 (the alternative cost) and
	// #664 / ADR 0073 §5 (the optional costs) arrived a week apart and
	// shipped as two fields with the same lifecycle, the same two
	// clears, the same snapshot handling and the same stamp three
	// lines apart. They are one question — "how was the spell that
	// became this permanent cast?" — and CR 400.7d asks it once.
	//
	// Same lifecycle as NamedTribe and ChosenColor and for the same
	// reason: it belongs to the ENTRY, not to the card. Stamped in
	// executeEntryToBattlefieldLocked, cleared by MoveCard on the way
	// off the battlefield (CR 400.7), carried by clone and the
	// snapshot. Read through Escaped / CastProvenanceForEffect and,
	// for the kicker half, CardKickedTimes / CardPaidOptionalCost.
	//
	// Added in S42 (#653, #664).
	Provenance CastProvenance

	// FaceDownKind is WHY this object is face down (ADR 0069). Empty
	// exactly when FaceDown is false; the two are written only by
	// SetFaceDown / ClearFaceDown, so "face down with no rule
	// attached" is unrepresentable.
	//
	// It sits here with the other strings rather than beside FaceDown
	// in the bool block: a 16-byte string dropped into that block
	// strands a bool and fails TestCardHasNoInteriorPadding (#620 /
	// #633).
	//
	// The kind, not the zone, is what every face-down question is
	// answered from — who may look (faceDownViewersLocked), whether
	// there is a CR 708.2 body (FaceDownIsPermanent), whether the
	// catalog is silent (CatalogKey). A Card knows nothing about
	// where it is, and keeping the answers on the kind is what lets
	// them live on Card methods with no *Game in reach.
	FaceDownKind FaceDownKind

	// PrintedSelf is this card's OWN printed values, stashed when a
	// CR 707 copy effect overwrote the flat printed fields above.
	// nil — which is every card that is not a Clone-class permanent
	// — means the printed fields are the card's own and nothing has
	// to be undone.
	//
	// It exists because CR 400.7 makes a permanent that changes
	// zones a new object: the copy effect applied to the PERMANENT,
	// so a Clone that dies is a card named Clone in its owner's
	// graveyard, not a second Llanowar Elves. The battlefield-leave
	// branch of the layer listener restores from here, in the same
	// place it clears the effective cache.
	//
	// Carried by the snapshot and deep-copied by clone.go: which
	// card a permanent is a copy of is not derivable from anything
	// else, and a restore that lost it would resurrect every clone
	// on the board as a 0/0. Added in S16.5 (#159 / #335).
	PrintedSelf *PrintedValues

	// FaceDownListed is the CR 708.2 body an effect LISTED for this
	// face-down object — Cyber Conversion's "It's a 2/2 Cyberman
	// artifact creature", Yedora's "It's a Forest land". nil is
	// CR 708.2a's default nameless 2/2 (FaceDownBody), which is every
	// morph, manifest, cloak and every Ixidron'd permanent.
	//
	// It REPLACES the default body rather than decorating it: a
	// face-down Forest is not a creature at all. So it is read in
	// exactly one place, faceDownCharacteristic, the layer-0
	// baseline, and every reader downstream sees it for free.
	//
	// Per-OBJECT data beside the per-kind FaceDownKind, set and
	// cleared with it by SetFaceDownListed / ClearFaceDown — so
	// MoveCard's CR 400.7 clear drops it on every zone change and
	// turning the permanent face up drops it too. Never mutated
	// through the pointer: a writer replaces it whole, and
	// clone.go and the snapshot copy it anyway, PrintedSelf's
	// posture. ADR 0082's second 2026-09-23 amendment (#1270).
	FaceDownListed *FaceDownListing
	// StartingDefense is the printed defense a battle enters the
	// battlefield with (CR 310.4), parsed from Scryfall's `defense`
	// string at deck-import time. Zero for every other card type.
	//
	// The exact sibling of StartingLoyalty, for the exact same
	// reason and after the exact same bug: defense is printed data
	// like Power / Toughness / ManaCost, not card-effect data, and
	// while the only source was the catalog every battle outside the
	// opt-in catalog entered with zero defense counters and was
	// swept into the graveyard by the CR 704.5v/w SBA before anyone
	// could attack it. That was live on `main` for every battle a
	// player could import. The catalog's BattleSpec.Defense survives
	// as a fallback for cards with no printed data — tokens,
	// fixtures — see CatalogBattleDefense.
	//
	// Added in S27.
	StartingDefense int

	// ProtectorPlayerID is the opponent chosen to protect a battle
	// as it enters (CR 310.9a). uuid.Nil for every other card type,
	// and for a battle whose protector prompt has not been answered
	// yet.
	//
	// The protector, NOT the controller, is the player who defends
	// the battle: they are the one whose creatures may block an
	// attack on it, and they are the one player who may not attack
	// it. That inversion is the whole mechanic — a battle is cast by
	// one player and guarded by another — and it is why this cannot
	// be derived from Controller.
	//
	// Cleared when the battle leaves the battlefield, alongside
	// Tapped and the combat declarations: a battle that returns is a
	// new object and chooses a new protector (CR 400.7).
	//
	// Added in S27.
	ProtectorPlayerID uuid.UUID

	// NextUntapSkips records the effects that keep this permanent from
	// untapping during an untap step: one-shot next-untap-step markers
	// (ADR 0058 Decision 2) and, since #1313, holds that last "for as
	// long as" a CR 611.2 duration (UntapSkip.While). It is
	// battlefield state, not a copiable value.
	NextUntapSkips []UntapSkip

	// ClassLevel is the CR 716.2 level designation on a Class
	// permanent — the marker that switches its printed "Level N"
	// abilities on (ADR 0071).
	//
	// ZERO READS AS LEVEL 1. CR 716.2b: a Class permanent with no
	// level designation is level 1, so the zero value is the correct
	// reading for a Class nobody has levelled and for every card that
	// is not a Class at all. Always go through ClassLevelOf, never
	// the field.
	//
	// A DESIGNATION, not a counter, and that is the whole reason it
	// is a field rather than an entry in Counters: nothing
	// proliferates it, nothing doubles it, and no counter-removal
	// cost can spend it (CR 716.4 keeps level counters — the CR
	// 702.87 leveler mechanic — a separate thing).
	//
	// Not copiable (CR 716.2c): a copy of a level-3 Class is level 1.
	// That falls out of where it lives — CopiableValuesOf projects
	// printed characteristics and never looks here.
	//
	// Cleared when the permanent leaves the battlefield (CR 400.7),
	// alongside NamedTribe and ChosenColor. Carried by the snapshot.
	// Written by SetClassLevelForEffect and by nothing else. Added in
	// S46 (#757).
	ClassLevel int

	// PreparedBy is set on a CR 722.3c prepare copy only: the
	// permanent OBJECT — instance and CR 400.7 epoch — whose prepared
	// designation keeps this copy in exile and castable (ADR 0090).
	// Zero on every other card, and zero on a prepare copy that has
	// left exile: MoveCard clears it on every move, so a copy that is
	// cast, countered or put anywhere else no longer names anything
	// and can never become castable again.
	//
	// Read by prepareCopyPermissionLocked, which DERIVES the cast
	// permission from it on every query rather than storing one, and
	// by the CR 704.5e sweep that removes a copy whose permanent is
	// gone or unprepared. Carried by the snapshot.
	PreparedBy PermissionCardRef

	// HiddenBy is set on a card exiled face down by HIDEAWAY (CR
	// 702.75a, ADR 0091): the permanent OBJECT — instance and CR 400.7
	// epoch — whose hideaway exiled it. Two rules read it:
	//
	//   - who may look (FaceDownHidden's viewer is that permanent's
	//     controller, faceDownViewersLocked), and
	//   - which card is "the exiled card" of that permanent's linked
	//     ability (CR 607.2a, HiddenCardsForEffect).
	//
	// Card-carried rather than read back off the event log (the
	// effects package's exiled-with record), because the first reader
	// is the face-down viewer rule, which answers about a Card inside
	// the engine and has no log walk to hand; and because the link is
	// printed on the exiled card itself ("the permanent that exiled
	// this card"). It is exact the same way that record is since #1239:
	// the epoch pins the incarnation, so a hideaway land that is
	// bounced and replayed is a new object with no claim on the card
	// its earlier self hid.
	//
	// Cleared by MoveCard on every move, so a card that leaves exile
	// names nothing ever again. Carried by the snapshot.
	HiddenBy PermissionCardRef

	// Solved, Prepared and PrepareCopy live in the bool block at the
	// end of Card, for alignment.

	// effective is the cached post-layer-resolution characteristic
	// for this card on the battlefield. Populated by the layer
	// engine's recompute pass; nil ⇒ "no recompute has run since
	// this card last entered the battlefield" or "card is not on
	// the battlefield." Card.Effective() reads this when set, falls
	// back to printedCharacteristic() otherwise. Pointer (not value)
	// so the nil sentinel is cheap and the recompute can replace it
	// atomically without partial-update visibility. Added in S16
	// sub-PR 3.
	effective *Characteristic

	// --- bools --------------------------------------------------------
	//
	// Every bool on Card lives here, together, rather than beside the
	// mechanic that added it: a lone bool between two 8-byte fields
	// strands 7 bytes of padding, and six of them cost Card 32 bytes
	// (#35). Each still has a one-line pointer in its original section.
	// TestCardHasNoInteriorPadding fails if a new bool strands padding
	// anywhere else.

	// NeedsEffect records that this card's printed text describes
	// rules only a hand-written catalog Spec can carry out — it is
	// NOT "has oracle text" and it is NOT "is missing from the
	// catalog". A vanilla creature is false; a creature whose whole
	// text is enforced keywords is false; a Forest is false. See
	// coverage.go for the reasoning and NeedsCatalogEffect for the
	// derivation.
	//
	// Stamped by the deck importer from the Scryfall record, the
	// same road Keywords and StartingLoyalty travel, and joined with
	// catalog membership by game.Unimplemented — the one definition
	// the deck-upload summary, the card view and the stack view all
	// read, so they cannot disagree.
	//
	// False for cards that never went through deck import (tokens,
	// fixtures, the demo seed), which means they are never flagged.
	// Deliberate: a missed signal costs a player nothing they
	// weren't already going to learn, and a false one costs the
	// signal its credibility.
	NeedsEffect bool

	// Tapped is the usual MTG tap state. Only meaningful for cards on
	// the battlefield; ignored in other zones.
	Tapped bool

	// IsCommander marks a card as a commander for the Commander format.
	// Commanders live in the command zone at game start.
	IsCommander bool

	// FaceDown is the visual face-down flag (CR 406.3a / CR 708) —
	// "is there a back showing". Distinct from the KnownBy knowledge
	// set: a face-down creature is face-down to everyone visually,
	// but the player the rules let look at it (its controller for a
	// CR 708.5 permanent, its owner for a foretold card) is still in
	// its KnownBy set, so the hover-reveal works on their client.
	//
	// WHY it is face down is FaceDownKind, up with the other strings.
	// Set only through SetFaceDown / ClearFaceDown, never directly;
	// cleared by MoveCard on every zone change (CR 400.7, ADR 0069
	// decision 5) and set again by the destination if the
	// destination is itself a face-down state. Added in S13.5;
	// given a kind by ADR 0069.
	FaceDown bool

	// SummonedThisTurn is the summoning-sickness flag (CR 302.6).
	// Set true whenever the card enters the battlefield; cleared at
	// the start of the controller's untap step. Haste (CR 702.10)
	// is a read-time bypass in HasSummoningSickness, NOT a
	// clear-on-ETB — so a creature that gains haste mid-turn
	// becomes attackable immediately, and one that loses haste
	// mid-turn remains sick until next untap. Only meaningful on
	// the battlefield; ignored in other zones. Added in S18 sub-PR 2.
	SummonedThisTurn bool

	// MarkedLethalByDeathtouch is the S18 deathtouch-mark flag (CR
	// 702.2c — "any nonzero damage from a source with deathtouch
	// causes that damage to be marked as lethal"). Set true when
	// damage from a deathtouch source lands on this creature;
	// read by the lethal-damage SBA. Cleared at StepCleanup
	// alongside DamageMarked. Only meaningful on the battlefield.
	// Added in S18 sub-PR 3.
	MarkedLethalByDeathtouch bool

	// VariableToughness records that the printed toughness is not a
	// number the engine can use — "*", "1+*", "?", or missing from a
	// creature's record altogether — so Toughness is the importer's 0
	// stand-in rather than a printed 0. ToughnessIsKnown reads it: a
	// `*` creature whose characteristic-defining ability this engine
	// does not compute stays out of CR 704.5f even after it loses its
	// last counter, where a real printed 0/0 does not (#683). A `*`
	// the engine DOES compute is back in, because the layer pass
	// answers the question the stand-in was standing in for
	// (Characteristic.PTDefined, #690).
	//
	// Stamped by the deck importer (deck.variableToughness), per face
	// on Faces, and carried by copy effects with the rest of the
	// printed values: Clone-style copies (CopiableValuesOf) and token
	// copies (TokenCopyTemplate). A copy whose exception sets a
	// numeric toughness clears it. False for a card that prints no
	// toughness (lands, instants, artifacts), for the joined top-level
	// type line of a multi-face printing — SetFace(0) overwrites it
	// from the face — and for cards that never went through import
	// (tokens, fixtures, the demo seed).
	VariableToughness bool

	// LostLastCounter records that this object's counters went from
	// some to none: its last counter was removed (an effect, the
	// add_counter action) or cancelled (CR 704.5q). Removing a counter
	// from a card that has none does not set it. It is one of the ways
	// ToughnessIsKnown learns that a 0 is real — the engine WATCHED
	// the number arrive — and without it a Hangarback Walker whose one
	// +1/+1 counter cancels against a -1/-1 counter, or is removed by
	// an effect, would stay on the battlefield as a 0/0 no damage
	// could kill (#683).
	//
	// It is not the only way, and since #691 it is not the way that
	// covers a printed card: a printing behind the object answers the
	// same question for every object that has one, counter history or
	// none, so what this flag still earns its keep for is the objects
	// with NO printing — a token, a test fixture — whose counters have
	// come and gone.
	//
	// Per object, like Counters: cleared wherever Counters is reset
	// for a new object (leaving the battlefield, a token, a spell
	// copy).
	LostLastCounter bool

	// PrintedPTKnown records the opposite of VariableToughness for an
	// object with no printing behind it: the 0 in Toughness is the
	// number the card PRINTS, not the importer's stand-in. It exists
	// for the real printed 0/0 tokens — the living weapon Germ, which
	// is a 0/0 that stays alive only while an Equipment is attached to
	// it, and which without this would sit on the battlefield forever
	// after the Equipment left.
	//
	// A token template opts in by setting it (effects/tokens_table.go).
	// The default stays false, which keeps every body-less test
	// fixture and every template whose P/T the engine genuinely does
	// not know out of CR 704.5f — that skip is #683's and is not
	// weakened here.
	//
	// VariableToughness still wins: a `*` is a stand-in whatever a
	// template claims, and ToughnessIsKnown asks that question first.
	PrintedPTKnown bool

	// Solved is the CR 719.3 designation on a Case permanent — the
	// marker that switches its printed "Solved — [ability]" clauses
	// on (ADR 0071).
	//
	// Set by SolveCaseForEffect, from the "To solve" trigger's
	// resolution, and by nothing else. Once set it STAYS set for as
	// long as the permanent is on the battlefield (CR 719.3b): no
	// card unsolves a Case, and the condition that solved it going
	// false again changes nothing.
	//
	// Not copiable, and cleared when the permanent leaves the
	// battlefield (CR 400.7) — the same two sentences as ClassLevel,
	// up with the other non-bool fields, and for the same reasons.
	// Carried by the snapshot. Added in S46 (#757).
	Solved bool

	// Harnessed is the CR 701.64 designation on a permanent — the
	// marker that switches its printed "∞ — [ability]" clauses on
	// (CR 702.186b, ADR 0071 amendment #1321).
	//
	// Set by HarnessForEffect, from the "Harness [this permanent]"
	// activated ability's resolution, and by nothing else. Once set it
	// STAYS set for as long as the permanent is on the battlefield
	// (CR 701.64b: "it stays harnessed until it leaves the
	// battlefield") — the same two sentences as Solved, for the same
	// reasons: not copiable, cleared when the permanent leaves the
	// battlefield (CR 400.7), carried by the snapshot.
	Harnessed bool

	// Prepared is the CR 722.3a designation on a permanent with a
	// prepare spell (ADR 0090): while it is set, the permanent's
	// controller may cast the CR 722.3c copy of its prepare spell that
	// sits in exile, and casting that copy clears it (CR 601.2i).
	//
	// Set by becomePreparedLocked and by nothing else, which refuses a
	// permanent with no prepare spell and one that is already prepared
	// (both are CR 722.3a). A designation, not a counter and not a
	// characteristic: not copiable, cleared when the permanent leaves
	// the battlefield (CR 400.7) exactly as Solved is, and KEPT across
	// phasing (CR 702.26d) — a permanent "phases in prepared" and makes
	// a fresh copy as it does (CR 722.3c). Carried by the snapshot.
	Prepared bool

	// PrepareCopy marks the CR 722.3c copy of a prepare spell — the
	// object created in exile as a permanent becomes prepared. It is
	// NOT A CARD (CR 707.10, 722.3c), so the CR 704.5e sweep removes it
	// from every zone but the stack, and from exile too once the
	// permanent PreparedBy names is gone or unprepared; its spell, once
	// cast, is a StackItem with IsCopy set and ceases to exist as it
	// leaves the stack. It also keeps the copy wearing its prepare
	// spell in every zone: CR 722.3c makes those characteristics its
	// NORMAL ones, so MoveCard's CR 712.8a front-face reset skips it.
	// Carried by the snapshot.
	PrepareCopy bool

	// PhasedOutIndirect records that this permanent phased out WITH
	// the permanent it is attached to rather than on its own
	// (CR 702.26g, "phasing out indirectly"), and so "won't phase in
	// by itself, but instead phases in along with the permanent it's
	// attached to".
	//
	// Also CR 702.26h: an object that would phase out directly and
	// indirectly at the same time "just phases out indirectly", which
	// is why the mark is decided once for a whole batch in
	// phaseOutSetLocked rather than per card by whoever asked.
	//
	// Meaningful only while the card sits in Game.PhasedOut. Not
	// copiable — CR 707.2 copies characteristics, not status. See
	// phasing.go and ADR 0084. Added in S46 (#1199).
	PhasedOutIndirect bool

	// TapOnPhaseIn is Oubliette's "Tap that creature as it phases in
	// this way" — a rider on ONE phase-in, consumed as it fires.
	//
	// A bool on the card rather than a delayed trigger because the
	// phase-in it rides on is a turn-based action or a duration
	// ending, and neither uses the stack (CR 702.26a), so there is
	// nothing for a trigger to sit on. See phasing.go and ADR 0084.
	// Added in S46 (#1199).
	TapOnPhaseIn bool

	// AbilitiesLostOnRestore marks a card that a restore brought back
	// with FEWER catalog abilities than the restore point recorded
	// (#522): the binary that wrote the file knew abilities for this
	// card that the binary that read it does not — an entry removed,
	// renamed or refactored between two deploys. The owner's decision
	// on #515 is to restore the table anyway, say so loudly, and stop
	// presenting the card as automated: Unimplemented answers true for
	// it, so the client shows the `manual` chip, because whatever the
	// engine still does for it is no longer what it did when the game
	// was captured. A card that comes back with MORE abilities is not
	// flagged — a new build adding abilities is normal.
	//
	// Set only by restoreCard (snapshot.go, abilityShortfallOf) and
	// cleared by NOTHING. It survives zone changes, undo, and every
	// later snapshot, and goes away only when the card itself leaves
	// the game. Carried by the snapshot. Added in S33 (#522).
	AbilitiesLostOnRestore bool
}

// AddKnower marks `viewerID` as having seen this card. No-op for
// uuid.Nil (admin / spectator placeholder). Idempotent. Added in
// S13.5.
func (c *Card) AddKnower(viewerID uuid.UUID) {
	if viewerID == uuid.Nil {
		return
	}
	if c.KnownBy == nil {
		c.KnownBy = make(map[uuid.UUID]bool)
	}
	c.KnownBy[viewerID] = true
}

// AddKnowersAll marks every supplied viewer as a knower. Used by
// public-zone moves (battlefield, stack, exile, graveyard) and by
// reveal effects (Thoughtseize, Telepathy, scry).
func (c *Card) AddKnowersAll(viewerIDs []uuid.UUID) {
	for _, id := range viewerIDs {
		c.AddKnower(id)
	}
}

// ClearKnown drops every knower for this card. Used by shuffle and
// hidden-zone "lose track" cases.
func (c *Card) ClearKnown() {
	c.KnownBy = nil
}

// IsKnownTo reports whether `viewerID` is currently a knower of
// this card's identity. uuid.Nil (admin / spectator) always returns
// true so admin sessions see everything.
func (c *Card) IsKnownTo(viewerID uuid.UUID) bool {
	if viewerID == uuid.Nil {
		return true
	}
	if c.KnownBy == nil {
		return false
	}
	return c.KnownBy[viewerID]
}

// CurrentPower returns the card's combat-relevant power: the
// post-layer effective power (S16: anthems, CDAs, etc.) plus any
// +1/+1 counters, minus any -1/-1 counters. Reads via Effective()
// so layer-7c modifications (Glorious Anthem) and layer-7a CDAs
// (Tarmogoyf) flow through naturally without combat code needing
// to know about the layer engine.
//
// Effective().Power equals printed power for cards with no static
// abilities affecting them, so the pre-S16 behavior is preserved
// for the vast majority of cards. Negative results clamp to zero
// (a -3/-3 modifier on a 2/2 deals no damage, not negative damage).
//
// Caller responsibility: when invoked from a write mutation that
// followed a static-ability-relevant state change (cast a spell,
// move a permanent), call g.RecomputeLayersIfStaleLocked first so
// Effective() reflects the new state. Combat damage and the SBA
// loop both do this at their top.
func (c Card) CurrentPower() int {
	p := c.PowerForComparison()
	if p < 0 {
		return 0
	}
	return p
}

// PowerForComparison includes layers and counters without clamping negative
// values. Skulk compares actual power; CurrentPower clamps damage to zero.
func (c Card) PowerForComparison() int {
	p := c.Effective().Power
	if c.Counters != nil {
		p += c.Counters["+1/+1"]
		p -= c.Counters["-1/-1"]
	}
	return p
}

// CurrentToughness returns the card's combat-relevant toughness:
// the post-layer effective toughness (S16: anthems, CDAs) plus any
// +1/+1 counters, minus any -1/-1 counters. Used by the lethal-
// damage and 0-toughness SBAs (S13.1). May be zero or negative —
// callers compare against DamageMarked directly. NOT clamped (cf.
// CurrentPower) because the SBAs need to tell a real 0 from the
// importer's stand-in 0; ToughnessIsKnown below is where that
// distinction is made, and this number means nothing for an object
// it answers false for.
//
// Same caller responsibility as CurrentPower: ensure
// RecomputeLayersIfStaleLocked has been called for this game state.
func (c Card) CurrentToughness() int {
	t := c.Effective().Toughness
	if c.Counters != nil {
		t += c.Counters["+1/+1"]
		t -= c.Counters["-1/-1"]
	}
	return t
}

// ToughnessIsKnown reports whether the engine knows what this
// object's toughness IS. It is the only gate on the toughness
// state-based action, and the one precondition CR 704.5f has in this
// engine.
//
// CR 704.5f has no precondition in paper: a creature with toughness 0
// or less is put into its owner's graveyard, full stop. The gate
// exists because Card.Toughness is not always a toughness. The deck
// importer parses Scryfall's printed string with strconv.Atoi and
// writes 0 when that fails, so a `*` creature, a printing the
// importer could not resolve, a token template with no body and every
// test fixture that typed a creature line without a body all arrive
// carrying a 0 that means "no number here". Killing those on sight is
// the bug the skip has always existed to prevent; letting the skip
// cover a REAL 0 is #683, #690 and #691.
//
// The answer, in the order it is decided:
//
//  1. A nonzero Toughness, or any counter on the object, was never in
//     question. The fast path, and almost every permanent on almost
//     every board.
//
//  2. A layer 7a or 7b effect DEFINED the P/T in the current pass
//     (Characteristic.PTDefined). Consuming Aberration and Lord of
//     Extinction print `*`, but this engine computes them, so
//     Effective().Toughness is the answer and empty graveyards really
//     do make them 0/0 (#690). This is the branch that lets a CODED
//     characteristic-defining ability out of the skip while an
//     uncoded one stays in it, and the only one that reaches a token,
//     which has no printing behind it.
//
//  3. Otherwise a `*` toughness is the importer's stand-in and the
//     engine does NOT know the number (Card.VariableToughness). A
//     Mortivore nobody has coded keeps the skip, losing its last
//     counter included (#683).
//
//  4. A 0 the engine WATCHED arrive: the object's counters went from
//     some to none (Card.LostLastCounter, #683). Holds for tokens and
//     fixtures as well as for printings.
//
//  5. A token whose template declares its 0 is the printed one
//     (Card.PrintedPTKnown) — the living weapon Germ. Opt-in per
//     template, so the body-less fixtures below keep the skip.
//
//  6. A printing behind the object (ScryfallID). Its 0 came out of
//     Scryfall's printed toughness and parsed as a number, because
//     branch 3 already took every printing where it did not. That is
//     a real printed 0/0 — a Hangarback Walker cast for X=0, a
//     Wildwood Scourge that entered with no counters — and CR 704.5f
//     puts it into its owner's graveyard at once, as it does in paper
//     (#691).
//
// What stays skipped is the set with no printing, no computed P/T and
// no counter history: test fixtures that left the body at 0, and 0/0
// token templates. Both are objects the engine genuinely has no
// toughness for.
//
// Caller responsibility is CurrentToughness': the effective
// characteristic must be fresh, so call RecomputeLayersIfStaleLocked
// first. The state-based action loop does.
func (c Card) ToughnessIsKnown() bool {
	if c.Toughness != 0 || len(c.Counters) > 0 {
		return true
	}
	if c.effective != nil && c.effective.PTDefined {
		return true
	}
	if c.VariableToughness {
		return false
	}
	return c.PrintedPTKnown || c.LostLastCounter || c.ScryfallID != ""
}

// --- card-type predicates ------------------------------------
//
// These read the card's EFFECTIVE types — the post-CR-613 view the
// layer engine computes — not the printed type line. That is the
// whole of #255 / #258 / #344 / #348: before this, layer 4 was
// computed, projected onto the wire, and then invisible to combat,
// state-based actions, targeting and the catalog's own Creature()
// predicate, because every one of them landed here and here read
// Card.TypeLine.
//
// Three properties make the reroute safe:
//
//  1. Card.Effective() is a pure read of the cached resolution. It
//     never triggers a recompute, so a static ability's AppliesTo
//     predicate can call IsLand() from inside the layer pass
//     without re-entering it. A type predicate that DID kick off a
//     recompute would recurse through applyLayerLocked forever;
//     keeping Effective() passive is the invariant that forbids it.
//  2. Off the battlefield the cache is nil and these fall through
//     to the printed type line verbatim — byte-identical to the
//     pre-change behaviour for every card in a hand, library,
//     graveyard, exile or on the stack. CR 113.6: a static ability
//     only does anything while its source is on the battlefield,
//     so there is nothing for the effective view to say there.
//  3. Freshness is the caller's job, exactly as it already was for
//     CurrentPower / CurrentToughness: a write path that reads
//     types after a state change calls
//     g.RecomputeLayersIfStaleLocked first. Every read path goes
//     through ReadSnapshot, which does it for them.
//
// Callers that genuinely want the PRINTED type — CR 707.2 copiable
// values, a deck-construction check, anything that must not move
// when a Blood Moon lands — use the PrintedIs* accessors below.

// IsCreature reports whether the card is a creature right now,
// after continuous effects. Covers "Creature — Human Wizard" and
// "Legendary Artifact Creature — Golem" alike, plus a land a
// Layer-4 static has animated and a Theros god whose devotion gate
// is unmet. An empty type line with no layer effect on it returns
// false (placeholder cards from the demo seed are conservatively
// treated as non-creatures).
func (c Card) IsCreature() bool { return c.HasCardType("creature") }

// IsLand reports whether the card is a land after continuous
// effects.
func (c Card) IsLand() bool { return c.HasCardType("land") }

// IsInstant reports whether the card is an instant. Instants share
// the priority window with activated abilities — they're castable
// any time the caller holds priority.
func (c Card) IsInstant() bool { return c.HasCardType("instant") }

// IsSorcery reports whether the card is a sorcery. Sorceries are
// sorcery-speed only — main phase, stack empty, caller is the
// active player.
func (c Card) IsSorcery() bool { return c.HasCardType("sorcery") }

// IsArtifact reports whether the card is an artifact after
// continuous effects — Mycosynth Lattice's "all permanents are
// artifacts in addition to their other types" lands here.
func (c Card) IsArtifact() bool { return c.HasCardType("artifact") }

// IsEnchantment reports whether the card is an enchantment.
func (c Card) IsEnchantment() bool { return c.HasCardType("enchantment") }

// IsPlaneswalker reports whether the card is a planeswalker.
func (c Card) IsPlaneswalker() bool { return c.HasCardType("planeswalker") }

// IsBattle reports whether the card is a battle (post-MoM card type).
func (c Card) IsBattle() bool { return c.HasCardType("battle") }

// IsPermanent reports whether the card resolves to the battlefield.
// Per CR 110.4, the permanent types are artifact, creature,
// enchantment, land, planeswalker, and battle. Instants and sorceries
// are explicitly NOT permanents (they resolve to the graveyard).
func (c Card) IsPermanent() bool {
	return c.IsArtifact() ||
		c.IsCreature() ||
		c.IsEnchantment() ||
		c.IsLand() ||
		c.IsPlaneswalker() ||
		c.IsBattle()
}

// IsToken reports whether the object is a token (CR 111). Token type
// lines are stamped "Token Creature — Goblin" by the catalog's token
// templates and "Token" is not a card type, so the printed line is
// the test — the one effects.IsToken has always made, now here so the
// engine can make it too.
//
// A token is not a card (CR 108.2), and a token that has left the
// battlefield can't move to another zone or come back onto the
// battlefield (CR 111.8). Since #596 this is also what the CR 704.5d
// state-based action reads — a token in any zone but the battlefield
// ceases to exist at the next state check (token_existence.go). Until
// that check runs, a token tucked into a library (Chaos Warp) is
// still sitting there as an object, and a move that promises "a
// permanent card" out of a hidden zone refuses it with this.
func (c Card) IsToken() bool { return typeLineHas(c.TypeLine, "token") }

// HasCardType reports whether the card's effective card types
// include `lowerType`, which MUST be lowercase (every caller in
// this package passes a literal).
//
// The nil-cache branch is not only an optimisation that keeps a
// hot predicate allocation-free: it makes the off-battlefield
// answer bit-for-bit the pre-layer answer, so a card that never
// reaches the layer engine cannot change behaviour because of this
// file. The cached branch compares whole type tokens instead of
// searching for a substring, which is strictly more accurate —
// "Island" no longer contains a "land" type by accident of
// spelling.
// The face-down guard on the printed branch (and on the three
// accessors below) is ADR 0069 decision 3's stated cost: the VALUE of
// the CR 708.2 body has one definition, faceDownCharacteristic, but
// the READS are where they always were, because these accessors take
// a deliberate fast path off the printed fields when the layer cache
// is cold. Without it a manifested Forest still answers "land" to
// every predicate that runs before the first recompute.
func (c Card) HasCardType(lowerType string) bool {
	if c.effective == nil {
		if c.FaceDownIsPermanent() {
			return typeListHas(faceDownCharacteristic(c).Types, lowerType)
		}
		return typeLineHas(c.TypeLine, lowerType)
	}
	return typeListHas(c.effective.Types, lowerType)
}

// HasSubtype reports whether the card's effective subtypes include
// `subtype`, case-insensitively. This is the accessor a Layer-4
// land-type grant (Urborg, Tomb of Yawgmoth) becomes visible
// through: CR 305.6's intrinsic mana abilities key off the basic
// land TYPE, never off the Basic supertype.
//
// Changeling (CR 702.73a) is answered here rather than by writing
// ~345 subtypes into the Characteristic, for the reasons on
// HasAllCreatureTypes. It is checked AFTER the printed / effective
// list so an ordinary card pays only a slice scan, and it is checked
// on both branches because 702.73a works in every zone — a Woodland
// Changeling in a graveyard really is an Elf, which is what a tribal
// reanimator or a lord counting from exile has to see.
func (c Card) HasSubtype(subtype string) bool {
	// CR 708.2: a face-down permanent has no subtypes but the ones an
	// effect LISTED for it (#1270) — Cyber Conversion's Cyberman,
	// Yedora's Forest — so it is not a Human, not an Elf, and — the
	// reason this branch precedes the changeling check below — not
	// every creature type either: the card underneath's changeling is
	// text the object does not have.
	if c.FaceDownIsPermanent() {
		if c.effective != nil {
			return typeListHas(c.effective.Subtypes, subtype)
		}
		return typeListHas(faceDownCharacteristic(c).Subtypes, subtype)
	}
	if c.effective == nil {
		_, _, printed := ParseTypeLine(c.TypeLine)
		if typeListHas(printed, subtype) {
			return true
		}
	} else if typeListHas(c.effective.Subtypes, subtype) {
		return true
	}
	return IsCreatureType(subtype) && HasAllCreatureTypes(&c)
}

// HasSupertype reports whether the card's effective supertypes
// include `supertype` ("basic", "legendary", "snow"),
// case-insensitively. Effective, not printed, so a supertype an effect
// adds or removes is visible here — the reader landwalk's "nonbasic"
// needs (CR 205.4c: a land without the basic supertype is nonbasic,
// whatever land types it has).
//
// Added for nonbasic landwalk (#705); IsLegendary is the same read for
// the one supertype the legend rule asks about.
func (c Card) HasSupertype(supertype string) bool {
	if c.effective == nil {
		// CR 708.2: no supertypes either — a face-down legendary
		// permanent is not legendary, which is why two face-down
		// copies of the same legend can coexist.
		if c.FaceDownIsPermanent() {
			return false
		}
		super, _, _ := ParseTypeLine(c.TypeLine)
		return typeListHas(super, supertype)
	}
	return typeListHas(c.effective.Supertypes, supertype)
}

// --- printed card-type predicates -----------------------------
//
// The deliberate other half of the split. These read the printed
// type line and are immune to every continuous effect, which is
// what CR 707.2's copiable values and any "what does this card
// actually say" question need. A separate, explicitly named
// surface rather than a bool argument, so a call site's choice
// between the two is visible in the diff that makes it.

// PrintedIsCreature reports whether the PRINTED type line says
// creature, ignoring every continuous effect.
func (c Card) PrintedIsCreature() bool { return typeLineHas(c.TypeLine, "creature") }

// PrintedIsLand reports whether the PRINTED type line says land,
// ignoring every continuous effect.
func (c Card) PrintedIsLand() bool { return typeLineHas(c.TypeLine, "land") }

// typeListHas reports whether `types` holds `needle` as a whole
// token, case-insensitively.
func typeListHas(types []string, needle string) bool {
	for _, t := range types {
		if equalFoldASCII(t, needle) {
			return true
		}
	}
	return false
}

// equalFoldASCII is strings.EqualFold restricted to ASCII. Type and
// subtype names are ASCII in every Scryfall type line, and this
// runs once per candidate per predicate per snapshot, so the
// allocation-free byte loop earns its place.
func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if x >= 'A' && x <= 'Z' {
			x += 'a' - 'A'
		}
		if y >= 'A' && y <= 'Z' {
			y += 'a' - 'A'
		}
		if x != y {
			return false
		}
	}
	return true
}

// typeLineHas does a case-insensitive substring check against the
// given lowercase needle. The needle MUST be lowercase (callers in
// this file always pass a literal). Returns false for an empty
// TypeLine (placeholder cards from the demo seed are conservatively
// treated as no-type).
func typeLineHas(typeLine, lowerNeedle string) bool {
	if typeLine == "" {
		return false
	}
	n := len(lowerNeedle)
	for i := 0; i+n <= len(typeLine); i++ {
		match := true
		for j := 0; j < n; j++ {
			c := typeLine[i+j]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c != lowerNeedle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// NewCard constructs a fresh Card instance with a new InstanceID, owned
// and controlled by the given player.
func NewCard(name string, owner uuid.UUID) Card {
	return Card{
		InstanceID: uuid.New(),
		Name:       name,
		Owner:      owner,
		Controller: owner,
	}
}

// NewCommander is like NewCard but flags the card as a commander.
func NewCommander(name string, owner uuid.UUID) Card {
	c := NewCard(name, owner)
	c.IsCommander = true
	return c
}

// EffectiveColors returns the card's colors after continuous effects: the
// layer-5 result on the battlefield, otherwise the stamped Colors list when
// present, then the colored symbols found in ManaCost (hybrid "{W/U}"
// contributes both). Added in S20 sub-PR 1.
func (c Card) EffectiveColors() []string {
	if c.effective != nil {
		return c.effective.Colors
	}
	// CR 708.2: a face-down permanent is COLOURLESS, whatever the
	// card underneath costs or Scryfall stamped. Guarded on the
	// printed branch only; the layered branch above already reads the
	// projection through printedCharacteristic.
	if c.FaceDownIsPermanent() {
		return nil
	}
	if len(c.Colors) > 0 {
		return c.Colors
	}
	seen := map[byte]bool{}
	var out []string
	for i := 0; i < len(c.ManaCost); i++ {
		switch ch := c.ManaCost[i]; ch {
		case 'W', 'U', 'B', 'R', 'G':
			if !seen[ch] {
				seen[ch] = true
				out = append(out, string(ch))
			}
		}
	}
	return out
}

// HasColor reports whether the card is the given color ("B" for
// black, etc.). Added in S20 sub-PR 1.
func (c Card) HasColor(color string) bool {
	for _, col := range c.EffectiveColors() {
		if col == color {
			return true
		}
	}
	return false
}

// IsColorless reports whether the card has no colors. Added in S20
// sub-PR 1.
func (c Card) IsColorless() bool {
	return len(c.EffectiveColors()) == 0
}
