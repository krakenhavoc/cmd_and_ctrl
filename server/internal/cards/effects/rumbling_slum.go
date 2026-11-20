package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rumbling Slum — Creature — Elemental {1}{R}{G}{G}, 5/5 (Guildpact,
// #5599):
//
//	"At the beginning of your upkeep, this creature deals 1 damage to
//	 each player."
//
// Juzám Djinn's upkeep trigger with Flame Rift's reach
// (b38DamageEachPlayer): 1 damage to every player still in the game, its
// controller included, as one damage event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77275fc5-5f70-4d7f-938e-f32b5c19c532",
		Name:         "Rumbling Slum",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Rumbling Slum — 1 damage to each player", func(g *game.Game, item *game.StackItem) error {
				return b38DamageEachPlayer(NewContext(g, item), 1)
			}),
		},
	})
}
