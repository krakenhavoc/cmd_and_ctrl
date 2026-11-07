package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uchbenbak, the Great Mistake — Legendary Creature — Skeleton Horror
// {3}{U}{B}, 6/4 (The Lost Caverns of Ixalan, #4470):
//
//	"Vigilance, menace
//	 Descend 8 — {4}{U}{B}: Return this card from your graveyard to the
//	 battlefield with a finality counter on it. Activate only if there
//	 are eight or more permanent cards in your graveyard and only as a
//	 sorcery."
//
// A graveyard activation (Grim Repriser's zone) whose condition is the
// descend count (permanentCardsInGraveyard), which includes Uchbenbak
// itself. The return is Grim Reaper's Scythe's verb aimed at the source:
// the card comes back with the finality counter as it enters, and only
// a return by this ability adds one. A card that left the graveyard in
// response stays where it is. Sorcery timing is CR 602.5d; a creature
// with a finality counter that would die is exiled instead (CR 122.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36bfcc4a-7f66-4649-ab6b-18d360299d46",
		Name:            "Uchbenbak, the Great Mistake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Activated: []ActivatedAbility{{
			Label:        "{4}{U}{B}: Return this card from your graveyard to the battlefield with a finality counter on it. Activate only if there are eight or more permanent cards in your graveyard and only as a sorcery.",
			Cost:         ManaCost("{4}{U}{B}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return permanentCardsInGraveyard(g, controller) >= 8
			},
			Effect: returnThisFromGraveyardWithAFinalityCounter,
		}},
	})
}
