package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:              "a3d8c0f8-ebe8-4e28-bef1-e2a45c0d86d5",
		Name:                  "Galvanic Juggernaut",
		Completeness:          CompletenessFull,
		Static:                []game.StaticAbility{AttacksEachCombat()},
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID
			}, "Galvanic Juggernaut — untap this creature", func(g *game.Game, item *game.StackItem) error {
				return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	})
}
