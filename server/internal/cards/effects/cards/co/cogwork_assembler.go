package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cogwork Assembler — Artifact Creature — Assembly-Worker {3}, 2/3
// (Dominaria Remastered):
//
//	"{7}: Create a token that's a copy of target artifact. That token
//	 gains haste. Exile it at the beginning of the next end step."
//
// Flameshadow Conjuring's copy, haste and end-step exile
// (rprB06HasteAndEndStepTail), with an artifact target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01cebebb-050c-4b69-8a13-c569183a3cc3",
		Name:         "Cogwork Assembler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}: Create a token that's a copy of target artifact. That token gains haste. Exile it at the beginning of the next end step.",
			Cost:    effects.ManaCost("{7}"),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.CreateTokenCopiesTogether{
					Controller: ctx.Controller(),
					Copies:     []uuid.UUID{id},
					Then:       ctx.ItemTail(cogworkAssemblerRest),
				}.Apply(ctx)
			},
		}},
	})
}

var cogworkAssemblerRest = effects.CardTail("cogwork-assembler/haste-and-exile-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.RprB06HasteAndEndStepTail(ctx, o, "Cogwork Assembler", effects.ExileListedCardsBody)
})
