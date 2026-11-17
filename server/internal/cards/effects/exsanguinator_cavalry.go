package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exsanguinator Cavalry — Creature — Vampire Knight {2}{B}, 2/3 (Commander
// Legends of March of the Machine):
//
//	"Menace, lifelink
//	 Whenever a Knight you control deals combat damage to a player, put a
//	 +1/+1 counter on that creature and create a Blood token."
func init() {
	Register(Spec{
		OracleID:        "a867be6f-46ce-4a12-a827-9429bd94243f",
		Name:            "Exsanguinator Cavalry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, s82Mom09KnightDealtCombatDamage,
				"Exsanguinator Cavalry — +1/+1 counter on that Knight and create a Blood token",
				func(g *game.Game, item *game.StackItem) error {
					if err := s82Mom09CounterOnThatCreature(g, item); err != nil {
						return err
					}
					return CreateToken{Controller: item.Controller, Template: BloodToken(), N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
