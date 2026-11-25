package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// To the Slaughter — Instant {2}{B} (Shadows over Innistrad, #5543):
//
//	"Target player sacrifices a creature or planeswalker of their choice.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, instead that player sacrifices a creature and a
//	 planeswalker of their choice."
//
// The graveyard is read as the spell resolves (the spell is still on the
// stack, so it does not count). Without delirium the player picks one
// permanent that is a creature or a planeswalker; with it they are asked
// for a creature and then for a planeswalker, and sacrifice nothing for a
// type they do not control (CR 701.21a). The player chooses, so hexproof
// and protection on their permanents do not matter. A player who is no
// longer a legal target makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	const name = "To the Slaughter"
	effects.Register(effects.Spec{
		OracleID:     "518cf9ce-14aa-4f51-81bf-f8c2692a4244",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			if !effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), player,
					effects.SacrificeSpec("a creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
					name+" — sacrifice a creature or planeswalker")
				return nil
			}
			ctx.Game.PlayerSacrificesForEffect(ctx.Source(), player, effects.SacrificeSpec("a creature", effects.Creature()),
				name+" — delirium: sacrifice a creature")
			ctx.Game.PlayerSacrificesForEffect(ctx.Source(), player, effects.SacrificeSpec("a planeswalker", effects.Planeswalker()),
				name+" — delirium: sacrifice a planeswalker")
			return nil
		},
	})
}
