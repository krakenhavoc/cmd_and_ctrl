package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worldfire — Sorcery {6}{R}{R}{R} (Magic 2013, S82 OLD, #5568):
//
//	"Exile all permanents. Exile all cards from all hands and graveyards.
//	 Each player's life total becomes 1."
//
// Three instructions in printed order (CR 608.2c): Merciless Eviction's exile
// sweep over every permanent, tokens and lands included; one batch exiling every
// card in every hand and graveyard (a commander is exiled first and its owner
// asked afterwards, CR 903.9a); then each player's life total becomes 1 as a
// gain or loss of the difference (CR 119.5), so "can't gain life" leaves a
// player below 1 where they were.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae0b8c13-0a71-4a60-bf9f-6e2da9503e9c",
		Name:         "Worldfire",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepExile}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ExileAllMatching{Match: effects.Permanent()}.Apply(ctx)
			}),
			effects.Instruction(effects.S82OldB430ExileAllFromHandsAndGraveyards),
			effects.Instruction(effects.S82OldB430EachPlayerLifeBecomesOne),
		),
	})
}
