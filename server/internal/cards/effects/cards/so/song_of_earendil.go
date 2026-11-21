package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Song of Eärendil — Enchantment — Saga {3}{G}{U} (Tales of Middle-earth
// Commander):
//
//	"I — Scry 2, then draw two cards.
//	 II — Create a Treasure token and a 2/2 blue Bird creature token with
//	 flying.
//	 III — Put a flying counter on each creature you control without
//	 flying."
//
// Behold the Unspeakable's scry-then-draw (the draw waits for the scry's
// answer as a keyed tail); chapter II is two token creations; chapter III
// puts the counter on each creature you control that has no flying as the
// chapter resolves, read through the layered view (CR 613). The Saga is
// sacrificed after III by the engine.
//
// No simplification.
func init() {
	const name = "Song of Eärendil"
	effects.Register(effects.Spec{
		OracleID:     "440366dd-7654-4ede-aa47-7e8a7916e31e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2, Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "scry 2, then draw two cards"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Scry{N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 2)}.Apply(ctx)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a Treasure token and a 2/2 blue Bird creature token with flying"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CreateToken{Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Template: effects.TokenCard("2/2 blue Bird with flying"), N: 1}.Apply(ctx)
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "put a flying counter on each creature you control without flying"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && !game.HasKeyword(&c, "flying") {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (effects.AddCounter{Target: id, Kind: game.CounterFlying, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
