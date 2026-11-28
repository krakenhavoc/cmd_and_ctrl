package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrow Ghoul — Creature — Zombie {1}{B}, 4/4 (Weatherlight, S82 older sets batch 779, #5629):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you exile the top creature
//	 card of your graveyard."
//
// The Elder Dragons' upkeep tax paid in cards instead of mana (UpkeepPayUnless with a pay action,
// #6194): the card is not the payer's to pick, it is the topmost creature card of their graveyard
// (CR 404.2), and a graveyard with none can't pay (CR 118.3), so declining or being unable to pay
// sacrifices the Ghoul. The table is held in the upkeep until it is answered.
//
// No simplification.
func init() {
	pay := effects.ExileTopGraveyardCardPayment("creature")
	effects.Register(effects.Spec{
		OracleID:     "1610875f-48da-490c-a371-c439dc36f533",
		Name:         "Barrow Ghoul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Barrow Ghoul — sacrifice this creature unless you exile the top creature card of your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     pay.Words,
						Action:   pay.Action,
						Question: "Barrow Ghoul — exile the top creature card of your graveyard or sacrifice Barrow Ghoul?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
