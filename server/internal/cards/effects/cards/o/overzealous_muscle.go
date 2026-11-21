package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overzealous Muscle — Creature — Ogre Mercenary {4}{B}, 5/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever you commit a crime during your turn, this creature gains
//	 indestructible until end of turn. (Targeting opponents, anything
//	 they control, and/or cards in their graveyards is a crime. …)"
//
// The crime is the once-per-announcement event (CR 700.13), read only
// while its controller is the active player. There is no once-each-turn
// limit: each crime triggers it, though a second grant changes nothing.
// The grant is on this object (CR 702.12b), so one that left and came back
// is a new object and keeps none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cb63198-29b5-4607-bd4a-f449807fb764",
		Name:         "Overzealous Muscle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.YouCommittedCrime(ev, source) && effects.IsYourTurn(g, source.Controller)
			}, "Overzealous Muscle — gains indestructible until end of turn (crime)",
				effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Overzealous Muscle — indestructible")),
		},
	})
}
