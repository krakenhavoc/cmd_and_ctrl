package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Majestic Metamorphosis — Instant {2}{U} (Streets of New Capenna,
// #5491):
//
//	"Until end of turn, target artifact or creature becomes a 4/4 Angel
//	 artifact creature and gains flying.
//	 Draw a card."
//
// I Am Iron Man's shape with a subtype. "Becomes an Angel artifact creature"
// does not say "in addition", so it sets the target's card types to Artifact
// and Creature and its subtypes to Angel (CR 205.1a, layer 4): a type it
// loses (Enchantment) takes its subtypes with it, supertypes stay. The base
// 4/4 is layer 7b, so counters and other pumps still apply on top. Types,
// base size and flying are one effect for the turn, over the object that was
// targeted. The draw does not depend on the target: it happens whenever the
// spell resolves, which needs its target still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5779d9b2-4de1-4aa4-a56d-0da05146743a",
		Name:         "Majestic Metamorphosis",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ScopedEffectFor{
					Target: id,
					Mods: []game.Mod{
						game.SetTypesMod([]string{"Artifact", "Creature"}, []string{"Angel"}),
						game.SetBasePowerMod(4),
						game.SetBaseToughnessMod(4),
						game.AddKeywordsMod("flying"),
					},
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Majestic Metamorphosis — becomes a 4/4 Angel artifact creature with flying",
				}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
