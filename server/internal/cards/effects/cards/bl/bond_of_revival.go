package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bond of Revival — Sorcery {4}{B} (War of the Spark, #5520):
//
//	"Return target creature card from your graveyard to the battlefield. It
//	 gains haste until your next turn."
//
// Kardur's Vicious Return's chapter III without the counter: the return is a
// one-card ReturnFromGraveyardTogether whose tail gives the creature that
// actually entered haste until the controller's next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62cca25f-3bf7-4b67-97d5-280bf2544722",
		Name:         "Bond of Revival",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		Targets:      effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{
				Targets:    []uuid.UUID{id},
				Controller: ctx.Controller(),
				ThenTail:   bondOfRevivalHaste,
			}.Apply(ctx)
		},
	})
}

// bondOfRevivalHaste gives the creature that entered haste until its
// controller's next turn.
var bondOfRevivalHaste = effects.TogetherTail("together/s82-old-b170-bond-of-revival-haste", func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
	for _, id := range entered {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
			Label:    "Bond of Revival — haste until your next turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
