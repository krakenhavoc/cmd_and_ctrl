package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Struggle // Survive — Split card with aftermath:
//
//	Struggle {2}{R} Instant: "Struggle deals damage to target creature
//	 equal to the number of lands you control."
//	Survive {1}{G} Sorcery — Aftermath: "Each player shuffles their
//	 graveyard into their library."
//
// Two specs, the back half under "<oracle_id>#1" as Dusk // Dawn is; the
// engine restricts an aftermath half to a graveyard and exiles it as it
// resolves from there (CR 702.127a). Struggle counts lands as it
// resolves; a creature that has gone makes it fizzle (CR 608.2b).
//
// No simplification.
func init() {
	const oracle = "0b1f2522-960c-405f-9622-4cadbc7ba545"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Struggle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			lands := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsLand() {
					lands++
				}
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: lands}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Survive",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, p.ID); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
