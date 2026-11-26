package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frenzied Tilling — Sorcery {3}{R}{G} (Gatecrash, #5565):
//
//	"Destroy target land. Search your library for a basic land card, put that
//	 card onto the battlefield tapped, then shuffle."
//
// Two instructions: the destruction and, whether or not the land was
// destroyed (it may be indestructible), the caster's search. With the target
// gone by resolution the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6b915b3-b504-42fb-8059-62a3e8e224e1",
		Name:         "Frenzied Tilling",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		Purpose:      game.Purpose{Lands: 1, Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			for _, t := range legal {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Frenzied Tilling — a basic land",
			}.Apply(ctx)
		},
	})
}
