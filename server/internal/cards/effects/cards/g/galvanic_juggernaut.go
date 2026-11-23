package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Juggernaut — Artifact Creature — Juggernaut {4}, 5/5 (Innistrad
// Remastered, #5948):
//
//	"This creature attacks each combat if able.
//	 This creature doesn't untap during your untap step.
//	 Whenever another creature dies, untap this creature."
//
// The first two are Zurgo Helmsmasher's requirement and Basalt Monolith's untap
// restriction (CR 502.3). The trigger watches every creature dying but the
// Juggernaut itself, whoever's, and untaps the source if it is still on the
// battlefield (UntapTarget does nothing for a card that left).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "a3d8c0f8-ebe8-4e28-bef1-e2a45c0d86d5",
		Name:                  "Galvanic Juggernaut",
		Completeness:          effects.CompletenessFull,
		Static:                []game.StaticAbility{effects.AttacksEachCombat()},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID
			}, "Galvanic Juggernaut — untap this creature", func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
