package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "c1d8e97b-f34b-4e52-9795-a4433f26725a",
		Name:           "Doomskar Oracle",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Foretell("{W}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Doomskar Oracle — you gain 2 life",
				Do(GainLife{Amount: 2})),
		},
	})
}
