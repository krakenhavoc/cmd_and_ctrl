package game

import "github.com/google/uuid"

// Card is a single instance of a Magic card inside a running game. One
// physical card = one Card value; if a player plays two copies of the
// same printed card from their library, there are two Card instances,
// each with its own InstanceID.
//
// At S02 this struct is deliberately minimal: name + owner + a few
// runtime state flags. Scryfall data (mana cost, oracle text, image
// URLs, type line, etc.) arrives in S04 when the Scryfall pipeline
// lands, at which point Card will grow a CardData reference.
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

	// DamageMarked is the damage currently noted on the creature this
	// turn, used by the lethal-damage state-based action (CR 704.5g).
	// Combat damage and direct-damage spells (resolved manually) write
	// into this field; the cleanup-step turn-based action zeroes it.
	// Only meaningful for creatures on the battlefield. Added in S13.1.
	DamageMarked int

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

	// SummonedThisTurn and MarkedLethalByDeathtouch live in the bool
	// block at the end of Card, for alignment.

	// ExilePlay is the impulse-exile permission (S21 sub-PR 6):
	// "exile the top card of your library — you may play it this
	// turn". Meaningful only while the card is in exile, and only
	// for the player it names, who is usually not the owner. Zero
	// value means the card is inert exile like any other. Cleared
	// as the card leaves exile and swept at cleanup.
	ExilePlay ExilePlayPermission

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
	// up" is exactly CR 711.2.
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

	// AttachedAt is the CR 613.7d timestamp: an Equipment's or
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

	// PrintedSelf is this card's OWN printed values, stashed when a
	// CR 706 copy effect overwrote the flat printed fields above.
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
	// StartingDefense is the printed defense a battle enters the
	// battlefield with (CR 310.4), parsed from Scryfall's `defense`
	// string at deck-import time. Zero for every other card type.
	//
	// The exact sibling of StartingLoyalty, for the exact same
	// reason and after the exact same bug: defense is printed data
	// like Power / Toughness / ManaCost, not card-effect data, and
	// while the only source was the catalog every battle outside the
	// opt-in catalog entered with zero defense counters and was
	// swept into the graveyard by the CR 704.5p SBA before anyone
	// could attack it. That was live on `main` for every battle a
	// player could import. The catalog's BattleSpec.Defense survives
	// as a fallback for cards with no printed data — tokens,
	// fixtures — see CatalogBattleDefense.
	//
	// Added in S27.
	StartingDefense int

	// ProtectorPlayerID is the opponent chosen to protect a battle
	// as it enters (CR 310.5). uuid.Nil for every other card type,
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
	// TestCardAlignmentPaddingStaysSmall fails if a new bool is added
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

	// FaceDown is the visual face-down flag (CR 708) — morph,
	// manifest, mutate-bottom, set face-down by an effect. Distinct
	// from the KnownBy knowledge set: a face-down creature is
	// face-down to everyone visually, but the morph caster (and
	// anyone who saw it via Frantic Search-style reveal) still has
	// the card in their KnownBy set so the hover-reveal works on
	// their client. Added in S13.5.
	FaceDown bool

	// SummonedThisTurn is the summoning-sickness flag (CR 302.1).
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
	p := c.Effective().Power
	if c.Counters != nil {
		p += c.Counters["+1/+1"]
		p -= c.Counters["-1/-1"]
	}
	if p < 0 {
		return 0
	}
	return p
}

// CurrentToughness returns the card's combat-relevant toughness:
// the post-layer effective toughness (S16: anthems, CDAs) plus any
// +1/+1 counters, minus any -1/-1 counters. Used by the lethal-
// damage and 0-toughness SBAs (S13.1). May be zero or negative —
// callers compare against DamageMarked directly. NOT clamped (cf.
// CurrentPower) because the SBAs need to distinguish "printed 0/0
// placeholder" (Toughness == 0, no counters) from "reduced to 0/0
// by -1/-1 counters" (Toughness > 0 + counters).
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
func (c Card) HasCardType(lowerType string) bool {
	if c.effective == nil {
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

// EffectiveColors returns the card's colors: the stamped Colors list
// when present, otherwise the colored symbols found in ManaCost
// (hybrid "{W/U}" contributes both). Layer 5 color-changing effects
// aren't modelled yet; when they are, this is the seam. Added in
// S20 sub-PR 1.
func (c Card) EffectiveColors() []string {
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
