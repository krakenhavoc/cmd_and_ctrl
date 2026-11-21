package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stump Stomp // Burnwillow Clearing — modal double-faced card. This
// file is the FRONT face, Sorcery {1}{R/G}:
//
//	"Target creature you control deals damage equal to its power to
//	 target creature or planeswalker you don't control."
//
// The back face, Burnwillow Clearing, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1".
//
// Bite Down at sorcery speed: two target clauses read positionally,
// slot 0 the damage source (its power read as the spell resolves, its
// deathtouch and lifelink carried by the damage) and slot 1 the
// recipient, each with its own predicate (CR 601.2c, re-checked per slot
// at resolution, CR 608.2b). The creature does the damage, not the
// spell, so nothing is dealt back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb7b1284-0b2c-4b6a-a389-b2b932838083",
		Name:         "Stump Stomp",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetPermanent("target creature or planeswalker you don't control",
				effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()),
		),
		OnResolve: effects.OneSidedBite,
	})
}
