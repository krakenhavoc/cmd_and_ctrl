package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flagstones of Trokair — Legendary Land (EDHREC rank 3334):
//
//	"{T}: Add {W}.
//	 When Flagstones of Trokair is put into a graveyard from the
//	 battlefield, you may search your library for a Plains card, put
//	 it onto the battlefield tapped, then shuffle."
//
// The Plains that replaces itself when it dies — to a land wipe, a
// sacrifice, or the legend rule against a second copy. "Put into a
// graveyard from the battlefield" is cardDied: a bounce or an exile
// does not fire it, as printed. The "may" is a real prompt; a yes
// runs the S22 chooser over every land with the Plains type
// (b31FetchPlainsTapped — a Snow-Covered Plains or a Plains-typed
// dual both qualify, CR 205.3i), with the fetched card entering
// tapped through the search's own tapped flag.
//
// The mana ability is declared because the card has no Plains type
// of its own to derive it from.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f73979bb-91a5-4388-b70b-0cd7a4e14291",
		Name:         "Flagstones of Trokair",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisDies("Flagstones of Trokair — search for a Plains card, onto the battlefield tapped", b31FetchPlainsTapped), "Flagstones of Trokair — search your library for a Plains card?"),
		},
	})
}
