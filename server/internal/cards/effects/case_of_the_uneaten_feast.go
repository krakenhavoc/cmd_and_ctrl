package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Uneaten Feast — Enchantment — Case {W} (Murders at Karlov
// Manor, #4469):
//
//	"Whenever a creature you control enters, you gain 1 life.
//	 To solve — You've gained 5 or more life this turn. (If unsolved,
//	 solve at the beginning of your end step.)
//	 Solved — Sacrifice this Case: Creature cards in your graveyard gain
//	 "You may cast this card from your graveyard" until end of turn."
//
// The life trigger is Sickbay's. The solve condition reads the life its
// controller gained this turn, at the end step and again on resolution
// (CR 603.4). The solved ability is Past in Flames' set-at-resolution
// grant (GrantCastFromYourGraveyard): a one-shot continuous effect locks
// the cards it affects as it resolves (CR 611.2c), so only the creature
// cards in the graveyard then can be cast, at their mana cost and their
// normal timing, until end of turn; a creature card that reaches the
// graveyard later gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d8730b5-f7ba-48b0-928d-d8b7685347d3",
		Name:         "Case of the Uneaten Feast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl, "Case of the Uneaten Feast — a creature you control entered: you gain 1 life",
				Do(GainLife{Amount: 1})),
			ToSolve("Case of the Uneaten Feast — to solve: you've gained 5 or more life this turn",
				func(g *game.Game, controller, _ uuid.UUID) bool {
					return b15LifeGainedThisTurn(g, controller) >= 5
				}),
		},
		Activated: []ActivatedAbility{{
			Label:      "Sacrifice this Case: Creature cards in your graveyard gain \"You may cast this card from your graveyard\" until end of turn.",
			Purpose:    game.Purpose{Answers: game.AnswerValue},
			Cost:       SacrificeThis(),
			ActiveWhen: Solved(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantCastFromYourGraveyard{
					Filter: game.PermissionFilter{CreatureOnly: true},
					Label:  "Cast from your graveyard (Case of the Uneaten Feast)",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
