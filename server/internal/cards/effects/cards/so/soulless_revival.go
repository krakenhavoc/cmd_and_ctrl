package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulless Revival — Instant — Arcane {1}{B}:
//
//	"Return target creature card from your graveyard to your hand.
//	Splice onto Arcane {1}{B}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f00dcc4-38e7-4dde-96d4-f8d4c08b334a",
		Name:         "Soulless Revival",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{1}{B}"),
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
		},
	})
}
