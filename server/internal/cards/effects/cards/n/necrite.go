package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrite — Creature — Thrull {1}{B}{B}, 2/2 (Fallen Empires, S82 OLD b854,
// #5637):
//
//	"Whenever this creature attacks and isn't blocked, you may sacrifice it. If
//	 you do, destroy target creature defending player controls. It can't be
//	 regenerated."
//
// Goblin Grenadiers' shape with one target, chosen as the trigger goes on the
// stack (CR 603.3d): the question is asked as it resolves, and the destruction
// hangs off the sacrifice actually happening. The target is judged again as it
// is destroyed (CR 608.2b), and the destruction ignores regeneration (CR
// 701.19c).
//
// No simplification.
func init() {
	const label = "Necrite — you may sacrifice it. If you do, destroy target creature defending player controls. It can't be regenerated"
	effects.Register(effects.Spec{
		OracleID:     "e6ab8a04-e225-4ff0-bb8b-c80d0e6ce50c",
		Name:         "Necrite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(effects.WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
				return effects.MayChoice{
					Question: "Necrite — sacrifice it to destroy target creature defending player controls?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     necriteSacrificed,
				}.Apply(effects.NewContext(g, item))
			}), effects.TargetCreatureDefendingPlayerControls),
		},
	})
}

// necriteDestroy is "destroy target creature defending player controls. It can't be
// regenerated" once the Necrite really was sacrificed. An on-disk key.
var necriteDestroy = effects.CardTail("s82-old-b854/necrite-destroy", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
		return effects.DestroyTarget{Target: t.ID, CantBeRegenerated: true}.Apply(ctx)
	}
	return nil
})

var necriteSacrificed = effects.S82OldB367SacrificeAuraAnswer("confirm/s82-old-b854-necrite-sacrifice", necriteDestroy)
