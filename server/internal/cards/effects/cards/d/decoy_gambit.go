package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Decoy Gambit — {2}{U} Instant (Commander 2020):
//
//	"For each opponent, choose up to one target creature that player controls,
//	 then return that creature to its owner's hand unless its controller has
//	 you draw a card."
//
// Shellshock's one clause per opponent (PerOpponentClauses), each "up to one".
// For every target still legal as it resolves (CR 608.2b) the creature's
// controller is asked, on their own, whether to have the caster draw a card
// (Combustible Gearhulk's MayChoice); if they decline the creature is returned
// to its owner's hand. The creature is carried on the prompt, so each
// opponent's answer affects only their own creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94d9878e-9786-438c-8b33-c1dff730b748",
		Name:         "Decoy Gambit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.PerOpponentClauses(game.MaxPlayers-1, "creature", effects.Creature()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				if err := (effects.MayChoice{
					Player:   c.Controller,
					Question: "Decoy Gambit — have its caster draw a card? If you don't, " + c.Name + " is returned to its owner's hand.",
					Draws:    1, Drawer: ctx.Controller(),
					YesLabel: "They draw a card",
					NoLabel:  "Return " + c.Name + " to its owner's hand",
					Then:     decoyGambitAnswer,
					Carry:    []uuid.UUID{t.ID},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

var decoyGambitAnswer = effects.MayYesNo("confirm/decoy-gambit-draw-or-bounce",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		id := effects.CarryAt(r, 0)
		if id == uuid.Nil {
			return nil
		}
		return effects.BounceToHand{Target: id}.Apply(ctx)
	})
