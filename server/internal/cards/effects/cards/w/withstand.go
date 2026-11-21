package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Withstand — Instant {2}{W} (Guildpact, #5599):
//
//	"Prevent the next 3 damage that would be dealt to any target this
//	 turn.
//	 Draw a card."
//
// Mending Hands' shield (PreventNextDamage, a CR 615 prevention shield
// on the creature, player or planeswalker targeted) at 3, then the draw
// (effectThenDrawACard). The shield asks nothing, so the two run in
// printed order. A target that is illegal as it resolves makes the whole
// spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebf2816b-e042-43e9-8734-d3a896334c69",
		Name:         "Withstand",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Draws: 1},
		OnResolve: effects.EffectThenDrawACard(func(g *game.Game, item *game.StackItem) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.PreventNextDamage{Target: item.Targets[0].ID, Amount: 3, Label: "Withstand: prevent the next 3 damage"}.Apply(effects.NewContext(g, item))
		}),
	})
}
