package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hemlock Vial — Artifact {1}{B} (Assassin's Creed, #5948):
//
//	"When this artifact enters, you draw a card and you lose 1 life.
//	 {B}, {T}, Sacrifice this artifact: Each equipped creature and Equipment
//	 you control gains deathtouch until end of turn."
//
// The entry trigger is a draw and a life loss in the printed order. The
// ability's set is locked as it resolves (CR 611.2c): the creatures its
// controller controls that have an Equipment attached, and the Equipment they
// control, whether or not attached. The Vial is sacrificed as a cost, so the
// ability still resolves without it (CR 602.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d977dc5-807b-4fab-b44b-ccad5e3d63e3",
		Name:         "Hemlock Vial",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hemlock Vial — you draw a card and you lose 1 life", effects.B36DrawAndLoseOne),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Sacrifice this artifact: Each equipped creature and Equipment you control gains deathtouch until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GrantKeywordUntilEOT{
					Match: func(g *game.Game, controller uuid.UUID, c game.Card) bool {
						if c.Controller != controller {
							return false
						}
						return c.HasSubtype("Equipment") || effects.IsEquippedCreature(g, c)
					},
					Keywords: []string{"deathtouch"},
					Label:    "Hemlock Vial — deathtouch",
				}.Apply(ctx)
			},
		}},
	})
}
