package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ultimate Green Goblin — Legendary Creature — Goblin Villain {1}{B/R}{B/R},
// 5/4 (Marvel's Spider-Man, #4460):
//
//	"At the beginning of your upkeep, discard a card, then create a Treasure
//	 token.
//	 Mayhem {2}{B/R}"
//
// The discard is the controller's own choice and mandatory; the Treasure is
// made after it lands, whether or not a card was discarded (an empty hand
// discards nothing, CR 701.9a, and the "then" still follows). Mayhem is the
// keyword's alternative cost, castable from the graveyard the turn it was
// discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b5b43d01-fce6-4a00-9c19-7a7e2a09d833",
		Name:             "Ultimate Green Goblin",
		Completeness:     effects.CompletenessFull,
		CastableZonesIf:  []game.ConditionalCastZone{effects.MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{effects.Mayhem("{2}{B/R}")},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Ultimate Green Goblin — discard a card, then create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player: item.Controller,
						Source: item.SourceCardID,
						N:      1,
					}, ultimateGreenGoblinTreasure, nil)
				}),
		},
	})
}

// ultimateGreenGoblinTreasure is "then create a Treasure token", after the
// discard (an on-disk key: never renamed, never reused).
var ultimateGreenGoblinTreasure = game.RegisterDiscardThen("choose-cards/ultimate-green-goblin-treasure", func(g *game.Game, r game.DiscardsDone) error {
	return effects.CreateToken{Controller: r.Player, Template: effects.TreasureToken(), N: 1}.Apply(effects.KeyedContext(g, r.Player, r.Source))
})
