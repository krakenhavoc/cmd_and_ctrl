package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emrakul's Influence — Enchantment {2}{G}{G} (Eldritch Moon, S82 OLD,
// #5541):
//
//	"Whenever you cast an Eldrazi creature spell with mana value 7 or
//	 greater, draw two cards."
//
// Path of Annihilation's cast trigger: the spell's mana value read on the
// stack, {X} counted (CR 202.3e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92eb07ca-ec57-48c3-98ce-9e878bd968c1",
		Name:         "Emrakul's Influence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Creature(), effects.HasSubtype("Eldrazi"), effects.ManaValueGE(7)),
				"Emrakul's Influence — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
