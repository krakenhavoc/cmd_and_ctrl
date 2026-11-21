package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stockman, Mad Fly-entist — {4}{U} Legendary Creature — Insect Mutant
// Scientist 3/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 When Stockman enters, draw a card, then discard a card.
//	 Islandcycling {2} ({2}, Discard this card: Search your library for an
//	 Island card, reveal it, put it into your hand, then shuffle.)"
//
// Overwhelmed Archivist's enters loot and Lórien Revealed's Islandcycling
// (typecycling, CR 702.29e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "baa4cb1b-6fb5-46df-9cdd-54421ee0fc88",
		Name:            "Stockman, Mad Fly-entist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stockman, Mad Fly-entist — draw a card, then discard a card", effects.B35LootOne),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Islandcycling", "{2}", "an Island card", effects.IsLandWithSubtype("Island")),
		},
	})
}
