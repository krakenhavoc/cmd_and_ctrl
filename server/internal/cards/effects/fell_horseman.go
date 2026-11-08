package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fell Horseman // Deathly Ride — Creature — Zombie Knight {3}{B}, 3/3 //
// Sorcery — Adventure {1}{B} (Wilds of Eldraine, #4471):
//
//	Fell Horseman — "When this creature dies, put it on the bottom of its
//	 owner's library."
//	Deathly Ride — "Return target creature card from your graveyard to
//	 your hand. (Then exile this card. You may cast the creature later
//	 from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The dies trigger is Murderous Rider's
// (putThisOnTheBottomOfItsOwnersLibrary): it moves the card only while it
// is still the graveyard object that died (CR 400.7). Deathly Ride is
// Raise Dead's return (returnTargetFromGraveyardToHand); a target that
// left in response does nothing (CR 608.2b). The Adventure lifecycle is
// the engine's (CR 715.3d).
//
// No simplification.
const fellHorsemanOracleID = "fab8d954-807b-426f-8057-99a5d2fec618"

func init() {
	Register(Spec{
		OracleID:     fellHorsemanOracleID,
		Name:         "Fell Horseman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Fell Horseman — put it on the bottom of its owner's library", putThisOnTheBottomOfItsOwnersLibrary),
		},
	})
	Register(Spec{
		OracleID:     fellHorsemanOracleID + "#1",
		Name:         "Deathly Ride",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
