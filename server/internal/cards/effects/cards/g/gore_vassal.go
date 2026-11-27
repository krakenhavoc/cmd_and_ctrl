package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gore Vassal — Creature — Phyrexian Dog {2}{W}, 2/1 (Mirrodin Besieged,
// S82 OLD batch 464, #5576):
//
//	"Sacrifice this creature: Put a -1/-1 counter on target creature. Then
//	 if that creature's toughness is 1 or greater, regenerate it."
//
// The counter goes on first; the creature's toughness is read with it (CR
// 122.1a) and, at 1 or more, it gets a regeneration shield (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4f44a37-1e66-4f60-9af2-1e67ba58445b",
		Name:         "Gore Vassal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Put a -1/-1 counter on target creature. Then if that creature's toughness is 1 or greater, regenerate it.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok || !effects.OnBattlefield(g, id) || c.CurrentToughness() < 1 {
					return nil
				}
				return effects.Regenerate{Target: id}.Apply(ctx)
			},
		}},
	})
}
