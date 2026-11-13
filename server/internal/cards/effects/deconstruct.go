package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deconstruct — Sorcery {2}{G} (Mirrodin, #5606):
//
//	"Destroy target artifact. Add {G}{G}{G}."
//
// Destroy the artifact, then the mana, which stays in its caster's pool
// until the end of the phase (CR 106.4) for a second spell. The mana
// comes whether or not the artifact was destroyed (an indestructible
// one); an artifact that is no longer a legal target leaves the spell
// with nothing to do, mana included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36a8ceb1-148b-41e0-a7bf-ceb879bf08e7",
		Name:         "Deconstruct",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return AddMana{Produced: "{G}{G}{G}"}.Apply(ctx)
		},
	})
}
