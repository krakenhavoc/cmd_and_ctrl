package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rankle, Master of Pranks — Legendary Creature — Faerie Rogue {2}{B}{B},
// 3/3 (Outlaws of Thunder Junction Commander, #5368):
//
//	"Flying, haste
//	 Whenever Rankle deals combat damage to a player, choose any number —
//	 • Each player discards a card.
//	 • Each player loses 1 life and draws a card.
//	 • Each player sacrifices a creature of their choice."
//
// Rankle's Prank's bullets at one card, one life and one creature, on a
// combat-damage trigger whose modes are chosen as it goes on the stack
// (CR 603.3c). "Any number" is zero to three, each at most once (CR
// 700.2d); none at all removes the trigger (CR 700.2b). The chosen
// bullets resolve in printed order, each waiting for the prompts the one
// before asked (ModeSteps, CR 608.2c). The discard is a real choice over
// each hand in APNAP order; the life loss and draw are each living
// player's in turn, the controller included; the sacrifice is one
// creature per player, a player with none sacrificing nothing.
//
// No simplification.
func init() {
	t := On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
		"Rankle, Master of Pranks — choose any number", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseN("Choose any number", 0, 3,
		ModeSteps("Each player discards a card.", nil,
			Instruction(func(ctx *Context) error {
				ctx.Game.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
					Source: ctx.Source(), N: 1, Question: "Rankle, Master of Pranks — discard a card",
				})
				return nil
			})),
		ModeSteps("Each player loses 1 life and draws a card.", nil,
			Instruction(func(ctx *Context) error {
				for _, id := range tablePlayers(ctx) {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -1); err != nil {
						return err
					}
					if err := (DrawCards{Player: id, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})),
		ModeSteps("Each player sacrifices a creature of their choice.", nil,
			Instruction(func(ctx *Context) error {
				return jmpB1EachPlayerSacrificesN(1, "Rankle, Master of Pranks — choose a creature to sacrifice")(ctx.Game, ctx.Item)
			})),
	)
	Register(Spec{
		OracleID:        "b8619990-9dc2-4fcc-bc7e-457b77cd2a8e",
		Name:            "Rankle, Master of Pranks",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
