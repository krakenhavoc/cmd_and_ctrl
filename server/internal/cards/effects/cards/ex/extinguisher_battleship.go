package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extinguisher Battleship — Artifact — Spacecraft {8}, 10/10 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, destroy target noncreature permanent.
//	 Then this Spacecraft deals 4 damage to each creature.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 5+.)
//	 5+ | Flying, trample"
//
// The entry trigger is Woodfall Primus's "destroy target noncreature
// permanent" followed by Invasion of Karsus's damage to each creature
// (damageEachMatching, one damage event). The trigger has a target, so
// with no noncreature permanent to target it is removed as it would go on
// the stack and deals no damage (CR 603.3d), and if its target is gone by
// resolution it does nothing at all (CR 608.2b). The damage does not
// depend on the destruction: an indestructible target survives and the 4
// damage is still dealt. The Battleship enters with no charge counters,
// so it is not a creature and is not dealt the damage. It wipes the
// board, so the card declares a Sweep. Station, the 5+ creature threshold and its keywords are the
// shared constructors (Galvanizing Sawship).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cce3dcc3-57bb-4b95-8b70-337c67bb3c4e",
		Name:         "Extinguisher Battleship",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Extinguisher Battleship — destroy target noncreature permanent, then 4 damage to each creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
						return err
					}
					return effects.DamageEachMatching(ctx, effects.Creature(), 4)
				}), effects.TargetPermanent("target noncreature permanent", effects.Noncreature())),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(5, 10, 10),
			effects.ThresholdKeywords(5, "flying", "trample"),
		),
	})
}
