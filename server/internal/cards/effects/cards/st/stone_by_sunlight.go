package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone by Sunlight — Instant {1}{W} (The Hobbit, HOB):
//
//	"Choose one —
//	 • Destroy target creature with power 4 or greater.
//	 • Until end of turn, target creature becomes an artifact in addition to
//	   its other types and gains indestructible. (Damage and effects that
//	   say "destroy" don't destroy it.)"
//
// Each bullet carries its own target clause, and the engine applies the
// chosen one. The second bullet is two grants on the same target for the
// same duration: a layer-4 type add (Liquimetal Torque's) and indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7907bbf0-f6d9-4e5b-b5de-73cd94979f98",
		Name:         "Stone by Sunlight",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target creature with power 4 or greater.",
				effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4))),
			effects.ModeWithPurpose(effects.Mode("Until end of turn, target creature becomes an artifact in addition to its other types and gains indestructible.",
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible"))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := effects.LegalTargetCards(item, ctx.Game)
			if len(targets) == 0 {
				return nil
			}
			if ctx.HasMode(0) {
				return effects.DestroyTarget{Target: targets[0]}.Apply(ctx)
			}
			if err := (effects.ScopedEffectFor{
				Target:   targets[0],
				Mods:     []game.Mod{game.AddTypesMod("Artifact")},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Stone by Sunlight — becomes an artifact in addition to its other types",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   targets[0],
				Keywords: []string{"indestructible"},
				Label:    "Stone by Sunlight — indestructible until end of turn",
			}.Apply(ctx)
		},
	})
}
