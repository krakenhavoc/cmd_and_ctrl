package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Circling Vultures — Creature — Bird {B}, 3/2 (Weatherlight, S82 older sets batch 780, #5629):
//
//	"Flying
//	 You may discard this card any time you could cast an instant.
//	 At the beginning of your upkeep, sacrifice this creature unless you exile the top creature card of your graveyard."
//
// Flying rides PrintedKeywords. The discard is the CR 116.2e special action (#6224): free, with priority, any time its
// owner could cast an instant. The upkeep tax is paid in cards: the topmost creature card of the graveyard (#6194),
// declining or having none sacrifices the Vultures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7fd6fe21-8fb8-499b-9dd4-806d0d660265",
		Name:            "Circling Vultures",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{effects.DiscardAnyTimeYouCouldCastAnInstant()},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB780UpkeepSacrificeUnlessExileTopCreatureCard("Circling Vultures"),
		},
	})
}
