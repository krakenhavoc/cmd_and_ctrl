package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nanogene Conversion — Sorcery {3}{U} (Doctor Who):
//
//	"Choose target creature you control. Each other creature becomes a copy of
//	 that creature until end of turn, except it isn't legendary."
//
// A layer 1 copy effect (BecomeCopy, CR 707.9) until end of turn over every
// other creature on the battlefield as the spell resolves (the set is locked
// then, CR 611.2c), each taking the target's copiable values with the
// Legendary supertype removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c06d1127-a269-453f-9c06-e9c424998eb7",
		Name:         "Nanogene Conversion",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				var others []uuid.UUID
				for _, c := range MatchingBattlefield(ctx, Creature()) {
					if c.InstanceID != t.ID {
						others = append(others, c.InstanceID)
					}
				}
				return BecomeCopy{
					Targets: others,
					Of:      t.ID,
					Except:  func(v *game.PrintedValues) { v.RemoveSupertype("Legendary") },
					Label:   "Nanogene Conversion — becomes a copy of the chosen creature, except it isn't legendary",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
