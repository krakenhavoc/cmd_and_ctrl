package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grave Exchange — Sorcery {4}{B}{B} (Avacyn Restored, S82 OLD, #5570):
//
//	"Return target creature card from your graveyard to your hand. Target
//	 player sacrifices a creature of their choice."
//
// Morgue Theft's return and Chainer's Edict's sacrifice as two target clauses
// (CR 115.1d). Each is judged on its own at resolution: a card that left the
// graveyard in response loses only the return, a player who is no longer a
// legal target loses only the sacrifice (CR 608.2b), and the spell does
// nothing only when both are gone. The player picks the creature as it
// resolves, so hexproof and protection on their creatures don't matter.
//
// No simplification.
func init() {
	const name = "Grave Exchange"
	effects.Register(effects.Spec{
		OracleID:     "cc7d902e-c86e-4771-ba66-2f63e847cff1",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
			effects.TargetPlayer("target player"),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetPlayer && ctx.IsTargetLegal(t) {
				ctx.Game.PlayerSacrificesForEffect(item.SourceCardID, t.ID, effects.SacrificeSpec("a creature", effects.Creature()), name+" — sacrifice a creature")
			}
			return nil
		},
	})
}
