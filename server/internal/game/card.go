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

	// Owner is the player who brought this card to the game. Ownership
	// is fixed at deck-build time and never changes.
	Owner uuid.UUID

	// Controller is the player who currently controls the card. May
	// differ from Owner for stolen permanents, Control Magic effects,
	// etc. Always equal to Owner for cards not on the battlefield.
	Controller uuid.UUID

	// Tapped is the usual MTG tap state. Only meaningful for cards on
	// the battlefield; ignored in other zones.
	Tapped bool

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

	// IsCommander marks a card as a commander for the Commander format.
	// Commanders live in the command zone at game start.
	IsCommander bool

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
}

// CurrentPower returns the card's effective power: base printed
// power plus any +1/+1 counters, minus any -1/-1 counters. Other
// dynamic effects (auras, equipment, anthem effects) are not
// modeled — the sandbox handles them via manual life adjustments.
// Negative results clamp to zero (a -3/-3 modifier on a 2/2 deals
// no damage, not negative damage).
func (c Card) CurrentPower() int {
	p := c.Power
	if c.Counters != nil {
		p += c.Counters["+1/+1"]
		p -= c.Counters["-1/-1"]
	}
	if p < 0 {
		return 0
	}
	return p
}

// IsCreature reports whether the card's TypeLine identifies it as a
// creature. Case-insensitive substring check against "creature";
// covers "Creature — Human Wizard" and "Legendary Artifact Creature
// — Golem" alike. Empty TypeLine returns false (placeholder cards
// from the demo seed are conservatively treated as non-creatures).
func (c Card) IsCreature() bool {
	return typeLineHas(c.TypeLine, "creature")
}

// IsLand reports whether the card's TypeLine identifies it as a land.
func (c Card) IsLand() bool {
	return typeLineHas(c.TypeLine, "land")
}

// IsInstant reports whether the card is an instant. Instants share
// the priority window with activated abilities — they're castable
// any time the caller holds priority.
func (c Card) IsInstant() bool {
	return typeLineHas(c.TypeLine, "instant")
}

// IsSorcery reports whether the card is a sorcery. Sorceries are
// sorcery-speed only — main phase, stack empty, caller is the
// active player.
func (c Card) IsSorcery() bool {
	return typeLineHas(c.TypeLine, "sorcery")
}

// IsArtifact reports whether the card is an artifact. Artifacts are
// permanents (resolution route: battlefield).
func (c Card) IsArtifact() bool {
	return typeLineHas(c.TypeLine, "artifact")
}

// IsEnchantment reports whether the card is an enchantment.
func (c Card) IsEnchantment() bool {
	return typeLineHas(c.TypeLine, "enchantment")
}

// IsPlaneswalker reports whether the card is a planeswalker.
func (c Card) IsPlaneswalker() bool {
	return typeLineHas(c.TypeLine, "planeswalker")
}

// IsBattle reports whether the card is a battle (post-MoM card type).
func (c Card) IsBattle() bool {
	return typeLineHas(c.TypeLine, "battle")
}

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
