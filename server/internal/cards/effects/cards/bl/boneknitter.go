package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boneknitter — Creature — Zombie Cleric {1}{B}, 1/1 (Onslaught):
//
//	"{1}{B}: Regenerate target Zombie.
//	 Morph {2}{B}"
//
// Swarmyard's regenerate-a-type ability on a mana cost. The target is
// any Zombie permanent, any player's, the Boneknitter included (it is a
// Zombie). Face down it is a 2/2 with no text and no subtypes (CR 708.2),
// so it has no ability and is not a Zombie until it is turned face up
// for {2}{B}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "34b2f465-0b15-428a-a1c2-bcd41af07d01",
		Name:             "Boneknitter",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{B}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Regenerate target Zombie.",
			Cost:    effects.ManaCost("{1}{B}"),
			Targets: effects.TargetPermanent("target Zombie", effects.HasSubtype("Zombie")),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
