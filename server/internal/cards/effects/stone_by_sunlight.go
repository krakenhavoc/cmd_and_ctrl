package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7907bbf0-f6d9-4e5b-b5de-73cd94979f98",
		Name:         "Stone by Sunlight",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy target creature with power 4 or greater.",
				TargetCreature("target creature with power 4 or greater", PowerGE(4))),
			ModeWithPurpose(Mode("Until end of turn, target creature becomes an artifact in addition to its other types and gains indestructible.",
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible"))),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			targets := legalTargetCards(item, ctx.Game)
			if len(targets) == 0 {
				return nil
			}
			if ctx.HasMode(0) {
				return DestroyTarget{Target: targets[0]}.Apply(ctx)
			}
			if err := (ScopedEffectFor{
				Target:   targets[0],
				Mods:     []game.Mod{game.AddTypesMod("Artifact")},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Stone by Sunlight — becomes an artifact in addition to its other types",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   targets[0],
				Keywords: []string{"indestructible"},
				Label:    "Stone by Sunlight — indestructible until end of turn",
			}.Apply(ctx)
		},
	})
}
