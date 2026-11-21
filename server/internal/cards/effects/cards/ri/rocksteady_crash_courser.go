package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rocksteady, Crash Courser — {4}{G}{G} Legendary Creature — Rhino Mutant 7/7
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Rocksteady can't be blocked by more than one creature.
//	 Boars you control can't be blocked by more than one creature.
//	 Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Both block limits are Outland Colossus's MaxBlockers, the second over the
// Boars its controller controls (Michelangelo, Mutant BFF's scope), read as
// blockers are declared (CR 509.1b). Forestcycling is typecycling (CR
// 702.29e) for any card with the Forest subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9330a7b-8a65-4e04-a777-bf6512332850",
		Name:         "Rocksteady, Crash Courser",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.MaxBlockers(effects.OnSelf(), 1),
			effects.MaxBlockers(effects.OnMatching(effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Boar"))), 1),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
