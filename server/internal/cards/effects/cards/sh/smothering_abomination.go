package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{OracleID: "c3db28e5-6162-4642-a1c9-a7ace0737950", Name: "Smothering Abomination", Completeness: effects.CompletenessFull,
		PrintedKeywords: []string{"devoid", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Smothering Abomination — sacrifice a creature", func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a creature", effects.Creature()), "Smothering Abomination — sacrifice a creature")
				return nil
			}),
			effects.On(game.EventSacrifice, effects.CreatureActionByYou, "Smothering Abomination — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
