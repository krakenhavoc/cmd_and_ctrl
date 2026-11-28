package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brand of Ill Omen — Enchantment — Aura {3}{R} (Ice Age, S82 OLD batch 839,
// #5636):
//
//	"Enchant creature
//	 Cumulative upkeep {R}
//	 Enchanted creature's controller can't cast creature spells."
//
// A static can't-cast on the permanent, so its presence is the duration (CR 101.2):
// the player is the controller of the creature the Aura is attached to as the spell
// is cast, so moving the creature to another controller moves the ban, and a spell
// of any other type is unaffected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "947bd180-ee95-4db4-b5d9-432428018eb8",
		Name:         "Brand of Ill Omen",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Brand of Ill Omen — cumulative upkeep {R}", "{R}")},
		CastRestrictions: []game.CastRestriction{{
			Label: "Enchanted creature's controller can't cast creature spells.",
			Forbids: func(q game.CastQuery) bool {
				host := q.Game.AttachedHostOf(&q.Source)
				if host == nil || host.Controller != q.Controller {
					return false
				}
				return effects.MatchCastCard(q.Game, effects.Creature(), q.Controller, q.Card)
			},
		}},
	})
}
