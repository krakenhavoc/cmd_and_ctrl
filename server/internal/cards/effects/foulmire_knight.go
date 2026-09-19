package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foulmire Knight // Profane Insight — the catalog's first ADVENTURE
// card (CR 715), oracle 10c0f814…:
//
//	Foulmire Knight — Creature — Zombie Knight {B}, 1/1
//	  "Deathtouch"
//	Profane Insight — Instant — Adventure {2}{B}
//	  "You draw a card and you lose 1 life. (Then exile this card.
//	   You may cast the creature later from exile.)"
//
// # Why this card, of the 170 adventure oracle IDs
//
// #719 ships the MECHANIC, and the card that proves a mechanic should
// add nothing else to prove. Profane Insight is the cheapest printed
// Adventure half that needs no announce-time choice at all: no target,
// no mode, no X. Every other candidate the issue named — Stomp, Petty
// Theft, Swift End — is a targeted instant, and a targeted Adventure
// half needs one more thing that does not exist yet: the view
// publishes `target_mode` and `legal_targets` for the face that is UP
// (face 0, the creature), so a client asked to cast face 1 has no
// picker to open. The engine and the bot enumerator are already
// face-correct; it is the human client that would be short a prompt.
// See the note in AGENTS.md §7 and `client/src/lib/faces.ts`.
//
// So the two halves here are deliberately boring, and what the card
// demonstrates is the LIFECYCLE: cast Profane Insight from hand, it
// resolves and is exiled (CR 715.3d), and the permission that lands
// on it there lets its owner cast Foulmire Knight — and only Foulmire
// Knight — out of exile for as long as it stays (CR 715.4). See
// server/internal/game/adventure.go.
//
// # Two keys, one card
//
// Scryfall issues one oracle_id per CARD, so the creature and the
// Adventure share one and game.CatalogKey makes it composite: face 0
// keeps the bare oracle ID and face 1 takes "<oracle_id>#1". Both
// faces register, the way a Siege's two halves do, because the
// creature half is a real castable object whose coverage the catalog
// page should state rather than leave blank.
//
// No simplification.

// foulmireKnightOracleID is shared by both faces of the card.
const foulmireKnightOracleID = "10c0f814-e302-49ff-aded-1b8d8424fd60"

func init() {
	// Face 0 — the creature. Deathtouch is a printed keyword the
	// engine enforces on its own, so the entry exists to SAY so: a
	// card with an Adventure prints rules on its other half, which is
	// enough to make NeedsCatalogEffect true for the whole card, and
	// without this key the creature would wear the "unimplemented"
	// badge in hand while being completely playable.
	Register(Spec{
		OracleID:        foulmireKnightOracleID,
		Name:            "Foulmire Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})

	// Face 1 — the Adventure. An instant while it is on the stack
	// (CR 715.3a), so it resolves like any other instant; what is
	// different happens after OnResolve returns, in the engine's
	// CR 715.3d branch, and this file says nothing about it.
	Register(Spec{
		OracleID:     foulmireKnightOracleID + "#1",
		Name:         "Profane Insight",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
				return err
			}
			// "you lose 1 life" — life LOSS, not damage (CR 119.3),
			// so it is a life change rather than a DealDamage and no
			// prevention or redirection sees it.
			return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
		},
	})
}
