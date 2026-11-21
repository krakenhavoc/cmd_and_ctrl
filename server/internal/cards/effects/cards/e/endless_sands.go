package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endless Sands — Land — Desert:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Exile target creature you control.
//	 {4}, {T}, Sacrifice this land: Return each creature card exiled with
//	 this land to the battlefield under its owner's control."
//
// Northampton Farm's link: the exile is "exiled with this land"
// (ExileTarget.Linked, CR 607.2a), and the third ability reads that
// link through the object it was activated from (Context.SourceRef),
// since the sacrifice is a cost. Every creature card so exiled returns
// together under its owner's control; a card that has left exile, and a
// token, which ceased to exist (CR 111.7), do not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c4033d97-769f-4811-8b11-f85b8817b7a2",
		Name:          "Endless Sands",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Exile target creature you control.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.AvB7ExileTargetLinked,
			},
			{
				Label:   "{4}, {T}, Sacrifice this land: Return each creature card exiled with this land to the battlefield under its owner's control.",
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
						if c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.ReturnFromExileTogether{Targets: ids}.Apply(ctx)
				},
			},
		},
	})
}
