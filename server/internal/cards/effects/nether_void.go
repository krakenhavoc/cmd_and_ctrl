package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nether Void — World Enchantment {3}{B}:
//
//	"Whenever a player casts a spell, counter it unless that player
//	 pays {3}."
//
// Every player's spells, its controller's included, through
// WheneverAPlayerCastsCounterUnlessPays (cast_counter_unless.go): the
// caster is asked as the trigger resolves, and a spell already gone
// from the stack is not charged. A world permanent, so the world rule
// (CR 704.5k) puts it into its owner's graveyard when a newer one
// enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53bab610-1d27-4897-a4ad-cfecca82b811",
		Name:         "Nether Void",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAPlayerCastsCounterUnlessPays("Nether Void — counter it unless that player pays {3}", nil,
				func(*game.Game, game.Card) string { return "{3}" }),
		},
	})
}
