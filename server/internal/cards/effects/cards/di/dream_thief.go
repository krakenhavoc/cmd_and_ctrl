package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream Thief — Creature — Faerie Rogue {2}{U} 2/1:
//
//	"Flying
//	 When this creature enters, draw a card if you've cast another blue spell
//	 this turn."
//
// Not an intervening-if: the trigger always goes on the stack and the
// condition is read as it resolves. "Another" excludes the Dream Thief's own
// cast (it was a blue spell), through the turn's record of each cast spell's
// colours (game.CastAnotherSpellOfColorThisTurnForEffect); a Dream Thief put
// onto the battlefield without being cast counts every blue spell cast.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "55be0fd0-b371-4b9e-920f-3123bb37a1e0",
		Name:            "Dream Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dream Thief — draw a card if you've cast another blue spell this turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.S82OldB526CastAnotherOfColor(ctx, item.Controller, item.SourceCardID, "U") {
					return nil
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}
