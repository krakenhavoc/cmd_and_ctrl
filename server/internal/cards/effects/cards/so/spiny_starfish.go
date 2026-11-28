package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiny Starfish — Creature — Starfish {2}{U} 0/1 (Alliances, S82 OLD batch
// 819):
//
//	"{U}: Regenerate this creature.
//	 At the beginning of each end step, if this creature regenerated this
//	 turn, create a 0/1 blue Starfish creature token for each time it
//	 regenerated this turn."
//
// "Each time it regenerated" is the turn's regeneration events for this
// object (CR 701.19); an intervening "if" (CR 603.4), checked again as the
// ability resolves, and counted then.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "663d6c61-e269-4f2f-9d8f-49c4c23f05d8",
		Name:         "Spiny Starfish",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{U}")},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "this creature regenerated this turn",
				Test: func(ctx *effects.Context) bool {
					return effects.OldB819TimesRegeneratedThisTurn(ctx.Game, ctx.Source()) > 0
				},
			}, effects.On(game.EventBeginEndStep, effects.AllOf(effects.AnyPlayer,
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.OldB819TimesRegeneratedThisTurn(g, source.InstanceID) > 0
				}),
				"Spiny Starfish — create a 0/1 blue Starfish creature token for each time it regenerated this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.OldB819TimesRegeneratedThisTurn(g, item.SourceCardID)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: game.Card{
						Name: "Starfish", TypeLine: "Token Creature — Starfish", Colors: []string{"U"}, Power: 0, Toughness: 1,
					}, N: n}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
