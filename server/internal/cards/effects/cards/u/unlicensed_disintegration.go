package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unlicensed Disintegration — Instant {1}{B}{R} (Starter Commander Decks):
//
//	"Destroy target creature. If you control an artifact, Unlicensed
//	 Disintegration deals 3 damage to that creature's controller."
//
// Minion Missile's shape: the creature's controller is read before the
// destroy, while it is still on the battlefield (CR 608.2h), so an
// indestructible creature's controller is hit too. The artifact is looked
// for as the spell resolves, after the destruction (CR 608.2c). A target that
// has become illegal counters the spell (CR 608.2b), and nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eaa2ff39-c3cf-428b-a2f1-a6a99a2c806b",
		Name:         "Unlicensed Disintegration",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 || ts[0].Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(ts[0].ID)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
				return err
			}
			if len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Artifact())) == 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 3}.Apply(ctx)
		},
	})
}
