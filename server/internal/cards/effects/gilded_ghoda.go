package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilded Ghoda — Creature — Horse Mount {1}{R}:
//
//	"Whenever this creature attacks while saddled, create a Treasure
//	 token.
//	 Saddle 1"
//
// Saddle is Saddle(1) (CR 702.171, ADR 0071 amendment 2026-10-08); the
// attack trigger reads the saddled designation when the attack is
// declared. The Treasure is the catalog token, so it carries its mana
// ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f4939dc-6e87-47fa-8287-9da60d4b5db2",
		Name:         "Gilded Ghoda",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(1)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Gilded Ghoda — create a Treasure", func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
