package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Capricious Efreet — Creature — Efreet {4}{R}{R}, 6/4 (Commander 2013, S82
// older sets):
//
//	"At the beginning of your upkeep, choose target nonland permanent you
//	 control and up to two target nonland permanents you don't control.
//	 Destroy one of them at random."
//
// Two target clauses on one trigger: one nonland permanent of yours and up to
// two nonland permanents of other players. As the trigger resolves, every
// target that is still legal (CR 608.2b) is a candidate and one is picked
// through the engine's keyed random draw, then destroyed. Nothing happens if
// none is still legal.
//
// No simplification.
func init() {
	t := effects.AtYourUpkeep("Capricious Efreet — destroy one of the targets at random", func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard && t.ID != uuid.Nil {
				ids = append(ids, t.ID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
		if len(pick) == 0 {
			return nil
		}
		return effects.DestroyTarget{Target: pick[0]}.Apply(ctx)
	})
	effects.Register(effects.Spec{
		OracleID:     "9abd2286-23e9-49cd-be53-39423890f35c",
		Name:         "Capricious Efreet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(t, effects.Clauses(
				effects.TargetPermanent("target nonland permanent you control", effects.Nonland(), effects.YouControl()),
				effects.TargetPermanent("up to two target nonland permanents you don't control", effects.Nonland(), effects.OpponentControls()).WithCount(0, 2),
			)),
		},
	})
}
