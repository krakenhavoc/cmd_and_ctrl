package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Willow Geist — Creature — Treefolk Spirit {G}, 1/1 (Innistrad: Midnight
// Hunt, S82):
//
//	"Trample
//	 Whenever one or more cards leave your graveyard, put a +1/+1 counter
//	 on this creature.
//	 When this creature dies, you gain life equal to its power."
//
// The counter trigger is one per batch of cards leaving together (CR
// 603.2c, OncePerBatch over whenCardsLeaveYourGraveyard, Spirit of
// Resilience's), a cast from the graveyard included. The dies trigger
// reads the Geist's power as it last existed on the battlefield (CR
// 603.10a, 608.2h), counters and continuous effects included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a7096a7-ef1b-4015-b1eb-5c2174d7ec9a",
		Name:            "Willow Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, effects.WhenCardsLeaveYourGraveyard,
				"Willow Geist — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
			effects.WhenThisDies("Willow Geist — gain life equal to its power",
				effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectPower})),
		},
	})
}
