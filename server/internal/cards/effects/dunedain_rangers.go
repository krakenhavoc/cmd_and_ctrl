package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dúnedain Rangers — Creature — Human Ranger {3}{G}, 4/4:
//
//	"Landfall — Whenever a land you control enters, if you don't
//	 control a Ring-bearer, the Ring tempts you."
//
// The "if" is an intervening if (CR 603.4): it is checked as the land
// enters and again as the ability resolves, so with two lands entering
// at once the first tempt gives you a Ring-bearer and the second
// ability does nothing. "A Ring-bearer" is yours (CR 701.54e): on the
// battlefield, under your control, with the designation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7de2a364-4383-437a-b6d6-55e4721a0141",
		Name:         "Dúnedain Rangers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(LandEnteredUnderYourControl, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return game.RingBearerOf(g, source.Controller) == uuid.Nil
			}), "Dúnedain Rangers — the Ring tempts you", func(g *game.Game, item *game.StackItem) error {
				if game.RingBearerOf(g, item.Controller) != uuid.Nil {
					return nil
				}
				return TheRingTemptsYou{}.Apply(NewContext(g, item))
			}),
		},
	})
}
