package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog-Strider Ash — Creature — Treefolk Shaman {3}{G}, 2/4 (Lorwyn):
//
//	"Swampwalk
//	 Whenever a player casts a Goblin spell, you may pay {G}. If you do, you
//	 gain 2 life."
//
// Leshrac's Sigil's offer: any player's cast counts (TmcB8APlayerCast), the
// Goblin word read through the Ash's text (CR 612.2); resolving the trigger
// asks its controller to pay {G}, and the life follows only a paid cost.
// Swampwalk is the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e3d4dd61-d646-449c-84f4-0ff927ca6b0a",
		Name:            "Bog-Strider Ash",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfSubtype("Goblin")),
				"Bog-Strider Ash — you may pay {G}; if you do, you gain 2 life",
				effects.S82OldB549BogStriderAshOffer),
		},
	})
}
