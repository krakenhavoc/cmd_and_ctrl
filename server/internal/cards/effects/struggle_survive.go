package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     oracle,
		Name:         "Struggle",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			lands := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsLand() {
					lands++
				}
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: lands}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Survive",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := finaleShuffleGraveyardIntoLibrary(ctx, p.ID); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
