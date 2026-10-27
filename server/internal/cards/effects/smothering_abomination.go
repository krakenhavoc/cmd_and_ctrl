package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{OracleID: "c3db28e5-6162-4642-a1c9-a7ace0737950", Name: "Smothering Abomination", Completeness: CompletenessFull,
		PrintedKeywords: []string{"devoid", "flying"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Smothering Abomination — sacrifice a creature", func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, sacrificeSpec("a creature", Creature()), "Smothering Abomination — sacrifice a creature")
				return nil
			}),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			}, "Smothering Abomination — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
