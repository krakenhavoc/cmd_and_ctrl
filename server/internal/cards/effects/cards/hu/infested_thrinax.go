package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infested Thrinax — Creature — Lizard {3}{B}{G}, 4/4:
//
//	"Flash
//	 When this creature enters, until end of turn, whenever a nontoken
//	 creature you control dies, create a number of 1/1 green Saproling
//	 creature tokens equal to that creature's power."
//
// The enters trigger sets up a repeating delayed trigger (CR 603.7b) for
// the turn, Gnawing Crescendo's condition: a nontoken creature that left
// the battlefield for a graveyard under the controller's control, a nontoken
// as it last existed (CR 603.10a). Each such death makes as many Saprolings
// as that creature's power as it last existed (CR 608.2h), none for a
// creature of zero or less power.
func init() {
	const name = "Infested Thrinax"
	effects.Register(effects.Spec{
		OracleID:        "860b5e4c-45bd-4ba1-930a-36a1450ebd37",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — until end of turn, whenever a nontoken creature you control dies, create Saprolings equal to its power",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DelayedOnEvent{
						Label:     name + " — create a number of 1/1 green Saproling creature tokens equal to that creature's power",
						On:        []game.EventKind{game.EventLTB},
						Condition: effects.Std54NontokenCreatureYouControlDiedCondition,
						Body:      infestedThrinaxSaprolingsBody,
						Repeats:   true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// infestedThrinaxSaprolingsBody is the delayed trigger's body: Saprolings
// equal to the dead creature's power as it last existed. An on-disk key.
var infestedThrinaxSaprolingsBody = game.SimpleDelayedBody("infested-thrinax/saprolings-equal-to-power", func(g *game.Game, item *game.StackItem) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	n := effects.DepartedCreaturePower(g, item.Trigger.Event.CardID)
	if n <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: n}.Apply(effects.NewContext(g, item))
})
