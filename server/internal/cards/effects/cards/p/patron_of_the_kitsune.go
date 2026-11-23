package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:      "e75356a8-6268-4355-a0bd-114a5243d527",
		Name:          "Patron of the Kitsune",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Offering("Fox", effects.Subtype("Fox"))},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil
			}, "Patron of the Kitsune — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Patron of the Kitsune — gain 1 life?"),
		},
	})
}
