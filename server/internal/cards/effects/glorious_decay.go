package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glorious Decay — Instant {1}{G} (Secrets of Strixhaven, #4456):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Glorious Decay deals 4 damage to target creature with flying.
//	 • Exile target card from a graveyard. Draw a card."
//
// Each bullet has its own target, chosen as the spell is cast. The flying
// restriction is read through the layered view, so a granted flying counts.
// The third bullet's draw is its own sentence (Cremate's reading): it happens
// whether or not the exile did, but a bullet whose target has gone does
// nothing at all (CR 608.2b), draw included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac4277d7-d78a-4d6e-8d6d-11f9e56f8ae7",
		Name:         "Glorious Decay",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
			ModeDoing("Glorious Decay deals 4 damage to target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying")),
				socB20DamageToTheModeTarget(4)),
			ModeDoing("Exile target card from a graveyard. Draw a card.",
				TargetCardInGraveyard("target card from a graveyard"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}),
		),
	})
}
