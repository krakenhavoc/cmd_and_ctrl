package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Falkenrath Perforator — Creature — Vampire {1}{R}, 2/1 (Innistrad:
// Midnight Hunt, #5497):
//
//	"Whenever this creature attacks, it deals 1 damage to defending
//	 player."
//
// Generous Plunderer's attack trigger: the defending player is resolved
// once, when the attack is declared (a creature attacking a planeswalker
// makes that planeswalker's controller the defending player).
//
// No simplification.
func init() {
	const key = "Falkenrath Perforator — 1 damage to defending player"
	effects.Register(effects.Spec{
		OracleID:     "6e991808-fea4-46a8-9e87-7b5917f33ce8",
		Name:         "Falkenrath Perforator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			},
			Key: key,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, key)
				item.Params.Player = effects.B17DefendingPlayer(g, ev)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				defender := item.Params.Player
				if defender == uuid.Nil {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: defender, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
