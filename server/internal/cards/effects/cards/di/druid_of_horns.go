package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Druid of Horns — Creature — Human Druid {3}{G}, 2/3 (Core Set 2019):
//
//	"Whenever you cast an Aura spell that targets this creature, create a
//	 3/3 green Beast creature token."
//
// Heroic's shape, narrowed to Aura spells: it triggers on the cast, for a spell
// its controller cast whose card is an Aura and that names this creature as a
// target. A copy is not cast and an ability is not a spell, so neither
// triggers it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06c12fcc-1c08-40a0-a0da-e139c42f083f",
		Name:         "Druid of Horns",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				item := g.StackItemForEffect(ev.CardID)
				if item == nil || item.Kind != game.StackItemSpell {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !spell.IsAura() {
					return false
				}
				for _, t := range item.Targets {
					if t.Kind == game.TargetCard && t.ID == source.InstanceID {
						return true
					}
				}
				return false
			}, "Druid of Horns — create a 3/3 green Beast creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1})),
		},
	})
}
