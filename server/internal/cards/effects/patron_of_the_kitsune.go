package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Patron of the Kitsune — Legendary Creature — Spirit {4}{W}{W}, 5/6
// (Betrayers of Kamigawa, #5602):
//
//	"Fox offering (You may cast this spell any time you could cast an
//	 instant by sacrificing a Fox and paying the difference in mana costs
//	 between this and the sacrificed Fox. Mana cost includes color.)
//	 Whenever a creature attacks, you may gain 1 life."
//
// Fox offering is Offering (offering.go, CR 702.48). The trigger is
// Impulsive Maneuvers' "whenever a creature attacks": every attacking
// creature's declaration, any player's, is one trigger, with the "you
// may" asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "e75356a8-6268-4355-a0bd-114a5243d527",
		Name:          "Patron of the Kitsune",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offering("Fox", Subtype("Fox"))},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil
			}, "Patron of the Kitsune — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Patron of the Kitsune — gain 1 life?"),
		},
	})
}
