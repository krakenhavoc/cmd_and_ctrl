package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rootcast Apprenticeship — Sorcery {3}{G}:
//
//	"Choose three. You may choose the same mode more than once.
//	 • Put two +1/+1 counters on target creature.
//	 • Create a token that's a copy of target token you control.
//	 • Target player creates a 1/1 green Squirrel creature token.
//	 • Target opponent sacrifices a nontoken artifact of their choice."
//
// Mystic Confluence's CR 700.2d shape: the announced modes are a
// multiset, each occurrence with its own target. The Squirrel goes to
// the TARGET player, and the sacrifice is the opponent's own pick among
// their nontoken artifacts (not a target of the spell: hexproof or
// protection on an artifact doesn't save it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffe13100-99dd-41e3-82ba-eaead6fd0113",
		Name:         "Rootcast Apprenticeship",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			effects.ModeDoing("Put two +1/+1 counters on target creature.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}),
			effects.ModeDoing("Create a token that's a copy of target token you control.",
				effects.TargetPermanent("target token you control", effects.IsTokenPredicate(), effects.YouControl()),
				effects.TokenCopyTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Target player creates a 1/1 green Squirrel creature token.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.CreateToken{Controller: t.ID, Template: effects.TokenCard("1/1 green Squirrel"), N: 1}.Apply(ctx)
				}),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Tokens: 1})),
			effects.ModeDoing("Target opponent sacrifices a nontoken artifact of their choice.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == t.ID && c.IsArtifact() && !c.IsToken() {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.SacrificeChoice{
						Player:     t.ID,
						Candidates: ids,
						Question:   "Rootcast Apprenticeship — sacrifice a nontoken artifact",
					}.Apply(ctx)
				}),
		),
	})
}
