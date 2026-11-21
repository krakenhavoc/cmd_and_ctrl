package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riveteers Confluence — Sorcery {2}{B}{R}{G} (Streets of New Capenna
// Commander):
//
//	"Choose three. You may choose the same mode more than once.
//	 • You draw a card and you lose 1 life.
//	 • Riveteers Confluence deals 1 damage to each creature and
//	   planeswalker you don't control.
//	 • You may put a land card from your hand or graveyard onto the
//	   battlefield tapped."
//
// Fiery Confluence's shape (CR 700.2d): each occurrence runs in the order
// chosen, so the damage bullet chosen twice is two separate passes. The
// land bullet is Dread Tiller's pick over the lands in the hand and the
// graveyard together, none being the answer; the land is put, not played.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f0bd777-f79d-47cd-83e3-85b80d5e7942",
		Name:         "Riveteers Confluence",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			effects.ModeWithPurpose(effects.ModeDoing("You draw a card and you lose 1 life.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
				}), game.Purpose{Draws: 1, Answers: game.AnswerValue}),
			effects.ModeWithPurpose(effects.ModeDoing("Riveteers Confluence deals 1 damage to each creature and planeswalker you don't control.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()), 1)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}),
			effects.ModeDoing("You may put a land card from your hand or graveyard onto the battlefield tapped.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Cmdr2B25PutLandFromHandOrGraveyardTapped("Riveteers Confluence")(ctx.Game, item)
				}),
		),
	})
}
