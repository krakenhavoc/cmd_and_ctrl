package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Web-Warriors — {4}{G/W} Creature — Spider Hero, 4/3 (Marvel's Spider-Man,
// #4460):
//
//	"When this creature enters, put a +1/+1 counter on each other creature
//	 you control."
//
// The set of creatures is fixed as the ability resolves (CR 608.2c), the
// source excluded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b23e4f8-882d-4cce-9b61-171ae2cc3a18",
		Name:         "Web-Warriors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Web-Warriors — a +1/+1 counter on each other creature you control",
				effects.PutCounterOnEachOtherCreatureYouControl),
		},
	})
}
