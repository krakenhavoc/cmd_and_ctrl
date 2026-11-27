package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Synod Sanctum — {1} Artifact (Archenemy):
//
//	"{2}, {T}: Exile target permanent you control.
//	 {2}, Sacrifice this artifact: Return all cards exiled with this artifact
//	 to the battlefield under your control."
//
// The exile is linked to the Sanctum; the return reads the cards exiled with
// the Sanctum object that was sacrificed (its last-known object, CR 608.2h)
// and puts them all onto the battlefield together under the ability's
// controller (a token exiled this way is gone, CR 111.8).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "643ddf81-bbd3-4398-b5a7-1169c6e2b081",
		Name:         "Synod Sanctum",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Exile target permanent you control.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetPermanent("target permanent you control", effects.YouControl()),
				Effect:  effects.FearOfAbductionExile,
			},
			{
				Label:   "{2}, Sacrifice this artifact: Return all cards exiled with this artifact to the battlefield under your control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
						ids = append(ids, c.InstanceID)
					}
					if len(ids) == 0 {
						return nil
					}
					return effects.ReturnFromExileTogether{Targets: ids, Controller: item.Controller}.Apply(ctx)
				},
			},
		},
	})
}
