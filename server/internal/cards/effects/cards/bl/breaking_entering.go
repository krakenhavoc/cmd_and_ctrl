package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breaking // Entering — split card, oracle 043370fd-9cfe-47ce-8019-e915cee1ae95
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Breaking — Sorcery {U}{B}: "Target player mills eight cards."
//	Entering — Sorcery {4}{B}{R}: "Put a creature card from a graveyard onto the
//	           battlefield under your control. It gains haste until end of turn."
//	           Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Breaking under the bare oracle ID, Entering
// under "#1"); the fused cast is the engine's (CR 702.102), Breaking resolving
// first, so a fused Entering may take a creature that was just milled. Entering
// is Extract from Darkness's choice: not a target, made as the spell resolves
// from every graveyard, the chosen card entering under your control whoever owned
// it; it then gains haste until end of turn if it is on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "043370fd-9cfe-47ce-8019-e915cee1ae95",
		Name:         "Breaking",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			return effects.MillCards{Player: t.ID, N: 8}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "043370fd-9cfe-47ce-8019-e915cee1ae95#1",
		Name:         "Entering",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			candidates := effects.AllGraveyardsCardIDs(ctx, game.Card.IsCreature)
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Entering — put a creature card from a graveyard onto the battlefield under your control",
				Cards:    candidates,
				Min:      1,
				Max:      1,
				ThenKey:  enteringPicked,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// enteringPicked is the chosen creature card entering under the caster's control
// with haste until end of turn, keyed (#4253). An on-disk key.
var enteringPicked = effects.ItemChoice("choose-cards/entering-reanimate", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
		return err
	}
	if !effects.OnBattlefield(ctx.Game, id) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Entering — haste until end of turn"}.Apply(ctx)
})
