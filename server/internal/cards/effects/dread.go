package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dread — Creature — Elemental Incarnation {3}{B}{B}{B}, 6/6:
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 Whenever a creature deals damage to you, destroy it.
//	 When Dread is put into a graveyard from anywhere, shuffle it into
//	 its owner's library."
//
// Fear is the canonical keyword (CR 702.36). The second line is No
// Mercy's trigger: any creature's damage to Dread's controller, combat
// or not, one trigger per damaging creature, destroyed at resolution if
// it is still on the battlefield (not targeted; indestructible saves
// it). The last line is the graveyard trigger (#2038,
// WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "09f41fac-8285-457a-bb71-8c6a93af2df9",
		Name:            "Dread",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13CreatureDealtDamageToYou(ev, source, g)
			}, "Dread — destroy the creature that damaged you", destroyTheCreatureThatDealtTheDamage),
			WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn("Dread — shuffle it into its owner's library"),
		},
	})
}
