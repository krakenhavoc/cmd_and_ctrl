package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska's Stoneglare — Sorcery {4}{B}{G} (Guilds of Ravnica):
//
//	"Destroy target creature. You gain life equal to its toughness. You may
//	 search your library and/or graveyard for a card named Vraska, Regal
//	 Gorgon, reveal it, and put it into your hand. If you search your library
//	 this way, shuffle."
//
// Garruk, Apex Predator's -3 for the first two sentences: the toughness is
// read before the destroy (CR 608.2h) and the life is gained whether or not
// the creature died. The search is Basri's Aegis' keyed "you may" answer and
// waits behind the destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c39e2046-173a-44b6-8e93-b22ede35485d",
		Name:         "Vraska's Stoneglare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				ids := effects.LegalTargetCardIDs(ctx)
				if len(ids) == 0 {
					return nil
				}
				c, ok := ctx.Game.LookupCardForEffect(ids[0])
				if !ok {
					return nil
				}
				return ctx.Game.DestroyPermanentsKeyedForEffect(ids[:1], effects.WithN(ctx.Tail(effects.GarrukApexPredatorGainLife), c.CurrentToughness()))
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Vraska's Stoneglare", "Vraska, Regal Gorgon")),
		),
	})
}
