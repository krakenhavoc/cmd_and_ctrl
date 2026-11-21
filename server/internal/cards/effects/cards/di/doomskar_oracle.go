package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doomskar Oracle — Creature — Human Cleric {2}{W}, 3/2 (Kaldheim,
// #5505):
//
//	"Whenever you cast your second spell each turn, you gain 2 life.
//	 Foretell {W}"
//
// Thunder Drake's trigger (YouCastYourSecondSpellEachTurn) with a life
// gain, plus the shared foretell special action (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "c1d8e97b-f34b-4e52-9795-a4433f26725a",
		Name:           "Doomskar Oracle",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Doomskar Oracle — you gain 2 life",
				effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
