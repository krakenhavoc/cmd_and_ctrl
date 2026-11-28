package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cold Storage — Artifact {4} (Tempest, S82 OLD, #5628):
//
//	"{3}: Exile target creature you control.
//	 Sacrifice this artifact: Return each creature card exiled with this
//	 artifact to the battlefield under your control."
//
// The exile is linked to this object (ExileTarget.Linked), so a second Cold
// Storage's cards and a card that has since left exile are out. The sacrifice
// is a cost, so the return reads the artifact's last-known record (CR 608.2h),
// puts every linked creature card onto the battlefield as one entry, and does
// so under the activator's control, not the owner's. A token exiled this way
// has ceased to exist and does not return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e23bf6bd-bc97-425d-902a-c6b606e82b65",
		Name:         "Cold Storage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}: Exile target creature you control.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.ManaCost("{3}"),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return effects.ExileTarget{Target: t.ID, Linked: true}.Apply(ctx)
				},
			},
			{
				Label:   "Sacrifice this artifact: Return each creature card exiled with this artifact to the battlefield under your control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.SacrificeThis(),
				Effect:  effects.S82OldB761ReturnExiledCreaturesUnderYourControl,
			},
		},
	})
}
