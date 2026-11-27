package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scald — Enchantment {1}{R} (Urza's Saga, S82 OLD b723):
//
//	Whenever a player taps an Island for mana, this enchantment deals 1 damage to that player.
//
// Manabarbs' trigger on a mana ability, narrowed to an Island that was tapped (any player's, its controller's included); the damage is dealt by this enchantment to the player who tapped it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1de0b522-f12a-4597-9d5f-9a246443487e",
		Name:         "Scald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: effects.S82OldB723TappedAnIsland,
			Key:       "Scald — deal 1 damage to that player",
			Effect:    effects.S82OldB723DamageThatPlayer(1),
		}},
	})
}
