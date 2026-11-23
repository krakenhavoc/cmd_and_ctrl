package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// I Am Iron Man — {2}{U} Instant (Marvel Super Heroes, #4455):
//
//	"Until end of turn, target artifact or creature becomes an artifact
//	 creature with base power and toughness 4/4 and gains flying.
//	 Draw a card."
//
// "Becomes an artifact creature" does not say "in addition", so it sets the
// target's card types to Artifact and Creature (CR 205.1a): a type it loses
// (Enchantment) takes its subtypes with it, one it keeps (Artifact, Creature)
// keeps its own, and supertypes stay. The base 4/4 is layer 7b, so +1/+1
// counters and other pumps still apply on top. Types, base size and flying
// are one effect for the turn, over the object that was targeted. The draw
// does not depend on the target: it happens if the spell resolves at all,
// and a spell whose only target is gone does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "deb16a9a-2fa9-4ad3-bb20-7c16066e3d14",
		Name:         "I Am Iron Man",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.ScopedEffectFor{
					Target: id,
					Mods: []game.Mod{
						game.SetTypesMod([]string{"Artifact", "Creature"}, nil),
						game.SetBasePowerMod(4),
						game.SetBaseToughnessMod(4),
						game.AddKeywordsMod("flying"),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "I Am Iron Man — becomes a 4/4 artifact creature with flying",
				}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
