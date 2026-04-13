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

	// Counters is a generic per-card counter map (+1/+1, -1/-1, loyalty,
	// charge, fade, etc.). nil means no counters. S02 does not interpret
	// counters; they're just storage until rules enforcement grows.
	Counters map[string]int

	// IsCommander marks a card as a commander for the Commander format.
	// Commanders live in the command zone at game start.
	IsCommander bool
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
