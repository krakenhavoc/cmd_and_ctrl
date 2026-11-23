package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beacon of Immortality — Instant {5}{W}:
//
//	"Double target player's life total. Shuffle Beacon of Immortality
//	 into its owner's library."
//
// Doubling a life total is gaining (or, below zero, losing) life equal
// to it (CR 701.10e), so it runs the ordinary life-change path: a
// "can't gain life" stops it and Sanguine Bond sees the gain. The spell
// then goes into its owner's library rather than the graveyard, the way
// Blue Sun's Zenith does, and that library is shuffled.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfcdf563-c6e5-4a54-bb63-4aba219b94e2",
		Name:         "Beacon of Immortality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			// The shuffle into the library waits for a prompt the life change asks
			// (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if p := ctx.Game.PlayerByIDForEffect(t.ID); p != nil && p.Life != 0 {
						if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, t.ID, p.Life); err != nil {
							return err
						}
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				owner := ctx.Item.Owner
				return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{}, game.KeyedTail{Ref: effects.ThenShuffle, Player: owner})
			}),
		),
	})
}
