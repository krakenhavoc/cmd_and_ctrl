package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fell Horseman // Deathly Ride — Creature — Zombie Knight {3}{B}, 3/3 //
// Sorcery — Adventure {1}{B} (Wilds of Eldraine, #4471):
//
//	Fell Horseman — "When this creature dies, put it on the bottom of its
//	 owner's library."
//	Deathly Ride — "Return target creature card from your graveyard to your
//	 hand. (Then exile this card. You may cast the creature later from
//	 exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). The dies trigger is Murderous Rider's (putDiedThisOnBottomOfLibrary);
// Deathly Ride is Sam's Desperate Rescue's return without the Ring.
//
// No simplification.
const FellHorsemanOracleID = "fab8d954-807b-426f-8057-99a5d2fec618"

func init() {
	effects.Register(effects.Spec{
		OracleID:     FellHorsemanOracleID,
		Name:         "Fell Horseman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Fell Horseman — put it on the bottom of its owner's library",
				effects.PutDiedThisOnBottomOfLibrary),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     FellHorsemanOracleID + "#1",
		Name:         "Deathly Ride",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
