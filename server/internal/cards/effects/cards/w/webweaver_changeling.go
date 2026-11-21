package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Webweaver Changeling — Creature — Shapeshifter {3}{G}{G}, 3/5 (Modern
// Horizons 2 / Morningtide):
//
//	"Changeling (This card is every creature type.)
//	 Reach
//	 When this creature enters, if there are three or more creature cards
//	 in your graveyard, you gain 5 life."
//
// Changeling and reach are printed keywords. The trigger has an
// intervening "if" (CR 603.4): three or more creature cards in your
// graveyard as the creature enters and again as the ability resolves; a
// noncreature card does not count and neither does the Changeling itself
// (it is on the battlefield).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8cf9387-54df-45b1-ad8d-9de57ed8916b",
		Name:            "Webweaver Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling", "reach"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "there are three or more creature cards in your graveyard",
				Test: func(ctx *effects.Context) bool {
					return effects.CreatureCardsInGraveyard(ctx.Game, ctx.Controller()) >= 3
				},
			}, effects.WhenThisEnters("Webweaver Changeling — you gain 5 life", effects.Do(effects.GainLife{Amount: 5}))),
		},
	})
}
