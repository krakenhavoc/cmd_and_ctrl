package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cruel Ultimatum — Sorcery {U}{U}{B}{B}{B}{R}{R} (Breaking News, #5929):
//
//	"Target opponent sacrifices a creature of their choice, discards three
//	 cards, then loses 5 life. You return a creature card from your graveyard
//	 to your hand, draw three cards, then gain 5 life."
//
// Seven instructions in printed order (CR 608.2c), each one waiting for the
// prompt the one before it asked: the opponent's sacrifice and their
// discard are their own choices, and the creature card you return is yours.
// The return is a pick from the graveyard as the spell resolves (it does
// not target), so a graveyard with no creature card returns nothing and the
// rest still happens. A target that has become illegal makes the whole
// spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01294ac9-1a06-4ee8-b3b1-db1f07c9d94e",
		Name:         "Cruel Ultimatum",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: &game.TargetPurposes{{Slot: 0, Discards: 3, LifeLoss: 5}},
			Draws:   3, LifeGain: 5,
		},
		Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(cruelUltimatumSacrifice),
			effects.Instruction(cruelUltimatumDiscard),
			effects.Instruction(cruelUltimatumLoseLife),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Cruel Ultimatum — return a creature card from your graveyard to your hand",
					Match:    game.Card.IsCreature,
					Min:      1,
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 3}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: ctx.Controller(), Amount: 5}.Apply(ctx)
			}),
		),
	})
}

// Each instruction of the walk is its own function: the walk parks before
// every step until the prompt the step before it asked has been answered
// (#4012, CR 608.2c), so no step reads anything a prompted exit changes.
// The target is read afresh by each step (CR 608.2b).

func cruelUltimatumSacrifice(ctx *effects.Context) error {
	opp := effects.TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	ctx.Game.PlayerSacrificesForEffect(ctx.Item.SourceCardID, opp, effects.SacrificeSpec("a creature", effects.Creature()), "Cruel Ultimatum — sacrifice a creature")
	return nil
}

func cruelUltimatumDiscard(ctx *effects.Context) error {
	opp := effects.TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp, Source: ctx.Item.SourceCardID, N: 3})
	return nil
}

func cruelUltimatumLoseLife(ctx *effects.Context) error {
	opp := effects.TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), opp, -5)
}
