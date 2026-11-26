package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blaze Commando — Creature — Minotaur Soldier {3}{R}{W}, 5/3 (Dragon's Maze,
// S82 OLD b406, #5564):
//
//	"Whenever an instant or sorcery spell you control deals damage, create two
//	 1/1 red and white Soldier creature tokens with haste."
//
// Satyr Firedancer's spell-damage condition (S82OldB406InstantOrSorceryYouControl
// DealtDamage) without the opponent clause. One trigger for each spell's damage
// instance (BatchTotalDamageDealt, CR 603.2c): damage that spell deals to
// several recipients at the same time is one occurrence, as the ruling states,
// while a second damage instruction is another.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92e2f982-f8b9-4c3b-ab50-14c025370e5c",
		Name:         "Blaze Commando",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			AppliesTo:  effects.S82OldB406InstantOrSorceryYouControlDealtDamage,
			Key:        "Blaze Commando — create two 1/1 red and white Soldier creature tokens with haste",
			AtBatchEnd: effects.ThatMuchDamageDealtAtLeast(1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 red and white Soldier with haste"), N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
