package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fates' Reversal — Sorcery {1}{B}:
//
//	"Return up to one target creature card from your graveyard to your
//	 hand. Venture into the dungeon."
//
// With no target chosen the spell still resolves and ventures; with a
// target that has left the graveyard it does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cd88e87-6585-437a-bcad-b494fffc0544",
		Name:         "Fates' Reversal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.VentureIntoTheDungeon{}.Apply(ctx)
		},
	})
}
