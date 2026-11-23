package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Never // Return — split card with aftermath, oracle
// bf276236-18a6-4a48-b47d-5227832d26e6:
//
//	Never — Sorcery {1}{B}{B}: "Destroy target creature or planeswalker."
//	Return — Sorcery {3}{B}, aftermath: "Exile target card from a graveyard.
//	         Create a 2/2 black Zombie creature token."
//
// Insult // Injury's registration: Never under the bare oracle ID, Return
// under "#1" (ADR 0103). Aftermath is the engine's (CR 702.127a). Return is
// one spell with one target: if the card has left its graveyard by
// resolution the spell is countered on resolution and makes no Zombie
// (CR 608.2b); otherwise the exile and the token are both done.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf276236-18a6-4a48-b47d-5227832d26e6",
		Name:         "Never",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
	effects.Register(effects.Spec{
		OracleID:     "bf276236-18a6-4a48-b47d-5227832d26e6#1",
		Name:         "Return",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
